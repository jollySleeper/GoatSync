package model

// AccessLevel represents the access level a member has to a collection.
// These values must match the Python AccessLevels enum.
type AccessLevel int

const (
	AccessLevelReadOnly  AccessLevel = 0 // Can only read items
	AccessLevelAdmin     AccessLevel = 1 // Can manage members and write
	AccessLevelReadWrite AccessLevel = 2 // Can read and write items
)

// CollectionMember represents a user's membership in a collection.
// This enables sharing collections between users with different access levels.
//
// Django model reference:
//
//	class CollectionMember(models.Model):
//	    stoken = models.OneToOneField(Stoken, on_delete=PROTECT, null=True)
//	    collection = models.ForeignKey(Collection, on_delete=CASCADE)
//	    user = models.ForeignKey(AUTH_USER_MODEL, on_delete=CASCADE)
//	    encryptionKey = models.BinaryField()
//	    collectionType = models.ForeignKey(CollectionType, on_delete=PROTECT, null=True)
//	    accessLevel = models.IntegerField(choices=AccessLevels.choices, default=READ_ONLY)
//
//	    class Meta:
//	        unique_together = ("user", "collection")
type CollectionMember struct {
	ID               uint        `gorm:"primaryKey"`
	CollectionID     uint        `gorm:"not null;index"`           // Foreign key to Collection
	UserID           uint        `gorm:"not null;index"`           // Foreign key to User
	StokenID         *uint       `gorm:"unique"`                   // Optional one-to-one with Stoken
	EncryptionKey    []byte      `gorm:"type:bytea;not null"`      // Encrypted collection key for this member
	CollectionTypeID *uint       `gorm:"index"`                    // Optional foreign key to CollectionType
	AccessLevel      AccessLevel `gorm:"default:0"`                // Default: read-only

	// Relations
	Collection     *Collection     `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE"`
	User           *User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Stoken         *Stoken         `gorm:"foreignKey:StokenID;constraint:OnDelete:RESTRICT"`
	CollectionType *CollectionType `gorm:"foreignKey:CollectionTypeID;constraint:OnDelete:RESTRICT"`
}

// TableName specifies the table name for GORM
func (CollectionMember) TableName() string {
	return "django_collectionmember"
}

// IsAdmin returns true if the member has admin access level
func (m *CollectionMember) IsAdmin() bool {
	return m.AccessLevel == AccessLevelAdmin
}

// CanWrite returns true if the member can write to the collection
func (m *CollectionMember) CanWrite() bool {
	return m.AccessLevel == AccessLevelAdmin || m.AccessLevel == AccessLevelReadWrite
}

// CollectionMemberRemoved tracks users who have been removed from a collection.
// This is used to notify clients that they should delete local data for this collection.
//
// Django model reference:
//
//	class CollectionMemberRemoved(models.Model):
//	    stoken = models.OneToOneField(Stoken, on_delete=PROTECT, null=True)
//	    collection = models.ForeignKey(Collection, on_delete=CASCADE)
//	    user = models.ForeignKey(AUTH_USER_MODEL, on_delete=CASCADE)
//
//	    class Meta:
//	        unique_together = ("user", "collection")
type CollectionMemberRemoved struct {
	ID           uint  `gorm:"primaryKey"`
	CollectionID uint  `gorm:"not null;index"`
	UserID       uint  `gorm:"not null;index"`
	StokenID     *uint `gorm:"unique"`

	// Relations
	Collection *Collection `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE"`
	User       *User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Stoken     *Stoken     `gorm:"foreignKey:StokenID;constraint:OnDelete:RESTRICT"`
}

// TableName specifies the table name for GORM
func (CollectionMemberRemoved) TableName() string {
	return "django_collectionmemberremoved"
}

