package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateAddress_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/addresses") {
			t.Errorf("path = %q, want /tenants/tenant_123/addresses", r.URL.Path)
		}

		var req GenerateAddressRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.UserID != "user_123" {
			t.Errorf("UserID = %q, want user_123", req.UserID)
		}
		if req.Chain != "ethereum" {
			t.Errorf("Chain = %q, want ethereum", req.Chain)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"address":"0xabc123","path":"m/44'/60'/0'/0/0","chain":"ethereum","token":"usdt"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	modeFixed := int16(0)
	req := &GenerateAddressRequest{
		UserID: "user_123",
		Chain:  "ethereum",
		Token:  "usdt",
		Mode:   &modeFixed,
	}
	resp, err := client.GenerateAddress(context.Background(), req)

	if err != nil {
		t.Fatalf("GenerateAddress() failed: %v", err)
	}
	if resp.Address != "0xabc123" {
		t.Errorf("Address = %q, want 0xabc123", resp.Address)
	}
	if resp.Chain != "ethereum" {
		t.Errorf("Chain = %q, want ethereum", resp.Chain)
	}
	if resp.Token != "usdt" {
		t.Errorf("Token = %q, want usdt", resp.Token)
	}
	if resp.Path != "m/44'/60'/0'/0/0" {
		t.Errorf("Path = %q, want m/44'/60'/0'/0/0", resp.Path)
	}
}

func TestGenerateAddress_Error400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid chain","code":"INVALID_CHAIN"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	modeFixed := int16(0)
	req := &GenerateAddressRequest{
		UserID: "user_123",
		Chain:  "invalid_chain",
		Token:  "usdt",
		Mode:   &modeFixed,
	}
	_, err := client.GenerateAddress(context.Background(), req)

	if err == nil {
		t.Fatal("GenerateAddress() should return error for 400")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "INVALID_CHAIN" {
		t.Errorf("Code = %q, want INVALID_CHAIN", apiErr.Code)
	}
	if apiErr.StatusCode() != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode())
	}
}

func TestGenerateAddress_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	modeFixed := int16(0)
	req := &GenerateAddressRequest{
		UserID: "user_123",
		Chain:  "ethereum",
		Token:  "usdt",
		Mode:   &modeFixed,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GenerateAddress(ctx, req)
	if err == nil {
		t.Fatal("GenerateAddress() should return error for cancelled context")
	}
}

func TestCreateOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/orders") {
			t.Errorf("path = %q, want /tenants/tenant_123/orders", r.URL.Path)
		}

		var req CreateOrderRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.OrderID != "merchant_order_123" {
			t.Errorf("OrderID = %q, want merchant_order_123", req.OrderID)
		}
		if req.Amount != "100.50" {
			t.Errorf("Amount = %q, want 100.50", req.Amount)
		}
		if req.Chain != "polygon" {
			t.Errorf("Chain = %q, want polygon", req.Chain)
		}
		if req.UserID != "user_789" {
			t.Errorf("UserID = %q, want user_789", req.UserID)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"order_id":"merchant_order_123","payment_address":"0xdef456","amount":"100.50","chain":"polygon","token":"usdc","expires_at":"2026-04-03T12:00:00Z","status":"pending"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreateOrderRequest{
		OrderID:  "merchant_order_123",
		Amount:   "100.50",
		Chain:    "polygon",
		Token:    "usdc",
		UserID:   "user_789",
		Metadata: map[string]interface{}{"ref": "inv-001"},
	}
	resp, err := client.CreateOrder(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateOrder() failed: %v", err)
	}
	if resp.OrderID != "merchant_order_123" {
		t.Errorf("OrderID = %q, want merchant_order_123", resp.OrderID)
	}
	if resp.PaymentAddress != "0xdef456" {
		t.Errorf("PaymentAddress = %q, want 0xdef456", resp.PaymentAddress)
	}
	if resp.Status != "pending" {
		t.Errorf("Status = %q, want pending", resp.Status)
	}
}

func TestCreateOrder_Error400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid amount","code":"INVALID_AMOUNT"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreateOrderRequest{
		OrderID:  "order_456",
		Amount:   "invalid",
		Chain:    "polygon",
		Token:    "usdc",
		UserID:   "user_123",
		Metadata: map[string]interface{}{},
	}
	_, err := client.CreateOrder(context.Background(), req)

	if err == nil {
		t.Fatal("CreateOrder() should return error for 400")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "INVALID_AMOUNT" {
		t.Errorf("Code = %q, want INVALID_AMOUNT", apiErr.Code)
	}
}

