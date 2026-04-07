package sdk

import "fmt"

// APIError represents an error response from the API.
type APIError struct {
	status    int
	Code      string
	Message   string
	RequestID string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s: %s (request_id: %s)", e.status, e.Code, e.Message, e.RequestID)
}

// StatusCode returns the HTTP status code.
func (e *APIError) StatusCode() int {
	return e.status
}

// IsRetryable returns true if the error represents a retryable condition.
// 5xx errors and 429 (rate limit) are retryable.
// 4xx errors (except 429) are not retryable.
func (e *APIError) IsRetryable() bool {
	// 5xx: server errors are retryable
	if e.status >= 500 && e.status < 600 {
		return true
	}
	// 429: rate limit is retryable (client should retry after backoff)
	if e.status == 429 {
		return true
	}
	// 4xx: client errors are NOT retryable (except 429)
	return false
}

// Predefined API errors
var (
	ErrUnauthorized = &APIError{
		status:  401,
		Code:    "UNAUTHORIZED",
		Message: "Invalid API key or signature",
	}

	ErrNotFound = &APIError{
		status:  404,
		Code:    "NOT_FOUND",
		Message: "Resource not found",
	}

	ErrRateLimit = &APIError{
		status:  429,
		Code:    "RATE_LIMIT",
		Message: "Too many requests",
	}

	ErrInsufficientBalance = &APIError{
		status:  400,
		Code:    "INSUFFICIENT_BALANCE",
		Message: "Insufficient balance",
	}
)
