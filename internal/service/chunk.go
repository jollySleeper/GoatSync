package service

import (
	"context"
	"io"

	"goatsync/internal/model"
	"goatsync/internal/repository"
	"goatsync/internal/storage"
	pkgerrors "goatsync/pkg/errors"
)

// ChunkService handles chunk business logic
type ChunkService struct {
	chunkRepo      repository.ChunkRepository
	collectionRepo repository.CollectionRepository
	memberRepo     repository.MemberRepository
	storage        *storage.FileStorage
}

// NewChunkService creates a new chunk service
func NewChunkService(
	chunkRepo repository.ChunkRepository,
	collectionRepo repository.CollectionRepository,
	memberRepo repository.MemberRepository,
	storage *storage.FileStorage,
) *ChunkService {
	return &ChunkService{
		chunkRepo:      chunkRepo,
		collectionRepo: collectionRepo,
		memberRepo:     memberRepo,
		storage:        storage,
	}
}

// UploadChunk uploads a new chunk
func (s *ChunkService) UploadChunk(
	ctx context.Context,
	collectionUID, itemUID, chunkUID string,
	userID uint,
	data io.Reader,
) error {
	// Get collection
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return err
	}
	if col == nil {
		return pkgerrors.ErrNotMember
	}

	// Check write access
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

	// Check if chunk already exists
	existing, err := s.chunkRepo.GetByUID(ctx, col.ID, chunkUID)
	if err != nil {
		return err
	}
	if existing != nil {
		return pkgerrors.ErrChunkExists
	}

	// Save to filesystem
	chunkPath := s.storage.ChunkPath(col.OwnerID, collectionUID, chunkUID)
	if err := s.storage.SaveChunkFromReader(col.OwnerID, collectionUID, chunkUID, data); err != nil {
		return err
	}

	// Create database record
	chunk := &model.CollectionItemChunk{
		UID:          chunkUID,
		CollectionID: col.ID,
		ChunkFile:    chunkPath,
	}
	return s.chunkRepo.Create(ctx, chunk)
}

// DownloadChunk downloads a chunk
func (s *ChunkService) DownloadChunk(
	ctx context.Context,
	collectionUID, itemUID, chunkUID string,
	userID uint,
) ([]byte, error) {
	// Get collection
	col, err := s.collectionRepo.GetByUID(ctx, collectionUID)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Check access
	member, err := s.memberRepo.GetByUserAndCollection(ctx, userID, col.ID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, pkgerrors.ErrNotMember
	}

	// Get chunk from database
	chunk, err := s.chunkRepo.GetByUID(ctx, col.ID, chunkUID)
	if err != nil {
		return nil, err
	}
	if chunk == nil {
		return nil, pkgerrors.ErrChunkNoContent
	}

	// Load from filesystem
	return s.storage.LoadChunk(col.OwnerID, collectionUID, chunkUID)
}

