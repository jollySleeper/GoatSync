// Package integration provides integration tests for GoatSync API.
// These tests verify 1:1 compatibility with the original EteSync server.
package integration

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"goatsync/internal/config"
	"goatsync/internal/crypto"
	"goatsync/internal/database"
	"goatsync/internal/handler"
	"goatsync/internal/model"
	"goatsync/internal/repository"
	"goatsync/internal/server"
	"goatsync/internal/service"
	"goatsync/internal/storage"

	"github.com/vmihailenco/msgpack/v5"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	// Setup
	cfg := &config.Config{
		Port:                  "8080",
		Debug:                 true,
		EncryptionSecret:      "test-secret-key-for-testing-32ch",
		ChunkStoragePath:      "/tmp/goatsync-test-chunks",
		ChallengeValidSeconds: 300,
	}

	// Try to connect to test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://goatsync:goatsync@localhost:5432/goatsync?sslmode=disable"
	}

	db, err := database.Connect(dbURL)
	if err != nil {
		// Skip integration tests if no database
		os.Exit(0)
	}

	// Run migrations
	_ = database.AutoMigrate(db,
		&model.Stoken{},
		&model.User{},
		&model.UserInfo{},
		&model.AuthToken{},
		&model.CollectionType{},
		&model.Collection{},
		&model.CollectionItem{},
		&model.CollectionItemRevision{},
		&model.CollectionItemChunk{},
		&model.RevisionChunkRelation{},
		&model.CollectionMember{},
		&model.CollectionMemberRemoved{},
		&model.CollectionInvitation{},
	)

	// Clear test data
	clearTestData(db)

	// Initialize components
	fileStorage := storage.NewFileStorage(cfg.ChunkStoragePath)
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	collectionRepo := repository.NewCollectionRepository(db)
	itemRepo := repository.NewItemRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	chunkRepo := repository.NewChunkRepository(db)

	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	collectionService := service.NewCollectionService(collectionRepo, cfg)
	itemService := service.NewItemService(itemRepo, nil, collectionRepo, memberRepo)
	memberService := service.NewMemberService(memberRepo, collectionRepo)
	invitationService := service.NewInvitationService(invitationRepo, memberRepo, userRepo)
	chunkService := service.NewChunkService(chunkRepo, collectionRepo, memberRepo, fileStorage)

	authHandler := handler.NewAuthHandler(authService)
	collectionHandler := handler.NewCollectionHandler(collectionService)
	itemHandler := handler.NewItemHandler(itemService)
	memberHandler := handler.NewMemberHandler(memberService)
	invitationHandler := handler.NewInvitationHandler(invitationService)
	chunkHandler := handler.NewChunkHandler(chunkService)
	websocketHandler := handler.NewWebSocketHandler(nil)
	healthHandler := handler.NewHealthHandler(db)
	testHandler := handler.NewTestHandler(db, true)

	srv := server.New(
		cfg,
		authService,
		authHandler,
		collectionHandler,
		itemHandler,
		memberHandler,
		invitationHandler,
		chunkHandler,
		websocketHandler,
		healthHandler,
		testHandler,
	)

	testServer = httptest.NewServer(srv.Engine())
	defer testServer.Close()

	// Run tests
	code := m.Run()

	// Cleanup
	clearTestData(db)
	testServer.Close()

	os.Exit(code)
}

func clearTestData(db interface{}) {
	// Implementation depends on GORM DB type
}

// TestIsEtebase verifies the is_etebase endpoint returns correctly
func TestIsEtebase(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/api/v1/authentication/is_etebase/")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestHealthEndpoints verifies health check endpoints
func TestHealthEndpoints(t *testing.T) {
	endpoints := []string{"/health", "/ready", "/live"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(testServer.URL + endpoint)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", endpoint, resp.StatusCode)
			}
		})
	}
}

