package model

// CollectionItem represents an item within a collection.
// Each item can have multiple revisions, with one marked as "current".
//
// Django model reference:
//
//	class CollectionItem(models.Model):
//	    uid = models.CharField(max_length=43)
//	    collection = models.ForeignKey(Collection, on_delete=CASCADE)
//	    version = models.PositiveSmallIntegerField()
//	    encryptionKey = models.BinaryField(null=True)
//
//	    class Meta:
//	        unique_together = ("uid", "collection")
type CollectionItem struct {
	ID            uint   `gorm:"primaryKey"`
	UID           string `gorm:"size:43;not null;index"`      // Item UID
	CollectionID  uint   `gorm:"not null;index"`              // Foreign key to Collection
	Version       uint16 `gorm:"not null"`                    // Item version
	EncryptionKey []byte `gorm:"type:bytea"`                  // Optional encryption key

	// Relations
	Collection *Collection               `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE"`
	Revisions  []CollectionItemRevision `gorm:"foreignKey:ItemID"`
}

// TableName specifies the table name for GORM
func (CollectionItem) TableName() string {
	return "django_collectionitem"
}

// Unique constraint: (uid, collection_id)
// This is handled by GORM migration with composite index
func (CollectionItem) UniqueIndexes() []string {
	return []string{"idx_item_uid_collection UNIQUE (uid, collection_id)"}
}

