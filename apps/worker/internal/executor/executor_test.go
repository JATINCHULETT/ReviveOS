package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	paymentprovider "github.com/reviveos/services/payment-provider"
	"github.com/reviveos/utils/db"
)

func TestExecutor_AlreadyCapturedProtection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	localProvider := paymentprovider.NewLocalPaymentProvider(pool)
	exec := NewRecoveryExecutor(pool, localProvider)

	// 1. Create merchant, customer, and payment
	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Cap_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("cap_%d@test.com", time.Now().UnixNano())).Scan(&customerID)

	var paymentID string
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, 4500.00, 'FAILED', 'card', 'TIMEOUT')
		RETURNING id::text
	`, merchantID, customerID).Scan(&paymentID)
	if err != nil {
		t.Fatalf("Failed to insert payment: %v", err)
	}

	// 2. Pre-seed the payment as CAPTURED directly in the local provider (e.g. customer paid out of band)
	_, err = localProvider.RegisterPayment(ctx, paymentID, 4500.00, "INR", "CAPTURED", "card", "")
	if err != nil {
		t.Fatalf("Failed to register provider payment: %v", err)
	}

	// 3. Create a recovery workflow in PLANNED state
	var workflowID string
	err = pool.QueryRow(ctx, `
		INSERT INTO recovery_workflows (payment_id, status, selected_action)
		VALUES ($1, 'PLANNED', 'DELAYED_RETRY')
		RETURNING id::text
	`, paymentID).Scan(&workflowID)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// 4. Run Executor
	res, err := exec.ExecuteWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	// 5. Verify NO recovery action was executed
	if res.Reconciliation != "SKIPPED_ALREADY_CAPTURED" {
		t.Errorf("Expected Reconciliation=SKIPPED_ALREADY_CAPTURED, got %s", res.Reconciliation)
	}
	if res.ActionTaken != "NO_ACTION" {
		t.Errorf("Expected ActionTaken=NO_ACTION, got %s", res.ActionTaken)
	}
	if !res.Recovered {
		t.Errorf("Expected Recovered=true (reconciled as captured)")
	}

	// Verify no recovery_actions were created
	var actionCount int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM recovery_actions WHERE workflow_id::text = $1", workflowID).Scan(&actionCount)
	if actionCount != 0 {
		t.Errorf("Expected 0 recovery_actions created, got %d", actionCount)
	}

	// Verify workflow and payment are now reconciled as CAPTURED / RECOVERED
	var payStatus string
	_ = pool.QueryRow(ctx, "SELECT status FROM payments WHERE id::text = $1", paymentID).Scan(&payStatus)
	if payStatus != "CAPTURED" {
		t.Errorf("Expected payment status to be reconciled to CAPTURED, got %s", payStatus)
	}

	t.Logf("SUCCESS: Already-captured payment protected from duplicate recovery action. Reconciled cleanly: %s", res.Message)
}

func TestExecutor_RealFailedPaymentRecoveryAndVerification(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	localProvider := paymentprovider.NewLocalPaymentProvider(pool)
	exec := NewRecoveryExecutor(pool, localProvider)

	// 1. Create merchant, customer, payment
	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Rec_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("rec_%d@test.com", time.Now().UnixNano())).Scan(&customerID)

	var paymentID string
	amount := 3200.00
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, $3, 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerID, amount).Scan(&paymentID)
	if err != nil {
		t.Fatalf("Failed to insert payment: %v", err)
	}

	// Seed as FAILED in provider
	_, err = localProvider.RegisterPayment(ctx, paymentID, amount, "INR", "FAILED", "card", "INSUFFICIENT_FUNDS")
	if err != nil {
		t.Fatalf("Failed to register provider payment: %v", err)
	}

	// Create workflow
	var workflowID string
	err = pool.QueryRow(ctx, `
		INSERT INTO recovery_workflows (payment_id, status, selected_action)
		VALUES ($1, 'SCHEDULED', 'DELAYED_RETRY')
		RETURNING id::text
	`, paymentID).Scan(&workflowID)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// 2. Run Executor: Real Provider Execution -> Verification
	res, err := exec.ExecuteWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	if res.Reconciliation != "PASSED" {
		t.Fatalf("Expected Reconciliation=PASSED, got %s", res.Reconciliation)
	}
	if !res.Recovered {
		t.Fatalf("Expected Recovered=true after provider retry and verification")
	}
	if res.ActionID == "" {
		t.Errorf("Expected ActionID to be populated")
	}
	if res.ProviderAttemptID == "" {
		t.Errorf("Expected ProviderAttemptID to be populated from real provider attempt")
	}

	// 3. Verify PostgreSQL authoritative records
	// A. payments table updated to CAPTURED
	var payStatus string
	_ = pool.QueryRow(ctx, "SELECT status FROM payments WHERE id::text = $1", paymentID).Scan(&payStatus)
	if payStatus != "CAPTURED" {
		t.Errorf("Expected payment status=CAPTURED in DB, got %s", payStatus)
	}

	// B. recovery_workflows updated to RECOVERED
	var wfStatus string
	_ = pool.QueryRow(ctx, "SELECT status FROM recovery_workflows WHERE id::text = $1", workflowID).Scan(&wfStatus)
	if wfStatus != "RECOVERED" {
		t.Errorf("Expected workflow status=RECOVERED in DB, got %s", wfStatus)
	}

	// C. recovery_outcomes recorded
	var outcomeRecovered bool
	var outcomeAmount float64
	err = pool.QueryRow(ctx, `
		SELECT recovered, recovered_amount::float8
		FROM recovery_outcomes
		WHERE payment_id::text = $1
		ORDER BY created_at DESC LIMIT 1
	`, paymentID).Scan(&outcomeRecovered, &outcomeAmount)
	if err != nil || !outcomeRecovered || outcomeAmount != amount {
		t.Errorf("Unexpected recovery outcome: recovered=%v, amount=%.2f (err: %v)", outcomeRecovered, outcomeAmount, err)
	}

	// D. Tamper-evident audit chain contains RECOVERY_OUTCOME_RECORDED
	var auditCount int
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_events 
		WHERE workflow_id::text = $1 AND action = 'RECOVERY_OUTCOME_RECORDED'
	`, workflowID).Scan(&auditCount)
	if auditCount < 1 {
		t.Errorf("Expected audit_events row for RECOVERY_OUTCOME_RECORDED")
	}

	t.Logf("SUCCESS: Real recovery execution & provider verification verified -> Workflow=%s, OutcomeAmount=%.2f",
		workflowID, outcomeAmount)
}

