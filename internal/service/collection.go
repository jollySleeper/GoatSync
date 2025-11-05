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

