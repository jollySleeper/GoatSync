package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"
	pkgerrors "goatsync/pkg/errors"

	"gorm.io/gorm"
)

// stokenRepository implements StokenRepository using GORM
type stokenRepository struct {
	db *gorm.DB
}

// NewStokenRepository creates a new stoken repository
func NewStokenRepository(db *gorm.DB) StokenRepository {
	return &stokenRepository{db: db}
}

// Create creates a new stoken
// The UID is automatically generated in the BeforeCreate hook
func (r *stokenRepository) Create(ctx context.Context) (*model.Stoken, error) {
	stoken := &model.Stoken{}
	if err := r.db.WithContext(ctx).Create(stoken).Error; err != nil {
		return nil, err
	}
	return stoken, nil
}

// GetByID retrieves a stoken by ID
func (r *stokenRepository) GetByID(ctx context.Context, id uint) (*model.Stoken, error) {
	var stoken model.Stoken
	err := r.db.WithContext(ctx).First(&stoken, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &stoken, nil
}

// GetByUID retrieves a stoken by UID
// Returns ErrBadStoken if the UID is provided but not found
func (r *stokenRepository) GetByUID(ctx context.Context, uid string) (*model.Stoken, error) {
	if uid == "" {
		return nil, nil // Empty UID means "from the beginning"
	}

	var stoken model.Stoken
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&stoken).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, pkgerrors.ErrBadStoken
	}
	if err != nil {
		return nil, err
	}
	return &stoken, nil
}

// StokenFilter provides helper methods for stoken-based filtering
type StokenFilter struct {
	db *gorm.DB
}

// NewStokenFilter creates a new stoken filter helper
func NewStokenFilter(db *gorm.DB) *StokenFilter {
	return &StokenFilter{db: db}
}

// FilterByStokenAndLimit filters a query by stoken and applies limit+1 to check if done.
// It returns the filtered results, the new stoken, and whether we've reached the end.
//
// Usage pattern:
//
//	filter := repository.NewStokenFilter(db)
//	results, newStoken, done, err := filter.FilterCollections(ctx, userID, stokenUID, limit)
func (f *StokenFilter) FilterQuery(
	ctx context.Context,
	query *gorm.DB,
	stokenUID string,
	limit int,
	maxStokenField string, // e.g., "max_stoken" - the annotated field name
) (*gorm.DB, *model.Stoken, error) {
	// If stokenUID is provided, get the stoken object
	var stokenObj *model.Stoken
	if stokenUID != "" {
		var err error
		stokenObj, err = NewStokenRepository(f.db).GetByUID(ctx, stokenUID)
		if err != nil {
			return nil, nil, err
		}
	}

	// Apply stoken filter if we have one
	if stokenObj != nil {
		query = query.Where(maxStokenField+" > ?", stokenObj.ID)
	}

	return query, stokenObj, nil
}

// GetNewStoken retrieves the stoken for the maximum stoken ID in the result set.
// If no results, returns the original stoken or nil.
func (f *StokenFilter) GetNewStoken(ctx context.Context, maxStokenID uint) (*model.Stoken, error) {
	if maxStokenID == 0 {
		return nil, nil
	}

	var stoken model.Stoken
	err := f.db.WithContext(ctx).First(&stoken, maxStokenID).Error
	if err != nil {
		return nil, err
	}
	return &stoken, nil
}

