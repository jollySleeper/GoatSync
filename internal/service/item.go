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

// ItemBatchIn represents an item in a batch request
type ItemBatchIn struct {
	UID     string     `msgpack:"uid"`
	Version uint16     `msgpack:"version"`
	Etag    *string    `msgpack:"etag,omitempty"`
	Content ContentIn  `msgpack:"content"`
}

// ContentIn represents item content in a batch request
type ContentIn struct {
	UID     string   `msgpack:"uid"`
	Meta    []byte   `msgpack:"meta"`
	Deleted bool     `msgpack:"deleted"`
	Chunks  []string `msgpack:"chunks,omitempty"`
}

// ItemBatchRequest is the request for batch item operations
type ItemBatchRequest struct {
	Items []ItemBatchIn `msgpack:"items"`
}

// ItemTransactionRequest is the request for transaction item operations
type ItemTransactionRequest struct {
	Items []ItemBatchIn `msgpack:"items"`
	Deps  *Dependencies `msgpack:"deps,omitempty"`
}

// Dependencies represents item dependencies for transactions
type Dependencies struct {
	Stoken string `msgpack:"stoken,omitempty"`
}

// BatchItems processes a batch of item updates (no etag validation)
func (s *ItemService) BatchItems(
	ctx context.Context,
	collectionUID string,
	userID uint,
	req *ItemBatchRequest,
) error {
	// Get collection and verify write access
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return err
	}
	if col == nil {
		return pkgerrors.ErrNotMember
	}

	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return err
	}
	if member == nil {
		return pkgerrors.ErrNotMember
	}
	if !member.CanWrite() {
		return pkgerrors.ErrNoWriteAccess
	}

	// Process each item
	for _, itemIn := range req.Items {
		if err := s.processItemUpdate(ctx, col.ID, &itemIn); err != nil {
			return err
		}
	}

	return nil
}

// TransactionItems processes a batch of item updates with etag validation
func (s *ItemService) TransactionItems(
	ctx context.Context,
	collectionUID string,
	userID uint,
	req *ItemTransactionRequest,
) error {
	// Get collection and verify write access
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return err
	}
	if col == nil {
		return pkgerrors.ErrNotMember
	}

	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return err
	}
	if member == nil {
		return pkgerrors.ErrNotMember
	}
	if !member.CanWrite() {
		return pkgerrors.ErrNoWriteAccess
	}

	// Process each item with etag validation
	for _, itemIn := range req.Items {
		// Validate etag if provided
		if itemIn.Etag != nil {
			existing, err := s.itemRepo.GetByUID(ctx, col.ID, itemIn.UID)
			if err != nil {
				return err
			}
			if existing != nil {
				currentEtag := s.getItemEtag(existing)
				if currentEtag != *itemIn.Etag {
					return pkgerrors.NewWrongEtagError(*itemIn.Etag, currentEtag)
				}
			}
		}

		if err := s.processItemUpdate(ctx, col.ID, &itemIn); err != nil {
			return err
		}
	}

	return nil
}

// FetchUpdatesRequest is the request for fetching item updates
type FetchUpdatesRequest struct {
	Items []ItemFetchIn `msgpack:"items"`
}

// ItemFetchIn represents an item in a fetch updates request
type ItemFetchIn struct {
	UID  string `msgpack:"uid"`
	Etag string `msgpack:"etag"`
}

// FetchUpdatesResponse is the response for fetch updates
type FetchUpdatesResponse struct {
	Data []ItemOut `msgpack:"data"`
}

// RevisionListResponse is the response for listing revisions
type RevisionListResponse struct {
	Data     []RevisionOut `msgpack:"data"`
	Iterator *string       `msgpack:"iterator,omitempty"`
	Done     bool          `msgpack:"done"`
}

// RevisionOut represents a revision in API responses
type RevisionOut struct {
	UID     string `msgpack:"uid"`
	Meta    []byte `msgpack:"meta"`
	Deleted bool   `msgpack:"deleted"`
}

// GetItemRevisions retrieves revisions for an item
func (s *ItemService) GetItemRevisions(
	ctx context.Context,
	collectionUID, itemUID string,
	userID uint,
	iterator string,
	limit int,
) (*RevisionListResponse, error) {
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

	// Get revisions
	revisions, err := s.revisionRepo.ListForItem(ctx, item.ID, limit)
	if err != nil {
		return nil, err
	}

	data := make([]RevisionOut, len(revisions))
	for i, rev := range revisions {
		data[i] = RevisionOut{
			UID:     rev.UID,
			Meta:    rev.Meta,
			Deleted: rev.Deleted,
		}
	}

	// Set iterator if there are more results
	var iteratorStr *string
	done := len(revisions) < limit
	if !done && len(revisions) > 0 {
		lastUID := revisions[len(revisions)-1].UID
		iteratorStr = &lastUID
	}

	return &RevisionListResponse{
		Data:     data,
		Iterator: iteratorStr,
		Done:     done,
	}, nil
}

// FetchUpdates returns items that have changed since the given etags
func (s *ItemService) FetchUpdates(
	ctx context.Context,
	collectionUID string,
	userID uint,
	req *FetchUpdatesRequest,
) (*FetchUpdatesResponse, error) {
	// Get collection and verify access
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, pkgerrors.ErrNotMember
	}

	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Check each item for changes
	var changed []ItemOut
	for _, itemIn := range req.Items {
		item, err := s.itemRepo.GetByUID(ctx, col.ID, itemIn.UID)
		if err != nil {
			return nil, err
		}
		if item == nil {
			continue
		}

		currentEtag := s.getItemEtag(item)
		if currentEtag != itemIn.Etag {
			changed = append(changed, s.itemToOut(item))
		}
	}

	return &FetchUpdatesResponse{Data: changed}, nil
}

func (s *ItemService) processItemUpdate(ctx context.Context, collectionID uint, itemIn *ItemBatchIn) error {
	// Get or create item
	item, err := s.itemRepo.GetByUID(ctx, collectionID, itemIn.UID)
	if err != nil {
		return err
	}

	if item == nil {
		// Create new item
		item = &model.CollectionItem{
			UID:          itemIn.UID,
			CollectionID: collectionID,
			Version:      itemIn.Version,
		}
		if err := s.itemRepo.Create(ctx, item); err != nil {
			return err
		}
	} else {
		// Update version
		item.Version = itemIn.Version
		if err := s.itemRepo.Update(ctx, item); err != nil {
			return err
		}
	}

	// Create new revision
	if s.revisionRepo != nil {
		revision := &model.CollectionItemRevision{
			UID:     itemIn.Content.UID,
			ItemID:  item.ID,
			Meta:    itemIn.Content.Meta,
			Deleted: itemIn.Content.Deleted,
		}
		if err := s.revisionRepo.Create(ctx, revision); err != nil {
			return err
		}
	}

	return nil
}

func (s *ItemService) getItemEtag(item *model.CollectionItem) string {
	if len(item.Revisions) > 0 {
		for _, rev := range item.Revisions {
			if rev.Current != nil && *rev.Current {
				return rev.UID
			}
		}
	}
	return ""
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