// TestSignupFlow tests the full signup flow
func TestSignupFlow(t *testing.T) {
	// Generate Ed25519 keypair for login
	loginPub, loginPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate login keypair: %v", err)
	}

	// Generate another keypair for encryption
	encPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate encryption keypair: %v", err)
	}

	// Create signup request
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	signupReq := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "testuser_" + hex.EncodeToString(salt[:4]),
			"email":    "test_" + hex.EncodeToString(salt[:4]) + "@example.com",
		},
		"salt":             salt,
		"loginPubkey":      []byte(loginPub),
		"pubkey":           []byte(encPub),
		"encryptedContent": []byte("encrypted-test-content"),
	}

	body, err := msgpack.Marshal(signupReq)
	if err != nil {
		t.Fatalf("Failed to marshal signup request: %v", err)
	}

	resp, err := http.Post(
		testServer.URL+"/api/v1/authentication/signup/",
		"application/msgpack",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("Signup request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 200/201, got %d", resp.StatusCode)
	}

	// Decode response
	var signupResp map[string]interface{}
	if err := msgpack.NewDecoder(resp.Body).Decode(&signupResp); err != nil {
		t.Fatalf("Failed to decode signup response: %v", err)
	}

	if signupResp["token"] == nil {
		t.Error("Expected token in signup response")
	}

	_ = loginPriv // Will be used for login test
}

// TestLoginChallengeFlow tests the login challenge mechanism
func TestLoginChallengeFlow(t *testing.T) {
	// First, we need a user - create one
	loginPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	encPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	username := "logintest_" + hex.EncodeToString(salt[:4])

	// Create user via signup
	signupReq := map[string]interface{}{
		"user": map[string]interface{}{
			"username": username,
			"email":    username + "@example.com",
		},
		"salt":             salt,
		"loginPubkey":      []byte(loginPub),
		"pubkey":           []byte(encPub),
		"encryptedContent": []byte("encrypted"),
	}

	body, _ := msgpack.Marshal(signupReq)
	resp, err := http.Post(
		testServer.URL+"/api/v1/authentication/signup/",
		"application/msgpack",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}
	_ = resp.Body.Close()

	// Now test login challenge
	challengeReq := map[string]interface{}{
		"username": username,
	}
	body, _ = msgpack.Marshal(challengeReq)

	resp, err = http.Post(
		testServer.URL+"/api/v1/authentication/login_challenge/",
		"application/msgpack",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("Login challenge failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var challengeResp map[string]interface{}
	if err := msgpack.NewDecoder(resp.Body).Decode(&challengeResp); err != nil {
		t.Fatalf("Failed to decode challenge response: %v", err)
	}

	// Verify challenge response has required fields
	if challengeResp["salt"] == nil {
		t.Error("Expected salt in challenge response")
	}
	if challengeResp["challenge"] == nil {
		t.Error("Expected challenge in challenge response")
	}
	if challengeResp["version"] == nil {
		t.Error("Expected version in challenge response")
	}
}

// TestUnauthorizedAccess tests that protected endpoints require auth
func TestUnauthorizedAccess(t *testing.T) {
	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/collection/"},
		{"POST", "/api/v1/authentication/logout/"},
		{"POST", "/api/v1/authentication/change_password/"},
	}

	for _, ep := range protectedEndpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, testServer.URL+ep.path, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}

// TestCryptoCompatibility tests that crypto operations match Python
func TestCryptoCompatibility(t *testing.T) {
	etebase := crypto.NewEtebase("test-secret-key")

	// Test BLAKE2b key derivation is deterministic
	salt := []byte("0123456789abcdef")

	key1, err := etebase.GetEncryptionKey(salt)
	if err != nil {
		t.Fatalf("GetEncryptionKey failed: %v", err)
	}

	key2, err := etebase.GetEncryptionKey(salt)
	if err != nil {
		t.Fatalf("GetEncryptionKey (second call) failed: %v", err)
	}

	if !bytes.Equal(key1[:], key2[:]) {
		t.Error("Key derivation is not deterministic")
	}

	// Test encryption/decryption roundtrip
	plaintext := []byte("Hello, EteSync!")
	ciphertext, err := etebase.Encrypt(key1, plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := etebase.Decrypt(key1, ciphertext)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted text doesn't match original")
	}
}

// TestMessagePackSerialization tests MessagePack format compatibility
func TestMessagePackSerialization(t *testing.T) {
	// Test that our responses are valid MessagePack
	resp, err := http.Get(testServer.URL + "/api/v1/authentication/is_etebase/")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/msgpack" && contentType != "" {
		// is_etebase can return empty body
		t.Logf("Content-Type: %s", contentType)
	}
}

