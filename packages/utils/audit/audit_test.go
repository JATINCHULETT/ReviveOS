package audit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
)

func TestOutcomeTracker_AuditChainIntegrity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	tracker := NewOutcomeTracker(pool)

	// 1. Create a merchant, customer, payment, and workflow in DB
	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("M_Aud_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("aud_%d@test.com", time.Now().UnixNano())).Scan(&customerID)

	var paymentID string
	_ = pool.QueryRow(ctx, "INSERT INTO payments (merchant_id, customer_id, amount, status) VALUES ($1, $2, 3500.00, 'FAILED') RETURNING id::text", merchantID, customerID).Scan(&paymentID)

	var workflowID string
	_ = pool.QueryRow(ctx, "INSERT INTO recovery_workflows (payment_id, status) VALUES ($1, 'ANALYZING') RETURNING id::text", paymentID).Scan(&workflowID)

	// 2. Append real lifecycle audit events
	lifecycleEvents := []AuditEvent{
		{
			WorkflowID: workflowID,
			Actor:      "system:webhook",
			Action:     "WEBHOOK_RECEIVED",
			Metadata:   map[string]interface{}{"payment_id": paymentID, "event_id": "evt_test_1"},
		},
		{
			WorkflowID: workflowID,
			Actor:      "worker:classifier",
			Action:     "FAILURE_CLASSIFIED",
			Metadata:   map[string]interface{}{"category": "INSUFFICIENT_FUNDS", "code": "INSUFFICIENT_FUNDS"},
		},
		{
			WorkflowID: workflowID,
			Actor:      "worker:ai",
			Action:     "AI_DECISION_CREATED",
			Metadata:   map[string]interface{}{"model": "deepseek-r1:1.5b", "action": "DELAYED_RETRY", "confidence": 0.85},
		},
		{
			WorkflowID: workflowID,
			Actor:      "worker:policy",
			Action:     "POLICY_EVALUATED",
			Metadata:   map[string]interface{}{"decision": "ALLOW", "reason": "Passed policy checks"},
		},
		{
			WorkflowID: workflowID,
			Actor:      "worker:executor",
			Action:     "RECOVERY_EXECUTED",
			Metadata:   map[string]interface{}{"attempt": 1, "action": "DELAYED_RETRY"},
		},
		{
			WorkflowID: workflowID,
			Actor:      "worker:verifier",
			Action:     "RECOVERY_OUTCOME_RECORDED",
			Metadata:   map[string]interface{}{"verified": true, "status": "CAPTURED", "recovered_amount": 3500.00},
		},
	}

	for _, ev := range lifecycleEvents {
		if err := tracker.AppendAuditLog(ctx, ev); err != nil {
			t.Fatalf("Failed to append audit log for action %s: %v", ev.Action, err)
		}
	}

	// 3. Cryptographically verify the complete audit hash chain
	valid, err := tracker.VerifyAuditChain(ctx, workflowID)
	if err != nil {
		t.Fatalf("VerifyAuditChain failed: %v", err)
	}
	if !valid {
		t.Fatalf("Expected audit chain to be valid, but returned false")
	}

	t.Logf("SUCCESS: Audit hash chain verified across %d lifecycle events for workflow %s", len(lifecycleEvents), workflowID)

	// 4. Test Tamper Detection: Tampering with a database row should break the chain
	var lastEventID string
	_ = pool.QueryRow(ctx, `
		SELECT event_id::text FROM audit_events 
		WHERE workflow_id::text = $1 
		ORDER BY timestamp DESC LIMIT 1
	`, workflowID).Scan(&lastEventID)

	// Modify payload_hash fraudulently in DB
	_, err = pool.Exec(ctx, `
		UPDATE audit_events 
		SET payload_hash = 'tampered_fake_hash_0000000000000000000000000000000000000000'
		WHERE event_id::text = $1
	`, lastEventID)
	if err != nil {
		t.Fatalf("Failed to mutate test record: %v", err)
	}

	// VerifyAuditChain must now detect the tampering and fail
	tamperValid, tamperErr := tracker.VerifyAuditChain(ctx, workflowID)
	if tamperValid || tamperErr == nil {
		t.Fatalf("Tampering was NOT detected by VerifyAuditChain (valid=%v, err=%v)", tamperValid, tamperErr)
	}

	t.Logf("SUCCESS: Tamper-evident ledger successfully caught forged hash: %v", tamperErr)
}
