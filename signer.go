package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"
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
// The signature is computed as HMAC-SHA256(canonicalized_payload, webhookSecret) and hex-encoded.
//
// The payload is canonicalized by re-marshaling the JSON with sorted keys to ensure
// consistent signature verification regardless of JSON key ordering from the source.
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
	canonicalized, err := canonicalizeJSON(payload)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(canonicalized)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// canonicalizeJSON re-marshals JSON with sorted keys for consistent hashing.
func canonicalizeJSON(data []byte) ([]byte, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

// DefaultMaxTimeDrift is the default maximum allowed time difference for request timestamps.
// This matches the payment gateway's 5 minute tolerance.
const DefaultMaxTimeDrift = 5 * time.Minute

// VerifyAPISignature verifies HMAC-SHA256 signature for incoming API requests.
// This function matches the payment gateway's signature verification middleware exactly.
//
// The signature is computed as: HMAC-SHA256(method + path + timestamp + body, apiSecret)
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path without query string (e.g., "/api/v1/payouts")
//   - timestamp: Unix timestamp string from X-Timestamp header
//   - body: Raw request body (empty string for GET requests)
//   - signature: The value from X-Signature header
//   - apiSecret: The tenant's API secret
//
// Returns true if the signature is valid, false otherwise.
// Timestamp freshness must be validated separately using VerifyTimestamp.
//
// Example:
//
//	func apiHandler(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//	    timestamp := r.Header.Get("X-Timestamp")
//	    signature := r.Header.Get("X-Signature")
//
//	    if !sdk.VerifyTimestamp(timestamp, sdk.DefaultMaxTimeDrift) {
//	        http.Error(w, "request expired", http.StatusUnauthorized)
//	        return
//	    }
//	    if !sdk.VerifyAPISignature(r.Method, r.URL.Path, timestamp, string(body), signature, apiSecret) {
//	        http.Error(w, "invalid signature", http.StatusUnauthorized)
//	        return
//	    }
//	    // Process request...
//	}
func VerifyAPISignature(method, path, timestamp, body, signature, apiSecret string) bool {
	expected := Sign(method, path, timestamp, body, apiSecret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifyTimestamp checks if a request timestamp is within the allowed time drift.
// This matches the payment gateway's timestamp validation logic.
//
// Parameters:
//   - timestampStr: Unix timestamp string from X-Timestamp header
//   - maxDrift: Maximum allowed time difference (use DefaultMaxTimeDrift for gateway compatibility)
//
// Returns true if the timestamp is valid (within drift), false otherwise.
func VerifyTimestamp(timestampStr string, maxDrift time.Duration) bool {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}

	return diff <= int64(maxDrift.Seconds())
}
