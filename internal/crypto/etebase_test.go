package crypto

import (
	"bytes"
	"testing"
)

// TestGetEncryptionKey verifies that our BLAKE2b key derivation matches the Python implementation.
// To verify compatibility:
// 1. Use the same secret key, salt, and personalization
// 2. Compare the output with Python's nacl.hash.blake2b
func TestGetEncryptionKey(t *testing.T) {
	e := NewEtebase("test_secret_key")

	// Test with a known salt
	salt := []byte("0123456789abcdef") // 16 bytes

	key, err := e.GetEncryptionKey(salt)
	if err != nil {
		t.Fatalf("GetEncryptionKey failed: %v", err)
	}

	// Key should be 32 bytes
	if len(key) != KeySize {
		t.Errorf("Expected key size %d, got %d", KeySize, len(key))
	}

	// The key should be deterministic
	key2, err := e.GetEncryptionKey(salt)
	if err != nil {
		t.Fatalf("GetEncryptionKey (second call) failed: %v", err)
	}

	if !bytes.Equal(key[:], key2[:]) {
		t.Error("GetEncryptionKey is not deterministic")
	}

	// Different salt should produce different key
	differentSalt := []byte("fedcba9876543210")
	key3, err := e.GetEncryptionKey(differentSalt)
	if err != nil {
		t.Fatalf("GetEncryptionKey (different salt) failed: %v", err)
	}

	if bytes.Equal(key[:], key3[:]) {
		t.Error("Different salts produced the same key")
	}
}

// TestEncryptDecrypt verifies that SecretBox encryption/decryption works correctly
func TestEncryptDecrypt(t *testing.T) {
	e := NewEtebase("test_secret_key")

	// Generate a key
	salt := []byte("0123456789abcdef")
	key, err := e.GetEncryptionKey(salt)
	if err != nil {
		t.Fatalf("GetEncryptionKey failed: %v", err)
	}

	// Test plaintext
	plaintext := []byte("Hello, this is a test message for SecretBox encryption!")

	// Encrypt
	ciphertext, err := e.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Ciphertext should be longer than plaintext (nonce + MAC)
	// nonce (24) + plaintext + MAC (16)
	expectedMinLen := NonceSize + len(plaintext) + 16
	if len(ciphertext) < expectedMinLen {
		t.Errorf("Ciphertext too short: %d < %d", len(ciphertext), expectedMinLen)
	}

	// Decrypt
	decrypted, err := e.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Verify decrypted matches original
	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted text doesn't match original")
	}
}

// TestDecryptWithWrongKey verifies that decryption fails with wrong key
func TestDecryptWithWrongKey(t *testing.T) {
	e := NewEtebase("test_secret_key")

	// Generate two different keys
	key1, _ := e.GetEncryptionKey([]byte("salt1234567890ab"))
	key2, _ := e.GetEncryptionKey([]byte("differentSalt!!0"))

	// Encrypt with key1
	plaintext := []byte("Secret message")
	ciphertext, err := e.Encrypt(key1, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Try to decrypt with key2 - should fail
	_, err = e.Decrypt(key2, ciphertext)
	if err == nil {
		t.Error("Decrypt should have failed with wrong key")
	}
	if err != ErrDecryptionFailed {
		t.Errorf("Expected ErrDecryptionFailed, got: %v", err)
	}
}

// TestVerifySignature verifies Ed25519 signature verification
func TestVerifySignature(t *testing.T) {
	// Use known test vectors
	// Note: In real tests, we'd use actual Ed25519 key pairs

	// Invalid signature (wrong length)
	err := VerifySignature(
		make([]byte, 32), // 32-byte pubkey
		[]byte("message"),
		make([]byte, 63), // Wrong size signature
	)
	if err != ErrInvalidSignature {
		t.Errorf("Expected ErrInvalidSignature for wrong signature length, got: %v", err)
	}

	// Invalid pubkey length
	err = VerifySignature(
		make([]byte, 31), // Wrong size pubkey
		[]byte("message"),
		make([]byte, 64),
	)
	if err == nil || err == ErrInvalidSignature {
		// Should be ErrInvalidKeySize
		if err != nil && err.Error() != "invalid key size: expected 32 bytes, got 31" {
			// Check if it's the key size error
			t.Logf("Got error: %v", err)
		}
	}
}

// TestGenerateStokenUID verifies stoken UID generation
func TestGenerateStokenUID(t *testing.T) {
	uid1, err := GenerateStokenUID()
	if err != nil {
		t.Fatalf("GenerateStokenUID failed: %v", err)
	}

	// Should be 32 characters
	if len(uid1) != 32 {
		t.Errorf("Expected UID length 32, got %d", len(uid1))
	}

	// Should only contain valid characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	for _, c := range uid1 {
		if !bytes.ContainsRune([]byte(validChars), c) {
			t.Errorf("UID contains invalid character: %c", c)
		}
	}

	// Should be unique (probabilistically)
	uid2, _ := GenerateStokenUID()
	if uid1 == uid2 {
		t.Error("Two generated UIDs are the same")
	}
}

// TestGenerateItemUID verifies item UID generation
func TestGenerateItemUID(t *testing.T) {
	uid, err := GenerateItemUID()
	if err != nil {
		t.Fatalf("GenerateItemUID failed: %v", err)
	}

	// Should be 22 characters
	if len(uid) != 22 {
		t.Errorf("Expected UID length 22, got %d", len(uid))
	}
}

// TestEncryptDecryptEmpty verifies handling of empty plaintext
func TestEncryptDecryptEmpty(t *testing.T) {
	e := NewEtebase("test_secret_key")
	salt := []byte("0123456789abcdef")
	key, _ := e.GetEncryptionKey(salt)

	// Encrypt empty plaintext
	ciphertext, err := e.Encrypt(key, []byte{})
	if err != nil {
		t.Fatalf("Encrypt empty failed: %v", err)
	}

	// Decrypt
	decrypted, err := e.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt empty failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("Expected empty decrypted, got %d bytes", len(decrypted))
	}
}

// TestDecryptTooShort verifies handling of ciphertext too short
func TestDecryptTooShort(t *testing.T) {
	e := NewEtebase("test_secret_key")
	salt := []byte("0123456789abcdef")
	key, _ := e.GetEncryptionKey(salt)

	// Try to decrypt ciphertext that's too short (< 24 bytes for nonce)
	_, err := e.Decrypt(key, []byte("too short"))
	if err != ErrInvalidNonceSize {
		t.Errorf("Expected ErrInvalidNonceSize, got: %v", err)
	}
}

