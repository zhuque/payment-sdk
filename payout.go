package sdk

import (
	"context"
	"fmt"
)

// CreatePayout creates a new payout request.
// This operation is NOT idempotent and should NOT be automatically retried.
func (c *Client) CreatePayout(ctx context.Context, req *CreatePayoutRequest) (*PayoutResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/payouts", c.tenantID)
	var resp PayoutResponse
	if err := c.Call(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPayout retrieves the status of a payout by ID.
func (c *Client) GetPayout(ctx context.Context, payoutID string) (*PayoutResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/payouts/%s", c.tenantID, payoutID)
	var resp PayoutResponse
	if err := c.Call(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPayouts retrieves a list of payouts with pagination.
func (c *Client) ListPayouts(ctx context.Context, limit, offset int) (*ListPayoutsResponse, error) {
	path := fmt.Sprintf("/api/v1/tenants/%s/payouts?limit=%d&offset=%d", c.tenantID, limit, offset)
	var resp ListPayoutsResponse
	if err := c.Call(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
