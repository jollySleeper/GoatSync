package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// chunkRepository implements ChunkRepository using GORM
type chunkRepository struct {
	db *gorm.DB
}

// NewChunkRepository creates a new chunk repository
func NewChunkRepository(db *gorm.DB) ChunkRepository {
	return &chunkRepository{db: db}
}

// Create creates a new chunk
func (r *chunkRepository) Create(ctx context.Context, chunk *model.CollectionItemChunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

// GetByUID retrieves a chunk by UID within a collection
func (r *chunkRepository) GetByUID(ctx context.Context, collectionID uint, uid string) (*model.CollectionItemChunk, error) {
	var chunk model.CollectionItemChunk
	err := r.db.WithContext(ctx).
		Where("collection_id = ? AND uid = ?", collectionID, uid).
		First(&chunk).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &chunk, err
}

// Delete deletes a chunk
func (r *chunkRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.CollectionItemChunk{}, id).Error
}

