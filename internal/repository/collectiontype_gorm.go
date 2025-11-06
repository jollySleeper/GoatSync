package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// collectionTypeRepository implements CollectionTypeRepository using GORM
type collectionTypeRepository struct {
	db *gorm.DB
}

// NewCollectionTypeRepository creates a new collection type repository
func NewCollectionTypeRepository(db *gorm.DB) CollectionTypeRepository {
	return &collectionTypeRepository{db: db}
}

// Create creates a new collection type
func (r *collectionTypeRepository) Create(ctx context.Context, colType *model.CollectionType) error {
	return r.db.WithContext(ctx).Create(colType).Error
}

// GetByUID retrieves a collection type by UID
func (r *collectionTypeRepository) GetByUID(ctx context.Context, uid []byte) (*model.CollectionType, error) {
	var colType model.CollectionType
	err := r.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&colType).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &colType, err
}

// GetOrCreate gets an existing type or creates a new one
func (r *collectionTypeRepository) GetOrCreate(ctx context.Context, ownerID uint, uid []byte) (*model.CollectionType, error) {
	var colType model.CollectionType
	err := r.db.WithContext(ctx).
		Where("owner_id = ? AND uid = ?", ownerID, uid).
		First(&colType).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		colType = model.CollectionType{
			OwnerID: ownerID,
			UID:     uid,
		}
		if err := r.db.WithContext(ctx).Create(&colType).Error; err != nil {
			return nil, err
		}
		return &colType, nil
	}

	return &colType, err
}

