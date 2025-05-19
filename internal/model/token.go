package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// AuthToken represents an authentication token for API access.
// Tokens are created on login and deleted on logout.
//
// Django model reference (token_auth):
//
//	class AuthToken(models.Model):
//	    key = models.CharField(max_length=40, primary_key=True)
//	    user = models.ForeignKey(AUTH_USER_MODEL, on_delete=CASCADE)
//	    created = models.DateTimeField(auto_now_add=True)
type AuthToken struct {
	Key       string    `gorm:"primaryKey;size:40"`    // 40-character hex token
	UserID    uint      `gorm:"not null;index"`        // Foreign key to User
	CreatedAt time.Time `gorm:"autoCreateTime"`        // Token creation time

	// Relations
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (AuthToken) TableName() string {
	return "django_authtoken_authtoken"
}

// BeforeCreate generates a random token key if not provided
func (t *AuthToken) BeforeCreate(tx *gorm.DB) error {
	if t.Key == "" {
		key, err := generateTokenKey()
		if err != nil {
			return err
		}
		t.Key = key
	}
	return nil
}

// generateTokenKey generates a 40-character hex token (20 random bytes)
func generateTokenKey() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsExpired checks if the token has expired (optional expiry feature)
// Note: Original EteSync tokens don't expire, but we can add this as an extension
func (t *AuthToken) IsExpired(maxAge time.Duration) bool {
	if maxAge == 0 {
		return false // No expiry
	}
	return time.Since(t.CreatedAt) > maxAge
}

