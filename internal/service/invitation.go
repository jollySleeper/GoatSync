package service

import (
	"context"

	"goatsync/internal/model"
	"goatsync/internal/repository"
	pkgerrors "goatsync/pkg/errors"
)

// InvitationService handles invitation business logic
type InvitationService struct {
	invitationRepo repository.InvitationRepository
	memberRepo     repository.MemberRepository
	userRepo       repository.UserRepository
}

// NewInvitationService creates a new invitation service
func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	memberRepo repository.MemberRepository,
	userRepo repository.UserRepository,
) *InvitationService {
	return &InvitationService{
		invitationRepo: invitationRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
	}
}

// InvitationOut represents an invitation in API responses
type InvitationOut struct {
	UID                 string `msgpack:"uid"`
	FromUsername        string `msgpack:"fromUsername,omitempty"`
	FromPubkey          []byte `msgpack:"fromPubkey,omitempty"`
	SignedEncryptionKey []byte `msgpack:"signedEncryptionKey"`
	AccessLevel         string `msgpack:"accessLevel"`
	Username            string `msgpack:"username,omitempty"`
	CollectionUID       string `msgpack:"collection,omitempty"`
}

// InvitationListResponse is the response for listing invitations
type InvitationListResponse struct {
	Data []InvitationOut `msgpack:"data"`
	Done bool            `msgpack:"done"`
}

// ListIncoming lists incoming invitations for a user
func (s *InvitationService) ListIncoming(ctx context.Context, userID uint) (*InvitationListResponse, error) {
	invitations, err := s.invitationRepo.ListIncoming(ctx, userID)
	if err != nil {
		return nil, err
	}

	data := make([]InvitationOut, len(invitations))
	for i, inv := range invitations {
		data[i] = InvitationOut{
			UID:                 inv.UID,
			SignedEncryptionKey: inv.SignedEncryptionKey,
			AccessLevel:         accessLevelToString(inv.AccessLevel),
		}
		if inv.FromMember != nil && inv.FromMember.User != nil {
			data[i].FromUsername = inv.FromMember.User.Username
			if inv.FromMember.User.UserInfo != nil {
				data[i].FromPubkey = inv.FromMember.User.UserInfo.Pubkey
			}
		}
	}

	return &InvitationListResponse{
		Data: data,
		Done: true,
	}, nil
}

// ListOutgoing lists outgoing invitations from a user
func (s *InvitationService) ListOutgoing(ctx context.Context, collectionID, userID uint) (*InvitationListResponse, error) {
	// Get member for this user
	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, collectionID)
	if err != nil {
		return nil, err
	}
	if member == nil || !member.IsAdmin() {
		return nil, pkgerrors.ErrAdminRequired
	}

	invitations, err := s.invitationRepo.ListOutgoing(ctx, member.ID)
	if err != nil {
		return nil, err
	}

	data := make([]InvitationOut, len(invitations))
	for i, inv := range invitations {
		data[i] = InvitationOut{
			UID:                 inv.UID,
			SignedEncryptionKey: inv.SignedEncryptionKey,
			AccessLevel:         accessLevelToString(inv.AccessLevel),
		}
		if inv.User != nil {
			data[i].Username = inv.User.Username
		}
	}

	return &InvitationListResponse{
		Data: data,
		Done: true,
	}, nil
}

// GetIncoming retrieves a single incoming invitation
func (s *InvitationService) GetIncoming(ctx context.Context, uid string, userID uint) (*InvitationOut, error) {
	inv, err := s.invitationRepo.GetByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.UserID != userID {
		return nil, pkgerrors.ErrNotMember
	}

	out := &InvitationOut{
		UID:                 inv.UID,
		SignedEncryptionKey: inv.SignedEncryptionKey,
		AccessLevel:         accessLevelToString(inv.AccessLevel),
	}
	if inv.FromMember != nil && inv.FromMember.User != nil {
		out.FromUsername = inv.FromMember.User.Username
	}

	return out, nil
}

// RejectInvitation rejects (deletes) an incoming invitation
func (s *InvitationService) RejectInvitation(ctx context.Context, uid string, userID uint) error {
	inv, err := s.invitationRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	if inv == nil || inv.UserID != userID {
		return pkgerrors.ErrNotMember
	}

	return s.invitationRepo.Delete(ctx, inv.ID)
}

// AcceptInvitation accepts an invitation and creates a membership
func (s *InvitationService) AcceptInvitation(ctx context.Context, uid string, userID uint, encryptionKey []byte) error {
	inv, err := s.invitationRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}
	if inv == nil || inv.UserID != userID {
		return pkgerrors.ErrNotMember
	}

	// Create membership
	member := &model.CollectionMember{
		CollectionID:  inv.FromMember.CollectionID,
		UserID:        userID,
		EncryptionKey: encryptionKey,
		AccessLevel:   inv.AccessLevel,
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return err
	}

	// Delete the invitation
	return s.invitationRepo.Delete(ctx, inv.ID)
}

