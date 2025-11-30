package service

import (
	"context"
	"testing"

	"goatsync/internal/config"
	"goatsync/internal/model"
)

// MockUserRepository is a mock implementation for testing
type MockUserRepository struct {
	users map[string]*model.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{users: make(map[string]*model.User)}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User, userInfo *model.UserInfo) error {
	user.UserInfo = userInfo
	m.users[user.Username] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return m.users[username], nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *MockUserRepository) UpdateUserInfo(ctx context.Context, userInfo *model.UserInfo) error {
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	for username, u := range m.users {
		if u.ID == id {
			delete(m.users, username)
			return nil
		}
	}
	return nil
}

// MockTokenRepository is a mock implementation for testing
type MockTokenRepository struct {
	tokens map[string]*model.AuthToken
}

func NewMockTokenRepository() *MockTokenRepository {
	return &MockTokenRepository{tokens: make(map[string]*model.AuthToken)}
}

func (m *MockTokenRepository) Create(ctx context.Context, token *model.AuthToken) error {
	if token.Key == "" {
		token.Key = "mock-token-key"
	}
	m.tokens[token.Key] = token
	return nil
}

func (m *MockTokenRepository) GetByKey(ctx context.Context, key string) (*model.AuthToken, error) {
	return m.tokens[key], nil
}

func (m *MockTokenRepository) Delete(ctx context.Context, key string) error {
	delete(m.tokens, key)
	return nil
}

func (m *MockTokenRepository) DeleteAllForUser(ctx context.Context, userID uint) error {
	for key, token := range m.tokens {
		if token.UserID == userID {
			delete(m.tokens, key)
		}
	}
	return nil
}

func TestAuthService_LoginChallenge_UserNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	tokenRepo := NewMockTokenRepository()
	cfg := &config.Config{
		EncryptionSecret:      "test-secret-key-32-chars-long!!",
		ChallengeValidSeconds: 300,
	}

	svc := NewAuthService(userRepo, tokenRepo, cfg)

	_, err := svc.LoginChallenge(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestAuthService_Signup_Success(t *testing.T) {
	userRepo := NewMockUserRepository()
	tokenRepo := NewMockTokenRepository()
	cfg := &config.Config{
		EncryptionSecret:      "test-secret-key-32-chars-long!!",
		ChallengeValidSeconds: 300,
	}

	svc := NewAuthService(userRepo, tokenRepo, cfg)

	req := &SignupRequest{
		User: SignupUser{
			Username: "testuser",
			Email:    "test@example.com",
		},
		Salt:             []byte("0123456789abcdef"),
		LoginPubkey:      make([]byte, 32),
		Pubkey:           make([]byte, 32),
		EncryptedContent: []byte("encrypted"),
	}

	resp, err := svc.Signup(context.Background(), req)
	if err != nil {
		t.Fatalf("Signup failed: %v", err)
	}

	if resp.Token == "" {
		t.Error("Expected token in response")
	}
	if resp.User.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", resp.User.Username)
	}
}

func TestAuthService_Signup_UserExists(t *testing.T) {
	userRepo := NewMockUserRepository()
	tokenRepo := NewMockTokenRepository()
	cfg := &config.Config{
		EncryptionSecret:      "test-secret-key-32-chars-long!!",
		ChallengeValidSeconds: 300,
	}

	// Add existing user
	userRepo.users["existinguser"] = &model.User{
		ID:       1,
		Username: "existinguser",
		Email:    "existing@example.com",
	}

	svc := NewAuthService(userRepo, tokenRepo, cfg)

	req := &SignupRequest{
		User: SignupUser{
			Username: "existinguser",
			Email:    "new@example.com",
		},
		Salt:             []byte("0123456789abcdef"),
		LoginPubkey:      make([]byte, 32),
		Pubkey:           make([]byte, 32),
		EncryptedContent: []byte("encrypted"),
	}

	_, err := svc.Signup(context.Background(), req)
	if err == nil {
		t.Error("Expected error for existing user")
	}
}

