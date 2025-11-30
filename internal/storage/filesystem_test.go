package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorage_ChunkPath(t *testing.T) {
	fs := NewFileStorage("/data/chunks")

	tests := []struct {
		userID        uint
		collectionUID string
		chunkUID      string
		expected      string
	}{
		{1, "col123", "abcdef123456", "/data/chunks/user_1/col123/ab/cdef123456"},
		{2, "col456", "xy", "/data/chunks/user_2/col456/xy"},
		{3, "col789", "a", "/data/chunks/user_3/col789/a"},
	}

	for _, tt := range tests {
		t.Run(tt.chunkUID, func(t *testing.T) {
			got := fs.ChunkPath(tt.userID, tt.collectionUID, tt.chunkUID)
			if got != tt.expected {
				t.Errorf("ChunkPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFileStorage_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "goatsync_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := NewFileStorage(tempDir)

	// Test data
	userID := uint(1)
	collectionUID := "test-collection"
	chunkUID := "test-chunk-12345"
	data := []byte("Hello, this is test chunk data!")

	// Save chunk
	err = fs.SaveChunk(userID, collectionUID, chunkUID, data)
	if err != nil {
		t.Fatalf("SaveChunk failed: %v", err)
	}

	// Verify file exists
	if !fs.ChunkExists(userID, collectionUID, chunkUID) {
		t.Error("ChunkExists returned false after save")
	}

	// Load chunk
	loaded, err := fs.LoadChunk(userID, collectionUID, chunkUID)
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	if string(loaded) != string(data) {
		t.Errorf("Loaded data doesn't match: got %s, want %s", loaded, data)
	}

	// Delete chunk
	err = fs.DeleteChunk(userID, collectionUID, chunkUID)
	if err != nil {
		t.Fatalf("DeleteChunk failed: %v", err)
	}

	// Verify file is deleted
	if fs.ChunkExists(userID, collectionUID, chunkUID) {
		t.Error("ChunkExists returned true after delete")
	}
}

func TestFileStorage_LoadNonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goatsync_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := NewFileStorage(tempDir)

	// Try to load non-existent chunk
	data, err := fs.LoadChunk(1, "col", "nonexistent")
	if err != nil {
		t.Errorf("LoadChunk returned error for non-existent file: %v", err)
	}
	if data != nil {
		t.Error("LoadChunk should return nil for non-existent file")
	}
}

func TestFileStorage_DirectoryCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goatsync_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fs := NewFileStorage(tempDir)

	// Save chunk in nested directory structure
	err = fs.SaveChunk(1, "collection", "chunk123456", []byte("data"))
	if err != nil {
		t.Fatalf("SaveChunk failed: %v", err)
	}

	// Verify directory structure was created
	expectedDir := filepath.Join(tempDir, "user_1", "collection", "ch")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("Expected directory %s was not created", expectedDir)
	}
}

