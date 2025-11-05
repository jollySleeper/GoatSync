package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// itemRepository implements ItemRepository using GORM
type itemRepository struct {
	db         *gorm.DB
	stokenRepo StokenRepository
}

// NewItemRepository creates a new item repository
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{
		db:         db,
		stokenRepo: NewStokenRepository(db),
	}
}

// Create creates a new collection item
func (r *itemRepository) Create(ctx context.Context, item *model.CollectionItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetByID retrieves an item by ID
func (r *itemRepository) GetByID(ctx context.Context, id uint) (*model.CollectionItem, error) {
	var item model.CollectionItem
	err := r.db.WithContext(ctx).
		Preload("Revisions", "current = ?", true).
		First(&item, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

// GetByUID retrieves an item by UID within a collection
func (r *itemRepository) GetByUID(ctx context.Context, collectionID uint, uid string) (*model.CollectionItem, error) {
	var item model.CollectionItem
	err := r.db.WithContext(ctx).
		Preload("Revisions", "current = ?", true).
		Where("collection_id = ? AND uid = ?", collectionID, uid).
		First(&item).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &item, err
}

// ListForCollection lists items for a collection with pagination
func (r *itemRepository) ListForCollection(
	ctx context.Context,
	collectionID uint,
	stoken string,
	limit int,
) (items []model.CollectionItem, newStoken *model.Stoken, done bool, err error) {
	if limit <= 0 {
		limit = 50
	}

	// Get stoken object if provided
	var stokenObj *model.Stoken
	if stoken != "" {
		stokenObj, err = r.stokenRepo.GetByUID(ctx, stoken)
		if err != nil {
			return nil, nil, false, err
		}
	}

	query := r.db.WithContext(ctx).
		Model(&model.CollectionItem{}).
		Where("collection_id = ?", collectionID).
		Preload("Revisions", "current = ?", true)

	// Apply stoken filter
	if stokenObj != nil {
		query = query.
			Joins("LEFT JOIN django_collectionitemrevision ON django_collectionitemrevision.item_id = django_collectionitem.id").
			Where("django_collectionitemrevision.stoken_id > ?", stokenObj.ID).
			Group("django_collectionitem.id")
	}

	err = query.
		Order("id ASC").
		Limit(limit + 1).
		Find(&items).Error

	if err != nil {
		return nil, nil, false, err
	}

	// Check if done
	if len(items) > limit {
		items = items[:limit]
		done = false
	} else {
		done = true
	}

	// Create new stoken
	if len(items) > 0 {
		newStoken, err = r.stokenRepo.Create(ctx)
		if err != nil {
			return nil, nil, false, err
		}
	}

	return items, newStoken, done, nil
}

// Update updates an existing item
func (r *itemRepository) Update(ctx context.Context, item *model.CollectionItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete deletes an item
func (r *itemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.CollectionItem{}, id).Error
}

