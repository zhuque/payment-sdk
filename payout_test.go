package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePayout_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/payouts") {
			t.Errorf("path = %q, want /tenants/tenant_123/payouts", r.URL.Path)
		}

		var req CreatePayoutRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.MerchantOrderID != "order_123" {
			t.Errorf("MerchantOrderID = %q, want order_123", req.MerchantOrderID)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"payout_id":"payout_123","merchant_order_id":"order_123","to_address":"0x123","amount":"100.00","chain":"ethereum","token":"usdt","status":"pending","created_at":"2026-04-03T10:00:00Z","updated_at":"2026-04-03T10:00:00Z"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreatePayoutRequest{
		MerchantOrderID: "order_123",
		ToAddress:       "0x123",
		Amount:          "100.00",
		Chain:           "ethereum",
		Token:           "usdt",
	}
	resp, err := client.CreatePayout(context.Background(), req)

	if err != nil {
		t.Fatalf("CreatePayout() failed: %v", err)
	}
	if resp.PayoutID != "payout_123" {
		t.Errorf("PayoutID = %q, want payout_123", resp.PayoutID)
	}
	if resp.MerchantOrderID != "order_123" {
		t.Errorf("MerchantOrderID = %q, want order_123", resp.MerchantOrderID)
	}
	if resp.Status != "pending" {
		t.Errorf("Status = %q, want pending", resp.Status)
	}
}

func TestCreatePayout_Error400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Insufficient balance","code":"INSUFFICIENT_BALANCE"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreatePayoutRequest{
		MerchantOrderID: "order_123",
		ToAddress:       "0x123",
		Amount:          "100.00",
		Chain:           "ethereum",
		Token:           "usdt",
	}
	_, err := client.CreatePayout(context.Background(), req)

	if err == nil {
		t.Fatal("CreatePayout() should return error for 400")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "INSUFFICIENT_BALANCE" {
		t.Errorf("Code = %q, want INSUFFICIENT_BALANCE", apiErr.Code)
	}
	if apiErr.StatusCode() != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode())
	}
}

func TestCreatePayout_Error401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Invalid credentials","code":"UNAUTHORIZED"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreatePayoutRequest{
		MerchantOrderID: "order_123",
		ToAddress:       "0x123",
		Amount:          "100.00",
		Chain:           "ethereum",
		Token:           "usdt",
	}
	_, err := client.CreatePayout(context.Background(), req)

	if err == nil {
		t.Fatal("CreatePayout() should return error for 401")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "UNAUTHORIZED" {
		t.Errorf("Code = %q, want UNAUTHORIZED", apiErr.Code)
	}
	if apiErr.StatusCode() != 401 {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode())
	}
}

func TestCreatePayout_Error500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error","code":"INTERNAL_ERROR"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreatePayoutRequest{
		MerchantOrderID: "order_123",
		ToAddress:       "0x123",
		Amount:          "100.00",
		Chain:           "ethereum",
		Token:           "usdt",
	}
	_, err := client.CreatePayout(context.Background(), req)

	if err == nil {
		t.Fatal("CreatePayout() should return error for 500")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %q, want INTERNAL_ERROR", apiErr.Code)
	}
	if apiErr.StatusCode() != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode())
	}
}

func TestGetPayout_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/payouts/payout_123") {
			t.Errorf("path = %q, want /tenants/tenant_123/payouts/payout_123", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"payout_id":"payout_123","merchant_order_id":"order_123","to_address":"0x123","amount":"100.00","chain":"ethereum","token":"usdt","status":"completed","created_at":"2026-04-03T10:00:00Z","updated_at":"2026-04-03T11:00:00Z","tx_hash":"0xabc"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	resp, err := client.GetPayout(context.Background(), "payout_123")

	if err != nil {
		t.Fatalf("GetPayout() failed: %v", err)
	}
	if resp.PayoutID != "payout_123" {
		t.Errorf("PayoutID = %q, want payout_123", resp.PayoutID)
	}
	if resp.Status != "completed" {
		t.Errorf("Status = %q, want completed", resp.Status)
	}
	if resp.TxHash == nil || *resp.TxHash != "0xabc" {
		t.Errorf("TxHash = %v, want 0xabc", resp.TxHash)
	}
}

func TestGetPayout_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Payout not found","code":"NOT_FOUND"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	_, err := client.GetPayout(context.Background(), "payout_999")

	if err == nil {
		t.Fatal("GetPayout() should return error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode() != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode())
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", apiErr.Code)
	}
}

func TestListPayouts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/payouts") {
			t.Errorf("path = %q, want /tenants/tenant_123/payouts", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %q, want 10", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "20" {
			t.Errorf("offset = %q, want 20", r.URL.Query().Get("offset"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"payouts":[{"payout_id":"payout_1","merchant_order_id":"order_1","to_address":"0x1","amount":"50.00","chain":"polygon","token":"usdc","status":"completed","created_at":"2026-04-03T10:00:00Z","updated_at":"2026-04-03T10:30:00Z"}],"total":100,"limit":10,"offset":20}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	resp, err := client.ListPayouts(context.Background(), 10, 20)

	if err != nil {
		t.Fatalf("ListPayouts() failed: %v", err)
	}
	if resp.Total != 100 {
		t.Errorf("Total = %d, want 100", resp.Total)
	}
	if resp.Limit != 10 {
		t.Errorf("Limit = %d, want 10", resp.Limit)
	}
	if resp.Offset != 20 {
		t.Errorf("Offset = %d, want 20", resp.Offset)
	}
	if len(resp.Payouts) != 1 {
		t.Errorf("len(Payouts) = %d, want 1", len(resp.Payouts))
	}
	if len(resp.Payouts) > 0 && resp.Payouts[0].PayoutID != "payout_1" {
		t.Errorf("Payouts[0].PayoutID = %q, want payout_1", resp.Payouts[0].PayoutID)
	}
}

func TestListPayouts_Pagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		if limit != "50" {
			t.Errorf("limit = %q, want 50", limit)
		}
		if offset != "100" {
			t.Errorf("offset = %q, want 100", offset)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"payouts":[],"total":150,"limit":50,"offset":100}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	resp, err := client.ListPayouts(context.Background(), 50, 100)

	if err != nil {
		t.Fatalf("ListPayouts() failed: %v", err)
	}
	if resp.Limit != 50 {
		t.Errorf("Limit = %d, want 50", resp.Limit)
	}
	if resp.Offset != 100 {
		t.Errorf("Offset = %d, want 100", resp.Offset)
	}
}
