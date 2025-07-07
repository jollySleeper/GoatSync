// Package repository provides data access layer interfaces and implementations.
// Interfaces are defined where consumed (in service layer), but we also define them
// here for documentation and reference.
package repository

import (
	"context"

	"goatsync/internal/model"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	// Create creates a new user with associated UserInfo
	Create(ctx context.Context, user *model.User, userInfo *model.UserInfo) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id uint) (*model.User, error)

	// GetByUsername retrieves a user by username (case-insensitive)
	GetByUsername(ctx context.Context, username string) (*model.User, error)

	// GetByEmail retrieves a user by email (case-insensitive)
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *model.User) error

	// UpdateUserInfo updates a user's UserInfo
	UpdateUserInfo(ctx context.Context, userInfo *model.UserInfo) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id uint) error
}

// TokenRepository defines the interface for auth token data access.
type TokenRepository interface {
	// Create creates a new auth token
	Create(ctx context.Context, token *model.AuthToken) error

	// GetByKey retrieves a token by its key
	GetByKey(ctx context.Context, key string) (*model.AuthToken, error)

	// Delete deletes a token by its key
	Delete(ctx context.Context, key string) error

	// DeleteAllForUser deletes all tokens for a user
	DeleteAllForUser(ctx context.Context, userID uint) error
}

// StokenRepository defines the interface for sync token data access.
type StokenRepository interface {
	// Create creates a new stoken
	Create(ctx context.Context) (*model.Stoken, error)

	// GetByID retrieves a stoken by ID
	GetByID(ctx context.Context, id uint) (*model.Stoken, error)

	// GetByUID retrieves a stoken by UID
	GetByUID(ctx context.Context, uid string) (*model.Stoken, error)
}

// CollectionRepository defines the interface for collection data access.
type CollectionRepository interface {
	// Create creates a new collection
	Create(ctx context.Context, collection *model.Collection) error

	// GetByID retrieves a collection by ID
	GetByID(ctx context.Context, id uint) (*model.Collection, error)

	// GetByUID retrieves a collection by UID
	GetByUID(ctx context.Context, uid string) (*model.Collection, error)

	// ListForUser lists collections for a user with pagination
	ListForUser(ctx context.Context, userID uint, stoken string, limit int) (
		collections []model.Collection,
		newStoken *model.Stoken,
		done bool,
		err error,
	)

	// ListByTypes lists collections filtered by collection types
	ListByTypes(ctx context.Context, userID uint, typeUIDs [][]byte, stoken string, limit int) (
		collections []model.Collection,
		newStoken *model.Stoken,
		done bool,
		err error,
	)

	// Update updates an existing collection
	Update(ctx context.Context, collection *model.Collection) error

	// Delete deletes a collection
	Delete(ctx context.Context, id uint) error
}

// ItemRepository defines the interface for collection item data access.
type ItemRepository interface {
	// Create creates a new collection item
	Create(ctx context.Context, item *model.CollectionItem) error

	// GetByID retrieves an item by ID
	GetByID(ctx context.Context, id uint) (*model.CollectionItem, error)

	// GetByUID retrieves an item by UID within a collection
	GetByUID(ctx context.Context, collectionID uint, uid string) (*model.CollectionItem, error)

	// ListForCollection lists items for a collection with pagination
	ListForCollection(ctx context.Context, collectionID uint, stoken string, limit int) (
		items []model.CollectionItem,
		newStoken *model.Stoken,
		done bool,
		err error,
	)

	// Update updates an existing item
	Update(ctx context.Context, item *model.CollectionItem) error

	// Delete deletes an item
	Delete(ctx context.Context, id uint) error
}

// RevisionRepository defines the interface for revision data access.
type RevisionRepository interface {
	// Create creates a new revision (also creates associated stoken)
	Create(ctx context.Context, revision *model.CollectionItemRevision) error

	// GetByUID retrieves a revision by UID
	GetByUID(ctx context.Context, uid string) (*model.CollectionItemRevision, error)

	// GetCurrentForItem retrieves the current revision for an item
	GetCurrentForItem(ctx context.Context, itemID uint) (*model.CollectionItemRevision, error)

	// ListForItem lists all revisions for an item
	ListForItem(ctx context.Context, itemID uint, limit int) ([]model.CollectionItemRevision, error)
}

// MemberRepository defines the interface for collection member data access.
type MemberRepository interface {
	// Create creates a new collection member
	Create(ctx context.Context, member *model.CollectionMember) error

	// GetByID retrieves a member by ID
	GetByID(ctx context.Context, id uint) (*model.CollectionMember, error)

	// GetByUserAndCollection retrieves a member by user ID and collection ID
	GetByUserAndCollection(ctx context.Context, userID, collectionID uint) (*model.CollectionMember, error)

	// ListForCollection lists all members of a collection
	ListForCollection(ctx context.Context, collectionID uint) ([]model.CollectionMember, error)

	// Update updates an existing member
	Update(ctx context.Context, member *model.CollectionMember) error

	// Delete removes a member from a collection (also creates MemberRemoved record)
	Delete(ctx context.Context, id uint) error

	// GetRemovedMemberships lists removed memberships for a user since a stoken
	GetRemovedMemberships(ctx context.Context, userID uint, stoken string) ([]model.CollectionMemberRemoved, error)
}

// InvitationRepository defines the interface for invitation data access.
type InvitationRepository interface {
	// Create creates a new invitation
	Create(ctx context.Context, invitation *model.CollectionInvitation) error

	// GetByID retrieves an invitation by ID
	GetByID(ctx context.Context, id uint) (*model.CollectionInvitation, error)

	// GetByUID retrieves an invitation by UID
	GetByUID(ctx context.Context, uid string) (*model.CollectionInvitation, error)

	// ListIncoming lists incoming invitations for a user
	ListIncoming(ctx context.Context, userID uint) ([]model.CollectionInvitation, error)

	// ListOutgoing lists outgoing invitations from a member
	ListOutgoing(ctx context.Context, memberID uint) ([]model.CollectionInvitation, error)

	// Delete deletes an invitation
	Delete(ctx context.Context, id uint) error

	// DeleteForCollection deletes all invitations for a collection
	DeleteForCollection(ctx context.Context, collectionID uint) error
}

// ChunkRepository defines the interface for chunk data access.
type ChunkRepository interface {
	// Create creates a new chunk
	Create(ctx context.Context, chunk *model.CollectionItemChunk) error

	// GetByUID retrieves a chunk by UID within a collection
	GetByUID(ctx context.Context, collectionID uint, uid string) (*model.CollectionItemChunk, error)

	// Delete deletes a chunk
	Delete(ctx context.Context, id uint) error
}

// CollectionTypeRepository defines the interface for collection type data access.
type CollectionTypeRepository interface {
	// Create creates a new collection type
	Create(ctx context.Context, colType *model.CollectionType) error

	// GetByUID retrieves a collection type by UID
	GetByUID(ctx context.Context, uid []byte) (*model.CollectionType, error)

	// GetOrCreate gets an existing type or creates a new one
	GetOrCreate(ctx context.Context, ownerID uint, uid []byte) (*model.CollectionType, error)
}

