# Payment Gateway SDK for Go

Official Go HTTP client SDK for the Payment Gateway. This SDK provides a simple interface for payout and collection operations across multiple blockchains.

## Installation

```bash
go get github.com/yourorg/payment-gateway/sdk
```

## Quick Start

### Client Initialization

```go
import (
    "context"
    "time"
    "github.com/yourorg/payment-gateway/sdk"
)

client := sdk.NewClient(
    "tenant_123", 
    "your_api_key", 
    "your_api_secret",
    sdk.WithBaseURL("https://api.example.com"),
    sdk.WithTimeout(30*time.Second),
)
ctx := context.Background()
```

### Create a Payout

```go
resp, err := client.CreatePayout(ctx, &sdk.CreatePayoutRequest{
    MerchantOrderID: "order_12345",
    ToAddress:       "0x123...",
    Amount:          "100.50",
    Chain:           "polygon",
    Token:           "usdt",
})
```

### Generate a Payment Address

```go
mode := int16(0) // 0: fixed, 1: one-time
resp, err := client.GenerateAddress(ctx, &sdk.GenerateAddressRequest{
    UserID: "user_789",
    Chain:  "ethereum",
    Token:  "usdc",
    Mode:   &mode,
})
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithBaseURL(url string)` | Sets the API base URL | `http://localhost:8080` |
| `WithTimeout(d time.Duration)` | Sets the HTTP request timeout | `30s` |
| `WithHTTPClient(hc *http.Client)` | Uses a custom HTTP client | `&http.Client{Timeout: 30s}` |

## API Reference

### Payout Operations

#### `CreatePayout(ctx, req)`
Creates a new payout request. This operation is NOT idempotent.
- **Request**: `CreatePayoutRequest` (MerchantOrderID, ToAddress, Amount, Chain, Token)
- **Response**: `PayoutResponse` (PayoutID, Status, Amount, etc.)

#### `GetPayout(ctx, payoutID)`
Retrieves the status of a specific payout.
- **Response**: `PayoutResponse` (includes `TxHash` and `ErrorMessage` if available)

#### `ListPayouts(ctx, limit, offset)`
Lists payouts for the tenant with pagination.
- **Response**: `ListPayoutsResponse`

### Collection Operations

#### `GenerateAddress(ctx, req)`
Generates a payment address for a user.
- **Request**: `GenerateAddressRequest` (UserID, Chain, Token, Mode)
- **Response**: `GenerateAddressResponse` (Address, Path, etc.)

#### `CreateOrder(ctx, req)`
Creates a payment order with an expiration time. Each call generates a new address.
- **Request**: `CreateOrderRequest` (OrderID, Amount, Chain, Token, UserID, Metadata)
- **Response**: `CreateOrderResponse` (OrderID, PaymentAddress, ExpiresAt, etc.)

#### `GetBalance(ctx, chain, token)`
Gets the balance for a specific chain and token.
- **Response**: `BalanceResponse` (Balance, FrozenBalance, AvailableBalance)

#### `ListBalances(ctx)`
Lists all balances across all chains and tokens for the tenant.
- **Response**: `ListBalancesResponse`

## Error Handling

The SDK returns `*sdk.APIError` for API-level errors (HTTP 4xx/5xx).

```go
var apiErr *sdk.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Code: %s, Message: %s\n", 
        apiErr.StatusCode(), apiErr.Code, apiErr.Message)
    
    if apiErr.IsRetryable() {
        // Implement exponential backoff
    }
}
```

## Supported Chains & Tokens

| Chain | Tokens |
|-------|--------|
| `ethereum` | `usdt`, `usdc` |
| `polygon` | `usdt`, `usdc` |
| `bsc` | `usdt`, `usdc` |
| `arbitrum` | `usdt`, `usdc` |
| `optimism` | `usdt`, `usdc` |
| `base` | `usdt`, `usdc` |
| `tron` | `usdt`, `usdc` |

## Best Practices

1. **Idempotency**: Payout and Order creation are NOT idempotent. Use your own `MerchantOrderID` and track it to avoid duplicate operations.
2. **Timeouts**: Use a reasonable timeout (default 30s) for payment operations.
3. **Security**: Never hardcode your `API_SECRET`. Use environment variables or a secrets manager.
4. **Retry Logic**: Only retry if `apiErr.IsRetryable()` returns true.
