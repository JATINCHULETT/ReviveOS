package policy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/reviveos/schemas"
	aiprovider "github.com/reviveos/services/ai-provider"
	"github.com/reviveos/utils/db"
)

func TestPolicyEngine_DBIntegration_Allow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewEngine(pool)

	// Create merchant and custom policy in DB
	var merchantID string
	err = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Pol_%d", time.Now().UnixNano())).Scan(&merchantID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO policies (merchant_id, max_retries, max_contacts, confidence_threshold, amount_threshold)
		VALUES ($1, 4, 3, 0.65, 75000.00)
		ON CONFLICT (merchant_id) DO UPDATE
		SET max_retries = 4, confidence_threshold = 0.65, amount_threshold = 75000.00
	`, merchantID)
	if err != nil {
		t.Fatalf("Failed to insert merchant policy: %v", err)
	}

	// Verify policy loaded from DB
	pol := engine.GetMerchantPolicy(ctx, merchantID)
	if pol.MaxRetries != 4 || pol.ConfidenceThreshold != 0.65 || pol.AmountThreshold != 75000.00 {
		t.Errorf("DB policy values mismatch: %+v", pol)
	}

	// Test normal allowed decision
	event := schemas.PaymentFailureEvent{
		Amount:        5000.00,
		Currency:      "INR",
		FailureCode:   "INSUFFICIENT_FUNDS",
		AttemptNumber: 1,
	}
	aiRec := aiprovider.AIRecommendation{
		RecommendedAction: "DELAYED_RETRY",
		Confidence:        0.85,
	}

	decision, err := engine.Evaluate(ctx, merchantID, event, aiRec)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if decision.Decision != DecisionAllow {
		t.Errorf("Expected DecisionAllow, got %s (Reason: %s)", decision.Decision, decision.Reason)
	}
	t.Logf("SUCCESS: Policy allowed -> Decision: %s, Reason: %s", decision.Decision, decision.Reason)
}

func TestPolicyEngine_RetryLimitBlock(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewEngine(pool)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Retry_%d", time.Now().UnixNano())).Scan(&merchantID)

	_, _ = pool.Exec(ctx, `
		INSERT INTO policies (merchant_id, max_retries, confidence_threshold, amount_threshold)
		VALUES ($1, 2, 0.70, 50000.00)
		ON CONFLICT (merchant_id) DO UPDATE SET max_retries = 2
	`, merchantID)

	// Attempt exceeds max_retries (3 > 2)
	event := schemas.PaymentFailureEvent{
		Amount:        1000.00,
		AttemptNumber: 3,
	}
	aiRec := aiprovider.AIRecommendation{
		RecommendedAction: "DELAYED_RETRY",
		Confidence:        0.90,
	}

	decision, err := engine.Evaluate(ctx, merchantID, event, aiRec)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if decision.Decision != DecisionBlock {
		t.Errorf("Expected DecisionBlock for retry limit exceeded, got %s", decision.Decision)
	}
	t.Logf("SUCCESS: Retry limit blocked -> Decision: %s, Reason: %s", decision.Decision, decision.Reason)
}

func TestPolicyEngine_ConfidenceThresholdEscalate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewEngine(pool)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Conf_%d", time.Now().UnixNano())).Scan(&merchantID)

	_, _ = pool.Exec(ctx, `
		INSERT INTO policies (merchant_id, max_retries, confidence_threshold, amount_threshold)
		VALUES ($1, 3, 0.80, 50000.00)
		ON CONFLICT (merchant_id) DO UPDATE SET confidence_threshold = 0.80
	`, merchantID)

	// AI Confidence 0.60 < Policy Threshold 0.80
	event := schemas.PaymentFailureEvent{Amount: 2000.00, AttemptNumber: 1}
	aiRec := aiprovider.AIRecommendation{
		RecommendedAction: "DELAYED_RETRY",
		Confidence:        0.60,
	}

	decision, err := engine.Evaluate(ctx, merchantID, event, aiRec)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if decision.Decision != DecisionEscalate {
		t.Errorf("Expected DecisionEscalate for low confidence, got %s", decision.Decision)
	}
	t.Logf("SUCCESS: Confidence threshold escalated -> Decision: %s, Reason: %s", decision.Decision, decision.Reason)
}

func TestPolicyEngine_HighValueThresholdEscalate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewEngine(pool)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_High_%d", time.Now().UnixNano())).Scan(&merchantID)

	_, _ = pool.Exec(ctx, `
		INSERT INTO policies (merchant_id, max_retries, confidence_threshold, amount_threshold)
		VALUES ($1, 3, 0.70, 25000.00)
		ON CONFLICT (merchant_id) DO UPDATE SET amount_threshold = 25000.00
	`, merchantID)

	// Amount 50000 > Threshold 25000
	event := schemas.PaymentFailureEvent{Amount: 50000.00, AttemptNumber: 1}
	aiRec := aiprovider.AIRecommendation{
		RecommendedAction: "DELAYED_RETRY",
		Confidence:        0.95,
	}

	decision, err := engine.Evaluate(ctx, merchantID, event, aiRec)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if decision.Decision != DecisionEscalate {
		t.Errorf("Expected DecisionEscalate for high value amount, got %s", decision.Decision)
	}
	t.Logf("SUCCESS: High value escalated -> Decision: %s, Reason: %s", decision.Decision, decision.Reason)
}

func TestPolicyEngine_CustomerOptOutBlock(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewEngine(pool)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Opt_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email, communication_opt_out) VALUES ($1, 'optout@test.com', true) RETURNING id::text", merchantID).Scan(&customerID)

	var paymentID string
	_ = pool.QueryRow(ctx, "INSERT INTO payments (merchant_id, customer_id, amount, status) VALUES ($1, $2, 1500.00, 'FAILED') RETURNING id::text", merchantID, customerID).Scan(&paymentID)

	event := schemas.PaymentFailureEvent{
		PaymentID:     paymentID,
		Amount:        1500.00,
		AttemptNumber: 1,
	}
	aiRec := aiprovider.AIRecommendation{
		RecommendedAction: "CUSTOMER_NOTIFICATION",
		Confidence:        0.90,
	}

	decision, err := engine.Evaluate(ctx, merchantID, event, aiRec)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if decision.Decision != DecisionBlock {
		t.Errorf("Expected DecisionBlock for opted out customer, got %s", decision.Decision)
	}
	t.Logf("SUCCESS: Customer opt-out blocked -> Decision: %s, Reason: %s", decision.Decision, decision.Reason)
}

func TestPolicyEngine_DuplicateActionBlock(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewEngine(pool)

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Dup_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, 'dup@test.com') RETURNING id::text", merchantID).Scan(&customerID)

	var paymentID string
	_ = pool.QueryRow(ctx, "INSERT INTO payments (merchant_id, customer_id, amount, status) VALUES ($1, $2, 3000.00, 'FAILED') RETURNING id::text", merchantID, customerID).Scan(&paymentID)

	var workflowID string
	_ = pool.QueryRow(ctx, "INSERT INTO recovery_workflows (payment_id, status) VALUES ($1, 'SCHEDULED') RETURNING id::text", paymentID).Scan(&workflowID)

	// Pre-insert an executed PAYMENT_LINK action
	_, _ = pool.Exec(ctx, `
		INSERT INTO recovery_actions (workflow_id, action_type, status, attempt)
		VALUES ($1, 'PAYMENT_LINK', 'EXECUTED', 1)
	`, workflowID)

	event := schemas.PaymentFailureEvent{
		PaymentID:     paymentID,
		Amount:        3000.00,
		AttemptNumber: 2,
	}
	aiRec := aiprovider.AIRecommendation{
		RecommendedAction: "PAYMENT_LINK",
		Confidence:        0.90,
	}

	decision, err := engine.Evaluate(ctx, merchantID, event, aiRec)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if decision.Decision != DecisionBlock {
		t.Errorf("Expected DecisionBlock for duplicate PAYMENT_LINK, got %s", decision.Decision)
	}
	t.Logf("SUCCESS: Duplicate action blocked -> Decision: %s, Reason: %s", decision.Decision, decision.Reason)
}
