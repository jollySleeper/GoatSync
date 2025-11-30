package handler

import (
	"net/http"

	"goatsync/internal/model"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// InvitationHandler handles invitation endpoints
type InvitationHandler struct {
	Base
	invitationService *service.InvitationService
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(invitationService *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

// ListIncoming handles GET /api/v1/invitation/incoming/
func (h *InvitationHandler) ListIncoming(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	resp, err := h.invitationService.ListIncoming(c.Request.Context(), user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// GetIncoming handles GET /api/v1/invitation/incoming/:invitation_uid/
func (h *InvitationHandler) GetIncoming(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	invitationUID := c.Param("invitation_uid")

	if invitationUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing invitation UID"))
		return
	}

	resp, err := h.invitationService.GetIncoming(c.Request.Context(), invitationUID, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// RejectIncoming handles DELETE /api/v1/invitation/incoming/:invitation_uid/
func (h *InvitationHandler) RejectIncoming(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	invitationUID := c.Param("invitation_uid")

	if invitationUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing invitation UID"))
		return
	}

	err := h.invitationService.RejectInvitation(c.Request.Context(), invitationUID, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

// AcceptRequest is the request body for accepting an invitation
type AcceptRequest struct {
	CollectionType []byte `msgpack:"collectionType"`
	EncryptionKey  []byte `msgpack:"encryptionKey"`
}

// AcceptIncoming handles POST /api/v1/invitation/incoming/:invitation_uid/accept/
func (h *InvitationHandler) AcceptIncoming(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	invitationUID := c.Param("invitation_uid")

	if invitationUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing invitation UID"))
		return
	}

	var req AcceptRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		h.HandleError(c, pkgerrors.ErrInvalidRequest)
		return
	}

	err := h.invitationService.AcceptInvitation(c.Request.Context(), invitationUID, user.ID, req.EncryptionKey)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

// ListOutgoing handles GET /api/v1/invitation/outgoing/
func (h *InvitationHandler) ListOutgoing(c *gin.Context) {
	user := c.MustGet("user").(*model.User)

	resp, err := h.invitationService.ListOutgoing(c.Request.Context(), user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// DeleteOutgoing handles DELETE /api/v1/invitation/outgoing/:invitation_uid/
func (h *InvitationHandler) DeleteOutgoing(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	invitationUID := c.Param("invitation_uid")

	if invitationUID == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing invitation UID"))
		return
	}

	err := h.invitationService.DeleteOutgoing(c.Request.Context(), invitationUID, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

// FetchUserForInviteRequest is the request body for fetching user info for invite
type FetchUserForInviteRequest struct {
	Username string `msgpack:"username"`
}

// FetchUserForInvite handles POST /api/v1/invitation/outgoing/fetch_user_profile/
func (h *InvitationHandler) FetchUserForInvite(c *gin.Context) {
	var req FetchUserForInviteRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		return
	}

	resp, err := h.invitationService.FetchUserForInvite(c.Request.Context(), req.Username)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

