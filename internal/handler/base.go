// Package handler provides HTTP handlers for the GoatSync API.
// Handlers are thin - they parse requests, call services, and format responses.
package handler

import (
	"io"
	"net/http"

	"goatsync/internal/codec"
	pkgerrors "goatsync/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Base provides shared utilities for all handlers
type Base struct{}

// ParseMsgpack parses a MessagePack request body into the given struct
func (h *Base) ParseMsgpack(c *gin.Context, v interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return codec.Unmarshal(body, v)
}

// RespondMsgpack sends a MessagePack response
func (h *Base) RespondMsgpack(c *gin.Context, status int, data interface{}) {
	packed, err := codec.Marshal(data)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(status, "application/msgpack", packed)
}

// RespondEmpty sends an empty response with the given status
func (h *Base) RespondEmpty(c *gin.Context, status int) {
	c.Status(status)
}

// HandleError handles an error and sends the appropriate response
func (h *Base) HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's an Etebase error
	if eteErr, ok := err.(*pkgerrors.EtebaseError); ok {
		h.RespondMsgpack(c, eteErr.StatusCode, eteErr)
		return
	}

	// Generic error - return 500
	c.AbortWithStatus(http.StatusInternalServerError)
}

// GetAuthToken extracts the auth token from the Authorization header
func (h *Base) GetAuthToken(c *gin.Context) string {
	return c.GetHeader("Authorization")
}

// GetHost returns the Host header from the request
func (h *Base) GetHost(c *gin.Context) string {
	return c.Request.Host
}

