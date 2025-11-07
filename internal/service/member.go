package service

import (
	"context"

	"goatsync/internal/model"
	"goatsync/internal/repository"
	pkgerrors "goatsync/pkg/errors"
)

// MemberService handles member business logic
type MemberService struct {
	memberRepo     repository.MemberRepository
	collectionRepo repository.CollectionRepository
}

// NewMemberService creates a new member service
func NewMemberService(
	memberRepo repository.MemberRepository,
	collectionRepo repository.CollectionRepository,
) *MemberService {
	return &MemberService{
		memberRepo:     memberRepo,
		collectionRepo: collectionRepo,
	}
}

// MemberOut represents a member in API responses
type MemberOut struct {
	Username    string `msgpack:"username"`
	AccessLevel string `msgpack:"accessLevel"`
}

// MemberListResponse is the response for listing members
type MemberListResponse struct {
	Data   []MemberOut `msgpack:"data"`
	Done   bool        `msgpack:"done"`
}

// ListMembers lists all members of a collection
func (s *MemberService) ListMembers(ctx context.Context, collectionUID string, userID uint) (*MemberListResponse, error) {
	// Get collection
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Check if user is admin
	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return nil, err
	}
	if member == nil || !member.IsAdmin() {
		return nil, pkgerrors.ErrAdminRequired
	}

	// List members
	members, err := s.memberRepo.ListForCollection(ctx, col.ID)
	if err != nil {
		return nil, err
	}

	data := make([]MemberOut, len(members))
	for i, m := range members {
		data[i] = MemberOut{
			Username:    m.User.Username,
			AccessLevel: accessLevelToString(m.AccessLevel),
		}
	}

	return &MemberListResponse{
		Data: data,
		Done: true,
	}, nil
}

// UpdateMemberAccess updates a member's access level
func (s *MemberService) UpdateMemberAccess(ctx context.Context, collectionUID, username string, userID uint, newAccessLevel model.AccessLevel) error {
	// Get collection
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return err
	}
	if col == nil {
		return pkgerrors.ErrNotMember
	}

	// Check if user is admin
	adminMember, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return err
	}
	if adminMember == nil || !adminMember.IsAdmin() {
		return pkgerrors.ErrAdminRequired
	}

	// TODO: Get target user by username and update their access level
	return nil
}

// RemoveMember removes a member from a collection
func (s *MemberService) RemoveMember(ctx context.Context, collectionUID, username string, userID uint) error {
	// Get collection
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return err
	}
	if col == nil {
		return pkgerrors.ErrNotMember
	}

	// Check if user is admin
	adminMember, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return err
	}
	if adminMember == nil || !adminMember.IsAdmin() {
		return pkgerrors.ErrAdminRequired
	}

	// TODO: Get target user by username and remove them
	return nil
}

// LeaveCollection allows a user to leave a collection
func (s *MemberService) LeaveCollection(ctx context.Context, collectionUID string, userID uint) error {
	// Get collection
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return err
	}
	if col == nil {
		return pkgerrors.ErrNotMember
	}

	// Get member
	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return err
	}
	if member == nil {
		return pkgerrors.ErrNotMember
	}

	// Can't leave if you're the owner (admin)
	if member.IsAdmin() && col.OwnerID == userID {
		return pkgerrors.ErrAdminRequired.WithDetail("Owner cannot leave collection")
	}

	return s.memberRepo.Delete(ctx, member.ID)
}

func accessLevelToString(level model.AccessLevel) string {
	switch level {
	case model.AccessLevelAdmin:
		return "admin"
	case model.AccessLevelReadWrite:
		return "readWrite"
	default:
		return "readOnly"
	}
}

