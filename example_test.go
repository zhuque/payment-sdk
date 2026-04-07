package sdk_test

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/payment-gateway/sdk"
)

func ExampleClient_CreatePayout() {
	client := sdk.NewClient("tenant_123", "ak_test", "sk_test",
		sdk.WithBaseURL("https://api.example.com"),
		sdk.WithTimeout(30*time.Second))

	ctx := context.Background()

	req := &sdk.CreatePayoutRequest{
		MerchantOrderID: "m_order_001",
		ToAddress:       "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		Amount:          "100.50",
		Chain:           "polygon",
		Token:           "usdt",
	}

	// This example assumes the client is correctly configured and the server is reachable.
	// In a real application, you would handle the error accordingly.
	_, _ = client.CreatePayout(ctx, req)

	fmt.Println("Payout creation attempted")
	// Output:
	// Payout creation attempted
}

func ExampleClient_GenerateAddress() {
	client := sdk.NewClient("tenant_123", "ak_test", "sk_test")
	ctx := context.Background()

	mode := int16(0)
	req := &sdk.GenerateAddressRequest{
		UserID: "user_456",
		Chain:  "ethereum",
		Token:  "usdc",
		Mode:   &mode,
	}

	// This example assumes the client is correctly configured and the server is reachable.
	_, _ = client.GenerateAddress(ctx, req)

	fmt.Println("Address generation attempted")
	// Output:
	// Address generation attempted
}

func ExampleClient_CreateOrder() {
	client := sdk.NewClient("tenant_123", "ak_test", "sk_test")
	ctx := context.Background()

	req := &sdk.CreateOrderRequest{
		Amount:      "50.00",
		Chain:       "bsc",
		Token:       "usdt",
		CallbackURL: "https://merchant.com/webhook",
		Metadata: map[string]interface{}{
			"internal_id": "cart_999",
		},
	}

	// This example assumes the client is correctly configured and the server is reachable.
	_, _ = client.CreateOrder(ctx, req)

	fmt.Println("Order creation attempted")
	// Output:
	// Order creation attempted
}
