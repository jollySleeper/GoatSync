package service

import (
	"context"

	"goatsync/internal/model"
	"goatsync/internal/repository"
	pkgerrors "goatsync/pkg/errors"
)

// ItemService handles item business logic
type ItemService struct {
	itemRepo       repository.ItemRepository
	revisionRepo   repository.RevisionRepository
	collectionRepo repository.CollectionRepository
	memberRepo     repository.MemberRepository
}

// NewItemService creates a new item service
func NewItemService(
	itemRepo repository.ItemRepository,
	revisionRepo repository.RevisionRepository,
	collectionRepo repository.CollectionRepository,
	memberRepo repository.MemberRepository,
) *ItemService {
	return &ItemService{
		itemRepo:       itemRepo,
		revisionRepo:   revisionRepo,
		collectionRepo: collectionRepo,
		memberRepo:     memberRepo,
	}
}

// ItemListResponse is the response for listing items
type ItemListResponse struct {
	Data   []ItemOut `msgpack:"data"`
	Stoken *string   `msgpack:"stoken,omitempty"`
	Done   bool      `msgpack:"done"`
}

// ListItems lists items in a collection
func (s *ItemService) ListItems(
	ctx context.Context,
	collectionUID string,
	userID uint,
	stoken string,
	limit int,
) (*ItemListResponse, error) {
	// Get collection and verify access
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Verify membership
	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// List items
	items, newStoken, done, err := s.itemRepo.ListForCollection(ctx, col.ID, stoken, limit)
	if err != nil {
		return nil, err
	}

	data := make([]ItemOut, len(items))
	for i, item := range items {
		data[i] = s.itemToOut(&item)
	}

	var stokenStr *string
	if newStoken != nil {
		stokenStr = &newStoken.UID
	}

	return &ItemListResponse{
		Data:   data,
		Stoken: stokenStr,
		Done:   done,
	}, nil
}

// GetItem retrieves a single item
func (s *ItemService) GetItem(
	ctx context.Context,
	collectionUID, itemUID string,
	userID uint,
) (*ItemOut, error) {
	// Get collection and verify access
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Verify membership
	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Get item
	item, err := s.itemRepo.GetByUID(ctx, col.ID, itemUID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, pkgerrors.ErrNotMember
	}

	out := s.itemToOut(item)
	return &out, nil
}

func (s *ItemService) itemToOut(item *model.CollectionItem) ItemOut {
	out := ItemOut{
		UID:     item.UID,
		Version: item.Version,
	}

	// Get current revision
	if len(item.Revisions) > 0 {
		for _, rev := range item.Revisions {
			if rev.Current != nil && *rev.Current {
				out.Etag = rev.UID
				out.Content = ContentOut{
					UID:     rev.UID,
					Meta:    rev.Meta,
					Deleted: rev.Deleted,
				}
				break
			}
		}
	}

	return out
}

