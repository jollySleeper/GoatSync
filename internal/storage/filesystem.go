// Package storage provides file storage for chunks.
package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileStorage handles chunk file storage on the filesystem
type FileStorage struct {
	basePath string
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{basePath: basePath}
}

// ChunkPath returns the path for a chunk file
// Format: {basePath}/user_{userID}/{collectionUID}/{uidPrefix}/{uidRest}
func (s *FileStorage) ChunkPath(userID uint, collectionUID, chunkUID string) string {
	if len(chunkUID) < 2 {
		return filepath.Join(s.basePath, fmt.Sprintf("user_%d", userID), collectionUID, chunkUID)
	}
	prefix := chunkUID[:2]
	rest := chunkUID[2:]
	return filepath.Join(s.basePath, fmt.Sprintf("user_%d", userID), collectionUID, prefix, rest)
}

// SaveChunk saves chunk data to the filesystem
func (s *FileStorage) SaveChunk(userID uint, collectionUID, chunkUID string, data []byte) error {
	path := s.ChunkPath(userID, collectionUID, chunkUID)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	return nil
}

// LoadChunk loads chunk data from the filesystem
func (s *FileStorage) LoadChunk(userID uint, collectionUID, chunkUID string) ([]byte, error) {
	path := s.ChunkPath(userID, collectionUID, chunkUID)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read chunk: %w", err)
	}

	return data, nil
}

// DeleteChunk deletes a chunk file from the filesystem
func (s *FileStorage) DeleteChunk(userID uint, collectionUID, chunkUID string) error {
	path := s.ChunkPath(userID, collectionUID, chunkUID)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete chunk: %w", err)
	}

	return nil
}

// ChunkExists checks if a chunk file exists
func (s *FileStorage) ChunkExists(userID uint, collectionUID, chunkUID string) bool {
	path := s.ChunkPath(userID, collectionUID, chunkUID)
	_, err := os.Stat(path)
	return err == nil
}

// SaveChunkFromReader saves chunk data from an io.Reader
func (s *FileStorage) SaveChunkFromReader(userID uint, collectionUID, chunkUID string, reader io.Reader) error {
	path := s.ChunkPath(userID, collectionUID, chunkUID)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create chunk file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	return nil
}

