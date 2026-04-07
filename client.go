package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client is the HTTP client for the Payment Gateway API.
type Client struct {
	baseURL    string
	tenantID   string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// NewClient creates a new API client with functional options.
// Default timeout is 30 seconds (conservative for payment operations).
// Default baseURL is http://localhost:8080 for local development.
//
// Example:
//
//	client := NewClient("tenant_123", "ak_test", "sk_test",
//	    WithBaseURL("https://api.example.com"),
//	    WithTimeout(10*time.Second))
func NewClient(tenantID, apiKey, apiSecret string, opts ...Option) *Client {
	c := &Client{
		baseURL:    "http://localhost:8080",
		tenantID:   tenantID,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithBaseURL sets the base URL for the API.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient = &http.Client{Timeout: timeout}
	}
}

// Call makes an authenticated HTTP request to the API.
// It automatically adds X-API-Key, X-Timestamp, and X-Signature headers.
// Context cancellation is propagated to the HTTP request.
//
// Returns APIError for HTTP 4xx/5xx responses.
func (c *Client) Call(ctx context.Context, method, path string, reqBody, result interface{}) error {
	// Marshal request body to JSON
	var bodyBytes []byte
	var err error
	if reqBody != nil {
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Build full URL
	url := c.baseURL + path

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	// Generate timestamp and signature
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	req.Header.Set("X-Timestamp", timestamp)

	bodyStr := string(bodyBytes)
	signature := Sign(method, path, timestamp, bodyStr, c.apiSecret)
	req.Header.Set("X-Signature", signature)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return &APIError{
				status:    resp.StatusCode,
				Code:      errResp.Code,
				Message:   errResp.Error,
				RequestID: resp.Header.Get("X-Request-ID"),
			}
		}
		return &APIError{
			status:    resp.StatusCode,
			Code:      "HTTP_ERROR",
			Message:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
			RequestID: resp.Header.Get("X-Request-ID"),
		}
	}

	// Unmarshal successful response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
