package middleware

import (
	"net/http"

	"goatsync/internal/codec"
	"goatsync/internal/service"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// respondError sends an error response in MessagePack format
func respondError(c *gin.Context, err *pkgerrors.EtebaseError) {
	packed, _ := codec.Marshal(err)
	c.AbortWithStatusJSON(err.StatusCode, packed)
}

// RequireAuth returns a middleware that validates the auth token
// and sets the user in the context
func RequireAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenKey := c.GetHeader("Authorization")
		if tokenKey == "" {
			respondError(c, pkgerrors.ErrInvalidToken)
			return
		}

		user, err := authService.GetUserByToken(c.Request.Context(), tokenKey)
		if err != nil {
			if eteErr, ok := err.(*pkgerrors.EtebaseError); ok {
				respondError(c, eteErr)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		if user == nil {
			respondError(c, pkgerrors.ErrInvalidToken)
			return
		}

		// Set user and token in context for use by handlers
		c.Set("user", user)
		c.Set("token", tokenKey)

		c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from the gin context
func GetUserFromContext(c *gin.Context) interface{} {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user
}

// GetTokenFromContext retrieves the auth token from the gin context
func GetTokenFromContext(c *gin.Context) string {
	token, exists := c.Get("token")
	if !exists {
		return ""
	}
	return token.(string)
}

