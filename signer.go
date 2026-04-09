package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign generates HMAC-SHA256 signature for API authentication.
// It concatenates method + path + timestamp + body (no delimiters),
// computes HMAC-SHA256 using the API secret, and returns hex-encoded string.
//
// Example:
//
//	signature := Sign("POST", "/api/v1/payouts", "1234567890", `{"amount":"100"}`, "your_secret")
//
// The signature must be sent in X-Signature header along with X-API-Key and X-Timestamp.
func Sign(method, path, timestamp, body, apiSecret string) string {
	message := method + path + timestamp + body
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of a webhook payload.
// The signature is computed as HMAC-SHA256(payload, webhookSecret) and hex-encoded.
//
// Parameters:
//   - payload: the raw JSON body received from the webhook request
//   - signature: the value from X-Signature header
//   - webhookSecret: your webhook secret configured in the payment gateway
//
// Returns true if the signature is valid, false otherwise.
//
// Example:
//
//	func webhookHandler(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//	    signature := r.Header.Get("X-Signature")
//	    if !sdk.VerifyWebhookSignature(body, signature, "your_webhook_secret") {
//	        http.Error(w, "invalid signature", http.StatusUnauthorized)
//	        return
//	    }
//	    // Process webhook...
//	}
func VerifyWebhookSignature(payload []byte, signature, webhookSecret string) bool {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