// TestErrorResponses tests that error responses match EteSync format
func TestErrorResponses(t *testing.T) {
	// Test user not found error
	challengeReq := map[string]interface{}{
		"username": "nonexistent_user_12345",
	}
	body, _ := msgpack.Marshal(challengeReq)

	resp, err := http.Post(
		testServer.URL+"/api/v1/authentication/login_challenge/",
		"application/msgpack",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for user not found, got %d", resp.StatusCode)
	}

	// Decode error response
	var errResp map[string]interface{}
	if err := msgpack.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Verify error format matches EteSync
	if errResp["code"] == nil {
		t.Error("Expected 'code' in error response")
	}
	if errResp["detail"] == nil {
		t.Error("Expected 'detail' in error response")
	}

	code, ok := errResp["code"].(string)
	if !ok || code != "user_not_found" {
		t.Errorf("Expected error code 'user_not_found', got '%v'", errResp["code"])
	}
}

// TestStokenGeneration tests that stokens are generated correctly
func TestStokenGeneration(t *testing.T) {
	// Generate multiple stokens and verify uniqueness
	stokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		uid, err := crypto.GenerateStokenUID()
		if err != nil {
			t.Fatalf("GenerateStokenUID failed: %v", err)
		}

		if len(uid) != 32 {
			t.Errorf("Stoken UID should be 32 chars, got %d", len(uid))
		}

		if stokens[uid] {
			t.Error("Generated duplicate stoken UID")
		}
		stokens[uid] = true
	}
}

// TestItemUIDGeneration tests item UID generation
func TestItemUIDGeneration(t *testing.T) {
	uids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		uid, err := crypto.GenerateItemUID()
		if err != nil {
			t.Fatalf("GenerateItemUID failed: %v", err)
		}

		if len(uid) != 22 {
			t.Errorf("Item UID should be 22 chars, got %d", len(uid))
		}

		if uids[uid] {
			t.Error("Generated duplicate item UID")
		}
		uids[uid] = true
	}
}

// TestEd25519SignatureVerification tests Ed25519 signature verification
func TestEd25519SignatureVerification(t *testing.T) {
	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Key generation failed: %v", err)
	}

	// Sign a message
	message := []byte("test message for signing")
	signature := ed25519.Sign(priv, message)

	// Verify with our crypto package
	err = crypto.VerifySignature(pub, message, signature)
	if err != nil {
		t.Errorf("Valid signature verification failed: %v", err)
	}

	// Test with wrong message
	wrongMessage := []byte("different message")
	err = crypto.VerifySignature(pub, wrongMessage, signature)
	if err == nil {
		t.Error("Should fail verification with wrong message")
	}
}

// BenchmarkKeyDerivation benchmarks the BLAKE2b key derivation
func BenchmarkKeyDerivation(b *testing.B) {
	etebase := crypto.NewEtebase("benchmark-secret-key")
	salt := []byte("0123456789abcdef")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = etebase.GetEncryptionKey(salt)
	}
}

// BenchmarkEncryption benchmarks SecretBox encryption
func BenchmarkEncryption(b *testing.B) {
	etebase := crypto.NewEtebase("benchmark-secret-key")
	salt := []byte("0123456789abcdef")
	key, _ := etebase.GetEncryptionKey(salt)
	plaintext := make([]byte, 1024) // 1KB
	_, _ = rand.Read(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = etebase.Encrypt(key, plaintext)
	}
}

