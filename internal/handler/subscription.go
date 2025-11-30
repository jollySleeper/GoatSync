package handler

import (
	"net/http"

	"goatsync/internal/model"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// SubscriptionHandler handles subscription ticket endpoints
type SubscriptionHandler struct {
	Base
	wsHandler *WebSocketHandler
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(wsHandler *WebSocketHandler) *SubscriptionHandler {
	return &SubscriptionHandler{
		wsHandler: wsHandler,
	}
}

// TicketResponse is the response for subscription ticket
type TicketResponse struct {
	Ticket string `msgpack:"ticket"`
}

// GetTicket handles POST /api/v1/collection/:collection_uid/item/subscription-ticket/
func (h *SubscriptionHandler) GetTicket(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	if collectionUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing collection UID"))
		return
	}

	// TODO: Verify user has access to collection
	// For now, just create the ticket
	ticket := h.wsHandler.CreateTicket(c.Request.Context(), user.ID, 0) // TODO: Get collection ID

	h.RespondMsgpack(c, http.StatusOK, TicketResponse{Ticket: ticket})
}

