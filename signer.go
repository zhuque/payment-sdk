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
