// Package errors defines error types for the Etebase protocol.
// Error codes and details MUST match the Python implementation exactly
// for compatibility with existing clients.
package errors

import (
	"fmt"
	"net/http"
)

// EtebaseError represents an error in the Etebase protocol.
// The error is serialized to MessagePack and returned to the client.
type EtebaseError struct {
	Code       string          `json:"code" msgpack:"code"`
	Detail     string          `json:"detail" msgpack:"detail"`
	Field      string          `json:"field,omitempty" msgpack:"field,omitempty"`
	Errors     []*EtebaseError `json:"errors,omitempty" msgpack:"errors,omitempty"`
	StatusCode int             `json:"-" msgpack:"-"` // HTTP status code, not serialized
}

// Error implements the error interface
func (e *EtebaseError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Detail, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

// WithField returns a copy of the error with the specified field
func (e *EtebaseError) WithField(field string) *EtebaseError {
	return &EtebaseError{
		Code:       e.Code,
		Detail:     e.Detail,
		Field:      field,
		StatusCode: e.StatusCode,
	}
}

// WithDetail returns a copy of the error with a custom detail message
func (e *EtebaseError) WithDetail(detail string) *EtebaseError {
	return &EtebaseError{
		Code:       e.Code,
		Detail:     detail,
		Field:      e.Field,
		StatusCode: e.StatusCode,
	}
}

// ============================================================================
// Authentication Errors (401, 400, 409)
// ============================================================================

// ErrUserNotFound is returned when the user doesn't exist
var ErrUserNotFound = &EtebaseError{
	Code:       "user_not_found",
	Detail:     "User not found",
	StatusCode: http.StatusUnauthorized,
}

// ErrUserNotInit is returned when the user exists but hasn't completed setup
var ErrUserNotInit = &EtebaseError{
	Code:       "user_not_init",
	Detail:     "User not properly init",
	StatusCode: http.StatusUnauthorized,
}

// ErrBadSignature is returned when login signature verification fails
var ErrBadSignature = &EtebaseError{
	Code:       "login_bad_signature",
	Detail:     "Wrong password for user.",
	StatusCode: http.StatusUnauthorized,
}

// ErrWrongAction is returned when the action in the login response doesn't match
var ErrWrongAction = &EtebaseError{
	Code:       "wrong_action",
	Detail:     "Expected different action",
	StatusCode: http.StatusBadRequest,
}

// ErrChallengeExpired is returned when the login challenge has expired
var ErrChallengeExpired = &EtebaseError{
	Code:       "challenge_expired",
	Detail:     "Login challenge has expired",
	StatusCode: http.StatusBadRequest,
}

// ErrWrongUser is returned when the challenge was issued for a different user
var ErrWrongUser = &EtebaseError{
	Code:       "wrong_user",
	Detail:     "This challenge is for the wrong user",
	StatusCode: http.StatusBadRequest,
}

// ErrWrongHost is returned when the host in the login response doesn't match
var ErrWrongHost = &EtebaseError{
	Code:       "wrong_host",
	Detail:     "Found wrong host name",
	StatusCode: http.StatusBadRequest,
}

// ErrUserExists is returned when trying to create a user that already exists
var ErrUserExists = &EtebaseError{
	Code:       "user_exists",
	Detail:     "User already exists",
	StatusCode: http.StatusConflict,
}

// ErrInvalidToken is returned for invalid or expired auth tokens
var ErrInvalidToken = &EtebaseError{
	Code:       "invalid_token",
	Detail:     "Invalid token",
	StatusCode: http.StatusUnauthorized,
}

// ============================================================================
// Collection/Sync Errors (400, 409)
// ============================================================================

// ErrBadStoken is returned when the sync token is invalid
var ErrBadStoken = &EtebaseError{
	Code:       "bad_stoken",
	Detail:     "Invalid stoken.",
	StatusCode: http.StatusBadRequest,
}

// ErrStaleStoken is returned when the sync token is too old
var ErrStaleStoken = &EtebaseError{
	Code:       "stale_stoken",
	Detail:     "Stoken is too old",
	StatusCode: http.StatusConflict,
}

// ErrWrongEtag is returned when the etag doesn't match (optimistic locking failure)
var ErrWrongEtag = &EtebaseError{
	Code:       "wrong_etag",
	Detail:     "Wrong etag",
	StatusCode: http.StatusConflict,
}

// ErrUniqueUID is returned when trying to create an item with a duplicate UID
var ErrUniqueUID = &EtebaseError{
	Code:       "unique_uid",
	Detail:     "Collection with this uid already exists",
	StatusCode: http.StatusConflict,
}

// ============================================================================
// Permission Errors (403)
// ============================================================================

// ErrAdminRequired is returned when an admin operation is attempted by a non-admin
var ErrAdminRequired = &EtebaseError{
	Code:       "admin_access_required",
	Detail:     "Only collection admins can perform this operation.",
	StatusCode: http.StatusForbidden,
}

// ErrNoWriteAccess is returned when write access is required but not available
var ErrNoWriteAccess = &EtebaseError{
	Code:       "no_write_access",
	Detail:     "You need write access to write to this collection",
	StatusCode: http.StatusForbidden,
}

// ErrNotMember is returned when the user is not a member of the collection
var ErrNotMember = &EtebaseError{
	Code:       "not_member",
	Detail:     "You are not a member of this collection",
	StatusCode: http.StatusForbidden,
}

// ============================================================================
// Chunk Errors (400, 409)
// ============================================================================

// ErrChunkExists is returned when trying to upload a chunk that already exists
var ErrChunkExists = &EtebaseError{
	Code:       "chunk_exists",
	Detail:     "Chunk already exists.",
	StatusCode: http.StatusConflict,
}

// ErrChunkNoContent is returned when trying to create a chunk without content
var ErrChunkNoContent = &EtebaseError{
	Code:       "chunk_no_content",
	Detail:     "Tried to create a new chunk without content",
	StatusCode: http.StatusBadRequest,
}

// ============================================================================
// Invitation Errors (400, 409)
// ============================================================================

// ErrNoSelfInvite is returned when a user tries to invite themselves
var ErrNoSelfInvite = &EtebaseError{
	Code:       "no_self_invite",
	Detail:     "Inviting yourself is not allowed",
	StatusCode: http.StatusBadRequest,
}

// ErrInvitationExists is returned when an invitation already exists
var ErrInvitationExists = &EtebaseError{
	Code:       "invitation_exists",
	Detail:     "Invitation already exists",
	StatusCode: http.StatusConflict,
}

// ErrAlreadyMember is returned when trying to invite someone who is already a member
var ErrAlreadyMember = &EtebaseError{
	Code:       "already_member",
	Detail:     "User is already a member of this collection",
	StatusCode: http.StatusConflict,
}

// ============================================================================
// Server/Feature Errors (501)
// ============================================================================

// ErrNotSupported is returned when a feature requires configuration that isn't present
var ErrNotSupported = &EtebaseError{
	Code:       "not_supported",
	Detail:     "This feature is not supported by this server",
	StatusCode: http.StatusNotImplemented,
}

// ErrRedisRequired is returned when a feature requires Redis but it's not configured
var ErrRedisRequired = &EtebaseError{
	Code:       "not_supported",
	Detail:     "This end-point requires Redis to be configured",
	StatusCode: http.StatusNotImplemented,
}

// ErrDashboardNotConfigured is returned when dashboard URL function isn't configured
var ErrDashboardNotConfigured = &EtebaseError{
	Code:       "not_supported",
	Detail:     "This server doesn't have a user dashboard.",
	StatusCode: http.StatusNotImplemented,
}

// ============================================================================
// Validation Errors (400)
// ============================================================================

// ErrInvalidRequest is returned for general validation errors
var ErrInvalidRequest = &EtebaseError{
	Code:       "invalid_request",
	Detail:     "Invalid request",
	StatusCode: http.StatusBadRequest,
}

// ErrMissingField is returned when a required field is missing
var ErrMissingField = &EtebaseError{
	Code:       "missing_field",
	Detail:     "Required field is missing",
	StatusCode: http.StatusBadRequest,
}

// NewValidationError creates a new validation error with field information
func NewValidationError(field, message string) *EtebaseError {
	return &EtebaseError{
		Code:       "validation_error",
		Detail:     message,
		Field:      field,
		StatusCode: http.StatusBadRequest,
	}
}

// NewWrongEtagError creates an ErrWrongEtag with specific expected/got values
func NewWrongEtagError(expected, got string) *EtebaseError {
	return &EtebaseError{
		Code:       "wrong_etag",
		Detail:     fmt.Sprintf("Wrong etag. Expected %s got %s", expected, got),
		StatusCode: http.StatusConflict,
	}
}

// NewWrongHostError creates an ErrWrongHost with specific host information
func NewWrongHostError(expected, got string) *EtebaseError {
	return &EtebaseError{
		Code:       "wrong_host",
		Detail:     fmt.Sprintf("Found wrong host name. Got: \"%s\" expected: \"%s\"", got, expected),
		StatusCode: http.StatusBadRequest,
	}
}

// NewWrongActionError creates an ErrWrongAction with the expected action
func NewWrongActionError(expected string) *EtebaseError {
	return &EtebaseError{
		Code:       "wrong_action",
		Detail:     fmt.Sprintf("Expected \"%s\" but got something else", expected),
		StatusCode: http.StatusBadRequest,
	}
}

