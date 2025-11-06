package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// invitationRepository implements InvitationRepository using GORM
type invitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository creates a new invitation repository
func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

// Create creates a new invitation
func (r *invitationRepository) Create(ctx context.Context, invitation *model.CollectionInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// GetByID retrieves an invitation by ID
func (r *invitationRepository) GetByID(ctx context.Context, id uint) (*model.CollectionInvitation, error) {
	var inv model.CollectionInvitation
	err := r.db.WithContext(ctx).
		Preload("FromMember").
		Preload("FromMember.Collection").
		Preload("User").
		First(&inv, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

// GetByUID retrieves an invitation by UID
func (r *invitationRepository) GetByUID(ctx context.Context, uid string) (*model.CollectionInvitation, error) {
	var inv model.CollectionInvitation
	err := r.db.WithContext(ctx).
		Preload("FromMember").
		Preload("FromMember.Collection").
		Preload("User").
		Where("uid = ?", uid).
		First(&inv).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

// ListIncoming lists incoming invitations for a user
func (r *invitationRepository) ListIncoming(ctx context.Context, userID uint) ([]model.CollectionInvitation, error) {
	var invitations []model.CollectionInvitation
	err := r.db.WithContext(ctx).
		Preload("FromMember").
		Preload("FromMember.Collection").
		Where("user_id = ?", userID).
		Find(&invitations).Error
	return invitations, err
}

// ListOutgoing lists outgoing invitations from a member
func (r *invitationRepository) ListOutgoing(ctx context.Context, memberID uint) ([]model.CollectionInvitation, error) {
	var invitations []model.CollectionInvitation
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("from_member_id = ?", memberID).
		Find(&invitations).Error
	return invitations, err
}

// Delete deletes an invitation
func (r *invitationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.CollectionInvitation{}, id).Error
}

// DeleteForCollection deletes all invitations for a collection
func (r *invitationRepository) DeleteForCollection(ctx context.Context, collectionID uint) error {
	return r.db.WithContext(ctx).
		Where("from_member_id IN (SELECT id FROM django_collectionmember WHERE collection_id = ?)", collectionID).
		Delete(&model.CollectionInvitation{}).Error
}

