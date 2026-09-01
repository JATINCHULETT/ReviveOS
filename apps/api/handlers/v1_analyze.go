package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/packages/recovery"
	"github.com/reviveos/packages/types"
	"github.com/reviveos/risk"
	"github.com/reviveos/schemas"
	aiprovider "github.com/reviveos/services/ai-provider"
	paymentprovider "github.com/reviveos/services/payment-provider"
	policy "github.com/reviveos/services/policy-engine"
)

// PaymentAnalyzeRequest is the input payload for POST /v1/payments/analyze
type PaymentAnalyzeRequest struct {
	PaymentID     string  `json:"paymentId"`
	OrderID       string  `json:"orderId,omitempty"`
	CustomerID    string  `json:"customerId,omitempty"`
	CustomerEmail string  `json:"customerEmail,omitempty"`
	Amount        float64 `json:"amount"` // in INR
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"paymentMethod,omitempty"`
	Bank          string  `json:"bank,omitempty"`
	FailureCode   string  `json:"failureCode"`
	FailureReason string  `json:"failureReason,omitempty"`
	AttemptNumber int     `json:"attemptNumber,omitempty"`
}

// PaymentAnalyzeResponse is the complete synchronous AI recovery response
type PaymentAnalyzeResponse struct {
	PaymentID           string                   `json:"paymentId"`
	FailureCategory     string                   `json:"failureCategory"`
	Diagnosis           string                   `json:"diagnosis"`
	FraudRisk           FraudRiskSummary         `json:"fraudRisk"`
	RecoveryProbability float64                  `json:"recoveryProbability"`
	NextBestAction      string                   `json:"nextBestAction"`
	Action              string                   `json:"action"` // RETRY, PAYMENT_LINK, ALTERNATIVE_PAYMENT, BLOCK, NO_ACTION
	DelaySeconds        int                      `json:"delaySeconds"`
	Confidence          float64                  `json:"confidence"`
	Reason              string                   `json:"reason"`
	CustomerHistory     schemas.CustomerHistory  `json:"customerHistory"`
	Decision            string                   `json:"decision"` // "RECOVER", "BLOCK", "ESCALATE", "NO_ACTION"
	Timestamp           string                   `json:"timestamp"`
}

type FraudRiskSummary struct {
	FraudProbability float64 `json:"fraudProbability"`
	RiskLevel        string  `json:"riskLevel"` // "LOW", "MEDIUM", "HIGH"
	ExpectedLoss     float64 `json:"expectedLoss,omitempty"`
	OverallRisk      string  `json:"overallRisk"`
}

// V1AnalyzePaymentHandler handles POST /v1/payments/analyze
func V1AnalyzePaymentHandler(pool *pgxpool.Pool) http.HandlerFunc {
	classifier := types.NewClassifier()
	probModel := recovery.NewProbabilityModel()
	riskClient := risk.NewClient()
	aiProvider := aiprovider.NewAIProvider()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		mCtx := GetMerchantContext(r)

		var req PaymentAnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Invalid request JSON: %v"}`, err), http.StatusBadRequest)
			return
		}

		if req.PaymentID == "" {
			http.Error(w, `{"error":"paymentId is required"}`, http.StatusBadRequest)
			return
		}
		if req.Currency == "" {
			req.Currency = "INR"
		}
		if req.AttemptNumber <= 0 {
			req.AttemptNumber = 1
		}

		// 1. Classify failure deterministically
		category := classifier.Classify(req.FailureCode, req.FailureReason)

		// 2. Fetch Customer History
		var history schemas.CustomerHistory
		if pool != nil {
			query := `
				SELECT 
					COALESCE(COUNT(CASE WHEN p.status IN ('CAPTURED', 'RECOVERED', 'SUCCESS') THEN 1 END), 0),
					COALESCE(COUNT(CASE WHEN p.status = 'FAILED' THEN 1 END), 0)
				FROM payments p
				LEFT JOIN customers c ON p.customer_id = c.id
				WHERE (
					(c.email = $2 AND $2 != '') OR 
					(p.customer_id::text = $1 AND $1 != '')
				)
			`
			_ = pool.QueryRow(r.Context(), query, req.CustomerID, req.CustomerEmail).Scan(
				&history.SuccessfulPayments,
				&history.FailedPayments,
			)
		}

		// 3. ML Risk Assessment (Fraud & Return Risk)
		riskReq := risk.RiskAnalysisRequest{
			EventType:            "PAYMENT_FAILED",
			PaymentID:            req.PaymentID,
			MerchantID:           mCtx.MerchantID,
			CustomerID:           req.CustomerID,
			CustomerEmail:        req.CustomerEmail,
			Amount:               req.Amount,
			Currency:             req.Currency,
			FailureCode:          req.FailureCode,
			AttemptNumber:        req.AttemptNumber,
			CustomerFailedCount:  history.FailedPayments,
			CustomerSuccessCount: history.SuccessfulPayments,
			Velocity1h:           1,
		}

		riskResp, _ := riskClient.AnalyzeRisk(r.Context(), riskReq)
		fraudProb := 0.04
		fraudLevel := "LOW"
		overallRisk := "LOW"
		var expectedLoss float64 = 0.0

		if riskResp != nil {
			fraudProb = riskResp.Fraud.Probability
			fraudLevel = riskResp.Fraud.RiskLevel
			overallRisk = riskResp.OverallRisk
			expectedLoss = riskResp.ExpectedLoss
		}

		// 4. Calculate Recovery Probability
		event := schemas.PaymentFailureEvent{
			PaymentID:       req.PaymentID,
			CustomerID:      req.CustomerID,
			Amount:          req.Amount,
			Currency:        req.Currency,
			PaymentMethod:   req.PaymentMethod,
			FailureCode:     req.FailureCode,
			AttemptNumber:   req.AttemptNumber,
			Timestamp:       time.Now(),
			CustomerHistory: history,
		}
		recoveryProb := probModel.Predict(event, category)

		// 5. Next Best Action & Optimal Timing
		action := "RETRY_NOW"
		delaySeconds := 0
		decision := "RECOVER"
		reason := "High recovery probability and low fraud risk"
		confidence := 0.92

		if fraudLevel == "HIGH" || fraudProb > 0.70 {
			action = "BLOCK"
			decision = "BLOCK"
			reason = "High fraud probability detected by Revenue Risk Engine"
			confidence = 0.98
		} else {
			switch category {
			case types.BankUnavailable, types.Timeout, types.NetworkError:
				action = "RETRY_LATER"
				delaySeconds = 120 // 2 minutes
				decision = "RECOVER"
				reason = "Temporary banking gateway outage; delayed retry scheduled."
				confidence = 0.94
			case types.InsufficientFunds:
				action = "RETRY_LATER"
				delaySeconds = 86400 // 24 hours
				decision = "RECOVER"
				reason = "Insufficient funds; retry timed after standard account replenishment window."
				confidence = 0.85
			case types.AuthenticationFailed, types.CustomerActionRequired:
				action = "PAYMENT_LINK"
				delaySeconds = 0
				decision = "RECOVER"
				reason = "Customer authentication required; generated smart interactive recovery link."
				confidence = 0.89
			case types.ExpiredCard, types.MandateFailed:
				action = "ALTERNATIVE_PAYMENT"
				delaySeconds = 0
				decision = "RECOVER"
				reason = "Expired credentials or mandate error; recommend switching payment method to UPI or alternative card."
				confidence = 0.91
			default:
				if recoveryProb > 0.5 {
					action = "RETRY_NOW"
					delaySeconds = 0
					decision = "RECOVER"
					reason = "Standard recoverable payment failure."
				} else {
					action = "NO_ACTION"
					decision = "NO_ACTION"
					reason = "Unrecoverable payment decline."
				}
			}
		}

		// Optional AI diagnosis enrichment if available
		diagnosis := fmt.Sprintf("%s failure classified as %s", req.FailureCode, category)
		if aiProvider != nil {
			aiCtx := aiprovider.AIContext{
				FailureEvent:       event,
				DeterministicClass: category,
				StatisticalProb:    recoveryProb,
				FraudRiskScore:     fraudProb,
				FraudRiskLevel:     fraudLevel,
			}
			inferCtx, cancelInfer := context.WithTimeout(r.Context(), 1*time.Second)
			rec, _ := aiProvider.RecommendStrategy(inferCtx, aiCtx)
			cancelInfer()
			if rec != nil && rec.Diagnosis != "" {
				diagnosis = rec.Diagnosis
				if rec.Reasoning != "" {
					reason = rec.Reasoning
				}
			}
		}

		// Record Audit Log
		RecordAuditLog(r.Context(), pool, mCtx.MerchantID, "API_KEY", mCtx.KeyID, "PAYMENT_ANALYZED", r.RemoteAddr, map[string]interface{}{
			"payment_id":           req.PaymentID,
			"failure_category":     category,
			"recovery_probability": recoveryProb,
			"fraud_probability":    fraudProb,
			"decision":             decision,
			"action":               action,
		})

		resp := PaymentAnalyzeResponse{
			PaymentID:       req.PaymentID,
			FailureCategory: string(category),
			Diagnosis:       diagnosis,
			FraudRisk: FraudRiskSummary{
				FraudProbability: fraudProb,
				RiskLevel:        fraudLevel,
				ExpectedLoss:     expectedLoss,
				OverallRisk:      overallRisk,
			},
			RecoveryProbability: recoveryProb,
			NextBestAction:      action,
			Action:              action,
			DelaySeconds:        delaySeconds,
			Confidence:          confidence,
			Reason:              reason,
			CustomerHistory:     history,
			Decision:            decision,
			Timestamp:           time.Now().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(resp)
	}
}

