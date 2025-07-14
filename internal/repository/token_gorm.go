package repository

import (
	"context"
	"errors"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// tokenRepository implements TokenRepository using GORM
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

// Create creates a new auth token
func (r *tokenRepository) Create(ctx context.Context, token *model.AuthToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByKey retrieves a token by its key
func (r *tokenRepository) GetByKey(ctx context.Context, key string) (*model.AuthToken, error) {
	var token model.AuthToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.UserInfo").
		Where("key = ?", key).
		First(&token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Delete deletes a token by its key
func (r *tokenRepository) Delete(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("key = ?", key).Delete(&model.AuthToken{}).Error
}

// DeleteAllForUser deletes all tokens for a user
func (r *tokenRepository) DeleteAllForUser(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.AuthToken{}).Error
}

