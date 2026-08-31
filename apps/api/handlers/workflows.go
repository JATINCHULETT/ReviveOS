package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	paymentprovider "github.com/reviveos/services/payment-provider"
	"github.com/reviveos/utils/audit"
)

type WorkflowSummary struct {
	ID                  string     `json:"id"`
	PaymentID           string     `json:"payment_id"`
	MerchantID          *string    `json:"merchant_id,omitempty"`
	Status              string     `json:"status"`
	RecoveryProbability float64    `json:"recovery_probability"`
	SelectedAction      string     `json:"selected_action"`
	ScheduledAt         *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Payment fields
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	PaymentStatus     string  `json:"payment_status"`
	PaymentMethod     string  `json:"payment_method"`
	FailureCode       string  `json:"failure_code"`
	RazorpayPaymentID string  `json:"razorpay_payment_id"`

	// Customer fields
	CustomerID          string `json:"customer_id"`
	CustomerEmail       string `json:"customer_email"`
	CustomerPhone       string `json:"customer_phone"`
	CommunicationOptOut bool   `json:"communication_opt_out"`

	// Derived metrics
	AttemptsCount int  `json:"attempts_count"`
	IsRecovered   bool `json:"is_recovered"`

	// Customer lifetime memory & learning stats
	CustomerSuccessCount int `json:"customer_success_count"`
	CustomerFailedCount  int `json:"customer_failed_count"`

	// Revenue Risk Intelligence
	FraudProbability  float64 `json:"fraud_probability"`
	ReturnProbability float64 `json:"return_probability"`
	OverallRisk       string  `json:"overall_risk"`
	ExpectedLoss      float64 `json:"expected_loss"`
	RiskAction        string  `json:"risk_action"`
}

type RiskAssessmentItem struct {
	ID                string    `json:"id"`
	PaymentID         string    `json:"payment_id"`
	WorkflowID        string    `json:"workflow_id"`
	EventType         string    `json:"event_type"`
	FraudProbability  float64   `json:"fraud_probability"`
	FraudRiskLevel    string    `json:"fraud_risk_level"`
	ReturnProbability float64   `json:"return_probability"`
	ReturnRiskLevel   string    `json:"return_risk_level"`
	OverallRiskLevel  string    `json:"overall_risk_level"`
	ExpectedLoss      float64   `json:"expected_loss"`
	RecommendedAction string    `json:"recommended_action"`
	Reason            string    `json:"reason"`
	ModelVersion      string    `json:"model_version"`
	CreatedAt         time.Time `json:"created_at"`
}

