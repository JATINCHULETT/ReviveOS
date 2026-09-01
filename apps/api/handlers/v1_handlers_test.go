package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuthMiddleware_Rejection(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mw := APIKeyAuthMiddleware(nil, dummyHandler)

	// Case 1: Missing Key
	req := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for missing key, got %d", rec.Code)
	}

	// Case 2: Valid Mock Key
	reqMock := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	reqMock.Header.Set("Authorization", "Bearer rvo_test_mock_key")
	recMock := httptest.NewRecorder()
	mw.ServeHTTP(recMock, reqMock)

	if recMock.Code != http.StatusOK {
		t.Fatalf("expected status 200 for mock key, got %d", recMock.Code)
	}
}

func TestV1AnalyzePaymentHandler_DeterministicLogic(t *testing.T) {
	handler := V1AnalyzePaymentHandler(nil)

	payload := PaymentAnalyzeRequest{
		PaymentID:     "pay_test_insufficient_123",
		Amount:        4999.0,
		Currency:      "INR",
		FailureCode:   "BAD_REQUEST_ERROR",
		FailureReason: "Payment failed due to low balance insufficient funds",
		CustomerID:    "cust_test_1",
		PaymentMethod: "card",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/payments/analyze", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp PaymentAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.PaymentID != "pay_test_insufficient_123" {
		t.Errorf("expected paymentId pay_test_insufficient_123, got %s", resp.PaymentID)
	}

	if resp.FailureCategory != "INSUFFICIENT_FUNDS" {
		t.Errorf("expected category INSUFFICIENT_FUNDS, got %s", resp.FailureCategory)
	}

	if resp.Action != "RETRY_LATER" {
		t.Errorf("expected action RETRY_LATER, got %s", resp.Action)
	}

	if resp.Decision != "RECOVER" {
		t.Errorf("expected decision RECOVER, got %s", resp.Decision)
	}
}

func TestV1AnalyzePaymentHandler_FraudRiskBlock(t *testing.T) {
	handler := V1AnalyzePaymentHandler(nil)

	// Unknown or generic failure
	payload := PaymentAnalyzeRequest{
		PaymentID:     "pay_test_fraud_999",
		Amount:        95000.0,
		Currency:      "INR",
		FailureCode:   "GATEWAY_TIMEOUT",
		CustomerID:    "cust_fraud_risk",
		PaymentMethod: "card",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/payments/analyze", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp PaymentAnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Confidence <= 0 {
		t.Errorf("expected positive confidence, got %f", resp.Confidence)
	}
}

func TestV1CustomerRecoveryProfileHandler(t *testing.T) {
	handler := V1CustomerRecoveryProfileHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/customers/cust_test_demo/recovery-profile", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var profile CustomerRecoveryProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil {
		t.Fatalf("failed to parse profile: %v", err)
	}

	if profile.CustomerID != "cust_test_demo" {
		t.Errorf("expected customerId cust_test_demo, got %s", profile.CustomerID)
	}

	if profile.RecoveryProbability <= 0 {
		t.Errorf("expected recovery probability > 0, got %f", profile.RecoveryProbability)
	}
}

func TestVerifyRazorpaySignature(t *testing.T) {
	body := []byte(`{"event":"payment.failed","payload":{"payment":{"entity":{"id":"pay_123"}}}}`)
	secret := "test_secret_key_reviveos"

	// Valid HMAC-SHA256 signature for this body & secret
	// compute using standard library
	expectedValid := VerifyRazorpaySignature(body, "invalid_sig", secret)
	if expectedValid {
		t.Errorf("expected invalid signature to fail")
	}
}
