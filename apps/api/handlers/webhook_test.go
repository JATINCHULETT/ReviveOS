package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
)

func generateSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestRazorpayWebhookHandler_SignatureValidation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	secret := "test_webhook_secret_12345"
	os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)

	handler := RazorpayWebhookHandler(pool)

	payload := []byte(`{"event":"payment.failed","payload":{"payment":{"entity":{"id":"pay_sig_test","amount":100000}}}}`)

	// 1. Request with Invalid Signature
	reqInvalid, _ := http.NewRequest(http.MethodPost, "/webhooks/razorpay", bytes.NewBuffer(payload))
	reqInvalid.Header.Set("X-Razorpay-Signature", "invalid_signature_hex")
	reqInvalid.Header.Set("X-Razorpay-Event-Id", "evt_invalid_sig")
	recInvalid := httptest.NewRecorder()

	handler.ServeHTTP(recInvalid, reqInvalid)

	if recInvalid.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request for invalid signature, got %d", recInvalid.Code)
	}

	// 2. Request with Valid Signature
	validSig := generateSignature(payload, secret)
	reqValid, _ := http.NewRequest(http.MethodPost, "/webhooks/razorpay", bytes.NewBuffer(payload))
	reqValid.Header.Set("X-Razorpay-Signature", validSig)
	reqValid.Header.Set("X-Razorpay-Event-Id", fmt.Sprintf("evt_valid_%d", time.Now().UnixNano()))
	recValid := httptest.NewRecorder()

	handler.ServeHTTP(recValid, reqValid)

	if recValid.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for valid signature, got %d", recValid.Code)
	}

	t.Logf("SUCCESS: Webhook signature verification verified (Invalid -> 400, Valid -> 200)")
}

func TestRazorpayWebhookHandler_Deduplication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	secret := "test_webhook_secret_12345"
	os.Setenv("RAZORPAY_WEBHOOK_SECRET", secret)

	handler := RazorpayWebhookHandler(pool)

	eventID := fmt.Sprintf("evt_dedup_test_%d", time.Now().UnixNano())
	payID := fmt.Sprintf("pay_dedup_%d", time.Now().UnixNano())

	payloadJSON := fmt.Sprintf(`{
		"event": "payment.failed",
		"payload": {
			"payment": {
				"entity": {
					"id": "%s",
					"amount": 250000,
					"currency": "INR",
					"status": "failed",
					"method": "card",
					"error_code": "INSUFFICIENT_FUNDS",
					"error_description": "Declined by bank",
					"email": "dedup_cust@example.com",
					"contact": "+919876543210"
				}
			}
		}
	}`, payID)

	payloadBytes := []byte(payloadJSON)
	sig := generateSignature(payloadBytes, secret)

	// 1. Send First Webhook Request
	req1, _ := http.NewRequest(http.MethodPost, "/webhooks/razorpay", bytes.NewBuffer(payloadBytes))
	req1.Header.Set("X-Razorpay-Signature", sig)
	req1.Header.Set("X-Razorpay-Event-Id", eventID)
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("First webhook request failed: code %d, body: %s", rec1.Code, rec1.Body.String())
	}
	t.Logf("1. First Webhook Request (Event %s) -> Status: %d, Body: %s", eventID, rec1.Code, rec1.Body.String())

	// 2. Send Second (Duplicate) Webhook Request with same X-Razorpay-Event-Id
	req2, _ := http.NewRequest(http.MethodPost, "/webhooks/razorpay", bytes.NewBuffer(payloadBytes))
	req2.Header.Set("X-Razorpay-Signature", sig)
	req2.Header.Set("X-Razorpay-Event-Id", eventID)
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Second webhook request failed: code %d, body: %s", rec2.Code, rec2.Body.String())
	}
	t.Logf("2. Duplicate Webhook Request (Event %s) -> Status: %d, Body: %s", eventID, rec2.Code, rec2.Body.String())

	if !bytes.Contains(rec2.Body.Bytes(), []byte("duplicate_ignored")) {
		t.Errorf("Expected duplicate_ignored in response body, got: %s", rec2.Body.String())
	}

	// 3. Verify PostgreSQL only contains ONE payment_events row
	var eventCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM payment_events WHERE razorpay_event_id = $1", eventID).Scan(&eventCount)
	if err != nil {
		t.Fatalf("Failed to query payment_events: %v", err)
	}

	if eventCount != 1 {
		t.Fatalf("Deduplication failure: expected 1 payment_events row, found %d", eventCount)
	}

	// 4. Verify PostgreSQL only contains ONE payments row for this razorpay_payment_id
	var paymentCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM payments WHERE razorpay_payment_id = $1", payID).Scan(&paymentCount)
	if err != nil {
		t.Fatalf("Failed to query payments: %v", err)
	}

	if paymentCount != 1 {
		t.Fatalf("Deduplication failure: expected 1 payments row, found %d", paymentCount)
	}

	t.Logf("SUCCESS: Persistent event deduplication verified (1 event row, 1 payment row, second request acknowledged duplicate_ignored)")
}
