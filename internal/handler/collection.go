package handler

import (
	"net/http"
	"strconv"

	"goatsync/internal/model"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// CollectionHandler handles collection endpoints
type CollectionHandler struct {
	Base
	collectionService *service.CollectionService
}

// NewCollectionHandler creates a new collection handler
func NewCollectionHandler(collectionService *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// List handles GET /api/v1/collection/
func (h *CollectionHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	// Parse query params
	stoken := c.Query("stoken")
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.collectionService.ListCollections(c.Request.Context(), user.ID, stoken, limit)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Get handles GET /api/v1/collection/:collection_uid/
func (h *CollectionHandler) Get(c *gin.Context) {
	collectionUID := c.Param("collection_uid")
	if collectionUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing collection UID"))
		return
	}

	col, err := h.collectionService.GetCollection(c.Request.Context(), collectionUID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	if col == nil {
		h.HandleError(c, pkgerrors.ErrNotMember)
		return
	}

	// TODO: Convert to proper response format
	h.RespondMsgpack(c, http.StatusOK, col)
}

// ListMulti handles POST /api/v1/collection/list_multi/
func (h *CollectionHandler) ListMulti(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	var req service.ListMultiRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		return
	}

	// Parse query params
	stoken := c.Query("stoken")
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.collectionService.ListMultiCollections(c.Request.Context(), user.ID, req.CollectionTypes, stoken, limit)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Create handles POST /api/v1/collection/
func (h *CollectionHandler) Create(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	var req service.CollectionCreateRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		return
	}

	resp, err := h.collectionService.CreateCollection(c.Request.Context(), user.ID, &req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusCreated, resp)
}

