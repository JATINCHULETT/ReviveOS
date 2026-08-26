package paymentprovider

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
)

// Ensure LocalPaymentProvider satisfies PaymentProvider interface at compile-time
var _ PaymentProvider = (*LocalPaymentProvider)(nil)

func TestLocalPaymentProvider_EndToEndFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	provider := NewLocalPaymentProvider(pool)

	// 1. Create a merchant and a failed payment in application database
	var merchantID string
	err = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("Prov_M_%d", time.Now().UnixNano())).Scan(&merchantID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	var customerID string
	err = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("cust_%d@test.com", time.Now().UnixNano())).Scan(&customerID)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	var paymentID string
	initialAmount := 2999.00
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method, failure_code)
		VALUES ($1, $2, $3, 'INR', 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerID, initialAmount).Scan(&paymentID)
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	// 2. GetPayment (Tests initialization from application table & reading provider state)
	status, err := provider.GetPayment(ctx, paymentID)
	if err != nil {
		t.Fatalf("GetPayment failed: %v", err)
	}

	if status.PaymentID != paymentID {
		t.Errorf("Expected PaymentID %s, got %s", paymentID, status.PaymentID)
	}
	if status.Status != "FAILED" {
		t.Errorf("Expected status FAILED, got %s", status.Status)
	}
	if status.Amount != initialAmount {
		t.Errorf("Expected amount %f, got %f", initialAmount, status.Amount)
	}
	if status.Captured {
		t.Errorf("Expected Captured to be false for FAILED payment")
	}
	if status.ProviderPaymentID == "" {
		t.Errorf("Expected non-empty ProviderPaymentID")
	}
	t.Logf("GetPayment initial status verified: ID=%s, Status=%s, ProvID=%s", status.PaymentID, status.Status, status.ProviderPaymentID)

	// 3. CreateRetryAttempt (Executes real provider-side retry)
	retryResult, err := provider.CreateRetryAttempt(ctx, paymentID, initialAmount)
	if err != nil {
		t.Fatalf("CreateRetryAttempt failed: %v", err)
	}

	if retryResult.Status != "SUCCESS" {
		t.Errorf("Expected retry status SUCCESS for retryable INSUFFICIENT_FUNDS, got %s (Error: %s)", retryResult.Status, retryResult.ErrorMessage)
	}
	if retryResult.AttemptID == "" {
		t.Errorf("Expected non-empty AttemptID")
	}
	t.Logf("CreateRetryAttempt result: AttemptID=%s, Status=%s, Amount=%.2f", retryResult.AttemptID, retryResult.Status, retryResult.Amount)

	// 4. VerifyPayment (Asks provider for authoritative verified state)
	verifiedStatus, err := provider.VerifyPayment(ctx, paymentID)
	if err != nil {
		t.Fatalf("VerifyPayment failed: %v", err)
	}

	if verifiedStatus.Status != "CAPTURED" {
		t.Errorf("Expected verified provider status to be CAPTURED after successful retry, got %s", verifiedStatus.Status)
	}
	if !verifiedStatus.Captured {
		t.Errorf("Expected verified status.Captured to be true")
	}
	t.Logf("VerifyPayment verified provider state: ID=%s, Status=%s, Captured=%v", verifiedStatus.PaymentID, verifiedStatus.Status, verifiedStatus.Captured)

	// 5. Inspect database tables directly
	// A. Verify local_provider_payments
	var (
		dbStatus        string
		dbAttemptsCount int
	)
	err = pool.QueryRow(ctx, "SELECT status, attempts_count FROM local_provider_payments WHERE payment_id = $1", paymentID).Scan(&dbStatus, &dbAttemptsCount)
	if err != nil {
		t.Fatalf("Failed to query local_provider_payments: %v", err)
	}
	if dbStatus != "CAPTURED" || dbAttemptsCount != 1 {
		t.Errorf("Expected local_provider_payments (status=CAPTURED, attempts=1), got (%s, %d)", dbStatus, dbAttemptsCount)
	}

	// B. Verify local_provider_attempts
	var attemptCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM local_provider_attempts WHERE payment_id = $1 AND status = 'SUCCESS'", paymentID).Scan(&attemptCount)
	if err != nil || attemptCount != 1 {
		t.Errorf("Expected 1 successful attempt in local_provider_attempts, got %d (err: %v)", attemptCount, err)
	}

	// C. Verify application payments table was NOT modified by provider (Provider separation rule)
	var appPaymentStatus string
	err = pool.QueryRow(ctx, "SELECT status FROM payments WHERE id::text = $1", paymentID).Scan(&appPaymentStatus)
	if err != nil {
		t.Fatalf("Failed to query payments table: %v", err)
	}
	if appPaymentStatus != "FAILED" {
		t.Errorf("Expected application payments table to remain FAILED until workflow reconciler updates it, got %s", appPaymentStatus)
	}
	t.Logf("SUCCESS: Application payments table remained untouched (status=%s) while provider state transitioned to CAPTURED", appPaymentStatus)
}

