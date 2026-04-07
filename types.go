package sdk

import (
	"time"

	"github.com/shopspring/decimal"
)

// ============================================================================
// PAYOUT API TYPES
// ============================================================================

// CreatePayoutRequest represents the request to create a payout
type CreatePayoutRequest struct {
	MerchantOrderID string `json:"merchant_order_id"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	Chain           string `json:"chain"`
	Token           string `json:"token"`
}

// CreatePayoutResponse represents the response after creating a payout
type CreatePayoutResponse struct {
	PayoutID        string          `json:"payout_id"`
	MerchantOrderID string          `json:"merchant_order_id"`
	ToAddress       string          `json:"to_address"`
	Amount          decimal.Decimal `json:"amount"`
	Chain           string          `json:"chain"`
	Token           string          `json:"token"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
}

// PayoutResponse represents a payout in list/get responses
type PayoutResponse struct {
	PayoutID        string          `json:"payout_id"`
	MerchantOrderID string          `json:"merchant_order_id"`
	ToAddress       string          `json:"to_address"`
	Amount          decimal.Decimal `json:"amount"`
	Chain           string          `json:"chain"`
	Token           string          `json:"token"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	TxHash          *string         `json:"tx_hash,omitempty"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ListPayoutsResponse represents the response for listing payouts
type ListPayoutsResponse struct {
	Payouts []*PayoutResponse `json:"payouts"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

// ============================================================================
// COLLECTION API TYPES - ADDRESS GENERATION
// ============================================================================

// GenerateAddressRequest represents the request to generate a payment address
type GenerateAddressRequest struct {
	UserID string `json:"user_id"`
	Chain  string `json:"chain"`
	Token  string `json:"token"`
	Mode   *int16 `json:"mode"` // nullable, defaults to 0 (fixed)
}

// GenerateAddressResponse represents the response after generating an address
type GenerateAddressResponse struct {
	Address string `json:"address"`
	Path    string `json:"path"`
	Chain   string `json:"chain"`
	Token   string `json:"token"`
}

// ============================================================================
// COLLECTION API TYPES - ORDER CREATION
// ============================================================================

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	Amount      string                 `json:"amount"`
	Chain       string                 `json:"chain"`
	Token       string                 `json:"token"`
	CallbackURL string                 `json:"callback_url"`
	Metadata    map[string]interface{} `json:"metadata"`
	UserID      string                 `json:"user_id"`
}

// CreateOrderResponse represents the response after creating an order
type CreateOrderResponse struct {
	OrderID        string          `json:"order_id"`
	PaymentAddress string          `json:"payment_address"`
	Amount         decimal.Decimal `json:"amount"`
	Chain          string          `json:"chain"`
	Token          string          `json:"token"`
	ExpiresAt      time.Time       `json:"expires_at"`
	Status         string          `json:"status"`
}

// ============================================================================
// COLLECTION API TYPES - BALANCE
// ============================================================================

// BalanceResponse represents a balance for a chain/token pair
type BalanceResponse struct {
	Chain            string          `json:"chain"`
	Token            string          `json:"token"`
	Balance          decimal.Decimal `json:"balance"`
	FrozenBalance    decimal.Decimal `json:"frozen_balance"`
	AvailableBalance decimal.Decimal `json:"available_balance"`
}

// ListBalancesResponse represents the response for listing balances
type ListBalancesResponse struct {
	Balances []*BalanceResponse `json:"balances"`
	Total    int                `json:"total"`
}

// ============================================================================
// ERROR RESPONSE TYPE
// ============================================================================

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
