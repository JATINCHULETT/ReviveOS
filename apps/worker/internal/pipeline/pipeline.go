package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/packages/recovery"
	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
	aiprovider "github.com/reviveos/services/ai-provider"
)

type AnalysisResult struct {
	WorkflowID       string                       `json:"workflow_id"`
	PaymentID        string                       `json:"payment_id"`
	CustomerID       string                       `json:"customer_id"`
	FailureCategory  types.FailureCategory        `json:"failure_category"`
	CustomerHistory  schemas.CustomerHistory      `json:"customer_history"`
	Probability      float64                      `json:"probability"`
	ModelVersion     string                       `json:"model_version"`
	PredictionID     string                       `json:"prediction_id"`
	AIDecisionID     string                       `json:"ai_decision_id"`
	AIRecommendation *aiprovider.AIRecommendation `json:"ai_recommendation"`
	FeaturesUsed     map[string]interface{}       `json:"features_used"`
}

type Pipeline struct {
	pool        *pgxpool.Pool
	classifier  *types.Classifier
	probModel   *recovery.ProbabilityModel
	aiProvider  aiprovider.Provider
}

func NewPipeline(pool *pgxpool.Pool) *Pipeline {
	return &Pipeline{
		pool:       pool,
		classifier: types.NewClassifier(),
		probModel:  recovery.NewProbabilityModel(),
		aiProvider: aiprovider.NewAIProvider(),
	}
}

func NewPipelineWithAI(pool *pgxpool.Pool, ai aiprovider.Provider) *Pipeline {
	return &Pipeline{
		pool:       pool,
		classifier: types.NewClassifier(),
		probModel:  recovery.NewProbabilityModel(),
		aiProvider: ai,
	}
}

// FetchCustomerHistory queries real historical payment outcomes for a customer from PostgreSQL.
func (p *Pipeline) FetchCustomerHistory(ctx context.Context, customerID, currentPaymentID string) (schemas.CustomerHistory, error) {
	var history schemas.CustomerHistory

	if p.pool == nil || customerID == "" {
		return history, fmt.Errorf("invalid pool or empty customerID")
	}

	query := `
		SELECT 
			COALESCE(COUNT(CASE WHEN status IN ('CAPTURED', 'RECOVERED', 'SUCCESS') THEN 1 END), 0) as successful_payments,
			COALESCE(COUNT(CASE WHEN status = 'FAILED' THEN 1 END), 0) as failed_payments
		FROM payments
		WHERE customer_id::text = $1 AND id::text != $2
	`

	err := p.pool.QueryRow(ctx, query, customerID, currentPaymentID).Scan(
		&history.SuccessfulPayments,
		&history.FailedPayments,
	)
	if err != nil {
		return history, fmt.Errorf("failed to query customer history for %s: %w", customerID, err)
	}

	return history, nil
}

