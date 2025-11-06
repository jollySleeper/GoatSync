package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// revisionRepository implements RevisionRepository using GORM
type revisionRepository struct {
	db         *gorm.DB
	stokenRepo StokenRepository
}

// NewRevisionRepository creates a new revision repository
func NewRevisionRepository(db *gorm.DB) RevisionRepository {
	return &revisionRepository{
		db:         db,
		stokenRepo: NewStokenRepository(db),
	}
}

// Create creates a new revision with associated stoken
func (r *revisionRepository) Create(ctx context.Context, revision *model.CollectionItemRevision) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create stoken for the revision
		stoken := &model.Stoken{}
		if err := tx.Create(stoken).Error; err != nil {
			return err
		}
		revision.StokenID = stoken.ID

		// Mark previous current revision as not current
		if err := tx.Model(&model.CollectionItemRevision{}).
			Where("item_id = ? AND current = ?", revision.ItemID, true).
			Update("current", nil).Error; err != nil {
			return err
		}

		// Set this revision as current
		current := true
		revision.Current = &current

		return tx.Create(revision).Error
	})
}

// GetByUID retrieves a revision by UID
func (r *revisionRepository) GetByUID(ctx context.Context, uid string) (*model.CollectionItemRevision, error) {
	var rev model.CollectionItemRevision
	err := r.db.WithContext(ctx).
		Preload("Chunks").
		Where("uid = ?", uid).
		First(&rev).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rev, err
}

// GetCurrentForItem retrieves the current revision for an item
func (r *revisionRepository) GetCurrentForItem(ctx context.Context, itemID uint) (*model.CollectionItemRevision, error) {
	var rev model.CollectionItemRevision
	err := r.db.WithContext(ctx).
		Preload("Chunks").
		Where("item_id = ? AND current = ?", itemID, true).
		First(&rev).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rev, err
}

// ListForItem lists all revisions for an item
func (r *revisionRepository) ListForItem(ctx context.Context, itemID uint, limit int) ([]model.CollectionItemRevision, error) {
	if limit <= 0 {
		limit = 50
	}

	var revisions []model.CollectionItemRevision
	err := r.db.WithContext(ctx).
		Preload("Chunks").
		Where("item_id = ?", itemID).
		Order("id DESC").
		Limit(limit).
		Find(&revisions).Error

	return revisions, err
}

