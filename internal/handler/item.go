package handler

import (
	"net/http"
	"strconv"

	"goatsync/internal/model"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ItemHandler handles item endpoints
type ItemHandler struct {
	Base
	itemService *service.ItemService
}

// NewItemHandler creates a new item handler
func NewItemHandler(itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
	}
}

// List handles GET /api/v1/collection/:collection_uid/item/
func (h *ItemHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	// Parse query params
	stoken := c.Query("stoken")
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	resp, err := h.itemService.ListItems(c.Request.Context(), collectionUID, user.ID, stoken, limit)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Get handles GET /api/v1/collection/:collection_uid/item/:item_uid/
func (h *ItemHandler) Get(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")
	itemUID := c.Param("item_uid")

	if itemUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing item UID"))
		return
	}

	resp, err := h.itemService.GetItem(c.Request.Context(), collectionUID, itemUID, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Batch handles POST /api/v1/collection/:collection_uid/item/batch/
func (h *ItemHandler) Batch(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	var req service.ItemBatchRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		return
	}

	err := h.itemService.BatchItems(c.Request.Context(), collectionUID, user.ID, &req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, gin.H{"status": "ok"})
}

// Transaction handles POST /api/v1/collection/:collection_uid/item/transaction/
func (h *ItemHandler) Transaction(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	var req service.ItemTransactionRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		return
	}

	err := h.itemService.TransactionItems(c.Request.Context(), collectionUID, user.ID, &req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, gin.H{"status": "ok"})
}

// FetchUpdates handles POST /api/v1/collection/:collection_uid/item/fetch_updates/
func (h *ItemHandler) FetchUpdates(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	var req service.FetchUpdatesRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		return
	}

	resp, err := h.itemService.FetchUpdates(c.Request.Context(), collectionUID, user.ID, &req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

