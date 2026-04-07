package sdk

import (
	"context"
	"fmt"
)

// ============================================================================
// COLLECTION API OPERATIONS
// ============================================================================

// GenerateAddress generates a payment address for a user.
// POST /api/v1/tenants/{tenantID}/addresses
//
// Example:
//
//	resp, err := client.GenerateAddress(ctx, &GenerateAddressRequest{
//	    UserID: "user_123",
//	    Chain:  "ethereum",
//	    Token:  "usdt",
//	    Mode:   &modeFixed, // 0: fixed, 1: one-time
//	})
func (c *Client) GenerateAddress(ctx context.Context, req *GenerateAddressRequest) (*GenerateAddressResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/addresses", c.tenantID)
	var resp GenerateAddressResponse
	if err := c.Call(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateOrder creates a new payment order.
// POST /api/v1/tenants/{tenantID}/orders
//
// This operation is NOT idempotent and should NOT be automatically retried.
// Each call creates a new order with a new payment address.
//
// Example:
//
//	resp, err := client.CreateOrder(ctx, &CreateOrderRequest{
//	    Amount:      "100.50",
//	    Chain:       "polygon",
//	    Token:       "usdc",
//	    CallbackURL: "https://example.com/webhook",
//	    Metadata:    map[string]interface{}{"order_id": "12345"},
//	})
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/orders", c.tenantID)
	var resp CreateOrderResponse
	if err := c.Call(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBalance retrieves the balance for a specific chain and token.
// GET /api/v1/tenants/{tenantID}/balances/{chain}/{token}
//
// Returns frozen, available, and total balances for the chain/token pair.
//
// Example:
//
//	resp, err := client.GetBalance(ctx, "ethereum", "usdt")
func (c *Client) GetBalance(ctx context.Context, chain, token string) (*BalanceResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/balances/%s/%s", c.tenantID, chain, token)
	var resp BalanceResponse
	if err := c.Call(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListBalances retrieves all balances for the tenant.
// GET /api/v1/tenants/{tenantID}/balances
//
// Example:
//
//	resp, err := client.ListBalances(ctx)
func (c *Client) ListBalances(ctx context.Context) (*ListBalancesResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/balances", c.tenantID)
	var resp ListBalancesResponse
	if err := c.Call(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