func TestLocalPaymentProvider_HardDecline_NoFakeSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	provider := NewLocalPaymentProvider(pool)

	testPaymentID := fmt.Sprintf("hard_decline_%d", time.Now().UnixNano())

	// Explicitly register a hard declined payment in provider (EXPIRED_CARD)
	_, err = provider.RegisterPayment(ctx, testPaymentID, 4500.00, "INR", "FAILED", "card", "EXPIRED_CARD")
	if err != nil {
		t.Fatalf("RegisterPayment failed: %v", err)
	}

	// Retry should NOT fake success for EXPIRED_CARD
	result, err := provider.CreateRetryAttempt(ctx, testPaymentID, 4500.00)
	if err != nil {
		t.Fatalf("CreateRetryAttempt returned unexpected error: %v", err)
	}

	if result.Status == "SUCCESS" {
		t.Fatalf("FAIL: Provider faked success for EXPIRED_CARD hard decline!")
	}

	if result.Status != "FAILED" {
		t.Errorf("Expected retry status FAILED, got %s", result.Status)
	}

	// Verify status remains FAILED
	verified, err := provider.VerifyPayment(ctx, testPaymentID)
	if err != nil {
		t.Fatalf("VerifyPayment failed: %v", err)
	}

	if verified.Status != "FAILED" || verified.Captured {
		t.Errorf("Expected provider state to remain FAILED with Captured=false, got status=%s, captured=%v", verified.Status, verified.Captured)
	}

	t.Logf("SUCCESS: Hard decline handled accurately without fake success (Status=%s, Reason=%s)",
		verified.Status, result.ErrorMessage)
}

func TestPaymentProvider_FactoryConfiguration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test default / empty
	prov1, err := NewPaymentProvider("", pool)
	if err != nil {
		t.Fatalf("Expected default provider to be local: %v", err)
	}
	if _, ok := prov1.(*LocalPaymentProvider); !ok {
		t.Fatalf("Expected *LocalPaymentProvider instance, got %T", prov1)
	}

	// Test explicit "local"
	prov2, err := NewPaymentProvider("local", pool)
	if err != nil || prov2 == nil {
		t.Fatalf("Expected local provider creation: %v", err)
	}

	// Test environment variable PAYMENT_PROVIDER=local
	os.Setenv("PAYMENT_PROVIDER", "local")
	prov3, err := NewPaymentProvider("", pool)
	if err != nil || prov3 == nil {
		t.Fatalf("Expected local provider from env var: %v", err)
	}

	// Test unknown provider
	_, err = NewPaymentProvider("stripe", pool)
	if err == nil {
		t.Fatalf("Expected error for unknown provider type")
	}

	t.Logf("SUCCESS: PaymentProvider factory configured and verified")
}
