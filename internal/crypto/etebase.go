// Package crypto implements EteSync/Etebase cryptographic operations.
//
// This package provides the cryptographic primitives used by EteSync:
//   - BLAKE2b key derivation with key, salt, and personalization
//   - NaCl SecretBox (XSalsa20-Poly1305) encryption/decryption
//   - Ed25519 signature verification
//
// IMPORTANT: These implementations MUST match the Python PyNaCl library exactly.
// Any deviation will cause authentication failures with existing clients.
package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/dchest/blake2b"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// KeySize is the size of encryption keys (32 bytes for SecretBox)
	KeySize = 32
	// NonceSize is the size of nonces for SecretBox (24 bytes)
	NonceSize = 24
	// SaltSize is the size of BLAKE2b salt (16 bytes)
	SaltSize = 16
	// PersonalizationSize is the maximum size of BLAKE2b personalization (16 bytes)
	PersonalizationSize = 16
)

var (
	// ErrDecryptionFailed is returned when SecretBox decryption fails
	ErrDecryptionFailed = errors.New("decryption failed: authentication error")
	// ErrInvalidSignature is returned when Ed25519 signature verification fails
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidKeySize is returned when a key has an incorrect size
	ErrInvalidKeySize = errors.New("invalid key size")
	// ErrInvalidNonceSize is returned when the ciphertext is too short to contain a nonce
	ErrInvalidNonceSize = errors.New("ciphertext too short: missing nonce")
)

// Personalization constants used by Etebase
var (
	// PersonalizationAuth is used for authentication challenge encryption
	PersonalizationAuth = []byte("etebase-auth")
)

// Etebase provides cryptographic operations for EteSync/Etebase protocol
type Etebase struct {
	secretKey string
}

// NewEtebase creates a new Etebase crypto instance with the given secret key.
// The secretKey should be the server's ENCRYPTION_SECRET from configuration.
func NewEtebase(secretKey string) *Etebase {
	return &Etebase{
		secretKey: secretKey,
	}
}

// GetEncryptionKey derives an encryption key using BLAKE2b with key, salt, and personalization.
//
// This matches the Python implementation:
//
//	def get_encryption_key(salt: bytes):
//	    key = nacl.hash.blake2b(settings.SECRET_KEY.encode(), encoder=nacl.encoding.RawEncoder)
//	    return nacl.hash.blake2b(
//	        b"",
//	        key=key,
//	        salt=salt[:nacl.hash.BLAKE2B_SALTBYTES],  # First 16 bytes
//	        person=b"etebase-auth",
//	        encoder=nacl.encoding.RawEncoder,
//	    )
func (e *Etebase) GetEncryptionKey(salt []byte) ([KeySize]byte, error) {
	var key [KeySize]byte

	// Step 1: Hash the secret key with BLAKE2b-512 to get a 64-byte hash
	// (default BLAKE2b output size), then use first 32 bytes as key
	// This matches: nacl.hash.blake2b(settings.SECRET_KEY.encode(), encoder=RawEncoder)
	// Note: PyNaCl's blake2b() defaults to 64-byte output (BLAKE2b-512)
	masterKeyHash := blake2b.Sum512([]byte(e.secretKey))
	masterKey := masterKeyHash[:KeySize] // Use first 32 bytes as the key

	// Step 2: Derive the encryption key using BLAKE2b with key, salt, and personalization
	// Prepare salt (must be exactly 16 bytes)
	var saltBytes [SaltSize]byte
	saltLen := len(salt)
	if saltLen > SaltSize {
		saltLen = SaltSize
	}
	copy(saltBytes[:], salt[:saltLen])

	// Prepare personalization (must be exactly 16 bytes, zero-padded)
	var person [PersonalizationSize]byte
	copy(person[:], PersonalizationAuth)

	// Create BLAKE2b hasher with key, salt, and personalization
	// Using github.com/dchest/blake2b which supports full BLAKE2b parameters
	h, err := blake2b.New(&blake2b.Config{
		Size:   KeySize,
		Key:    masterKey,
		Salt:   saltBytes[:],
		Person: person[:],
	})
	if err != nil {
		return key, fmt.Errorf("failed to create BLAKE2b hasher: %w", err)
	}

	// Hash empty input (b"" in Python)
	// h.Write is not needed for empty input, just get the sum
	sum := h.Sum(nil)
	copy(key[:], sum)

	return key, nil
}

// Encrypt encrypts plaintext using NaCl SecretBox (XSalsa20-Poly1305).
// Returns: nonce (24 bytes) || ciphertext (with 16-byte authentication tag)
//
// This matches the Python implementation:
//
//	box = nacl.secret.SecretBox(enc_key)
//	encrypted = box.encrypt(plaintext, encoder=nacl.encoding.RawEncoder)
//
// The output format is: nonce || ciphertext (where ciphertext includes the 16-byte MAC)
func (e *Etebase) Encrypt(key [KeySize]byte, plaintext []byte) ([]byte, error) {
	// Generate random 24-byte nonce
	var nonce [NonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt using SecretBox
	// secretbox.Seal appends the encrypted message to the output slice
	// Format: nonce || ciphertext (ciphertext includes MAC)
	ciphertext := secretbox.Seal(nonce[:], plaintext, &nonce, &key)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using NaCl SecretBox (XSalsa20-Poly1305).
// The ciphertext format should be: nonce (24 bytes) || encrypted data (with 16-byte MAC)
//
// This matches the Python implementation:
//
//	box = nacl.secret.SecretBox(enc_key)
//	plaintext = box.decrypt(ciphertext)
func (e *Etebase) Decrypt(key [KeySize]byte, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < NonceSize {
		return nil, ErrInvalidNonceSize
	}

	// Extract nonce from the beginning of ciphertext
	var nonce [NonceSize]byte
	copy(nonce[:], ciphertext[:NonceSize])

	// Decrypt the rest
	plaintext, ok := secretbox.Open(nil, ciphertext[NonceSize:], &nonce, &key)
	if !ok {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// VerifySignature verifies an Ed25519 signature.
//
// This matches the Python implementation:
//
//	verify_key = nacl.signing.VerifyKey(pubkey, encoder=nacl.encoding.RawEncoder)
//	verify_key.verify(message, signature)
//
// Parameters:
//   - pubkey: The Ed25519 public key (32 bytes)
//   - message: The message that was signed
//   - signature: The signature to verify (64 bytes)
//
// Returns nil if the signature is valid, ErrInvalidSignature otherwise.
func VerifySignature(pubkey, message, signature []byte) error {
	if len(pubkey) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidKeySize, ed25519.PublicKeySize, len(pubkey))
	}

	if len(signature) != ed25519.SignatureSize {
		return ErrInvalidSignature
	}

	if !ed25519.Verify(pubkey, message, signature) {
		return ErrInvalidSignature
	}

	return nil
}

// GenerateRandomBytes generates cryptographically secure random bytes.
func GenerateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// GenerateStokenUID generates a random UID for Stoken (sync token).
// Format: 32 characters from [a-zA-Z0-9-_]
//
// This matches the Python implementation:
//
//	def generate_stoken_uid():
//	    return get_random_string(32, allowed_chars="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")
func GenerateStokenUID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	const length = 32

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate stoken UID: %w", err)
	}

	for i := range bytes {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}

	return string(bytes), nil
}

// GenerateItemUID generates a random UID for collection items.
// Format: 22+ characters from [a-zA-Z0-9-_]
func GenerateItemUID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	const length = 22

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate item UID: %w", err)
	}

	for i := range bytes {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}

	return string(bytes), nil
}

