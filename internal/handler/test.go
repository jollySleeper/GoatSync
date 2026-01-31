package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TestHandler handles test/debug endpoints
type TestHandler struct {
	Base
	db    *gorm.DB
	debug bool
}

// NewTestHandler creates a new test handler
func NewTestHandler(db *gorm.DB, debug bool) *TestHandler {
	return &TestHandler{
		db:    db,
		debug: debug,
	}
}

// ResetResponse is the response for reset endpoint
type ResetResponse struct {
	Status string `json:"status"`
}

// Reset handles POST /api/v1/test/authentication/reset/
// Only available in DEBUG mode - clears all test data
func (h *TestHandler) Reset(c *gin.Context) {
	if !h.debug {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not available"})
		return
	}

	// Delete all data in reverse foreign key order
	tables := []string{
		"django_collectioninvitation",
		"django_collectionmemberremoved",
		"django_collectionmember",
		"django_revisionchunkrelation",
		"django_collectionitemchunk",
		"django_collectionitemrevision",
		"django_collectionitem",
		"django_collection",
		"django_collectiontype",
		"token_auth_authtoken",
		"django_userinfo",
		"myauth_user",
		"django_stoken",
	}

	for _, table := range tables {
		// Ignore errors - table might not exist
		_ = h.db.Exec("DELETE FROM " + table).Error
	}

	c.JSON(http.StatusOK, ResetResponse{Status: "reset"})
}