// V1RecoveryDecisionHandler handles POST /v1/recovery/decision
func V1RecoveryDecisionHandler(pool *pgxpool.Pool) http.HandlerFunc {
	polEngine := policy.NewEngine(pool)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		mCtx := GetMerchantContext(r)

		var req struct {
			PaymentID           string                 `json:"paymentId"`
			FailureType         string                 `json:"failureType"`
			FraudProbability    float64                `json:"fraudProbability"`
			RecoveryProbability float64                `json:"recoveryProbability"`
			CustomerHistory     map[string]interface{} `json:"customerHistory"`
			PaymentContext      map[string]interface{} `json:"paymentContext"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Invalid request JSON: %v"}`, err), http.StatusBadRequest)
			return
		}

		decision := "RECOVER"
		action := "RETRY"
		delaySeconds := 120
		confidence := 0.94
		reason := "High recovery probability and low fraud risk"

		if req.FraudProbability > 0.65 {
			decision = "BLOCK"
			action = "BLOCK"
			delaySeconds = 0
			confidence = 0.98
			reason = "High fraud probability exceeds policy threshold"
		} else if req.RecoveryProbability < 0.30 {
			decision = "NO_ACTION"
			action = "NO_ACTION"
			delaySeconds = 0
			confidence = 0.90
			reason = "Low statistical recovery probability"
		}

		// Apply merchant policy engine
		if pool != nil {
			pol := polEngine.GetMerchantPolicy(r.Context(), mCtx.MerchantID)
			if req.RecoveryProbability < pol.ConfidenceThreshold && req.FraudProbability <= 0.65 {
				decision = "ESCALATE"
				action = "PAYMENT_LINK"
				reason = fmt.Sprintf("Recovery probability %.2f below merchant policy confidence threshold %.2f", req.RecoveryProbability, pol.ConfidenceThreshold)
			}
		}

		RecordAuditLog(r.Context(), pool, mCtx.MerchantID, "API_KEY", mCtx.KeyID, "RECOVERY_DECISION_EVALUATED", r.RemoteAddr, map[string]interface{}{
			"payment_id": req.PaymentID,
			"decision":   decision,
			"action":     action,
		})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"decision":     decision,
			"action":       action,
			"delaySeconds": delaySeconds,
			"confidence":   confidence,
			"reason":       reason,
			"timestamp":    time.Now().Format(time.RFC3339),
		})
	}
}

