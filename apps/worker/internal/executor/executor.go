package executor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/packages/recovery"
	"github.com/reviveos/schemas"
	aiprovider "github.com/reviveos/services/ai-provider"
	paymentprovider "github.com/reviveos/services/payment-provider"
	policy "github.com/reviveos/services/policy-engine"
	"github.com/reviveos/utils/audit"
)

// ExecutionResult captures the complete verification lifecycle and outcome of a recovery execution.
type ExecutionResult struct {
	WorkflowID        string                         `json:"workflow_id"`
	PaymentID         string                         `json:"payment_id"`
	InitialStatus     string                         `json:"initial_status"`
	Reconciliation    string                         `json:"reconciliation"` // PASSED, SKIPPED_ALREADY_CAPTURED, SKIPPED_WORKFLOW_RESOLVED, BLOCKED_OPT_OUT, BLOCKED_MAX_RETRIES, BLOCKED_POLICY, ESCALATED_POLICY
	ActionTaken       string                         `json:"action_taken"`
	ActionID          string                         `json:"action_id,omitempty"`
	AttemptNumber     int                            `json:"attempt_number"`
	ProviderAttemptID string                         `json:"provider_attempt_id,omitempty"`
	VerifiedStatus    *paymentprovider.PaymentStatus `json:"verified_status,omitempty"`
	OutcomeRecorded   bool                           `json:"outcome_recorded"`
	Recovered         bool                           `json:"recovered"`
	Message           string                         `json:"message"`
}

// RecoveryExecutor orchestrates safe, reconciled recovery execution against a real PaymentProvider.
type RecoveryExecutor struct {
	pool         *pgxpool.Pool
	provider     paymentprovider.PaymentProvider
	policyEngine *policy.Engine
}

// NewRecoveryExecutor initializes a new RecoveryExecutor instance.
func NewRecoveryExecutor(pool *pgxpool.Pool, provider paymentprovider.PaymentProvider) *RecoveryExecutor {
	if provider == nil {
		var err error
		provider, err = paymentprovider.NewPaymentProvider("", pool)
		if err != nil {
			log.Printf("[Executor] Warning: failed to instantiate payment provider: %v", err)
		}
	}

	return &RecoveryExecutor{
		pool:         pool,
		provider:     provider,
		policyEngine: policy.NewEngine(pool),
	}
}

