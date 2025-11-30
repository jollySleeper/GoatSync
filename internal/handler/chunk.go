package handler

import (
	"net/http"

	"goatsync/internal/model"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ChunkHandler handles chunk endpoints
type ChunkHandler struct {
	Base
	chunkService *service.ChunkService
}

// NewChunkHandler creates a new chunk handler
func NewChunkHandler(chunkService *service.ChunkService) *ChunkHandler {
	return &ChunkHandler{
		chunkService: chunkService,
	}
}

// Upload handles PUT /api/v1/collection/:collection_uid/item/:item_uid/chunk/:chunk_uid/
func (h *ChunkHandler) Upload(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")
	itemUID := c.Param("item_uid")
	chunkUID := c.Param("chunk_uid")

	if chunkUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing chunk UID"))
		return
	}

	err := h.chunkService.UploadChunk(
		c.Request.Context(),
		collectionUID, itemUID, chunkUID,
		user.ID,
		c.Request.Body,
	)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusCreated)
}

// Download handles GET /api/v1/collection/:collection_uid/item/:item_uid/chunk/:chunk_uid/download/
func (h *ChunkHandler) Download(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")
	itemUID := c.Param("item_uid")
	chunkUID := c.Param("chunk_uid")

	if chunkUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing chunk UID"))
		return
	}

	data, err := h.chunkService.DownloadChunk(
		c.Request.Context(),
		collectionUID, itemUID, chunkUID,
		user.ID,
	)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", data)
}