type AIDecisionItem struct {
	ID                    string    `json:"id"`
	WorkflowID            string    `json:"workflow_id"`
	Provider              string    `json:"provider"`
	Model                 string    `json:"model"`
	PromptHash            *string   `json:"prompt_hash,omitempty"`
	RawResponse           *string   `json:"raw_response,omitempty"`
	Diagnosis             *string   `json:"diagnosis,omitempty"`
	RecommendedAction     *string   `json:"recommended_action,omitempty"`
	RecommendedDelayHours *int      `json:"recommended_delay_hours,omitempty"`
	Confidence            *float64  `json:"confidence,omitempty"`
	Recoverability        *float64  `json:"recoverability,omitempty"`
	Reasoning             *string   `json:"reasoning,omitempty"`
	InferenceDurationMs   *int      `json:"inference_duration_ms,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

type ModelPredictionItem struct {
	ID              string          `json:"id"`
	WorkflowID      string          `json:"workflow_id"`
	PaymentID       string          `json:"payment_id"`
	ModelVersion    string          `json:"model_version"`
	Probability     float64         `json:"probability"`
	FailureCategory *string         `json:"failure_category,omitempty"`
	FeaturesUsed    json.RawMessage `json:"features_used,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type RecoveryActionItem struct {
	ID          string     `json:"id"`
	WorkflowID  string     `json:"workflow_id"`
	ActionType  string     `json:"action_type"`
	Status      string     `json:"status"`
	Attempt     int        `json:"attempt"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	Result      *string    `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type RecoveryOutcomeItem struct {
	ID              string     `json:"id"`
	ActionID        *string    `json:"action_id,omitempty"`
	PaymentID       string     `json:"payment_id"`
	Recovered       bool       `json:"recovered"`
	RecoveredAmount *float64   `json:"recovered_amount,omitempty"`
	TimeToRecovery  *string    `json:"time_to_recovery,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AuditEventItem struct {
	EventID           string          `json:"event_id"`
	WorkflowID        string          `json:"workflow_id"`
	Timestamp         time.Time       `json:"timestamp"`
	Actor             string          `json:"actor"`
	Action            string          `json:"action"`
	PayloadHash       string          `json:"payload_hash"`
	PreviousEventHash *string         `json:"previous_event_hash,omitempty"`
	EventHash         string          `json:"event_hash"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

type WorkflowDetailResponse struct {
	Workflow         WorkflowSummary       `json:"workflow"`
	AIDecisions      []AIDecisionItem      `json:"ai_decisions"`
	ModelPredictions []ModelPredictionItem `json:"model_predictions"`
	RecoveryActions  []RecoveryActionItem  `json:"recovery_actions"`
	RecoveryOutcomes []RecoveryOutcomeItem `json:"recovery_outcomes"`
	AuditEvents      []AuditEventItem      `json:"audit_events"`
	RiskAssessments  []RiskAssessmentItem  `json:"risk_assessments"`
}

type WorkflowsListResponse struct {
	Data   []WorkflowSummary `json:"data"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func WorkflowsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/workflows")
		path = strings.TrimPrefix(path, "/")

		// Check for sub-actions like /workflows/{id}/approve, /workflows/{id}/reject, /workflows/{id}/override
		parts := strings.Split(path, "/")

		if r.Method == http.MethodPost {
			if len(parts) >= 2 {
				wfID := parts[0]
				action := parts[1]
				switch action {
				case "approve":
					approveWorkflow(r.Context(), pool, w, r, wfID)
					return
				case "reject":
					rejectWorkflow(r.Context(), pool, w, r, wfID)
					return
				case "override":
					overrideWorkflow(r.Context(), pool, w, r, wfID)
					return
				}
			}
			writeJSONError(w, http.StatusBadRequest, "invalid action endpoint", "expected /workflows/{id}/approve, /reject, or /override")
			return
		}

		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET, POST")
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "")
			return
		}

		// GET /workflows/interventions (returns all escalated / requires human review workflows)
		if path == "interventions" || path == "escalations" {
			getInterventions(r.Context(), pool, w, r)
			return
		}

		// If a specific ID is in URL path or query parameter, return single workflow
		if path != "" && !strings.Contains(path, "/") {
			getWorkflowByID(r.Context(), pool, w, path)
			return
		}
		if idParam := r.URL.Query().Get("id"); idParam != "" {
			getWorkflowByID(r.Context(), pool, w, idParam)
			return
		}

		listWorkflows(r.Context(), pool, w, r)
	}
}

func listWorkflows(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	// Automatically pull and sync payment links from Razorpay
	_ = paymentprovider.SyncRazorpayPaymentLinks(ctx, pool)

	query := r.URL.Query()

	limit := 50
	if l := query.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 200 {
			limit = val
		}
	}

	offset := 0
	if o := query.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	statusFilter := query.Get("status")
	merchantFilter := query.Get("merchant_id")
	paymentFilter := query.Get("payment_id")

	var conditions []string
	var args []interface{}
	argIdx := 1

	if statusFilter != "" && !strings.EqualFold(statusFilter, "all") {
		conditions = append(conditions, fmt.Sprintf("rw.status = $%d", argIdx))
		args = append(args, statusFilter)
		argIdx++
	}
	if merchantFilter != "" {
		conditions = append(conditions, fmt.Sprintf("rw.merchant_id::text = $%d", argIdx))
		args = append(args, merchantFilter)
		argIdx++
	}
	if paymentFilter != "" {
		conditions = append(conditions, fmt.Sprintf("(p.id::text = $%d OR p.razorpay_payment_id = $%d)", argIdx, argIdx))
		args = append(args, paymentFilter)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		JOIN customers c ON p.customer_id = c.id
		%s
	`, whereClause)

	var total int
	err := pool.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		log.Printf("ERROR: failed to count workflows: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query database", err.Error())
		return
	}

	// Query data
	argsWithPaging := append(args, limit, offset)
	dataSQL := fmt.Sprintf(`
		SELECT 
			rw.id::text,
			rw.payment_id::text,
			rw.merchant_id::text,
			rw.status,
			COALESCE(rw.recovery_probability, 0)::float8,
			COALESCE(rw.selected_action, ''),
			rw.scheduled_at,
			rw.created_at,
			rw.updated_at,
			p.amount::float8,
			p.currency,
			p.status,
			COALESCE(p.method, ''),
			COALESCE(p.failure_code, ''),
			COALESCE(p.razorpay_payment_id, ''),
			c.id::text,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			c.communication_opt_out,
			COALESCE((SELECT COUNT(*) FROM recovery_actions WHERE workflow_id = rw.id), 0),
			COALESCE((SELECT recovered FROM recovery_outcomes WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1), false),
			COALESCE(rw.fraud_probability, 0)::float8,
			COALESCE(rw.return_probability, 0)::float8,
			COALESCE(rw.overall_risk, 'LOW'),
			COALESCE(rw.expected_loss, 0)::float8,
			COALESCE(rw.risk_action, 'ALLOW')
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		JOIN customers c ON p.customer_id = c.id
		%s
		ORDER BY rw.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	rows, err := pool.Query(ctx, dataSQL, argsWithPaging...)
	if err != nil {
		log.Printf("ERROR: failed to query workflows list: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query workflows", err.Error())
		return
	}
	defer rows.Close()

	items := make([]WorkflowSummary, 0)
	for rows.Next() {
		var item WorkflowSummary
		var merchantID *string
		err := rows.Scan(
			&item.ID,
			&item.PaymentID,
			&merchantID,
			&item.Status,
			&item.RecoveryProbability,
			&item.SelectedAction,
			&item.ScheduledAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Amount,
			&item.Currency,
			&item.PaymentStatus,
			&item.PaymentMethod,
			&item.FailureCode,
			&item.RazorpayPaymentID,
			&item.CustomerID,
			&item.CustomerEmail,
			&item.CustomerPhone,
			&item.CommunicationOptOut,
			&item.AttemptsCount,
			&item.IsRecovered,
			&item.FraudProbability,
			&item.ReturnProbability,
			&item.OverallRisk,
			&item.ExpectedLoss,
			&item.RiskAction,
		)
		if err != nil {
			log.Printf("ERROR: failed to scan workflow row: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to parse workflow data", err.Error())
			return
		}
		item.MerchantID = merchantID
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("ERROR: rows iteration error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "error during data retrieval", err.Error())
		return
	}

	resp := WorkflowsListResponse{
		Data:   items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func getWorkflowByID(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, idStr string) {
	// Query main workflow summary
	workflowSQL := `
		SELECT 
			rw.id::text,
			rw.payment_id::text,
			rw.merchant_id::text,
			rw.status,
			COALESCE(rw.recovery_probability, 0)::float8,
			COALESCE(rw.selected_action, ''),
			rw.scheduled_at,
			rw.created_at,
			rw.updated_at,
			p.amount::float8,
			p.currency,
			p.status,
			COALESCE(p.method, ''),
			COALESCE(p.failure_code, ''),
			COALESCE(p.razorpay_payment_id, ''),
			c.id::text,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			c.communication_opt_out,
			COALESCE((SELECT COUNT(*) FROM recovery_actions WHERE workflow_id = rw.id), 0),
			COALESCE((SELECT recovered FROM recovery_outcomes WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1), false),
			COALESCE((SELECT COUNT(*) FROM payments p2 WHERE (p2.customer_id = c.id OR (c.email != '' AND p2.customer_id IN (SELECT id FROM customers WHERE email = c.email))) AND p2.status IN ('CAPTURED', 'RECOVERED', 'SUCCESS')), 0),
			COALESCE((SELECT COUNT(*) FROM payments p2 WHERE (p2.customer_id = c.id OR (c.email != '' AND p2.customer_id IN (SELECT id FROM customers WHERE email = c.email))) AND p2.status = 'FAILED'), 0),
			COALESCE(rw.fraud_probability, 0)::float8,
			COALESCE(rw.return_probability, 0)::float8,
			COALESCE(rw.overall_risk, 'LOW'),
			COALESCE(rw.expected_loss, 0)::float8,
			COALESCE(rw.risk_action, 'ALLOW')
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		JOIN customers c ON p.customer_id = c.id
		WHERE rw.id::text = $1 OR p.razorpay_payment_id = $1 OR p.id::text = $1
		ORDER BY rw.created_at DESC
		LIMIT 1
	`

	var item WorkflowSummary
	var merchantID *string
	err := pool.QueryRow(ctx, workflowSQL, idStr).Scan(
		&item.ID,
		&item.PaymentID,
		&merchantID,
		&item.Status,
		&item.RecoveryProbability,
		&item.SelectedAction,
		&item.ScheduledAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Amount,
		&item.Currency,
		&item.PaymentStatus,
		&item.PaymentMethod,
		&item.FailureCode,
		&item.RazorpayPaymentID,
		&item.CustomerID,
		&item.CustomerEmail,
		&item.CustomerPhone,
		&item.CommunicationOptOut,
		&item.AttemptsCount,
		&item.IsRecovered,
		&item.CustomerSuccessCount,
		&item.CustomerFailedCount,
		&item.FraudProbability,
		&item.ReturnProbability,
		&item.OverallRisk,
		&item.ExpectedLoss,
		&item.RiskAction,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "workflow not found", fmt.Sprintf("no workflow matching identifier '%s'", idStr))
			return
		}
		log.Printf("ERROR: failed to get workflow '%s': %v", idStr, err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query workflow", err.Error())
		return
	}
	item.MerchantID = merchantID

	actualWfID := item.ID

	// 1. Fetch AI Decisions
	aiDecisions := make([]AIDecisionItem, 0)
	aiRows, err := pool.Query(ctx, `
		SELECT 
			id::text,
			workflow_id::text,
			provider,
			model,
			prompt_hash,
			raw_response,
			diagnosis,
			recommended_action,
			recommended_delay_hours,
			confidence::float8,
			recoverability::float8,
			reasoning,
			inference_duration_ms,
			created_at
		FROM ai_decisions
		WHERE workflow_id::text = $1
		ORDER BY created_at ASC
	`, actualWfID)
	if err == nil {
		defer aiRows.Close()
		for aiRows.Next() {
			var ai AIDecisionItem
			if err := aiRows.Scan(
				&ai.ID,
				&ai.WorkflowID,
				&ai.Provider,
				&ai.Model,
				&ai.PromptHash,
				&ai.RawResponse,
				&ai.Diagnosis,
				&ai.RecommendedAction,
				&ai.RecommendedDelayHours,
				&ai.Confidence,
				&ai.Recoverability,
				&ai.Reasoning,
				&ai.InferenceDurationMs,
				&ai.CreatedAt,
			); err == nil {
				aiDecisions = append(aiDecisions, ai)
			}
		}
	}

	// 2. Fetch Model Predictions
	modelPredictions := make([]ModelPredictionItem, 0)
	mpRows, err := pool.Query(ctx, `
		SELECT 
			id::text,
			workflow_id::text,
			payment_id::text,
			model_version,
			probability::float8,
			failure_category,
			COALESCE(features_used, '{}'::jsonb),
			created_at
		FROM model_predictions
		WHERE workflow_id::text = $1
		ORDER BY created_at ASC
	`, actualWfID)
	if err == nil {
		defer mpRows.Close()
		for mpRows.Next() {
			var mp ModelPredictionItem
			var featBytes []byte
			if err := mpRows.Scan(
				&mp.ID,
				&mp.WorkflowID,
				&mp.PaymentID,
				&mp.ModelVersion,
				&mp.Probability,
				&mp.FailureCategory,
				&featBytes,
				&mp.CreatedAt,
			); err == nil {
				mp.FeaturesUsed = featBytes
				modelPredictions = append(modelPredictions, mp)
			}
		}
	}

	// 3. Fetch Recovery Actions
	recoveryActions := make([]RecoveryActionItem, 0)
	raRows, err := pool.Query(ctx, `
		SELECT 
			id::text,
			workflow_id::text,
			action_type,
			status,
			attempt,
			executed_at,
			result,
			created_at,
			updated_at
		FROM recovery_actions
		WHERE workflow_id::text = $1
		ORDER BY attempt ASC, created_at ASC
	`, actualWfID)
	if err == nil {
		defer raRows.Close()
		for raRows.Next() {
			var ra RecoveryActionItem
			if err := raRows.Scan(
				&ra.ID,
				&ra.WorkflowID,
				&ra.ActionType,
				&ra.Status,
				&ra.Attempt,
				&ra.ExecutedAt,
				&ra.Result,
				&ra.CreatedAt,
				&ra.UpdatedAt,
			); err == nil {
				recoveryActions = append(recoveryActions, ra)
			}
		}
	}

	// 4. Fetch Recovery Outcomes
	recoveryOutcomes := make([]RecoveryOutcomeItem, 0)
	roRows, err := pool.Query(ctx, `
		SELECT 
			id::text,
			action_id::text,
			payment_id::text,
			recovered,
			recovered_amount::float8,
			time_to_recovery::text,
			created_at
		FROM recovery_outcomes
		WHERE payment_id::text = $1
		ORDER BY created_at ASC
	`, item.PaymentID)
	if err == nil {
		defer roRows.Close()
		for roRows.Next() {
			var ro RecoveryOutcomeItem
			var actionID *string
			if err := roRows.Scan(
				&ro.ID,
				&actionID,
				&ro.PaymentID,
				&ro.Recovered,
				&ro.RecoveredAmount,
				&ro.TimeToRecovery,
				&ro.CreatedAt,
			); err == nil {
				ro.ActionID = actionID
				recoveryOutcomes = append(recoveryOutcomes, ro)
			}
		}
	}

	// 5. Fetch Audit Events
	auditEvents := make([]AuditEventItem, 0)
	auditRows, err := pool.Query(ctx, `
		SELECT 
			event_id::text,
			workflow_id::text,
			timestamp,
			actor,
			action,
			payload_hash,
			previous_event_hash,
			event_hash,
			COALESCE(metadata, '{}'::jsonb)
		FROM audit_events
		WHERE workflow_id::text = $1
		ORDER BY timestamp ASC
	`, actualWfID)
	if err == nil {
		defer auditRows.Close()
		for auditRows.Next() {
			var ae AuditEventItem
			var metaBytes []byte
			if err := auditRows.Scan(
				&ae.EventID,
				&ae.WorkflowID,
				&ae.Timestamp,
				&ae.Actor,
				&ae.Action,
				&ae.PayloadHash,
				&ae.PreviousEventHash,
				&ae.EventHash,
				&metaBytes,
			); err == nil {
				ae.Metadata = metaBytes
				auditEvents = append(auditEvents, ae)
			}
		}
	}

	// 6. Fetch Risk Assessments
	riskAssessments := make([]RiskAssessmentItem, 0)
	riskRows, err := pool.Query(ctx, `
		SELECT 
			id::text,
			payment_id::text,
			COALESCE(workflow_id::text, ''),
			event_type,
			fraud_probability::float8,
			fraud_risk_level,
			COALESCE(return_probability, 0)::float8,
			COALESCE(return_risk_level, 'LOW'),
			overall_risk_level,
			expected_loss::float8,
			recommended_action,
			COALESCE(reason, ''),
			COALESCE(model_version, 'fraud-rf-v1.0'),
			created_at
		FROM risk_assessments
		WHERE workflow_id::text = $1 OR payment_id::text = $2
		ORDER BY created_at ASC
	`, actualWfID, item.PaymentID)
	if err == nil {
		defer riskRows.Close()
		for riskRows.Next() {
			var ra RiskAssessmentItem
			if err := riskRows.Scan(
				&ra.ID,
				&ra.PaymentID,
				&ra.WorkflowID,
				&ra.EventType,
				&ra.FraudProbability,
				&ra.FraudRiskLevel,
				&ra.ReturnProbability,
				&ra.ReturnRiskLevel,
				&ra.OverallRiskLevel,
				&ra.ExpectedLoss,
				&ra.RecommendedAction,
				&ra.Reason,
				&ra.ModelVersion,
				&ra.CreatedAt,
			); err == nil {
				riskAssessments = append(riskAssessments, ra)
			}
		}
	}

	resp := WorkflowDetailResponse{
		Workflow:         item,
		AIDecisions:      aiDecisions,
		ModelPredictions: modelPredictions,
		RecoveryActions:  recoveryActions,
		RecoveryOutcomes: recoveryOutcomes,
		AuditEvents:      auditEvents,
		RiskAssessments:  riskAssessments,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSONError(w http.ResponseWriter, status int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]string{
		"error": message,
	}
	if detail != "" {
		body["detail"] = detail
	}
	_ = json.NewEncoder(w).Encode(body)
}

type ApproveWorkflowRequest struct {
	Action string `json:"action"`
	Notes  string `json:"notes"`
}

type RejectWorkflowRequest struct {
	Reason string `json:"reason"`
	Notes  string `json:"notes"`
}

type OverrideWorkflowRequest struct {
	Action     string `json:"action"`
	DelayHours int    `json:"delay_hours"`
	Notes      string `json:"notes"`
}

type InterventionItem struct {
	WorkflowSummary
	CustomerSuccessCount int     `json:"customer_success_count"`
	CustomerFailedCount  int     `json:"customer_failed_count"`
	LatestDiagnosis      string  `json:"latest_diagnosis"`
	LatestConfidence     float64 `json:"latest_confidence"`
	EscalationReason     string  `json:"escalation_reason"`
}

func getInterventions(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	_ = paymentprovider.SyncRazorpayPaymentLinks(ctx, pool)

	merchantFilter := r.URL.Query().Get("merchant_id")
	whereClause := "WHERE rw.status IN ('ESCALATED', 'REQUIRES_HUMAN_REVIEW')"
	var args []interface{}
	if merchantFilter != "" {
		whereClause += " AND rw.merchant_id::text = $1"
		args = append(args, merchantFilter)
	}

	querySQL := fmt.Sprintf(`
		SELECT 
			rw.id::text,
			rw.payment_id::text,
			rw.merchant_id::text,
			rw.status,
			COALESCE(rw.recovery_probability, 0)::float8,
			COALESCE(rw.selected_action, ''),
			rw.scheduled_at,
			rw.created_at,
			rw.updated_at,
			p.amount::float8,
			p.currency,
			p.status,
			COALESCE(p.method, ''),
			COALESCE(p.failure_code, ''),
			COALESCE(p.razorpay_payment_id, ''),
			c.id::text,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			c.communication_opt_out,
			COALESCE((SELECT COUNT(*) FROM recovery_actions WHERE workflow_id = rw.id), 0),
			COALESCE((SELECT recovered FROM recovery_outcomes WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1), false),
			COALESCE((SELECT COUNT(*) FROM payments p2 WHERE p2.customer_id = c.id AND p2.status IN ('CAPTURED', 'RECOVERED', 'SUCCESS')), 0),
			COALESCE((SELECT COUNT(*) FROM payments p2 WHERE p2.customer_id = c.id AND p2.status = 'FAILED'), 0),
			COALESCE(rw.fraud_probability, 0)::float8,
			COALESCE(rw.return_probability, 0)::float8,
			COALESCE(rw.overall_risk, 'LOW'),
			COALESCE(rw.expected_loss, 0)::float8,
			COALESCE(rw.risk_action, 'ALLOW'),
			COALESCE((SELECT diagnosis FROM ai_decisions WHERE workflow_id = rw.id ORDER BY created_at DESC LIMIT 1), ''),
			COALESCE((SELECT confidence FROM ai_decisions WHERE workflow_id = rw.id ORDER BY created_at DESC LIMIT 1), 0.75)::float8,
			COALESCE((SELECT metadata->>'reason' FROM audit_events WHERE workflow_id = rw.id AND action = 'POLICY_EVALUATED_ESCALATE' ORDER BY timestamp DESC LIMIT 1), 'Flagged for human operator review')
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		JOIN customers c ON p.customer_id = c.id
		%s
		ORDER BY rw.created_at DESC
	`, whereClause)

	rows, err := pool.Query(ctx, querySQL, args...)
	if err != nil {
		log.Printf("ERROR: failed to query interventions: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query interventions", err.Error())
		return
	}
	defer rows.Close()

	items := make([]InterventionItem, 0)
	for rows.Next() {
		var item InterventionItem
		var schedAt *time.Time
		err := rows.Scan(
			&item.ID,
			&item.PaymentID,
			&item.MerchantID,
			&item.Status,
			&item.RecoveryProbability,
			&item.SelectedAction,
			&schedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Amount,
			&item.Currency,
			&item.PaymentStatus,
			&item.PaymentMethod,
			&item.FailureCode,
			&item.RazorpayPaymentID,
			&item.CustomerID,
			&item.CustomerEmail,
			&item.CustomerPhone,
			&item.CommunicationOptOut,
			&item.AttemptsCount,
			&item.IsRecovered,
			&item.CustomerSuccessCount,
			&item.CustomerFailedCount,
			&item.FraudProbability,
			&item.ReturnProbability,
			&item.OverallRisk,
			&item.ExpectedLoss,
			&item.RiskAction,
			&item.LatestDiagnosis,
			&item.LatestConfidence,
			&item.EscalationReason,
		)
		if err == nil {
			item.ScheduledAt = schedAt
			items = append(items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  items,
		"total": len(items),
	})
}

func approveWorkflow(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request, wfID string) {
	var req ApproveWorkflowRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var (
		actualWfID    string
		paymentID     string
		merchantID    string
		wfStatus      string
		currentAction string
		amount        float64
		currency      string
		custEmail     string
		custPhone     string
		optOut        bool
		attemptsCount int
	)

	lookupErr := pool.QueryRow(ctx, `
		SELECT 
			rw.id::text,
			rw.payment_id::text,
			rw.merchant_id::text,
			rw.status,
			COALESCE(rw.selected_action, 'PAYMENT_LINK'),
			p.amount::float8,
			p.currency,
			COALESCE(c.email, ''),
			COALESCE(c.phone, ''),
			c.communication_opt_out,
			COALESCE((SELECT COUNT(*) FROM recovery_actions WHERE workflow_id = rw.id), 0)
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		JOIN customers c ON p.customer_id = c.id
		WHERE rw.id::text = $1 OR rw.payment_id::text = $1
		LIMIT 1
	`, wfID).Scan(
		&actualWfID,
		&paymentID,
		&merchantID,
		&wfStatus,
		&currentAction,
		&amount,
		&currency,
		&custEmail,
		&custPhone,
		&optOut,
		&attemptsCount,
	)

	if lookupErr != nil {
		writeJSONError(w, http.StatusNotFound, "workflow not found", lookupErr.Error())
		return
	}

	actionToExecute := req.Action
	if actionToExecute == "" {
		actionToExecute = currentAction
	}
	if actionToExecute == "" {
		actionToExecute = "PAYMENT_LINK"
	}

	if optOut && (actionToExecute == "PAYMENT_LINK" || actionToExecute == "CUSTOMER_NOTIFICATION") {
		writeJSONError(w, http.StatusBadRequest, "customer opted out", "customer has opted out of communication channels")
		return
	}

	// Trigger recovery via PaymentProvider
	provider, _ := paymentprovider.NewPaymentProvider("", pool)

	retryResult, retryErr := provider.CreateRetryAttemptWithCustomer(
		ctx,
		paymentID,
		amount,
		custEmail,
		custPhone,
		"",
	)

	currentAttempt := attemptsCount + 1
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
		actionResultStr = retryResult.PaymentLinkURL
		if actionResultStr == "" {
			actionResultStr = fmt.Sprintf("Attempt ID: %s", retryResult.AttemptID)
		}
	}

	// Record action
	var actionUUID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO recovery_actions (workflow_id, action_type, status, attempt, result, executed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id::text
	`, actualWfID, actionToExecute, actionStatus, currentAttempt, actionResultStr).Scan(&actionUUID)

	// Update workflow status to SCHEDULED / EXECUTING
	nextStatus := "SCHEDULED"
	if actionStatus == "EXECUTED" {
		nextStatus = "SCHEDULED"
	}
	_, _ = pool.Exec(ctx, `
		UPDATE recovery_workflows
		SET status = $1,
		    selected_action = $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id::text = $3
	`, nextStatus, actionToExecute, actualWfID)

	// Append Audit Log
	_ = audit.AppendAuditLog(ctx, pool, audit.AuditEvent{
		WorkflowID: actualWfID,
		Actor:      "human:operator",
		Action:     "WORKFLOW_MANUALLY_APPROVED",
		Metadata: map[string]interface{}{
			"operator":          "admin",
			"action_executed":   actionToExecute,
			"notes":             req.Notes,
			"payment_id":        paymentID,
			"attempt":           currentAttempt,
			"payment_link_url":  actionResultStr,
			"previous_status":   wfStatus,
		},
	})

	log.Printf("[HumanIntervention] Approved workflow %s for payment %s (Action: %s)", actualWfID, paymentID, actionToExecute)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "approved",
		"workflow_id":     actualWfID,
		"new_status":      nextStatus,
		"action_executed": actionToExecute,
		"result":          actionResultStr,
		"message":         "Workflow approved and recovery action executed successfully",
	})
}

