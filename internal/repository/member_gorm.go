package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// memberRepository implements MemberRepository using GORM
type memberRepository struct {
	db         *gorm.DB
	stokenRepo StokenRepository
}

// NewMemberRepository creates a new member repository
func NewMemberRepository(db *gorm.DB) MemberRepository {
	return &memberRepository{
		db:         db,
		stokenRepo: NewStokenRepository(db),
	}
}

// Create creates a new collection member
func (r *memberRepository) Create(ctx context.Context, member *model.CollectionMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create stoken for the member
		stoken := &model.Stoken{}
		if err := tx.Create(stoken).Error; err != nil {
			return err
		}
		member.StokenID = &stoken.ID
		return tx.Create(member).Error
	})
}

// GetByID retrieves a member by ID
func (r *memberRepository) GetByID(ctx context.Context, id uint) (*model.CollectionMember, error) {
	var member model.CollectionMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.UserInfo").
		First(&member, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &member, err
}

// GetByUserAndCollection retrieves a member by user ID and collection ID
func (r *memberRepository) GetByUserAndCollection(ctx context.Context, userID, collectionID uint) (*model.CollectionMember, error) {
	var member model.CollectionMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ? AND collection_id = ?", userID, collectionID).
		First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &member, err
}

// GetByUsernameAndCollection retrieves a member by username and collection ID
func (r *memberRepository) GetByUsernameAndCollection(ctx context.Context, username string, collectionID uint) (*model.CollectionMember, error) {
	var member model.CollectionMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Joins("JOIN myauth_user ON myauth_user.id = django_collectionmember.user_id").
		Where("myauth_user.username = ? AND collection_id = ?", username, collectionID).
		First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &member, err
}

// ListForCollection lists all members of a collection
func (r *memberRepository) ListForCollection(ctx context.Context, collectionID uint) ([]model.CollectionMember, error) {
	var members []model.CollectionMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("collection_id = ?", collectionID).
		Find(&members).Error
	return members, err
}

// Update updates an existing member
func (r *memberRepository) Update(ctx context.Context, member *model.CollectionMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create new stoken on update
		stoken := &model.Stoken{}
		if err := tx.Create(stoken).Error; err != nil {
			return err
		}
		member.StokenID = &stoken.ID
		return tx.Save(member).Error
	})
}

// Delete removes a member from a collection
func (r *memberRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var member model.CollectionMember
		if err := tx.First(&member, id).Error; err != nil {
			return err
		}

		// Create removed member record
		stoken := &model.Stoken{}
		if err := tx.Create(stoken).Error; err != nil {
			return err
		}

		removed := &model.CollectionMemberRemoved{
			CollectionID: member.CollectionID,
			UserID:       member.UserID,
			StokenID:     &stoken.ID,
		}
		if err := tx.Create(removed).Error; err != nil {
			return err
		}

		return tx.Delete(&member).Error
	})
}

// GetRemovedMemberships lists removed memberships for a user since a stoken
func (r *memberRepository) GetRemovedMemberships(ctx context.Context, userID uint, stoken string) ([]model.CollectionMemberRemoved, error) {
	query := r.db.WithContext(ctx).
		Model(&model.CollectionMemberRemoved{}).
		Where("user_id = ?", userID)

	if stoken != "" {
		stokenObj, err := r.stokenRepo.GetByUID(ctx, stoken)
		if err != nil {
			return nil, err
		}
		if stokenObj != nil {
			query = query.Where("stoken_id > ?", stokenObj.ID)
		}
	}

	var removed []model.CollectionMemberRemoved
	err := query.Find(&removed).Error
	return removed, err
}

