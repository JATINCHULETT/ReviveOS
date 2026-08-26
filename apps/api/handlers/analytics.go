package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecentOutcomeItem struct {
	PaymentID       string    `json:"payment_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	FailureCode     string    `json:"failure_code"`
	ActionType      string    `json:"action_type"`
	Status          string    `json:"status"`
	Recovered       bool      `json:"recovered"`
	RecoveredAmount float64   `json:"recovered_amount"`
	CreatedAt       time.Time `json:"created_at"`
}

type AnalyticsOverviewResponse struct {
	TotalPayments      int                 `json:"total_payments"`
	FailedPayments     int                 `json:"failed_payments"`
	TotalWorkflows     int                 `json:"total_workflows"`
	RecoveredWorkflows int                 `json:"recovered_workflows"`
	RecoveryRate       float64             `json:"recovery_rate"`
	RecoveredRevenue   float64             `json:"recovered_revenue"`
	TotalAtRiskRevenue float64             `json:"total_at_risk_revenue"`
	ActiveRecoveries   int                 `json:"active_recoveries"`
	PendingActions     int                 `json:"pending_actions"`
	CategoryBreakdown  map[string]int      `json:"category_breakdown"`
	RecentOutcomes     []RecentOutcomeItem `json:"recent_outcomes"`
}

func AnalyticsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "")
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/analytics")
		path = strings.TrimPrefix(path, "/")

		switch path {
		case "overview", "":
			getAnalyticsOverview(r.Context(), pool, w, r)
		default:
			writeJSONError(w, http.StatusNotFound, "analytics endpoint not found", fmt.Sprintf("no analytics view '%s'", path))
		}
	}
}

func getAnalyticsOverview(ctx context.Context, pool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	merchantFilter := r.URL.Query().Get("merchant_id")

	var query string
	var args []interface{}

	if merchantFilter != "" {
		query = `
			SELECT 
				COUNT(DISTINCT p.id) as total_payments,
				COUNT(DISTINCT CASE WHEN p.status = 'FAILED' THEN p.id END) as failed_payments,
				COUNT(DISTINCT rw.id) as total_workflows,
				COUNT(DISTINCT CASE WHEN rw.status = 'RECOVERED' OR ro.recovered = true THEN rw.id END) as recovered_workflows,
				COALESCE(SUM(CASE WHEN ro.recovered = true THEN ro.recovered_amount ELSE 0 END), 0)::float8 as recovered_revenue,
				COALESCE(SUM(CASE WHEN p.status = 'FAILED' THEN p.amount ELSE 0 END), 0)::float8 as total_at_risk_revenue,
				COUNT(DISTINCT CASE WHEN rw.status IN ('ANALYZING', 'PLANNED', 'SCHEDULED', 'EXECUTING', 'VERIFYING', 'PENDING_VERIFICATION', 'REQUIRES_HUMAN_REVIEW', 'ESCALATED') THEN rw.id END) as active_recoveries,
				(SELECT COUNT(*) FROM recovery_actions ra JOIN recovery_workflows rw2 ON ra.workflow_id = rw2.id WHERE ra.status = 'PENDING' AND rw2.merchant_id::text = $1) as pending_actions
			FROM payments p
			LEFT JOIN recovery_workflows rw ON rw.payment_id = p.id
			LEFT JOIN recovery_outcomes ro ON ro.payment_id = p.id
			WHERE p.merchant_id::text = $1
		`
		args = append(args, merchantFilter)
	} else {
		query = `
			SELECT 
				COUNT(DISTINCT p.id) as total_payments,
				COUNT(DISTINCT CASE WHEN p.status = 'FAILED' THEN p.id END) as failed_payments,
				COUNT(DISTINCT rw.id) as total_workflows,
				COUNT(DISTINCT CASE WHEN rw.status = 'RECOVERED' OR ro.recovered = true THEN rw.id END) as recovered_workflows,
				COALESCE(SUM(CASE WHEN ro.recovered = true THEN ro.recovered_amount ELSE 0 END), 0)::float8 as recovered_revenue,
				COALESCE(SUM(CASE WHEN p.status = 'FAILED' THEN p.amount ELSE 0 END), 0)::float8 as total_at_risk_revenue,
				COUNT(DISTINCT CASE WHEN rw.status IN ('ANALYZING', 'PLANNED', 'SCHEDULED', 'EXECUTING', 'VERIFYING', 'PENDING_VERIFICATION', 'REQUIRES_HUMAN_REVIEW', 'ESCALATED') THEN rw.id END) as active_recoveries,
				(SELECT COUNT(*) FROM recovery_actions WHERE status = 'PENDING') as pending_actions
			FROM payments p
			LEFT JOIN recovery_workflows rw ON rw.payment_id = p.id
			LEFT JOIN recovery_outcomes ro ON ro.payment_id = p.id
		`
	}

	var resp AnalyticsOverviewResponse
	resp.CategoryBreakdown = make(map[string]int)
	resp.RecentOutcomes = make([]RecentOutcomeItem, 0)

	err := pool.QueryRow(ctx, query, args...).Scan(
		&resp.TotalPayments,
		&resp.FailedPayments,
		&resp.TotalWorkflows,
		&resp.RecoveredWorkflows,
		&resp.RecoveredRevenue,
		&resp.TotalAtRiskRevenue,
		&resp.ActiveRecoveries,
		&resp.PendingActions,
	)
	if err != nil {
		log.Printf("ERROR: failed to query analytics overview: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query analytics overview", err.Error())
		return
	}

	if resp.TotalWorkflows > 0 {
		resp.RecoveryRate = float64(resp.RecoveredWorkflows) / float64(resp.TotalWorkflows)
	} else if resp.FailedPayments > 0 {
		resp.RecoveryRate = float64(resp.RecoveredWorkflows) / float64(resp.FailedPayments)
	} else {
		resp.RecoveryRate = 0.0
	}

	// Fetch category breakdown
	catQuery := `
		SELECT COALESCE(p.failure_code, 'UNKNOWN'), COUNT(p.id)
		FROM payments p
		GROUP BY p.failure_code
		ORDER BY COUNT(p.id) DESC
	`
	catRows, err := pool.Query(ctx, catQuery)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var code string
			var count int
			if err := catRows.Scan(&code, &count); err == nil && code != "" {
				resp.CategoryBreakdown[code] = count
			}
		}
	}

	// Fetch recent outcomes
	outcomeQuery := `
		SELECT 
			p.id::text,
			p.amount,
			p.currency,
			COALESCE(p.failure_code, 'UNKNOWN'),
			COALESCE(rw.selected_action, 'RETRY'),
			p.status,
			COALESCE(ro.recovered, false),
			COALESCE(ro.recovered_amount, 0),
			COALESCE(ro.created_at, p.created_at)
		FROM payments p
		LEFT JOIN recovery_workflows rw ON rw.payment_id = p.id
		LEFT JOIN recovery_outcomes ro ON ro.payment_id = p.id
		ORDER BY p.created_at DESC
		LIMIT 10
	`
	outRows, err := pool.Query(ctx, outcomeQuery)
	if err == nil {
		defer outRows.Close()
		for outRows.Next() {
			var item RecentOutcomeItem
			if err := outRows.Scan(
				&item.PaymentID,
				&item.Amount,
				&item.Currency,
				&item.FailureCode,
				&item.ActionType,
				&item.Status,
				&item.Recovered,
				&item.RecoveredAmount,
				&item.CreatedAt,
			); err == nil {
				resp.RecentOutcomes = append(resp.RecentOutcomes, item)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
