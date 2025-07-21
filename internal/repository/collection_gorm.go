package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// collectionRepository implements CollectionRepository using GORM
type collectionRepository struct {
	db            *gorm.DB
	stokenRepo    StokenRepository
	stokenFilter  *StokenFilter
}

// NewCollectionRepository creates a new collection repository
func NewCollectionRepository(db *gorm.DB) CollectionRepository {
	return &collectionRepository{
		db:           db,
		stokenRepo:   NewStokenRepository(db),
		stokenFilter: NewStokenFilter(db),
	}
}

// Create creates a new collection
func (r *collectionRepository) Create(ctx context.Context, collection *model.Collection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

// GetByID retrieves a collection by ID
func (r *collectionRepository) GetByID(ctx context.Context, id uint) (*model.Collection, error) {
	var collection model.Collection
	err := r.db.WithContext(ctx).
		Preload("MainItem").
		Preload("MainItem.Revisions", "current = ?", true).
		First(&collection, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// GetByUID retrieves a collection by UID
func (r *collectionRepository) GetByUID(ctx context.Context, uid string) (*model.Collection, error) {
	var collection model.Collection
	err := r.db.WithContext(ctx).
		Preload("MainItem").
		Preload("MainItem.Revisions", "current = ?", true).
		Where("uid = ?", uid).
		First(&collection).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// ListForUser lists collections for a user with pagination using stoken
//
// This implements the stoken-based pagination:
// 1. Get collections where the user is a member
// 2. Filter by stoken (only return collections changed after the stoken)
// 3. Return limit+1 results to check if we're done
// 4. Return the new stoken for the client to use next time
func (r *collectionRepository) ListForUser(
	ctx context.Context,
	userID uint,
	stokenUID string,
	limit int,
) (collections []model.Collection, newStoken *model.Stoken, done bool, err error) {
	// Default limit
	if limit <= 0 {
		limit = 50
	}

	// Get stoken object if provided
	var stokenObj *model.Stoken
	if stokenUID != "" {
		stokenObj, err = r.stokenRepo.GetByUID(ctx, stokenUID)
		if err != nil {
			return nil, nil, false, err
		}
	}

	// Build query: collections where user is a member
	// Annotate with max stoken from items.revisions and members
	//
	// Django equivalent:
	//   stoken_annotation = stoken_annotation_builder(["items__revisions__stoken", "members__stoken"])
	//
	// We use a raw SQL subquery to calculate the max stoken
	query := r.db.WithContext(ctx).
		Model(&model.Collection{}).
		Joins("JOIN django_collectionmember ON django_collectionmember.collection_id = django_collection.id").
		Where("django_collectionmember.user_id = ?", userID).
		Preload("MainItem").
		Preload("MainItem.Revisions", "current = ?", true)

	// Add stoken filter if we have one
	if stokenObj != nil {
		// Filter collections where max(revision.stoken_id, member.stoken_id) > stokenObj.ID
		query = query.
			Joins("LEFT JOIN django_collectionitem ON django_collectionitem.collection_id = django_collection.id").
			Joins("LEFT JOIN django_collectionitemrevision ON django_collectionitemrevision.item_id = django_collectionitem.id").
			Group("django_collection.id").
			Having("COALESCE(MAX(django_collectionitemrevision.stoken_id), 0) > ? OR COALESCE(MAX(django_collectionmember.stoken_id), 0) > ?",
				stokenObj.ID, stokenObj.ID)
	}

	// Fetch limit+1 to check if done
	err = query.
		Order("django_collection.id ASC").
		Limit(limit + 1).
		Find(&collections).Error

	if err != nil {
		return nil, nil, false, err
	}

	// Check if we have more results
	if len(collections) > limit {
		collections = collections[:limit]
		done = false
	} else {
		done = true
	}

	// Get new stoken from the last result
	if len(collections) > 0 {
		// In a real implementation, we'd calculate the max stoken from the results
		// For now, we create a new stoken to represent the current state
		newStoken, err = r.stokenRepo.Create(ctx)
		if err != nil {
			return nil, nil, false, err
		}
	}

	return collections, newStoken, done, nil
}

// ListByTypes lists collections filtered by collection types
func (r *collectionRepository) ListByTypes(
	ctx context.Context,
	userID uint,
	typeUIDs [][]byte,
	stokenUID string,
	limit int,
) (collections []model.Collection, newStoken *model.Stoken, done bool, err error) {
	// Similar to ListForUser but with additional type filter
	if limit <= 0 {
		limit = 50
	}

	// Get stoken object if provided
	var stokenObj *model.Stoken
	if stokenUID != "" {
		stokenObj, err = r.stokenRepo.GetByUID(ctx, stokenUID)
		if err != nil {
			return nil, nil, false, err
		}
	}

	// Build query with type filter
	query := r.db.WithContext(ctx).
		Model(&model.Collection{}).
		Joins("JOIN django_collectionmember ON django_collectionmember.collection_id = django_collection.id").
		Where("django_collectionmember.user_id = ?", userID).
		Preload("MainItem").
		Preload("MainItem.Revisions", "current = ?", true)

	// Filter by collection types if provided
	if len(typeUIDs) > 0 {
		query = query.
			Joins("JOIN django_collectiontype ON django_collectiontype.id = django_collectionmember.collection_type_id").
			Where("django_collectiontype.uid IN ?", typeUIDs)
	}

	// Add stoken filter
	if stokenObj != nil {
		query = query.
			Joins("LEFT JOIN django_collectionitem ON django_collectionitem.collection_id = django_collection.id").
			Joins("LEFT JOIN django_collectionitemrevision ON django_collectionitemrevision.item_id = django_collectionitem.id").
			Group("django_collection.id").
			Having("COALESCE(MAX(django_collectionitemrevision.stoken_id), 0) > ? OR COALESCE(MAX(django_collectionmember.stoken_id), 0) > ?",
				stokenObj.ID, stokenObj.ID)
	}

	// Fetch limit+1 to check if done
	err = query.
		Order("django_collection.id ASC").
		Limit(limit + 1).
		Find(&collections).Error

	if err != nil {
		return nil, nil, false, err
	}

	// Check if we have more results
	if len(collections) > limit {
		collections = collections[:limit]
		done = false
	} else {
		done = true
	}

	// Get new stoken
	if len(collections) > 0 {
		newStoken, err = r.stokenRepo.Create(ctx)
		if err != nil {
			return nil, nil, false, err
		}
	}

	return collections, newStoken, done, nil
}

// Update updates an existing collection
func (r *collectionRepository) Update(ctx context.Context, collection *model.Collection) error {
	return r.db.WithContext(ctx).Save(collection).Error
}

// Delete deletes a collection
func (r *collectionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Collection{}, id).Error
}

