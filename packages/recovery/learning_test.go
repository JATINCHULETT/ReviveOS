package recovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
)

func TestLearningEngine_ComputeRealOutcomeStatistics(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	engine := NewLearningEngine(pool)

	// 1. Create a merchant
	var merchantID string
	err = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Learn_%d", time.Now().UnixNano())).Scan(&merchantID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	var customerID string
	err = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("learn_%d@test.com", time.Now().UnixNano())).Scan(&customerID)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	// 2. Insert 3 Recovered Payments (INSUFFICIENT_FUNDS, ₹2000 each)
	for i := 0; i < 3; i++ {
		var payID, wfID, actID string
		_ = pool.QueryRow(ctx, `
			INSERT INTO payments (merchant_id, customer_id, amount, status, failure_code)
			VALUES ($1, $2, 2000.00, 'CAPTURED', 'INSUFFICIENT_FUNDS')
			RETURNING id::text
		`, merchantID, customerID).Scan(&payID)

		_ = pool.QueryRow(ctx, `
			INSERT INTO recovery_workflows (payment_id, status, selected_action)
			VALUES ($1, 'RECOVERED', 'DELAYED_RETRY')
			RETURNING id::text
		`, payID).Scan(&wfID)

		_ = pool.QueryRow(ctx, `
			INSERT INTO recovery_actions (workflow_id, action_type, status, attempt)
			VALUES ($1, 'DELAYED_RETRY', 'EXECUTED', 1)
			RETURNING id::text
		`, wfID).Scan(&actID)

		_, _ = pool.Exec(ctx, `
			INSERT INTO recovery_outcomes (action_id, payment_id, recovered, recovered_amount)
			VALUES ($1, $2, true, 2000.00)
		`, actID, payID)
	}

	// 3. Insert 1 Failed Workflow (INSUFFICIENT_FUNDS, ₹1000)
	var failPayID, failWfID, failActID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, failure_code)
		VALUES ($1, $2, 1000.00, 'FAILED', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerID).Scan(&failPayID)

	_ = pool.QueryRow(ctx, `
		INSERT INTO recovery_workflows (payment_id, status, selected_action)
		VALUES ($1, 'FAILED', 'DELAYED_RETRY')
		RETURNING id::text
	`, failPayID).Scan(&failWfID)

	_ = pool.QueryRow(ctx, `
		INSERT INTO recovery_actions (workflow_id, action_type, status, attempt)
		VALUES ($1, 'DELAYED_RETRY', 'FAILED', 1)
		RETURNING id::text
	`, failWfID).Scan(&failActID)

	_, _ = pool.Exec(ctx, `
		INSERT INTO recovery_outcomes (action_id, payment_id, recovered, recovered_amount)
		VALUES ($1, $2, false, 0)
	`, failActID, failPayID)

	// 4. Compute Learning Stats for this merchant from PostgreSQL
	stats, err := engine.ComputeLearningStats(ctx, merchantID)
	if err != nil {
		t.Fatalf("ComputeLearningStats failed: %v", err)
	}

	if stats.TotalWorkflows != 4 {
		t.Errorf("Expected 4 total workflows, got %d", stats.TotalWorkflows)
	}
	if stats.RecoveredWorkflows != 3 {
		t.Errorf("Expected 3 recovered workflows, got %d", stats.RecoveredWorkflows)
	}
	if stats.FailedWorkflows != 1 {
		t.Errorf("Expected 1 failed workflow, got %d", stats.FailedWorkflows)
	}
	if stats.TotalRecoveredAmount != 6000.00 {
		t.Errorf("Expected total recovered amount 6000.00, got %.2f", stats.TotalRecoveredAmount)
	}
	if stats.OverallRecoveryRate != 0.75 {
		t.Errorf("Expected overall recovery rate 0.75 (3/4), got %.4f", stats.OverallRecoveryRate)
	}

	// Verify Category Metric for INSUFFICIENT_FUNDS
	catMetric, ok := stats.CategoryStats["INSUFFICIENT_FUNDS"]
	if !ok {
		t.Fatalf("Expected category stats for INSUFFICIENT_FUNDS")
	}
	if catMetric.TotalCount != 4 || catMetric.Recovered != 3 || catMetric.RecoveryRate != 0.75 {
		t.Errorf("Unexpected category metric: %+v", catMetric)
	}

	t.Logf("SUCCESS: Real PostgreSQL learning stats verified -> Total=%d, Rec=%d, Rate=%.2f, Amount=%.2f",
		stats.TotalWorkflows, stats.RecoveredWorkflows, stats.OverallRecoveryRate, stats.TotalRecoveredAmount)
}
