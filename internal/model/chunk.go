package model

// CollectionItemChunk represents a chunk of binary data for a collection item.
// Large items are split into chunks for efficient transfer and storage.
//
// Django model reference:
//
//	class CollectionItemChunk(models.Model):
//	    uid = models.CharField(max_length=60)
//	    collection = models.ForeignKey(Collection, on_delete=CASCADE)
//	    chunkFile = models.FileField(upload_to=chunk_directory_path, max_length=150, unique=True)
//
//	    class Meta:
//	        unique_together = ("collection", "uid")
type CollectionItemChunk struct {
	ID           uint   `gorm:"primaryKey"`
	UID          string `gorm:"size:60;not null;index"`            // Chunk UID
	CollectionID uint   `gorm:"not null;index"`                    // Foreign key to Collection
	ChunkFile    string `gorm:"size:150;uniqueIndex;not null"`     // Path to chunk file

	// Relations
	Collection *Collection `gorm:"foreignKey:CollectionID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (CollectionItemChunk) TableName() string {
	return "django_collectionitemchunk"
}

// RevisionChunkRelation is a many-to-many relationship between revisions and chunks.
// A revision can have multiple chunks, and chunks can be shared between revisions.
//
// Django model reference:
//
//	class RevisionChunkRelation(models.Model):
//	    chunk = models.ForeignKey(CollectionItemChunk, on_delete=CASCADE)
//	    revision = models.ForeignKey(CollectionItemRevision, on_delete=CASCADE)
//
//	    class Meta:
//	        ordering = ("id",)
type RevisionChunkRelation struct {
	ID         uint `gorm:"primaryKey"`
	ChunkID    uint `gorm:"not null;index"`    // Foreign key to CollectionItemChunk
	RevisionID uint `gorm:"not null;index"`    // Foreign key to CollectionItemRevision

	// Relations
	Chunk    *CollectionItemChunk    `gorm:"foreignKey:ChunkID;constraint:OnDelete:CASCADE"`
	Revision *CollectionItemRevision `gorm:"foreignKey:RevisionID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (RevisionChunkRelation) TableName() string {
	return "django_revisionchunkrelation"
}

