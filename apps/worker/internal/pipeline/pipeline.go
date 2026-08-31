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
	"github.com/reviveos/risk"
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
	RiskAssessment   *risk.RiskAnalysisResponse   `json:"risk_assessment,omitempty"`
}

type Pipeline struct {
	pool        *pgxpool.Pool
	classifier  *types.Classifier
	probModel   *recovery.ProbabilityModel
	aiProvider  aiprovider.Provider
	riskClient  *risk.Client
}

func NewPipeline(pool *pgxpool.Pool) *Pipeline {
	return &Pipeline{
		pool:       pool,
		classifier: types.NewClassifier(),
		probModel:  recovery.NewProbabilityModel(),
		aiProvider: aiprovider.NewAIProvider(),
		riskClient: risk.NewClient(),
	}
}

func NewPipelineWithAI(pool *pgxpool.Pool, ai aiprovider.Provider) *Pipeline {
	return &Pipeline{
		pool:       pool,
		classifier: types.NewClassifier(),
		probModel:  recovery.NewProbabilityModel(),
		aiProvider: ai,
		riskClient: risk.NewClient(),
	}
}

// FetchCustomerHistory queries real historical payment outcomes for a customer from PostgreSQL by Email or ID.
func (p *Pipeline) FetchCustomerHistory(ctx context.Context, customerID, customerEmail, currentPaymentID string) (schemas.CustomerHistory, error) {
	var history schemas.CustomerHistory

	if p.pool == nil {
		return history, fmt.Errorf("invalid pool")
	}

	query := `
		SELECT 
			COALESCE(COUNT(CASE WHEN p.status IN ('CAPTURED', 'RECOVERED', 'SUCCESS') THEN 1 END), 0) as successful_payments,
			COALESCE(COUNT(CASE WHEN p.status = 'FAILED' THEN 1 END), 0) as failed_payments
		FROM payments p
		LEFT JOIN customers c ON p.customer_id = c.id
		WHERE (
			(c.email = $2 AND $2 != '') OR 
			(p.customer_id::text = $1 AND $1 != '')
		) 
		AND p.id::text != $3
	`

	err := p.pool.QueryRow(ctx, query, customerID, customerEmail, currentPaymentID).Scan(
		&history.SuccessfulPayments,
		&history.FailedPayments,
	)
	if err != nil {
		return history, fmt.Errorf("failed to query customer history for %s (%s): %w", customerID, customerEmail, err)
	}

	return history, nil
}

