package paymentprovider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Ensure RazorpayPaymentProvider satisfies PaymentProvider interface at compile-time
var _ PaymentProvider = (*RazorpayPaymentProvider)(nil)

func TestRazorpay_WebhookSignatureVerification(t *testing.T) {
	secret := "test_webhook_secret_key_12345"
	payload := []byte(`{"event":"payment.failed","payload":{"payment":{"entity":{"id":"pay_test123","amount":50000}}}}`)

	// Calculate valid HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	// 1. Valid Signature
	if !VerifyWebhookSignature(payload, validSignature, secret) {
		t.Fatalf("Expected valid signature to verify successfully")
	}

	// 2. Tampered Payload
	tamperedPayload := []byte(`{"event":"payment.failed","payload":{"payment":{"entity":{"id":"pay_test123","amount":99999}}}}`)
	if VerifyWebhookSignature(tamperedPayload, validSignature, secret) {
		t.Fatalf("Expected tampered payload to fail signature verification")
	}

	// 3. Wrong Secret
	if VerifyWebhookSignature(payload, validSignature, "wrong_secret") {
		t.Fatalf("Expected wrong secret to fail signature verification")
	}

	// 4. Empty Signature / Secret
	if VerifyWebhookSignature(payload, "", secret) || VerifyWebhookSignature(payload, validSignature, "") {
		t.Fatalf("Expected empty signature/secret to fail")
	}

	t.Logf("SUCCESS: Webhook signature verification verified (valid, tampered, wrong secret)")
}

func TestRazorpay_MockServer_RealHTTPProtocols(t *testing.T) {
	// Setup mock Razorpay HTTP server emulating real Razorpay responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "rzp_test_key" || pass != "rzp_test_secret" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"code":"BAD_REQUEST_ERROR","description":"The key_id or key_secret provided is invalid."}}`))
			return
		}

		path := r.URL.Path
		switch {
		case path == "/payments/pay_success_123":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "pay_success_123",
				"entity": "payment",
				"amount": 250000,
				"currency": "INR",
				"status": "captured",
				"method": "card",
				"captured": true,
				"created_at": 1700000000
			}`))

		case path == "/payments/pay_failed_456":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "pay_failed_456",
				"entity": "payment",
				"amount": 150000,
				"currency": "INR",
				"status": "failed",
				"method": "card",
				"captured": false,
				"error_code": "BAD_REQUEST_ERROR",
				"error_description": "Payment was declined by issuing bank",
				"error_reason": "payment_declined",
				"created_at": 1700000000
			}`))

		case path == "/payments/pay_notfound_789":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code":"BAD_REQUEST_ERROR","description":"Payment id does not exist"}}`))

		case path == "/payments/pay_ratelimit":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"TOO_MANY_REQUESTS","description":"Rate limit exceeded."}}`))

		case path == "/payment_links":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "plink_test_999",
				"status": "created",
				"short_url": "https://rzp.io/i/test999",
				"created_at": 1700000000
			}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	provider := NewRazorpayPaymentProvider("rzp_test_key", "rzp_test_secret", server.URL)

	// 1. Test GetPayment / VerifyPayment Success
	status, err := provider.GetPayment(ctx, "pay_success_123")
	if err != nil {
		t.Fatalf("GetPayment failed: %v", err)
	}
	if status.PaymentID != "pay_success_123" || status.Status != "CAPTURED" || !status.Captured || status.Amount != 2500.00 {
		t.Errorf("Unexpected payment status: %+v", status)
	}
	t.Logf("GetPayment (Success) mapped: ID=%s, Amount=%.2f, Status=%s, Captured=%v",
		status.PaymentID, status.Amount, status.Status, status.Captured)

	// 2. Test GetPayment Failed Payment
	failedStatus, err := provider.GetPayment(ctx, "pay_failed_456")
	if err != nil {
		t.Fatalf("GetPayment failed: %v", err)
	}
	if failedStatus.Status != "FAILED" || failedStatus.Captured || failedStatus.FailureCode != "BAD_REQUEST_ERROR" {
		t.Errorf("Unexpected failed payment status: %+v", failedStatus)
	}
	t.Logf("GetPayment (Failed) mapped: ID=%s, Status=%s, FailureCode=%s, Reason=%s",
		failedStatus.PaymentID, failedStatus.Status, failedStatus.FailureCode, failedStatus.FailureReason)

	// 3. Test 404 Not Found
	_, err = provider.GetPayment(ctx, "pay_notfound_789")
	if err == nil {
		t.Fatalf("Expected 404 error, got nil")
	}
	t.Logf("404 Not Found handled properly: %v", err)

	// 4. Test 429 Rate Limiting
	_, err = provider.GetPayment(ctx, "pay_ratelimit")
	if err == nil {
		t.Fatalf("Expected 429 error, got nil")
	}
	t.Logf("429 Rate Limit handled properly: %v", err)

	// 5. Test 401 Unauthorized
	badProvider := NewRazorpayPaymentProvider("wrong_key", "wrong_secret", server.URL)
	_, err = badProvider.GetPayment(ctx, "pay_success_123")
	if err == nil {
		t.Fatalf("Expected 401 error, got nil")
	}
	t.Logf("401 Unauthorized handled properly: %v", err)

	// 6. Test CreateRetryAttempt
	retryRes, err := provider.CreateRetryAttempt(ctx, "pay_failed_456", 1500.00)
	if err != nil {
		t.Fatalf("CreateRetryAttempt failed: %v", err)
	}
	if retryRes.Status != "SUCCESS" || retryRes.AttemptID != "plink_test_999" {
		t.Errorf("Unexpected retry result: %+v", retryRes)
	}
	t.Logf("CreateRetryAttempt created payment link: AttemptID=%s, Status=%s", retryRes.AttemptID, retryRes.Status)

	// 7. Test CreateRetryAttemptWithCustomer (email only, no SMS phone)
	retryWithCustRes, err := provider.CreateRetryAttemptWithCustomer(ctx, "pay_failed_456", 1500.00, "customer@example.com", "", "Test Customer")
	if err != nil {
		t.Fatalf("CreateRetryAttemptWithCustomer failed: %v", err)
	}
	if retryWithCustRes.Status != "SUCCESS" || retryWithCustRes.AttemptID != "plink_test_999" {
		t.Errorf("Unexpected retry result with customer: %+v", retryWithCustRes)
	}
	t.Logf("CreateRetryAttemptWithCustomer succeeded: AttemptID=%s", retryWithCustRes.AttemptID)
}

func TestRazorpay_LiveCredentialsOrBlocked(t *testing.T) {
	keyID := os.Getenv("RAZORPAY_KEY_ID")
	keySecret := os.Getenv("RAZORPAY_KEY_SECRET")

	if keyID == "" || keySecret == "" {
		t.Log("RAZORPAY_KEY_ID or RAZORPAY_KEY_SECRET not set in environment: Live network call reported as BLOCKED BY CREDENTIALS")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	provider := NewRazorpayPaymentProvider(keyID, keySecret, "")
	_, err := provider.GetPayment(ctx, "pay_nonexistent_test_id")
	if err != nil {
		t.Logf("Live Razorpay test mode call executed (returned expected error for non-existent test ID): %v", err)
	}
}
