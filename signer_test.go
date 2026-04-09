package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"
	"time"
)

func TestSign_EmptyBody(t *testing.T) {
	// Test GET request with empty body
	method := "GET"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := ""
	secret := "test_secret"

	signature := Sign(method, path, timestamp, body, secret)

	// Verify signature is non-empty hex string
	if signature == "" {
		t.Fatal("signature should not be empty")
	}
	if len(signature) != 64 { // SHA256 produces 32 bytes = 64 hex chars
		t.Errorf("signature length = %d, want 64", len(signature))
	}

	// Compute expected signature manually for verification
	message := method + path + timestamp + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	if signature != expected {
		t.Errorf("Sign() = %s, want %s", signature, expected)
	}
}

func TestSign_JSONBody(t *testing.T) {
	// Test POST request with JSON body
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := `{"merchant_order_id":"order123","to_address":"0x1234","amount":"100.50","chain":"ethereum","token":"usdt"}`
	secret := "test_secret"

	signature := Sign(method, path, timestamp, body, secret)

	// Verify signature format
	if len(signature) != 64 {
		t.Errorf("signature length = %d, want 64", len(signature))
	}

	// Verify signature is lowercase hex
	for _, c := range signature {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("signature contains non-hex char: %c", c)
		}
	}

	// Compute expected and compare
	message := method + path + timestamp + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	if signature != expected {
		t.Errorf("Sign() = %s, want %s", signature, expected)
	}
}

func TestSign_SpecialCharacters(t *testing.T) {
	// Test with special characters in body
	method := "POST"
	path := "/api/v1/tenants/2b123/addresses"
	timestamp := "1714567890"
	body := `{"user_id":"user@example.com","chain":"polygon","token":"usdc","data":"测试\n\t<>\"&"}`
	secret := "secret_with_$pecial_ch@rs!"

	signature := Sign(method, path, timestamp, body, secret)

	// Compute expected manually
	message := method + path + timestamp + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	if signature != expected {
		t.Errorf("Sign() = %s, want %s", signature, expected)
	}
}

func TestSign_DifferentMethods(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		path      string
		timestamp string
		body      string
		secret    string
	}{
		{
			name:      "GET with query params in path",
			method:    "GET",
			path:      "/api/v1/payouts?limit=10&offset=0",
			timestamp: "1234567890",
			body:      "",
			secret:    "secret",
		},
		{
			name:      "POST with nested path",
			method:    "POST",
			path:      "/api/v1/tenants/2b123/orders",
			timestamp: "1234567890",
			body:      `{"amount":"50"}`,
			secret:    "secret",
		},
		{
			name:      "PUT request",
			method:    "PUT",
			path:      "/api/v1/tenants/2b123/payouts/123",
			timestamp: "1234567890",
			body:      `{"status":"completed"}`,
			secret:    "secret",
		},
		{
			name:      "DELETE request",
			method:    "DELETE",
			path:      "/api/v1/tenants/2b123/addresses/456",
			timestamp: "1234567890",
			body:      "",
			secret:    "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := Sign(tt.method, tt.path, tt.timestamp, tt.body, tt.secret)

			// Verify format
			if len(signature) != 64 {
				t.Errorf("signature length = %d, want 64", len(signature))
			}

			// Compute expected
			message := tt.method + tt.path + tt.timestamp + tt.body
			mac := hmac.New(sha256.New, []byte(tt.secret))
			mac.Write([]byte(message))
			expected := hex.EncodeToString(mac.Sum(nil))

			if signature != expected {
				t.Errorf("Sign() = %s, want %s", signature, expected)
			}
		})
	}
}

func TestSign_ConsistencyAcrossCalls(t *testing.T) {
	// Same inputs should always produce same signature
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := `{"amount":"100"}`
	secret := "secret"

	sig1 := Sign(method, path, timestamp, body, secret)
	sig2 := Sign(method, path, timestamp, body, secret)
	sig3 := Sign(method, path, timestamp, body, secret)

	if sig1 != sig2 || sig2 != sig3 {
		t.Errorf("signatures are inconsistent: %s, %s, %s", sig1, sig2, sig3)
	}
}

func TestSign_SensitivityToInputChanges(t *testing.T) {
	// Changing any input should produce different signature
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := `{"amount":"100"}`
	secret := "secret"

	baseSig := Sign(method, path, timestamp, body, secret)

	// Change method
	if sig := Sign("GET", path, timestamp, body, secret); sig == baseSig {
		t.Error("changing method should change signature")
	}

	// Change path
	if sig := Sign(method, "/api/v2/payouts", timestamp, body, secret); sig == baseSig {
		t.Error("changing path should change signature")
	}

	// Change timestamp
	if sig := Sign(method, path, "9999999999", body, secret); sig == baseSig {
		t.Error("changing timestamp should change signature")
	}

	// Change body
	if sig := Sign(method, path, timestamp, `{"amount":"200"}`, secret); sig == baseSig {
		t.Error("changing body should change signature")
	}

	// Change secret
	if sig := Sign(method, path, timestamp, body, "different_secret"); sig == baseSig {
		t.Error("changing secret should change signature")
	}
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	payload := []byte(`{"event":"payment.confirmed","order_id":"order123","amount":"100.50"}`)
	secret := "webhook_secret_key"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	if !VerifyWebhookSignature(payload, signature, secret) {
		t.Error("VerifyWebhookSignature should return true for valid signature")
	}
}

