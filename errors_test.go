package sdk

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		status:    500,
		Code:      "INTERNAL_ERROR",
		Message:   "Something went wrong",
		RequestID: "req_123",
	}
	expected := "[500] INTERNAL_ERROR: Something went wrong (request_id: req_123)"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestAPIError_StatusCode(t *testing.T) {
	tests := []struct {
		name           string
		statusCodeVal  int
		expectedStatus int
	}{
		{"401 Unauthorized", 401, 401},
		{"404 Not Found", 404, 404},
		{"429 Rate Limit", 429, 429},
		{"500 Internal Server Error", 500, 500},
		{"502 Bad Gateway", 502, 502},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{status: tt.statusCodeVal}
			if got := err.StatusCode(); got != tt.expectedStatus {
				t.Errorf("StatusCode() = %d, want %d", got, tt.expectedStatus)
			}
		})
	}
}

func TestAPIError_IsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"500 Internal Server Error", 500, true},
		{"501 Not Implemented", 501, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"429 Rate Limit", 429, true},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"403 Forbidden", 403, false},
		{"404 Not Found", 404, false},
		{"409 Conflict", 409, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{status: tt.statusCode}
			if got := err.IsRetryable(); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         *APIError
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{"ErrUnauthorized", ErrUnauthorized, 401, "UNAUTHORIZED", "Invalid API key or signature"},
		{"ErrNotFound", ErrNotFound, 404, "NOT_FOUND", "Resource not found"},
		{"ErrRateLimit", ErrRateLimit, 429, "RATE_LIMIT", "Too many requests"},
		{"ErrInsufficientBalance", ErrInsufficientBalance, 400, "INSUFFICIENT_BALANCE", "Insufficient balance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode() != tt.wantStatus {
				t.Errorf("StatusCode() = %d, want %d", tt.err.StatusCode(), tt.wantStatus)
			}
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.wantCode)
			}
			if tt.err.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", tt.err.Message, tt.wantMessage)
			}
		})
	}
}

func TestErrRateLimit_IsRetryable(t *testing.T) {
	if !ErrRateLimit.IsRetryable() {
		t.Error("ErrRateLimit.IsRetryable() = false, want true")
	}
}

func TestErrUnauthorized_IsNotRetryable(t *testing.T) {
	if ErrUnauthorized.IsRetryable() {
		t.Error("ErrUnauthorized.IsRetryable() = true, want false")
	}
}

func TestErrorsIs(t *testing.T) {
	err := &APIError{
		status:  404,
		Code:    "NOT_FOUND",
		Message: "Not found",
	}

	// errors.Is should work with exact same pointer
	if !errors.Is(err, err) {
		t.Error("errors.Is(err, err) = false, want true")
	}

	// Different error object should not match
	otherErr := &APIError{
		status:  404,
		Code:    "NOT_FOUND",
		Message: "Not found",
	}
	if errors.Is(err, otherErr) {
		t.Error("errors.Is(err, otherErr) = true, want false")
	}
}

func TestAPIError_Implements_Error(t *testing.T) {
	var _ error = (*APIError)(nil)
}