// AnalyzePayment executes the failure classification, history extraction, probability estimation, AI inference, and persistence.
func (p *Pipeline) AnalyzePayment(ctx context.Context, paymentIDOrExternal string) (*AnalysisResult, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("db pool is nil")
	}

	// 1. Query Payment from PostgreSQL with Customer Email
	var (
		paymentID         string
		merchantID        string
		customerID        string
		customerEmail     string
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
			p.id::text,
			p.merchant_id::text,
			p.customer_id::text,
			COALESCE(c.email, ''),
			p.amount::float8,
			p.currency,
			p.status,
			COALESCE(p.method, ''),
			COALESCE(p.failure_code, ''),
			COALESCE(p.razorpay_payment_id, ''),
			p.created_at
		FROM payments p
		LEFT JOIN customers c ON p.customer_id = c.id
		WHERE p.id::text = $1 OR p.razorpay_payment_id = $1
		LIMIT 1
	`

	err := p.pool.QueryRow(ctx, paymentQuery, paymentIDOrExternal).Scan(
		&paymentID,
		&merchantID,
		&customerID,
		&customerEmail,
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

	// 3. Query Real Customer History from PostgreSQL by Email or ID
	history, err := p.FetchCustomerHistory(ctx, customerID, customerEmail, paymentID)
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

	// 6.5. REVENUE RISK ASSESSMENT STAGE (Fraud Detection & Return Risk Intelligence)
	riskReq := risk.RiskAnalysisRequest{
		EventType:            "PAYMENT_FAILED",
		PaymentID:            paymentID,
		MerchantID:           merchantID,
		CustomerID:           customerID,
		CustomerEmail:        customerEmail,
		Amount:               amount,
		Currency:             currency,
		FailureCode:          failureCode,
		AttemptNumber:        attemptNumber,
		CustomerFailedCount:  history.FailedPayments,
		CustomerSuccessCount: history.SuccessfulPayments,
		Velocity1h:           1,
	}

	riskResp, riskErr := p.riskClient.AnalyzeRisk(ctx, riskReq)
	if riskErr != nil || riskResp == nil {
		log.Printf("[Pipeline] Warning: risk analysis failed (%v), using default safe risk scores", riskErr)
	}

	if riskResp != nil {
		var returnProb float64 = 0.0
		var returnLevel string = "LOW"
		if riskResp.ReturnRisk != nil {
			returnProb = riskResp.ReturnRisk.Probability
			returnLevel = riskResp.ReturnRisk.RiskLevel
		}
		rawPayload, _ := json.Marshal(riskResp)

		_, _ = p.pool.Exec(ctx, `
			INSERT INTO risk_assessments (
				payment_id, workflow_id, merchant_id, event_type,
				fraud_probability, fraud_risk_level, return_probability, return_risk_level,
				overall_risk_level, expected_loss, recommended_action, reason, model_version, raw_payload, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, CURRENT_TIMESTAMP)
		`, paymentID, workflowID, merchantID, "PAYMENT_FAILED",
			riskResp.Fraud.Probability, riskResp.Fraud.RiskLevel, returnProb, returnLevel,
			riskResp.OverallRisk, riskResp.ExpectedLoss, riskResp.RecommendedAction, riskResp.Reason, riskResp.Fraud.ModelVersion, rawPayload)

		// Update workflow with risk intelligence
		_, _ = p.pool.Exec(ctx, `
			UPDATE recovery_workflows 
			SET fraud_probability = $1, return_probability = $2, overall_risk = $3, expected_loss = $4, risk_action = $5, updated_at = CURRENT_TIMESTAMP
			WHERE id::text = $6
		`, riskResp.Fraud.Probability, returnProb, riskResp.OverallRisk, riskResp.ExpectedLoss, riskResp.RecommendedAction, workflowID)
	}

	// 7. Calculate Statistical Recovery Probability
	probability := p.probModel.Predict(event, category)
	modelVersion := "logistic-v1"

	// 8. Prepare Features Used JSON
	fraudProbVal := 0.05
	fraudLevelVal := "LOW"
	if riskResp != nil {
		fraudProbVal = riskResp.Fraud.Probability
		fraudLevelVal = riskResp.Fraud.RiskLevel
	}

	features := map[string]interface{}{
		"amount":            amount,
		"currency":          currency,
		"category":          string(category),
		"failure_code":      failureCode,
		"attempt_number":    attemptNumber,
		"customer_success":  history.SuccessfulPayments,
		"customer_failures": history.FailedPayments,
		"fraud_risk_score":  fraudProbVal,
		"fraud_risk_level":  fraudLevelVal,
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
		FraudRiskScore:     fraudProbVal,
		FraudRiskLevel:     fraudLevelVal,
		EmpiricalStats: map[string]float64{
			"category_success_rate": categorySuccessRate,
			"customer_success_rate": func() float64 {
				tot := history.SuccessfulPayments + history.FailedPayments
				if tot > 0 {
					return float64(history.SuccessfulPayments) / float64(tot)
				}
				return 0.5
			}(),
			"fraud_risk_probability": fraudProbVal,
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

	// 13. Update Workflow with Selected Action & Probability & advance status to SCHEDULED
	_, err = p.pool.Exec(ctx, `
		UPDATE recovery_workflows
		SET recovery_probability = $1, 
		    selected_action = $2, 
		    status = CASE WHEN status = 'ANALYZING' THEN 'SCHEDULED' ELSE status END,
		    updated_at = CURRENT_TIMESTAMP
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
		RiskAssessment:   riskResp,
	}, nil
}