func TestVerifyWebhookSignature_Invalid(t *testing.T) {
	payload := []byte(`{"event":"payment.confirmed","order_id":"order123"}`)
	secret := "webhook_secret_key"
	invalidSignature := "0000000000000000000000000000000000000000000000000000000000000000"

	if VerifyWebhookSignature(payload, invalidSignature, secret) {
		t.Error("VerifyWebhookSignature should return false for invalid signature")
	}
}

func TestVerifyWebhookSignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"event":"payment.confirmed"}`)
	correctSecret := "correct_secret"
	wrongSecret := "wrong_secret"

	mac := hmac.New(sha256.New, []byte(correctSecret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	if VerifyWebhookSignature(payload, signature, wrongSecret) {
		t.Error("VerifyWebhookSignature should return false when secret doesn't match")
	}
}

func TestVerifyWebhookSignature_TamperedPayload(t *testing.T) {
	originalPayload := []byte(`{"event":"payment.confirmed","amount":"100"}`)
	tamperedPayload := []byte(`{"event":"payment.confirmed","amount":"999"}`)
	secret := "webhook_secret"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(originalPayload)
	signature := hex.EncodeToString(mac.Sum(nil))

	if VerifyWebhookSignature(tamperedPayload, signature, secret) {
		t.Error("VerifyWebhookSignature should return false for tampered payload")
	}
}

func TestVerifyWebhookSignature_EmptyPayload(t *testing.T) {
	payload := []byte{}
	secret := "webhook_secret"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	if !VerifyWebhookSignature(payload, signature, secret) {
		t.Error("VerifyWebhookSignature should handle empty payload")
	}
}

func TestVerifyWebhookSignature_TimingAttackResistance(t *testing.T) {
	payload := []byte(`{"event":"payment.confirmed"}`)
	secret := "webhook_secret"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	almostValidSig := "a" + validSig[1:]

	if VerifyWebhookSignature(payload, almostValidSig, secret) {
		t.Error("VerifyWebhookSignature should reject almost-valid signatures")
	}
}

func TestVerifyAPISignature_Valid(t *testing.T) {
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := `{"amount":"100"}`
	secret := "api_secret"

	signature := Sign(method, path, timestamp, body, secret)

	if !VerifyAPISignature(method, path, timestamp, body, signature, secret) {
		t.Error("VerifyAPISignature should return true for valid signature")
	}
}

func TestVerifyAPISignature_Invalid(t *testing.T) {
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := `{"amount":"100"}`
	secret := "api_secret"
	wrongSignature := "0000000000000000000000000000000000000000000000000000000000000000"

	if VerifyAPISignature(method, path, timestamp, body, wrongSignature, secret) {
		t.Error("VerifyAPISignature should return false for invalid signature")
	}
}

func TestVerifyAPISignature_WrongSecret(t *testing.T) {
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	body := `{"amount":"100"}`
	correctSecret := "correct_secret"
	wrongSecret := "wrong_secret"

	signature := Sign(method, path, timestamp, body, correctSecret)

	if VerifyAPISignature(method, path, timestamp, body, signature, wrongSecret) {
		t.Error("VerifyAPISignature should return false when secret doesn't match")
	}
}

func TestVerifyAPISignature_TamperedBody(t *testing.T) {
	method := "POST"
	path := "/api/v1/payouts"
	timestamp := "1234567890"
	originalBody := `{"amount":"100"}`
	tamperedBody := `{"amount":"999"}`
	secret := "api_secret"

	signature := Sign(method, path, timestamp, originalBody, secret)

	if VerifyAPISignature(method, path, timestamp, tamperedBody, signature, secret) {
		t.Error("VerifyAPISignature should return false for tampered body")
	}
}

func TestVerifyAPISignature_EmptyBody(t *testing.T) {
	method := "GET"
	path := "/api/v1/payouts/123"
	timestamp := "1234567890"
	body := ""
	secret := "api_secret"

	signature := Sign(method, path, timestamp, body, secret)

	if !VerifyAPISignature(method, path, timestamp, body, signature, secret) {
		t.Error("VerifyAPISignature should handle empty body (GET requests)")
	}
}

func TestVerifyTimestamp_Valid(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	if !VerifyTimestamp(timestamp, DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return true for current timestamp")
	}
}

func TestVerifyTimestamp_Expired(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)

	if VerifyTimestamp(timestamp, DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return false for expired timestamp (10 min old)")
	}
}

func TestVerifyTimestamp_Future(t *testing.T) {
	timestamp := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)

	if VerifyTimestamp(timestamp, DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return false for future timestamp (10 min ahead)")
	}
}

func TestVerifyTimestamp_InvalidFormat(t *testing.T) {
	if VerifyTimestamp("invalid", DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return false for invalid timestamp format")
	}

	if VerifyTimestamp("", DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return false for empty timestamp")
	}
}

func TestVerifyTimestamp_BoundaryConditions(t *testing.T) {
	now := time.Now().Unix()

	justWithinDrift := strconv.FormatInt(now-int64(DefaultMaxTimeDrift.Seconds()), 10)
	if !VerifyTimestamp(justWithinDrift, DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return true for timestamp exactly at max drift")
	}

	justOutsideDrift := strconv.FormatInt(now-int64(DefaultMaxTimeDrift.Seconds())-1, 10)
	if VerifyTimestamp(justOutsideDrift, DefaultMaxTimeDrift) {
		t.Error("VerifyTimestamp should return false for timestamp just beyond max drift")
	}
}
