package policy

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/schemas"
	"github.com/reviveos/services/ai-provider"
)

type DecisionType string

const (
	DecisionAllow    DecisionType = "ALLOW"
	DecisionModify   DecisionType = "MODIFY"
	DecisionBlock    DecisionType = "BLOCK"
	DecisionEscalate DecisionType = "ESCALATE"
)

type PolicyDecision struct {
	Decision       DecisionType
	Reason         string
	ModifiedAction aiprovider.AIRecommendation // Populated if Decision == MODIFY
}

type Engine struct {
	db *pgxpool.Pool
}

func NewEngine(db *pgxpool.Pool) *Engine {
	return &Engine{db: db}
}

// MerchantPolicy holds configuration thresholds from the policies table
type MerchantPolicy struct {
	MaxRetries          int
	MaxContacts         int
	ConfidenceThreshold float64
	AmountThreshold     float64
}

// GetMerchantPolicy loads policy from DB or returns safe defaults
func (e *Engine) GetMerchantPolicy(ctx context.Context, merchantID string) MerchantPolicy {
	// Safe defaults
	policy := MerchantPolicy{
		MaxRetries:          3,
		MaxContacts:         2,
		ConfidenceThreshold: 0.70,
		AmountThreshold:     50000.0,
	}

	if e.db == nil || merchantID == "" {
		return policy
	}

	err := e.db.QueryRow(ctx, `
		SELECT max_retries, max_contacts, confidence_threshold, amount_threshold
		FROM policies
		WHERE merchant_id::text = $1
		LIMIT 1
	`, merchantID).Scan(&policy.MaxRetries, &policy.MaxContacts, &policy.ConfidenceThreshold, &policy.AmountThreshold)
	if err != nil {
		// If merchant doesn't have a row yet, insert one with defaults for future lookups
		_, _ = e.db.Exec(ctx, `
			INSERT INTO policies (merchant_id, max_retries, max_contacts, confidence_threshold, amount_threshold)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (merchant_id) DO NOTHING
		`, merchantID, policy.MaxRetries, policy.MaxContacts, policy.ConfidenceThreshold, policy.AmountThreshold)
	}

	return policy
}

// Evaluate applies database-configured safety policies to the AI's recommendation.
func (e *Engine) Evaluate(ctx context.Context, merchantID string, event schemas.PaymentFailureEvent, aiRec aiprovider.AIRecommendation) (PolicyDecision, error) {
	// 1. Fetch real policy thresholds from DB
	pol := e.GetMerchantPolicy(ctx, merchantID)

	// 2. Check Customer Opt-Out in database
	if e.db != nil && event.PaymentID != "" {
		var optOut bool
		err := e.db.QueryRow(ctx, `
			SELECT c.communication_opt_out
			FROM customers c
			JOIN payments p ON p.customer_id = c.id
			WHERE p.id::text = $1
			LIMIT 1
		`, event.PaymentID).Scan(&optOut)
		if err == nil && optOut {
			if aiRec.RecommendedAction == "CUSTOMER_NOTIFICATION" || aiRec.RecommendedAction == "PAYMENT_LINK" {
				log.Printf("Policy: Customer opted out of communications. Blocking action %s", aiRec.RecommendedAction)
				return PolicyDecision{
					Decision: DecisionBlock,
					Reason:   "Customer has opted out of communications.",
				}, nil
			}
		}
	}

	// 3. Retry Limit Check
	if event.CustomerHistory.FailedPayments >= pol.MaxRetries || event.AttemptNumber > pol.MaxRetries {
		return PolicyDecision{
			Decision: DecisionBlock,
			Reason:   fmt.Sprintf("Retry limit (%d) reached for payment/customer.", pol.MaxRetries),
		}, nil
	}

	// 4. Duplicate Action Prevention
	if e.db != nil && event.PaymentID != "" {
		var duplicateCount int
		err := e.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM recovery_actions ra
			JOIN recovery_workflows rw ON ra.workflow_id = rw.id
			WHERE rw.payment_id::text = $1 AND ra.action_type = $2 AND ra.status = 'EXECUTED'
		`, event.PaymentID, string(aiRec.RecommendedAction)).Scan(&duplicateCount)
		if err == nil && duplicateCount > 0 && (aiRec.RecommendedAction == "PAYMENT_METHOD_UPDATE" || aiRec.RecommendedAction == "PAYMENT_LINK") {
			return PolicyDecision{
				Decision: DecisionBlock,
				Reason:   fmt.Sprintf("Duplicate action prevented: %s already executed for payment.", aiRec.RecommendedAction),
			}, nil
		}
	}

	// 5. Confidence Threshold Check (Escalation)
	if aiRec.Confidence < pol.ConfidenceThreshold {
		return PolicyDecision{
			Decision: DecisionEscalate,
			Reason:   fmt.Sprintf("AI confidence (%.2f) below merchant policy threshold (%.2f). Routing to human review.", aiRec.Confidence, pol.ConfidenceThreshold),
		}, nil
	}

	// 6. High-Value Threshold Check (Escalation)
	if event.Amount > pol.AmountThreshold {
		return PolicyDecision{
			Decision: DecisionEscalate,
			Reason:   fmt.Sprintf("Payment amount (%.2f) exceeds merchant high-value threshold (%.2f). Escalating for approval.", event.Amount, pol.AmountThreshold),
		}, nil
	}

	// 7. Action Modification (e.g. repeated immediate retry -> force delayed retry)
	if aiRec.RecommendedAction == "IMMEDIATE_RETRY" && event.AttemptNumber > 1 {
		modified := aiRec
		modified.RecommendedAction = "DELAYED_RETRY"
		modified.RecommendedDelayHours = 24
		return PolicyDecision{
			Decision:       DecisionModify,
			Reason:         "Multiple attempts recorded. Modifying IMMEDIATE_RETRY to DELAYED_RETRY (+24h).",
			ModifiedAction: modified,
		}, nil
	}

	// Passed all checks
	return PolicyDecision{
		Decision: DecisionAllow,
		Reason:   "Passed all database policy and safety checks.",
	}, nil
}