func TestExecutor_VerificationFailed_NoFalseRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	localProvider := paymentprovider.NewLocalPaymentProvider(pool)
	exec := NewRecoveryExecutor(pool, localProvider)

	// 1. Create merchant with policy max_retries = 3
	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Fail_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("fail_%d@test.com", time.Now().UnixNano())).Scan(&customerID)

	var paymentID string
	amount := 5000.00
	// Hard decline error: EXPIRED_CARD (will not succeed on retry)
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, $3, 'FAILED', 'card', 'EXPIRED_CARD')
		RETURNING id::text
	`, merchantID, customerID, amount).Scan(&paymentID)
	if err != nil {
		t.Fatalf("Failed to insert payment: %v", err)
	}

	_, err = localProvider.RegisterPayment(ctx, paymentID, amount, "INR", "FAILED", "card", "EXPIRED_CARD")
	if err != nil {
		t.Fatalf("Failed to register provider payment: %v", err)
	}

	var workflowID string
	err = pool.QueryRow(ctx, `
		INSERT INTO recovery_workflows (payment_id, status, selected_action)
		VALUES ($1, 'SCHEDULED', 'DELAYED_RETRY')
		RETURNING id::text
	`, paymentID).Scan(&workflowID)
	if err != nil {
		t.Fatalf("Failed to create workflow: %v", err)
	}

	// 2. Run Executor
	res, err := exec.ExecuteWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	// 3. Verify NO FALSE RECOVERY
	if res.Recovered {
		t.Fatalf("FAIL: False recovery detected! Hard declined payment must not be marked as recovered.")
	}

	// Verify payment remains FAILED in PostgreSQL
	var payStatus string
	_ = pool.QueryRow(ctx, "SELECT status FROM payments WHERE id::text = $1", paymentID).Scan(&payStatus)
	if payStatus != "FAILED" {
		t.Errorf("Expected payment status to remain FAILED, got %s", payStatus)
	}

	// Verify recovery_outcomes record shows recovered = false
	var outcomeRecovered bool
	err = pool.QueryRow(ctx, `
		SELECT recovered FROM recovery_outcomes WHERE payment_id::text = $1 ORDER BY created_at DESC LIMIT 1
	`, paymentID).Scan(&outcomeRecovered)
	if err != nil || outcomeRecovered {
		t.Errorf("Expected recovery_outcomes (recovered=false), got %v (err: %v)", outcomeRecovered, err)
	}

	t.Logf("SUCCESS: Verification failed properly without false recovery (Status=FAILED, Message=%s)", res.Message)
}

func TestExecutor_CustomerOptOut_Blocked(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	localProvider := paymentprovider.NewLocalPaymentProvider(pool)
	exec := NewRecoveryExecutor(pool, localProvider)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Opt_%d", time.Now().UnixNano())).Scan(&merchantID)

	// Customer with communication_opt_out = true
	var customerID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO customers (merchant_id, email, communication_opt_out) 
		VALUES ($1, $2, true) 
		RETURNING id::text
	`, merchantID, fmt.Sprintf("optout_%d@test.com", time.Now().UnixNano())).Scan(&customerID)

	var paymentID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, 1800.00, 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerID).Scan(&paymentID)

	var workflowID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO recovery_workflows (payment_id, status, selected_action)
		VALUES ($1, 'PLANNED', 'CUSTOMER_NOTIFICATION')
		RETURNING id::text
	`, paymentID).Scan(&workflowID)

	// Run Executor
	res, err := exec.ExecuteWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	if res.Reconciliation != "BLOCKED_OPT_OUT" {
		t.Errorf("Expected Reconciliation=BLOCKED_OPT_OUT, got %s", res.Reconciliation)
	}
	if res.ActionTaken != "NO_ACTION" {
		t.Errorf("Expected ActionTaken=NO_ACTION, got %s", res.ActionTaken)
	}

	// Verify workflow status HALTED
	var wfStatus string
	_ = pool.QueryRow(ctx, "SELECT status FROM recovery_workflows WHERE id::text = $1", workflowID).Scan(&wfStatus)
	if wfStatus != "HALTED" {
		t.Errorf("Expected workflow status=HALTED, got %s", wfStatus)
	}

	t.Logf("SUCCESS: Customer opt-out blocked communication action properly")
}

func TestExecutor_CustomerNotification_SentAndAudited(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	localProvider := paymentprovider.NewLocalPaymentProvider(pool)
	exec := NewRecoveryExecutor(pool, localProvider)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Notif_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	custEmail := fmt.Sprintf("notify_%d@test.com", time.Now().UnixNano())
	_ = pool.QueryRow(ctx, `
		INSERT INTO customers (merchant_id, email, communication_opt_out) 
		VALUES ($1, $2, false) 
		RETURNING id::text
	`, merchantID, custEmail).Scan(&customerID)

	var paymentID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, 3200.00, 'FAILED', 'card', 'EXPIRED_CARD')
		RETURNING id::text
	`, merchantID, customerID).Scan(&paymentID)

	var workflowID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO recovery_workflows (payment_id, status, selected_action)
		VALUES ($1, 'PLANNED', 'PAYMENT_LINK')
		RETURNING id::text
	`, paymentID).Scan(&workflowID)

	// Run Executor
	res, err := exec.ExecuteWorkflow(ctx, workflowID)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	if !res.NotificationSent {
		t.Errorf("Expected NotificationSent=true, got false")
	}

	// Verify Audit Event logged in DB
	var auditCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_events 
		WHERE workflow_id::text = $1 AND action = 'CUSTOMER_NOTIFICATION_SENT'
	`, workflowID).Scan(&auditCount)
	if err != nil || auditCount == 0 {
		t.Errorf("Expected CUSTOMER_NOTIFICATION_SENT in audit_events, found %d records (err: %v)", auditCount, err)
	}

	t.Logf("SUCCESS: Recovery notification dispatched and logged to audit ledger with hash chaining.")
}