// ExecuteWorkflow runs the comprehensive pre-execution checks, reconciliation, execution, and verification.
func (e *RecoveryExecutor) ExecuteWorkflow(ctx context.Context, workflowID string) (*ExecutionResult, error) {
	if e.pool == nil {
		return nil, fmt.Errorf("database connection pool is nil")
	}
	if e.provider == nil {
		return nil, fmt.Errorf("payment provider is not configured")
	}

	// 1. Load Workflow, Payment, and Customer Context from PostgreSQL
	var (
		paymentID           string
		merchantID          string
		customerID          string
		workflowStatus      string
		selectedAction      sql.NullString
		paymentAmount       float64
		paymentCurrency     string
		paymentStatus       string
		failureCode         sql.NullString
		communicationOptOut bool
		customerEmail       sql.NullString
		customerPhone       sql.NullString
		workflowCreatedAt   time.Time
	)

	query := `
		SELECT 
			rw.payment_id::text,
			p.merchant_id::text,
			p.customer_id::text,
			rw.status,
			rw.selected_action,
			p.amount::float8,
			p.currency,
			p.status,
			p.failure_code,
			c.communication_opt_out,
			c.email,
			c.phone,
			rw.created_at
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		JOIN customers c ON p.customer_id = c.id
		WHERE rw.id::text = $1
	`

	err := e.pool.QueryRow(ctx, query, workflowID).Scan(
		&paymentID,
		&merchantID,
		&customerID,
		&workflowStatus,
		&selectedAction,
		&paymentAmount,
		&paymentCurrency,
		&paymentStatus,
		&failureCode,
		&communicationOptOut,
		&customerEmail,
		&customerPhone,
		&workflowCreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("recovery workflow %s not found", workflowID)
		}
		return nil, fmt.Errorf("failed to query workflow context: %w", err)
	}

	res := &ExecutionResult{
		WorkflowID:    workflowID,
		PaymentID:     paymentID,
		InitialStatus: paymentStatus,
	}

	// 2. CHECK 1: Workflow State (Ensure workflow is not already resolved)
	if workflowStatus == "RECOVERED" || workflowStatus == "HALTED" || workflowStatus == "FAILED" {
		res.Reconciliation = "SKIPPED_WORKFLOW_RESOLVED"
		res.ActionTaken = "NO_ACTION"
		res.Message = fmt.Sprintf("Workflow %s is already in terminal state: %s", workflowID, workflowStatus)
		log.Printf("[Executor] %s", res.Message)
		return res, nil
	}

	// 3. CHECK 2: Reconciliation via PaymentProvider.GetPayment()
	providerStatus, err := e.provider.GetPayment(ctx, paymentID)
	if err != nil {
		log.Printf("[Executor] Warning: GetPayment returned error for %s: %v", paymentID, err)
	}

	// Check whether payment already succeeded on provider
	if providerStatus != nil && (providerStatus.Status == "CAPTURED" || providerStatus.Captured) {
		res.Reconciliation = "SKIPPED_ALREADY_CAPTURED"
		res.ActionTaken = "NO_ACTION"
		res.Recovered = true
		res.OutcomeRecorded = true
		res.VerifiedStatus = providerStatus
		res.Message = "Payment is already CAPTURED on provider; reconciled database state without executing recovery action."

		// Reconcile PostgreSQL state
		tx, err := e.pool.Begin(ctx)
		if err == nil {
			defer tx.Rollback(ctx)
			_, _ = tx.Exec(ctx, `UPDATE payments SET status = 'CAPTURED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, paymentID)
			_, _ = tx.Exec(ctx, `UPDATE recovery_workflows SET status = 'RECOVERED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)
			_, _ = tx.Exec(ctx, `
				INSERT INTO recovery_outcomes (payment_id, recovered, recovered_amount, time_to_recovery, created_at)
				VALUES ($1, true, $2, $3, CURRENT_TIMESTAMP)
			`, paymentID, paymentAmount, time.Since(workflowCreatedAt).String())
			_ = tx.Commit(ctx)
		}

		_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
			WorkflowID: workflowID,
			Actor:      "executor:reconciliation",
			Action:     "RECONCILIATION_ALREADY_CAPTURED",
			Metadata: map[string]interface{}{
				"provider_status": providerStatus.Status,
				"payment_id":      paymentID,
			},
		})

		log.Printf("[Executor] RECONCILIATION_PASSED: Payment %s was already CAPTURED on provider. Reconciled successfully.", paymentID)
		return res, nil
	}

	// 4. CHECK 3: Customer Opt-Out Check
	actionStr := "DELAYED_RETRY"
	if selectedAction.Valid && selectedAction.String != "" {
		actionStr = selectedAction.String
	}

	if communicationOptOut && (actionStr == "CUSTOMER_NOTIFICATION" || actionStr == "PAYMENT_LINK") {
		res.Reconciliation = "BLOCKED_OPT_OUT"
		res.ActionTaken = "NO_ACTION"
		res.Message = "Customer has opted out of communications; blocking communication action."

		_, _ = e.pool.Exec(ctx, `UPDATE recovery_workflows SET status = 'HALTED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)

		_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
			WorkflowID: workflowID,
			Actor:      "executor:policy",
			Action:     "POLICY_EVALUATED_OPT_OUT",
			Metadata: map[string]interface{}{
				"blocked_action": actionStr,
				"customer_id":    customerID,
			},
		})

		log.Printf("[Executor] BLOCKED_OPT_OUT: Customer %s opted out. Action %s blocked.", customerID, actionStr)
		return res, nil
	}

	// 5. CHECK 4: Retry Limits Check against Policy
	var previousAttemptsCount int
	err = e.pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*), 0)
		FROM recovery_actions
		WHERE workflow_id::text = $1
	`, workflowID).Scan(&previousAttemptsCount)
	if err != nil {
		previousAttemptsCount = 0
	}

	merchantPol := e.policyEngine.GetMerchantPolicy(ctx, merchantID)
	if previousAttemptsCount >= merchantPol.MaxRetries {
		res.Reconciliation = "BLOCKED_MAX_RETRIES"
		res.ActionTaken = "NO_ACTION"
		res.Message = fmt.Sprintf("Maximum retry attempts (%d/%d) reached according to merchant policy.", previousAttemptsCount, merchantPol.MaxRetries)

		_, _ = e.pool.Exec(ctx, `UPDATE recovery_workflows SET status = 'FAILED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)

		_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
			WorkflowID: workflowID,
			Actor:      "executor:policy",
			Action:     "MAX_RETRIES_EXCEEDED",
			Metadata: map[string]interface{}{
				"attempts":    previousAttemptsCount,
				"max_retries": merchantPol.MaxRetries,
			},
		})

		log.Printf("[Executor] BLOCKED_MAX_RETRIES: Workflow %s exceeded retry limit %d.", workflowID, merchantPol.MaxRetries)
		return res, nil
	}

	// 6. CHECK 5: Policy Engine Evaluation
	polDecision, err := e.policyEngine.Evaluate(ctx, merchantID, schemas.PaymentFailureEvent{
		PaymentID:   paymentID,
		Amount:      paymentAmount,
		Currency:    paymentCurrency,
		FailureCode: failureCode.String,
	}, aiprovider.AIRecommendation{
		RecommendedAction: recovery.ActionType(actionStr),
	})

	if err == nil && polDecision.Decision == policy.DecisionBlock {
		res.Reconciliation = "BLOCKED_POLICY"
		res.ActionTaken = "NO_ACTION"
		res.Message = fmt.Sprintf("Policy engine blocked execution: %s", polDecision.Reason)

		_, _ = e.pool.Exec(ctx, `UPDATE recovery_workflows SET status = 'HALTED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)

		_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
			WorkflowID: workflowID,
			Actor:      "executor:policy",
			Action:     "POLICY_EVALUATED_BLOCK",
			Metadata: map[string]interface{}{
				"reason": polDecision.Reason,
			},
		})
		return res, nil
	}

	// 7. ALL CHECKS PASSED -> EXECUTE RECOVERY ACTION
	res.Reconciliation = "PASSED"
	currentAttemptNumber := previousAttemptsCount + 1
	res.AttemptNumber = currentAttemptNumber
	res.ActionTaken = actionStr

	log.Printf("[Executor] RECONCILIATION_PASSED: Executing recovery action '%s' (Attempt %d) for Payment %s",
		actionStr, currentAttemptNumber, paymentID)

	// Record action in recovery_actions as PENDING
	var actionUUID string
	err = e.pool.QueryRow(ctx, `
		INSERT INTO recovery_actions (workflow_id, action_type, status, attempt, executed_at)
		VALUES ($1, $2, 'PENDING', $3, CURRENT_TIMESTAMP)
		RETURNING id::text
	`, workflowID, actionStr, currentAttemptNumber).Scan(&actionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to record recovery_action: %w", err)
	}
	res.ActionID = actionUUID

	_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
		WorkflowID: workflowID,
		Actor:      "executor:action",
		Action:     "RECOVERY_EXECUTED",
		Metadata: map[string]interface{}{
			"action_id":  actionUUID,
			"action":     actionStr,
			"attempt":    currentAttemptNumber,
			"payment_id": paymentID,
		},
	})

	// Call PaymentProvider.CreateRetryAttempt() for retry actions
	retryResult, retryErr := e.provider.CreateRetryAttempt(ctx, paymentID, paymentAmount)
	if retryResult != nil {
		res.ProviderAttemptID = retryResult.AttemptID
	}

	// Update recovery_actions with execution result
	actionStatus := "EXECUTED"
	var actionResultStr string
	if retryErr != nil || (retryResult != nil && retryResult.Status == "FAILED") {
		actionStatus = "FAILED"
		if retryResult != nil && retryResult.ErrorMessage != "" {
			actionResultStr = retryResult.ErrorMessage
		} else if retryErr != nil {
			actionResultStr = retryErr.Error()
		}
	} else if retryResult != nil {
		actionResultStr = retryResult.Status
	}

	_, _ = e.pool.Exec(ctx, `
		UPDATE recovery_actions
		SET status = $1, result = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id::text = $3
	`, actionStatus, actionResultStr, actionUUID)

	// 8. VERIFICATION via PaymentProvider.VerifyPayment()
	verifiedStatus, verifyErr := e.provider.VerifyPayment(ctx, paymentID)
	if verifyErr != nil {
		log.Printf("[Executor] Warning: VerifyPayment returned error: %v", verifyErr)
	}
	res.VerifiedStatus = verifiedStatus

	if verifiedStatus != nil && (verifiedStatus.Status == "CAPTURED" || verifiedStatus.Captured) {
		// GENUINE RECOVERY VERIFIED
		res.Recovered = true
		res.OutcomeRecorded = true
		res.Message = fmt.Sprintf("Recovery executed and VERIFIED as CAPTURED by provider (Attempt %d)", currentAttemptNumber)

		// Authoritative update of payments and recovery_workflows ONLY after verified by provider!
		tx, err := e.pool.Begin(ctx)
		if err == nil {
			defer tx.Rollback(ctx)
			_, _ = tx.Exec(ctx, `UPDATE payments SET status = 'CAPTURED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, paymentID)
			_, _ = tx.Exec(ctx, `UPDATE recovery_workflows SET status = 'RECOVERED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)
			_, _ = tx.Exec(ctx, `
				INSERT INTO recovery_outcomes (action_id, payment_id, recovered, recovered_amount, time_to_recovery, created_at)
				VALUES ($1, $2, true, $3, $4, CURRENT_TIMESTAMP)
			`, actionUUID, paymentID, paymentAmount, time.Since(workflowCreatedAt).String())
			_ = tx.Commit(ctx)
		}

		_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
			WorkflowID: workflowID,
			Actor:      "executor:verifier",
			Action:     "RECOVERY_OUTCOME_RECORDED",
			Metadata: map[string]interface{}{
				"verified":   true,
				"status":     "CAPTURED",
				"action_id":  actionUUID,
				"payment_id": paymentID,
			},
		})

		log.Printf("[Executor] RECOVERY_VERIFIED: Payment %s recovered (CAPTURED).", paymentID)
		return res, nil
	}

	// 9. VERIFICATION FAILED / NOT CAPTURED -> PREVENT FALSE RECOVERY
	res.Recovered = false
	res.OutcomeRecorded = true

	// Check if attempts exhausted
	if currentAttemptNumber >= merchantPol.MaxRetries {
		res.Message = fmt.Sprintf("Recovery attempt verified as FAILED. Max retries (%d/%d) exhausted; workflow marked FAILED.", currentAttemptNumber, merchantPol.MaxRetries)
		_, _ = e.pool.Exec(ctx, `UPDATE recovery_workflows SET status = 'FAILED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)
	} else {
		res.Message = fmt.Sprintf("Recovery attempt (%d/%d) verified as FAILED. Remaining attempts available; workflow SCHEDULED for next attempt.", currentAttemptNumber, merchantPol.MaxRetries)
		_, _ = e.pool.Exec(ctx, `UPDATE recovery_workflows SET status = 'SCHEDULED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1`, workflowID)
	}

	// Record failed outcome in recovery_outcomes
	_, _ = e.pool.Exec(ctx, `
		INSERT INTO recovery_outcomes (action_id, payment_id, recovered, recovered_amount, created_at)
		VALUES ($1, $2, false, 0, CURRENT_TIMESTAMP)
	`, actionUUID, paymentID)

	_ = audit.AppendAuditLog(ctx, e.pool, audit.AuditEvent{
		WorkflowID: workflowID,
		Actor:      "executor:verifier",
		Action:     "RECOVERY_OUTCOME_RECORDED",
		Metadata: map[string]interface{}{
			"verified":   true,
			"status":     "FAILED",
			"action_id":  actionUUID,
			"payment_id": paymentID,
			"attempts":   currentAttemptNumber,
		},
	})

	log.Printf("[Executor] RECOVERY_VERIFIED_FAILED: Payment %s remained FAILED. No false recovery recorded.", paymentID)
	return res, nil
}