// AnalyzePayment executes the failure classification, history extraction, probability estimation, AI inference, and persistence.
func (p *Pipeline) AnalyzePayment(ctx context.Context, paymentIDOrExternal string) (*AnalysisResult, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("db pool is nil")
	}

	// 1. Query Payment from PostgreSQL
	var (
		paymentID         string
		merchantID        string
		customerID        string
		amount            float64
		currency          string
		paymentStatus     string
		paymentMethod     string
		failureCode       string
		razorpayPaymentID string
		createdAt         time.Time
	)

	paymentQuery := `
		SELECT 
			id::text,
			merchant_id::text,
			customer_id::text,
			amount::float8,
			currency,
			status,
			COALESCE(method, ''),
			COALESCE(failure_code, ''),
			COALESCE(razorpay_payment_id, ''),
			created_at
		FROM payments
		WHERE id::text = $1 OR razorpay_payment_id = $1
		LIMIT 1
	`

	err := p.pool.QueryRow(ctx, paymentQuery, paymentIDOrExternal).Scan(
		&paymentID,
		&merchantID,
		&customerID,
		&amount,
		&currency,
		&paymentStatus,
		&paymentMethod,
		&failureCode,
		&razorpayPaymentID,
		&createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("payment not found for %s: %w", paymentIDOrExternal, err)
	}

	// 2. Ensure Recovery Workflow exists for this payment
	var workflowID string
	wfQuery := `
		SELECT id::text
		FROM recovery_workflows
		WHERE payment_id::text = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err = p.pool.QueryRow(ctx, wfQuery, paymentID).Scan(&workflowID)
	if err != nil {
		insertWfSQL := `
			INSERT INTO recovery_workflows (payment_id, merchant_id, status, created_at, updated_at)
			VALUES ($1, $2, 'ANALYZING', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING id::text
		`
		err = p.pool.QueryRow(ctx, insertWfSQL, paymentID, merchantID).Scan(&workflowID)
		if err != nil {
			return nil, fmt.Errorf("failed to create recovery workflow for payment %s: %w", paymentID, err)
		}
	}

	// 3. Query Real Customer History from PostgreSQL
	history, err := p.FetchCustomerHistory(ctx, customerID, paymentID)
	if err != nil {
		log.Printf("[Pipeline] Warning: failed to fetch customer history: %v, defaulting to 0", err)
	}

	// 4. Deterministic Failure Classification
	category := p.classifier.Classify(failureCode, "")

	// 5. Query Attempt Number
	var attemptCount int
	_ = p.pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(*), 0)
		FROM recovery_actions
		WHERE workflow_id::text = $1
	`, workflowID).Scan(&attemptCount)
	attemptNumber := attemptCount + 1

	// 6. Build Event with Real Features
	event := schemas.PaymentFailureEvent{
		PaymentID:       paymentID,
		CustomerID:      customerID,
		Amount:          amount,
		Currency:        currency,
		PaymentMethod:   paymentMethod,
		FailureCode:     failureCode,
		AttemptNumber:   attemptNumber,
		Timestamp:       createdAt,
		CustomerHistory: history,
	}

	// 7. Calculate Statistical Recovery Probability
	probability := p.probModel.Predict(event, category)
	modelVersion := "logistic-v1"

	// 8. Prepare Features Used JSON
	features := map[string]interface{}{
		"amount":            amount,
		"currency":          currency,
		"category":          string(category),
		"failure_code":      failureCode,
		"attempt_number":    attemptNumber,
		"customer_success":  history.SuccessfulPayments,
		"customer_failures": history.FailedPayments,
	}
	featuresJSON, _ := json.Marshal(features)

	// 9. Persist Prediction to model_predictions table
	var predictionID string
	insertPredSQL := `
		INSERT INTO model_predictions (
			workflow_id,
			payment_id,
			model_version,
			probability,
			failure_category,
			features_used,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING id::text
	`
	err = p.pool.QueryRow(
		ctx,
		insertPredSQL,
		workflowID,
		paymentID,
		modelVersion,
		probability,
		string(category),
		featuresJSON,
	).Scan(&predictionID)
	if err != nil {
		return nil, fmt.Errorf("failed to persist model prediction: %w", err)
	}

	// 10. Query Empirical Stats for AI Context
	var categorySuccessRate float64 = 0.65
	var statsRow struct {
		Recovered int
		Total     int
	}
	err = p.pool.QueryRow(ctx, `
		SELECT 
			COALESCE(COUNT(CASE WHEN ro.recovered = true THEN 1 END), 0),
			COALESCE(COUNT(rw.id), 0)
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		LEFT JOIN recovery_outcomes ro ON ro.payment_id = p.id
		WHERE p.failure_code = $1
	`, failureCode).Scan(&statsRow.Recovered, &statsRow.Total)
	if err == nil && statsRow.Total > 0 {
		categorySuccessRate = float64(statsRow.Recovered) / float64(statsRow.Total)
	}

	aiCtx := aiprovider.AIContext{
		FailureEvent:       event,
		DeterministicClass: category,
		StatisticalProb:    probability,
		EmpiricalStats: map[string]float64{
			"category_success_rate": categorySuccessRate,
			"customer_success_rate": func() float64 {
				tot := history.SuccessfulPayments + history.FailedPayments
				if tot > 0 {
					return float64(history.SuccessfulPayments) / float64(tot)
				}
				return 0.5
			}(),
		},
	}

	// 11. Execute AI Inference via Ollama / NVIDIA (or Safe Fallback if unavailable)
	var aiRec *aiprovider.AIRecommendation
	if p.aiProvider != nil {
		inferCtx, cancelInfer := context.WithTimeout(ctx, 15*time.Second)
		rec, err := p.aiProvider.RecommendStrategy(inferCtx, aiCtx)
		cancelInfer()
		if err != nil {
			log.Printf("[Pipeline] AI inference error (%v), activating safe fallback", err)
			aiRec = aiprovider.SafeFallback(aiCtx, err)
		} else {
			aiRec = rec
		}
	} else {
		aiRec = aiprovider.SafeFallback(aiCtx, nil)
	}

	// 12. Persist AI Decision to ai_decisions table in PostgreSQL
	var aiDecisionID string
	insertAIDecisionSQL := `
		INSERT INTO ai_decisions (
			workflow_id,
			provider,
			model,
			prompt_hash,
			raw_response,
			diagnosis,
			recommended_action,
			recommended_delay_hours,
			confidence,
			recoverability,
			reasoning,
			inference_duration_ms,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP)
		RETURNING id::text
	`
	err = p.pool.QueryRow(
		ctx,
		insertAIDecisionSQL,
		workflowID,
		aiRec.Provider,
		aiRec.Model,
		aiRec.PromptHash,
		aiRec.RawResponse,
		aiRec.Diagnosis,
		string(aiRec.RecommendedAction),
		aiRec.RecommendedDelayHours,
		aiRec.Confidence,
		aiRec.Recoverability,
		aiRec.Reasoning,
		aiRec.InferenceDurationMs,
	).Scan(&aiDecisionID)
	if err != nil {
		log.Printf("[Pipeline] ERROR persisting ai_decisions: %v", err)
	}

	// 13. Update Workflow with Selected Action & Probability
	_, err = p.pool.Exec(ctx, `
		UPDATE recovery_workflows
		SET recovery_probability = $1, selected_action = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id::text = $3
	`, probability, string(aiRec.RecommendedAction), workflowID)
	if err != nil {
		log.Printf("[Pipeline] Warning: failed to update workflow recovery_probability: %v", err)
	}

	log.Printf("[Pipeline] Complete for payment %s | Prob: %.4f | AI Action: %s | Conf: %.2f | Latency: %dms | AIDecisionID: %s",
		paymentID, probability, aiRec.RecommendedAction, aiRec.Confidence, aiRec.InferenceDurationMs, aiDecisionID)

	return &AnalysisResult{
		WorkflowID:       workflowID,
		PaymentID:        paymentID,
		CustomerID:       customerID,
		FailureCategory:  category,
		CustomerHistory:  history,
		Probability:      probability,
		ModelVersion:     modelVersion,
		PredictionID:     predictionID,
		AIDecisionID:     aiDecisionID,
		AIRecommendation: aiRec,
		FeaturesUsed:     features,
	}, nil
}
