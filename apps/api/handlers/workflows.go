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
}

type WorkflowsListResponse struct {
	Data   []WorkflowSummary `json:"data"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

func WorkflowsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "")
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/workflows")
		path = strings.TrimPrefix(path, "/")

		// If a specific ID is in URL path or query parameter, return single workflow
		if path != "" {
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
			COALESCE((SELECT recovered FROM recovery_outcomes WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1), false)
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
			COALESCE((SELECT recovered FROM recovery_outcomes WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1), false)
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

	resp := WorkflowDetailResponse{
		Workflow:         item,
		AIDecisions:      aiDecisions,
		ModelPredictions: modelPredictions,
		RecoveryActions:  recoveryActions,
		RecoveryOutcomes: recoveryOutcomes,
		AuditEvents:      auditEvents,
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
