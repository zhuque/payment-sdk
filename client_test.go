package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("tenant_123", "ak_test", "sk_test")
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "http://localhost:8080")
	}
	if client.tenantID != "tenant_123" {
		t.Errorf("tenantID = %q, want %q", client.tenantID, "tenant_123")
	}
	if client.apiKey != "ak_test" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "ak_test")
	}
	if client.apiSecret != "sk_test" {
		t.Errorf("apiSecret = %q, want %q", client.apiSecret, "sk_test")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, 30*time.Second)
	}
}

func TestWithBaseURL(t *testing.T) {
	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL("https://api.example.com"))
	if client.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://api.example.com")
	}
}

func TestWithTimeout(t *testing.T) {
	client := NewClient("tenant_123", "ak_test", "sk_test", WithTimeout(10*time.Second))
	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, 10*time.Second)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 15 * time.Second}
	client := NewClient("tenant_123", "ak_test", "sk_test", WithHTTPClient(customClient))
	if client.httpClient != customClient {
		t.Error("httpClient not set to custom client")
	}
	if client.httpClient.Timeout != 15*time.Second {
		t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, 15*time.Second)
	}
}

func TestCall_SetsAuthHeaders(t *testing.T) {
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	err := client.Call(context.Background(), "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("Call() failed: %v", err)
	}

	if receivedHeaders.Get("X-API-Key") != "ak_test" {
		t.Errorf("X-API-Key = %q, want %q", receivedHeaders.Get("X-API-Key"), "ak_test")
	}
	if receivedHeaders.Get("X-Timestamp") == "" {
		t.Error("X-Timestamp header not set")
	}
	if receivedHeaders.Get("X-Signature") == "" {
		t.Error("X-Signature header not set")
	}
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", receivedHeaders.Get("Content-Type"), "application/json")
	}
}

func TestCall_SignatureIsValid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract signature components
		method := r.Method
		path := r.URL.Path
		timestamp := r.Header.Get("X-Timestamp")
		signature := r.Header.Get("X-Signature")

		// Read body
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)

		// Recompute signature
		expectedSig := Sign(method, path, timestamp, body, "sk_test")
		if signature != expectedSig {
			t.Errorf("signature mismatch: got %q, want %q", signature, expectedSig)
			w.WriteHeader(401)
			w.Write([]byte(`{"error":"Invalid signature","code":"UNAUTHORIZED"}`))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	reqBody := map[string]string{"test": "value"}
	err := client.Call(context.Background(), "POST", "/test", reqBody, nil)
	if err != nil {
		t.Fatalf("Call() failed: %v", err)
	}
}

func TestCall_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second) // Slow response
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Call(ctx, "GET", "/test", nil, nil)
	if err == nil {
		t.Error("Call() should fail with context timeout")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("error should mention context deadline, got: %v", err)
	}
}

func TestCall_HTTPError_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"Not found","code":"NOT_FOUND"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	err := client.Call(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("Call() should return error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}

	if apiErr.StatusCode() != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode())
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "NOT_FOUND")
	}
	if apiErr.Message != "Not found" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Not found")
	}
}

func TestCall_HTTPError_500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "req_123")
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Internal error","code":"INTERNAL_ERROR"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	err := client.Call(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("Call() should return error for 500")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}

	if apiErr.StatusCode() != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode())
	}
	if apiErr.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "INTERNAL_ERROR")
	}
	if apiErr.RequestID != "req_123" {
		t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req_123")
	}
}

func TestCall_HTTPError_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	err := client.Call(context.Background(), "GET", "/test", nil, nil)

	if err == nil {
		t.Fatal("Call() should return error for 400")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}

	if apiErr.Code != "HTTP_ERROR" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "HTTP_ERROR")
	}
	if !strings.Contains(apiErr.Message, "400") {
		t.Errorf("Message should contain status code 400, got: %q", apiErr.Message)
	}
}

func TestCall_JSONMarshalling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		if req["test"] != "value" {
			t.Errorf("request body test = %q, want %q", req["test"], "value")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	reqBody := map[string]string{"test": "value"}
	var respBody map[string]string
	err := client.Call(context.Background(), "POST", "/test", reqBody, &respBody)

	if err != nil {
		t.Fatalf("Call() failed: %v", err)
	}

	if respBody["result"] != "success" {
		t.Errorf("response result = %q, want %q", respBody["result"], "success")
	}
}

func TestCall_NilRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		if len(bodyBytes) != 0 {
			t.Errorf("expected empty body, got %d bytes", len(bodyBytes))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	err := client.Call(context.Background(), "GET", "/test", nil, nil)

	if err != nil {
		t.Fatalf("Call() failed: %v", err)
	}
}

func TestCall_NilResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204) // No Content
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	err := client.Call(context.Background(), "DELETE", "/test", nil, nil)

	if err != nil {
		t.Fatalf("Call() failed: %v", err)
	}
}