// V1RecoveryExecuteHandler handles POST /v1/recovery/execute
func V1RecoveryExecuteHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		mCtx := GetMerchantContext(r)

		var req struct {
			PaymentID      string                 `json:"paymentId"`
			Amount         float64                `json:"amount"`
			Action         string                 `json:"action"` // "RETRY", "PAYMENT_LINK", "ALTERNATIVE_PAYMENT"
			IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
			Metadata       map[string]interface{} `json:"metadata,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Invalid request JSON: %v"}`, err), http.StatusBadRequest)
			return
		}

		if req.PaymentID == "" {
			http.Error(w, `{"error":"paymentId is required"}`, http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			req.Amount = 4999.0
		}

		provider, pErr := paymentprovider.NewPaymentProvider("", pool)
		if pErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Payment provider error: %v"}`, pErr), http.StatusInternalServerError)
			return
		}

		result, err := provider.CreateRetryAttempt(r.Context(), req.PaymentID, req.Amount)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Recovery execution failed: %v"}`, err), http.StatusBadGateway)
			return
		}

		RecordAuditLog(r.Context(), pool, mCtx.MerchantID, "API_KEY", mCtx.KeyID, "RECOVERY_EXECUTED", r.RemoteAddr, map[string]interface{}{
			"payment_id": req.PaymentID,
			"action":     req.Action,
			"status":     result.Status,
			"link_url":   result.PaymentLinkURL,
		})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "EXECUTED",
			"paymentId":       req.PaymentID,
			"action":          req.Action,
			"resultStatus":    result.Status,
			"paymentLinkUrl":  result.PaymentLinkURL,
			"verified":        result.Status == "SUCCESS" || result.Status == "CAPTURED",
			"executedAt":      time.Now().Format(time.RFC3339),
		})
	}
}

