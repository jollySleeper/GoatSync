package service

import (
	"context"

	"goatsync/internal/config"
	"goatsync/internal/model"
	"goatsync/internal/repository"
)

// CollectionService handles collection business logic
type CollectionService struct {
	collectionRepo repository.CollectionRepository
	cfg            *config.Config
}

// NewCollectionService creates a new collection service
func NewCollectionService(
	collectionRepo repository.CollectionRepository,
	cfg *config.Config,
) *CollectionService {
	return &CollectionService{
		collectionRepo: collectionRepo,
		cfg:            cfg,
	}
}

// CollectionListResponse is the response for listing collections
type CollectionListResponse struct {
	Data               []CollectionOut `msgpack:"data"`
	Stoken             *string         `msgpack:"stoken,omitempty"`
	Done               bool            `msgpack:"done"`
	RemovedMemberships []RemovedOut    `msgpack:"removedMemberships,omitempty"`
}

// CollectionOut represents a collection in API responses
type CollectionOut struct {
	Item        ItemOut `msgpack:"item"`
	AccessLevel string  `msgpack:"accessLevel"`
	Stoken      string  `msgpack:"stoken"`
}

// RemovedOut represents a removed membership
type RemovedOut struct {
	UID string `msgpack:"uid"`
}

// ItemOut represents an item in API responses
type ItemOut struct {
	UID     string      `msgpack:"uid"`
	Version uint16      `msgpack:"version"`
	Etag    string      `msgpack:"etag"`
	Content ContentOut  `msgpack:"content"`
}

// ContentOut represents item content in API responses
type ContentOut struct {
	UID     string     `msgpack:"uid"`
	Meta    []byte     `msgpack:"meta"`
	Deleted bool       `msgpack:"deleted"`
	Chunks  []ChunkOut `msgpack:"chunks,omitempty"`
}

// ChunkOut represents a chunk reference in API responses
type ChunkOut struct {
	UID string `msgpack:"uid"`
}

// ListCollections lists collections for a user
func (s *CollectionService) ListCollections(
	ctx context.Context,
	userID uint,
	stoken string,
	limit int,
) (*CollectionListResponse, error) {
	collections, newStoken, done, err := s.collectionRepo.ListForUser(ctx, userID, stoken, limit)
	if err != nil {
		return nil, err
	}

	// Convert to API response format
	data := make([]CollectionOut, len(collections))
	for i, col := range collections {
		data[i] = s.collectionToOut(&col)
	}

	var stokenStr *string
	if newStoken != nil {
		stokenStr = &newStoken.UID
	}

	return &CollectionListResponse{
		Data:   data,
		Stoken: stokenStr,
		Done:   done,
	}, nil
}

// GetCollection retrieves a single collection by UID
func (s *CollectionService) GetCollection(
	ctx context.Context,
	uid string,
) (*model.Collection, error) {
	return s.collectionRepo.GetByUID(ctx, uid)
}

// CollectionCreateRequest is the request for creating a collection
type CollectionCreateRequest struct {
	Collection CollectionIn `msgpack:"collection"`
	Item       ItemIn       `msgpack:"item"`
}

// CollectionIn represents collection data in requests
type CollectionIn struct {
	UID            string `msgpack:"uid"`
	CollectionType string `msgpack:"collectionType,omitempty"`
}

// ItemIn represents item data in requests
type ItemIn struct {
	UID     string          `msgpack:"uid"`
	Version uint16          `msgpack:"version"`
	Etag    *string         `msgpack:"etag,omitempty"`
	Content CollectionContent `msgpack:"content"`
}

// CollectionContent represents collection item content
type CollectionContent struct {
	UID     string   `msgpack:"uid"`
	Meta    []byte   `msgpack:"meta"`
	Chunks  []string `msgpack:"chunks,omitempty"`
}

// ListMultiRequest is the request for listing collections by types
type ListMultiRequest struct {
	CollectionTypes [][]byte `msgpack:"collectionTypes"`
}

// ListMultiCollections lists collections filtered by collection types
func (s *CollectionService) ListMultiCollections(
	ctx context.Context,
	userID uint,
	typeUIDs [][]byte,
	stoken string,
	limit int,
) (*CollectionListResponse, error) {
	collections, newStoken, done, err := s.collectionRepo.ListByTypes(ctx, userID, typeUIDs, stoken, limit)
	if err != nil {
		return nil, err
	}

	data := make([]CollectionOut, len(collections))
	for i, col := range collections {
		data[i] = s.collectionToOut(&col)
	}

	var stokenStr *string
	if newStoken != nil {
		stokenStr = &newStoken.UID
	}

	return &CollectionListResponse{
		Data:   data,
		Stoken: stokenStr,
		Done:   done,
	}, nil
}

// CreateCollection creates a new collection
func (s *CollectionService) CreateCollection(
	ctx context.Context,
	userID uint,
	req *CollectionCreateRequest,
) (*CollectionOut, error) {
	// Create the main item for the collection
	mainItem := &model.CollectionItem{
		UID:     req.Item.UID,
		Version: req.Item.Version,
	}

	// Create the collection
	col := &model.Collection{
		UID:      req.Collection.UID,
		OwnerID:  userID,
		MainItem: mainItem,
	}

	if err := s.collectionRepo.Create(ctx, col); err != nil {
		return nil, err
	}

	// Create membership for the owner (admin access)
	// Note: This would be handled by the repository or a separate call

	out := s.collectionToOut(col)
	return &out, nil
}

// collectionToOut converts a model.Collection to CollectionOut
func (s *CollectionService) collectionToOut(col *model.Collection) CollectionOut {
	out := CollectionOut{
		AccessLevel: "readWrite", // TODO: Get from membership
		Stoken:      "",          // TODO: Calculate from annotations
	}

	if col.MainItem != nil {
		out.Item = ItemOut{
			UID:     col.MainItem.UID,
			Version: col.MainItem.Version,
		}
	}

	return out
}

