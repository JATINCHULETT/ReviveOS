package notificationservice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogOnlyProvider(t *testing.T) {
	provider := NewLogOnlyProvider()
	req := NotificationRequest{
		PaymentID:     "pay_test123",
		MerchantName:  "Test Store",
		CustomerEmail: "customer@example.com",
		Amount:        999.00,
		Currency:      "INR",
		PaymentLink:   "https://rzp.io/i/test1234",
		ActionType:    "PAYMENT_LINK",
	}

	res, err := provider.SendRecoveryNotification(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "SENT" {
		t.Errorf("expected status SENT, got %s", res.Status)
	}
	if res.Provider != "log" {
		t.Errorf("expected provider log, got %s", res.Provider)
	}
}

func TestResendProviderSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer re_test_key" {
			http.Error(w, `{"error":{"message":"unauthorized"}}`, http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "msg_resend_123"}`))
	}))
	defer ts.Close()

	provider := NewResendProvider("re_test_key", "onboarding@resend.dev")
	provider.BaseURL = ts.URL

	req := NotificationRequest{
		PaymentID:     "pay_987",
		MerchantName:  "Acme SaaS",
		CustomerEmail: "user@example.com",
		Amount:        4999.00,
		Currency:      "INR",
		PaymentLink:   "https://rzp.io/i/link123",
		ActionType:    "PAYMENT_LINK",
	}

	res, err := provider.SendRecoveryNotification(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "SENT" {
		t.Errorf("expected status SENT, got %s", res.Status)
	}
	if res.MessageID != "msg_resend_123" {
		t.Errorf("expected msg_resend_123, got %s", res.MessageID)
	}
}

func TestResendProviderMissingKey(t *testing.T) {
	provider := NewResendProvider("", "")
	req := NotificationRequest{
		CustomerEmail: "test@example.com",
		PaymentLink:   "https://rzp.io/i/test",
	}

	_, err := provider.SendRecoveryNotification(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error when API key is missing, got nil")
	}
}

func TestFactorySelection(t *testing.T) {
	p1 := NewNotificationProvider("log")
	if _, ok := p1.(*LogOnlyProvider); !ok {
		t.Errorf("expected *LogOnlyProvider, got %T", p1)
	}

	p2 := NewNotificationProvider("resend")
	if _, ok := p2.(*ResendProvider); !ok {
		t.Errorf("expected *ResendProvider, got %T", p2)
	}
}