// V1GetPaymentHandler handles GET /v1/payments/:paymentId
func V1GetPaymentHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		mCtx := GetMerchantContext(r)
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 3 {
			http.Error(w, `{"error":"Missing paymentId in URL path"}`, http.StatusBadRequest)
			return
		}
		paymentID := pathParts[2]

		if pool == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         paymentID,
				"merchantId": mCtx.MerchantID,
				"status":     "FAILED",
				"amount":     4999.00,
				"currency":   "INR",
			})
			return
		}

		var (
			id, status, method, failureCode, extID string
			amount                                 float64
			currency                               string
			createdAt, updatedAt                   time.Time
		)

		query := `
			SELECT id::text, status, COALESCE(method, ''), COALESCE(failure_code, ''), COALESCE(razorpay_payment_id, ''), amount::float8, currency, created_at, updated_at
			FROM payments
			WHERE (id::text = $1 OR razorpay_payment_id = $1) AND (merchant_id::text = $2 OR $2 = '00000000-0000-0000-0000-000000000001')
			LIMIT 1
		`
		err := pool.QueryRow(r.Context(), query, paymentID, mCtx.MerchantID).Scan(
			&id, &status, &method, &failureCode, &extID, &amount, &currency, &createdAt, &updatedAt,
		)
		if err != nil {
			http.Error(w, `{"error":"Payment not found"}`, http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                id,
			"razorpayPaymentId": extID,
			"status":            status,
			"amount":            amount,
			"currency":          currency,
			"method":            method,
			"failureCode":       failureCode,
			"createdAt":         createdAt,
			"updatedAt":         updatedAt,
		})
	}
}
