package model

// CollectionInvitation represents an invitation to join a collection.
// A member with admin access can invite other users to join a collection.
//
// Django model reference:
//
//	class CollectionInvitation(models.Model):
//	    uid = models.CharField(max_length=43)
//	    version = models.PositiveSmallIntegerField(default=1)
//	    fromMember = models.ForeignKey(CollectionMember, on_delete=CASCADE)
//	    user = models.ForeignKey(AUTH_USER_MODEL, on_delete=CASCADE)
//	    signedEncryptionKey = models.BinaryField()
//	    accessLevel = models.IntegerField(choices=AccessLevels.choices, default=READ_ONLY)
//
//	    class Meta:
//	        unique_together = ("user", "fromMember")
type CollectionInvitation struct {
	ID                  uint        `gorm:"primaryKey"`
	UID                 string      `gorm:"size:43;not null;index"`           // Invitation UID
	Version             uint16      `gorm:"default:1"`                        // Protocol version
	FromMemberID        uint        `gorm:"not null;index"`                   // Foreign key to CollectionMember
	UserID              uint        `gorm:"not null;index"`                   // User being invited
	SignedEncryptionKey []byte      `gorm:"type:bytea;not null"`              // Signed encryption key
	AccessLevel         AccessLevel `gorm:"default:0"`                        // Default: read-only

	// Relations
	FromMember *CollectionMember `gorm:"foreignKey:FromMemberID;constraint:OnDelete:CASCADE"`
	User       *User             `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (CollectionInvitation) TableName() string {
	return "django_collectioninvitation"
}

