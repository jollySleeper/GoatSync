package handler

import (
	"net/http"

	"goatsync/internal/model"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// MemberHandler handles member endpoints
type MemberHandler struct {
	Base
	memberService *service.MemberService
}

// NewMemberHandler creates a new member handler
func NewMemberHandler(memberService *service.MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

// List handles GET /api/v1/collection/:collection_uid/member/
func (h *MemberHandler) List(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	resp, err := h.memberService.ListMembers(c.Request.Context(), collectionUID, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Remove handles DELETE /api/v1/collection/:collection_uid/member/:username/
func (h *MemberHandler) Remove(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")
	username := c.Param("username")

	if username == "" {
		h.HandleError(c, pkgerrors.ErrInvalidRequest.WithDetail("missing username"))
		return
	}

	err := h.memberService.RemoveMember(c.Request.Context(), collectionUID, username, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

// Leave handles POST /api/v1/collection/:collection_uid/member/leave/
func (h *MemberHandler) Leave(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	collectionUID := c.Param("collection_uid")

	err := h.memberService.LeaveCollection(c.Request.Context(), collectionUID, user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

