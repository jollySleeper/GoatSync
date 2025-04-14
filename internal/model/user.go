package model

import (
	"time"
)

// User represents a user in the system.
// Django uses AUTH_USER_MODEL which includes username, email, first_name, etc.
//
// Note: Django stores the original username casing in first_name field,
// while username is stored lowercase for case-insensitive lookup.
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;size:150;not null"` // Stored lowercase
	Email     string    `gorm:"uniqueIndex;size:254;not null"`
	FirstName string    `gorm:"size:150"` // Stores original username casing
	LastName  string    `gorm:"size:150"`
	IsActive  bool      `gorm:"default:true"`
	IsStaff   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"column:date_joined;autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	UserInfo *UserInfo `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "myauth_user"
}

// UserInfo contains user-specific cryptographic information for Etebase.
//
// Django model reference:
//
//	class UserInfo(models.Model):
//	    owner = models.OneToOneField(AUTH_USER_MODEL, on_delete=CASCADE, primary_key=True)
//	    version = models.PositiveSmallIntegerField(default=1)
//	    loginPubkey = models.BinaryField()
//	    pubkey = models.BinaryField()
//	    encryptedContent = models.BinaryField()
//	    salt = models.BinaryField()
type UserInfo struct {
	OwnerID          uint   `gorm:"primaryKey"`          // References User.ID
	Version          int    `gorm:"default:1"`           // Protocol version
	LoginPubkey      []byte `gorm:"type:bytea;not null"` // Ed25519 public key for login
	Pubkey           []byte `gorm:"type:bytea;not null"` // Main Ed25519 public key
	EncryptedContent []byte `gorm:"type:bytea;not null"` // Encrypted user data
	Salt             []byte `gorm:"type:bytea;not null"` // Salt for key derivation

	// Relation back to User
	Owner *User `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (UserInfo) TableName() string {
	return "django_userinfo"
}