func rejectWorkflow(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request, wfID string) {
	var req RejectWorkflowRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var actualWfID, paymentID, prevStatus string
	err := pool.QueryRow(ctx, `
		SELECT id::text, payment_id::text, status
		FROM recovery_workflows
		WHERE id::text = $1 OR payment_id::text = $1
		LIMIT 1
	`, wfID).Scan(&actualWfID, &paymentID, &prevStatus)

	if err != nil {
		writeJSONError(w, http.StatusNotFound, "workflow not found", err.Error())
		return
	}

	_, err = pool.Exec(ctx, `
		UPDATE recovery_workflows
		SET status = 'HALTED',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id::text = $1
	`, actualWfID)

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to halt workflow", err.Error())
		return
	}

	_ = audit.AppendAuditLog(ctx, pool, audit.AuditEvent{
		WorkflowID: actualWfID,
		Actor:      "human:operator",
		Action:     "WORKFLOW_MANUALLY_REJECTED",
		Metadata: map[string]interface{}{
			"operator":        "admin",
			"reason":          req.Reason,
			"notes":           req.Notes,
			"previous_status": prevStatus,
		},
	})

	log.Printf("[HumanIntervention] Rejected workflow %s (Status: HALTED)", actualWfID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "rejected",
		"workflow_id": actualWfID,
		"new_status":  "HALTED",
		"message":     "Workflow rejected and recovery halted",
	})
}

func overrideWorkflow(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request, wfID string) {
	var req OverrideWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Action == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid request", "action is required")
		return
	}

	var actualWfID string
	err := pool.QueryRow(ctx, "SELECT id::text FROM recovery_workflows WHERE id::text = $1 OR payment_id::text = $1 LIMIT 1", wfID).Scan(&actualWfID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "workflow not found", err.Error())
		return
	}

	_, err = pool.Exec(ctx, `
		UPDATE recovery_workflows
		SET selected_action = $1,
		    status = 'SCHEDULED',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id::text = $2
	`, req.Action, actualWfID)

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update workflow strategy", err.Error())
		return
	}

	_ = audit.AppendAuditLog(ctx, pool, audit.AuditEvent{
		WorkflowID: actualWfID,
		Actor:      "human:operator",
		Action:     "WORKFLOW_STRATEGY_OVERRIDDEN",
		Metadata: map[string]interface{}{
			"operator":    "admin",
			"new_action":  req.Action,
			"delay_hours": req.DelayHours,
			"notes":       req.Notes,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "overridden",
		"workflow_id": actualWfID,
		"new_action":  req.Action,
		"message":     "Workflow strategy updated successfully",
	})
}
