package repository

import (
	"context"
	"errors"
	"strings"

	"goatsync/internal/model"

	"gorm.io/gorm"
)

// userRepository implements UserRepository using GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user with associated UserInfo
func (r *userRepository) Create(ctx context.Context, user *model.User, userInfo *model.UserInfo) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store username lowercase for case-insensitive lookup
		// Original casing is stored in FirstName
		user.FirstName = user.Username
		user.Username = strings.ToLower(user.Username)
		user.Email = strings.ToLower(user.Email)

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Link UserInfo to User
		userInfo.OwnerID = user.ID
		if err := tx.Create(userInfo).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("UserInfo").
		First(&user, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username (case-insensitive)
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("UserInfo").
		Where("username = ?", strings.ToLower(username)).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email (case-insensitive)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("UserInfo").
		Where("email = ?", strings.ToLower(email)).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateUserInfo updates a user's UserInfo
func (r *userRepository) UpdateUserInfo(ctx context.Context, userInfo *model.UserInfo) error {
	return r.db.WithContext(ctx).Save(userInfo).Error
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

