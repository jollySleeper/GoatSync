package errors

import (
	"net/http"
	"strings"
	"testing"
)

func TestEtebaseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *EtebaseError
		expected string
	}{
		{
			name:     "basic error",
			err:      ErrUserNotFound,
			expected: "user_not_found: User not found",
		},
		{
			name: "error with field",
			err: &EtebaseError{
				Code:   "validation_error",
				Detail: "Invalid value",
				Field:  "username",
			},
			expected: "validation_error: Invalid value (field: username)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEtebaseError_WithField(t *testing.T) {
	original := ErrMissingField
	withField := original.WithField("email")

	if withField.Field != "email" {
		t.Errorf("Expected field 'email', got '%s'", withField.Field)
	}

	// Original should not be modified
	if original.Field != "" {
		t.Error("Original error was modified")
	}

	// Code should be preserved
	if withField.Code != original.Code {
		t.Errorf("Code not preserved: %s != %s", withField.Code, original.Code)
	}
}

func TestEtebaseError_WithDetail(t *testing.T) {
	original := ErrInvalidRequest
	withDetail := original.WithDetail("Custom error message")

	if withDetail.Detail != "Custom error message" {
		t.Errorf("Expected detail 'Custom error message', got '%s'", withDetail.Detail)
	}

	// Original should not be modified
	if original.Detail == "Custom error message" {
		t.Error("Original error was modified")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("password", "must be at least 8 characters")

	if err.Code != "validation_error" {
		t.Errorf("Expected code 'validation_error', got '%s'", err.Code)
	}
	if err.Field != "password" {
		t.Errorf("Expected field 'password', got '%s'", err.Field)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

func TestNewWrongEtagError(t *testing.T) {
	err := NewWrongEtagError("expected123", "got456")

	if err.Code != "wrong_etag" {
		t.Errorf("Expected code 'wrong_etag', got '%s'", err.Code)
	}
	if err.StatusCode != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, err.StatusCode)
	}
	if !strings.Contains(err.Detail, "expected123") || !strings.Contains(err.Detail, "got456") {
		t.Errorf("Detail should contain both etags: %s", err.Detail)
	}
}

func TestNewWrongHostError(t *testing.T) {
	err := NewWrongHostError("example.com", "evil.com")

	if err.Code != "wrong_host" {
		t.Errorf("Expected code 'wrong_host', got '%s'", err.Code)
	}
	if !strings.Contains(err.Detail, "example.com") {
		t.Errorf("Detail should contain expected host: %s", err.Detail)
	}
}

func TestNewWrongActionError(t *testing.T) {
	err := NewWrongActionError("login")

	if err.Code != "wrong_action" {
		t.Errorf("Expected code 'wrong_action', got '%s'", err.Code)
	}
	if !strings.Contains(err.Detail, "login") {
		t.Errorf("Detail should contain expected action: %s", err.Detail)
	}
}

func TestErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      *EtebaseError
		expected int
	}{
		{"ErrUserNotFound", ErrUserNotFound, http.StatusUnauthorized},
		{"ErrUserNotInit", ErrUserNotInit, http.StatusUnauthorized},
		{"ErrBadSignature", ErrBadSignature, http.StatusUnauthorized},
		{"ErrUserExists", ErrUserExists, http.StatusConflict},
		{"ErrBadStoken", ErrBadStoken, http.StatusBadRequest},
		{"ErrWrongEtag", ErrWrongEtag, http.StatusConflict},
		{"ErrAdminRequired", ErrAdminRequired, http.StatusForbidden},
		{"ErrNotMember", ErrNotMember, http.StatusForbidden},
		{"ErrNotSupported", ErrNotSupported, http.StatusNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode != tt.expected {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expected, tt.err.StatusCode)
			}
		})
	}
}

