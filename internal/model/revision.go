package model

// CollectionItemRevision represents a revision of a collection item.
// Items can have multiple revisions, but only one is marked as "current".
//
// Django model reference:
//
//	class CollectionItemRevision(models.Model):
//	    stoken = models.OneToOneField(Stoken, on_delete=PROTECT)
//	    uid = models.CharField(unique=True, max_length=43)
//	    item = models.ForeignKey(CollectionItem, on_delete=CASCADE)
//	    meta = models.BinaryField()
//	    current = models.BooleanField(default=True, null=True)
//	    deleted = models.BooleanField(default=False)
//
//	    class Meta:
//	        unique_together = ("item", "current")
type CollectionItemRevision struct {
	ID       uint   `gorm:"primaryKey"`
	UID      string `gorm:"uniqueIndex;size:43;not null"`     // Revision UID
	ItemID   uint   `gorm:"not null;index"`                   // Foreign key to CollectionItem
	StokenID uint   `gorm:"unique;not null"`                  // One-to-one with Stoken
	Meta     []byte `gorm:"type:bytea;not null"`              // Encrypted metadata
	Current  *bool  `gorm:"index"`                            // Nullable, only one can be true per item
	Deleted  bool   `gorm:"default:false"`                    // Soft delete flag

	// Relations
	Item   *CollectionItem `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE"`
	Stoken *Stoken         `gorm:"foreignKey:StokenID;constraint:OnDelete:RESTRICT"`
	Chunks []RevisionChunkRelation `gorm:"foreignKey:RevisionID"`
}

// TableName specifies the table name for GORM
func (CollectionItemRevision) TableName() string {
	return "django_collectionitemrevision"
}

