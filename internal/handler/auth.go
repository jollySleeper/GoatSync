package handler

import (
	"net/http"

	"goatsync/internal/model"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	Base
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// IsEtebase handles GET /api/v1/authentication/is_etebase/
// Returns 200 OK to indicate this is an Etebase server
func (h *AuthHandler) IsEtebase(c *gin.Context) {
	c.Status(http.StatusOK)
}

// LoginChallenge handles POST /api/v1/authentication/login_challenge/
func (h *AuthHandler) LoginChallenge(c *gin.Context) {
	var req service.LoginChallengeRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		h.HandleError(c, pkgerrors.ErrInvalidRequest)
		return
	}

	resp, err := h.authService.LoginChallenge(c.Request.Context(), req.Username)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Login handles POST /api/v1/authentication/login/
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		h.HandleError(c, pkgerrors.ErrInvalidRequest)
		return
	}

	host := h.GetHost(c)
	resp, err := h.authService.Login(c.Request.Context(), &req, host)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusOK, resp)
}

// Logout handles POST /api/v1/authentication/logout/
// Requires authentication
func (h *AuthHandler) Logout(c *gin.Context) {
	tokenKey := h.GetAuthToken(c)
	if tokenKey == "" {
		h.HandleError(c, pkgerrors.ErrInvalidToken)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), tokenKey); err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

// Signup handles POST /api/v1/authentication/signup/
func (h *AuthHandler) Signup(c *gin.Context) {
	var req service.SignupRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		h.HandleError(c, pkgerrors.ErrInvalidRequest)
		return
	}

	resp, err := h.authService.Signup(c.Request.Context(), &req)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondMsgpack(c, http.StatusCreated, resp)
}

// ChangePassword handles POST /api/v1/authentication/change_password/
// Requires authentication
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userVal, exists := c.Get("user")
	if !exists {
		h.HandleError(c, pkgerrors.ErrInvalidToken)
		return
	}
	user := userVal.(*model.User)

	var req service.ChangePasswordRequest
	if err := h.ParseMsgpack(c, &req); err != nil {
		h.HandleError(c, pkgerrors.ErrInvalidRequest)
		return
	}

	host := h.GetHost(c)
	if err := h.authService.ChangePassword(c.Request.Context(), user, &req, host); err != nil {
		h.HandleError(c, err)
		return
	}

	h.RespondEmpty(c, http.StatusNoContent)
}

// DashboardURL handles POST /api/v1/authentication/dashboard_url/
// Requires authentication
func (h *AuthHandler) DashboardURL(c *gin.Context) {
	// Dashboard URL is not supported in the default configuration
	h.HandleError(c, pkgerrors.ErrDashboardNotConfigured)
}