func TestCreateOrder_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	req := &CreateOrderRequest{
		OrderID:  "order_789",
		Amount:   "100.00",
		Chain:    "ethereum",
		Token:    "usdt",
		UserID:   "user_456",
		Metadata: map[string]interface{}{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.CreateOrder(ctx, req)
	if err == nil {
		t.Fatal("CreateOrder() should return error for cancelled context")
	}
}

func TestGetBalance_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/balances/ethereum/usdt") {
			t.Errorf("path = %q, want /tenants/tenant_123/balances/ethereum/usdt", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"chain":"ethereum","token":"usdt","balance":"1000.00","frozen_balance":"100.00","available_balance":"900.00"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	resp, err := client.GetBalance(context.Background(), "ethereum", "usdt")

	if err != nil {
		t.Fatalf("GetBalance() failed: %v", err)
	}
	if resp.Chain != "ethereum" {
		t.Errorf("Chain = %q, want ethereum", resp.Chain)
	}
	if resp.Token != "usdt" {
		t.Errorf("Token = %q, want usdt", resp.Token)
	}
	if resp.Balance.String() != "1000" {
		t.Errorf("Balance = %s, want 1000", resp.Balance.String())
	}
	if resp.FrozenBalance.String() != "100" {
		t.Errorf("FrozenBalance = %s, want 100", resp.FrozenBalance.String())
	}
	if resp.AvailableBalance.String() != "900" {
		t.Errorf("AvailableBalance = %s, want 900", resp.AvailableBalance.String())
	}
}

func TestGetBalance_Error404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Balance not found","code":"NOT_FOUND"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	_, err := client.GetBalance(context.Background(), "nonexistent", "token")

	if err == nil {
		t.Fatal("GetBalance() should return error for 404")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", apiErr.Code)
	}
	if apiErr.StatusCode() != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode())
	}
}

func TestGetBalance_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetBalance(ctx, "ethereum", "usdt")
	if err == nil {
		t.Fatal("GetBalance() should return error for cancelled context")
	}
}

func TestListBalances_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/tenants/tenant_123/balances") {
			t.Errorf("path = %q, want /tenants/tenant_123/balances", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"balances":[{"chain":"ethereum","token":"usdt","balance":"1000.00","frozen_balance":"100.00","available_balance":"900.00"},{"chain":"polygon","token":"usdc","balance":"500.00","frozen_balance":"0","available_balance":"500.00"}],"total":2}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	resp, err := client.ListBalances(context.Background())

	if err != nil {
		t.Fatalf("ListBalances() failed: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("Total = %d, want 2", resp.Total)
	}
	if len(resp.Balances) != 2 {
		t.Fatalf("len(Balances) = %d, want 2", len(resp.Balances))
	}

	if resp.Balances[0].Chain != "ethereum" {
		t.Errorf("Balances[0].Chain = %q, want ethereum", resp.Balances[0].Chain)
	}
	if resp.Balances[0].Token != "usdt" {
		t.Errorf("Balances[0].Token = %q, want usdt", resp.Balances[0].Token)
	}
	if resp.Balances[1].Chain != "polygon" {
		t.Errorf("Balances[1].Chain = %q, want polygon", resp.Balances[1].Chain)
	}
	if resp.Balances[1].Token != "usdc" {
		t.Errorf("Balances[1].Token = %q, want usdc", resp.Balances[1].Token)
	}
}

func TestListBalances_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"balances":[],"total":0}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	resp, err := client.ListBalances(context.Background())

	if err != nil {
		t.Fatalf("ListBalances() failed: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Total = %d, want 0", resp.Total)
	}
	if len(resp.Balances) != 0 {
		t.Errorf("len(Balances) = %d, want 0", len(resp.Balances))
	}
}

func TestListBalances_Error500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error","code":"INTERNAL_ERROR"}`))
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))
	_, err := client.ListBalances(context.Background())

	if err == nil {
		t.Fatal("ListBalances() should return error for 500")
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

func TestListBalances_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient("tenant_123", "ak_test", "sk_test", WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.ListBalances(ctx)
	if err == nil {
		t.Fatal("ListBalances() should return error for cancelled context")
	}
}
