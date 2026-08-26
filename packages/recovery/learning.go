package recovery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LearningStats represents empirical recovery performance aggregated from real database records.
type LearningStats struct {
	TotalWorkflows       int     `json:"total_workflows"`
	RecoveredWorkflows   int     `json:"recovered_workflows"`
	FailedWorkflows      int     `json:"failed_workflows"`
	OverallRecoveryRate  float64 `json:"overall_recovery_rate"`
	TotalRecoveredAmount float64 `json:"total_recovered_amount"`
	CategoryStats        map[string]CategoryMetric `json:"category_stats"`
	ActionStats          map[string]ActionMetric   `json:"action_stats"`
}

type CategoryMetric struct {
	TotalCount   int     `json:"total_count"`
	Recovered    int     `json:"recovered"`
	RecoveryRate float64 `json:"recovery_rate"`
}

type ActionMetric struct {
	TotalCount   int     `json:"total_count"`
	Successful   int     `json:"successful"`
	SuccessRate  float64 `json:"success_rate"`
}

// LearningEngine aggregates real recovery feedback from PostgreSQL to calibrate models.
type LearningEngine struct {
	pool *pgxpool.Pool
}

// NewLearningEngine creates a new LearningEngine instance.
func NewLearningEngine(pool *pgxpool.Pool) *LearningEngine {
	return &LearningEngine{pool: pool}
}

// ComputeLearningStats queries the database for real recovery outcomes.
func (le *LearningEngine) ComputeLearningStats(ctx context.Context, merchantID string) (*LearningStats, error) {
	if le.pool == nil {
		return nil, fmt.Errorf("database connection pool is nil")
	}

	stats := &LearningStats{
		CategoryStats: make(map[string]CategoryMetric),
		ActionStats:   make(map[string]ActionMetric),
	}

	// 1. Overall Workflow Outcomes
	var (
		totalWorkflows   int
		recoveredCount   int
		failedCount      int
		totalAmount      float64
	)

	filterClause := ""
	args := []interface{}{}
	if merchantID != "" {
		filterClause = "WHERE p.merchant_id::text = $1"
		args = append(args, merchantID)
	}

	query := fmt.Sprintf(`
		SELECT 
			COALESCE(COUNT(rw.id), 0) as total,
			COALESCE(COUNT(CASE WHEN rw.status = 'RECOVERED' THEN 1 END), 0) as recovered,
			COALESCE(COUNT(CASE WHEN rw.status = 'FAILED' THEN 1 END), 0) as failed,
			COALESCE(SUM(CASE WHEN ro.recovered = true THEN ro.recovered_amount ELSE 0 END), 0)::float8 as total_amount
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		LEFT JOIN recovery_outcomes ro ON ro.payment_id = p.id
		%s
	`, filterClause)

	err := le.pool.QueryRow(ctx, query, args...).Scan(
		&totalWorkflows,
		&recoveredCount,
		&failedCount,
		&totalAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query overall learning stats: %w", err)
	}

	stats.TotalWorkflows = totalWorkflows
	stats.RecoveredWorkflows = recoveredCount
	stats.FailedWorkflows = failedCount
	stats.TotalRecoveredAmount = totalAmount

	if totalWorkflows > 0 {
		stats.OverallRecoveryRate = float64(recoveredCount) / float64(totalWorkflows)
	}

	// 2. Failure Category Recovery Rates
	catQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(p.failure_code, 'UNKNOWN') as failure_code,
			COUNT(DISTINCT rw.id) as total,
			COUNT(DISTINCT CASE WHEN rw.status = 'RECOVERED' THEN rw.id END) as recovered
		FROM recovery_workflows rw
		JOIN payments p ON rw.payment_id = p.id
		%s
		GROUP BY p.failure_code
	`, filterClause)

	catRows, err := le.pool.Query(ctx, catQuery, args...)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var code string
			var tot, rec int
			if err := catRows.Scan(&code, &tot, &rec); err == nil {
				rate := 0.0
				if tot > 0 {
					rate = float64(rec) / float64(tot)
				}
				stats.CategoryStats[code] = CategoryMetric{
					TotalCount:   tot,
					Recovered:    rec,
					RecoveryRate: rate,
				}
			}
		}
	}

	// 3. Action Type Success Rates
	actionQuery := fmt.Sprintf(`
		SELECT 
			ra.action_type,
			COUNT(ra.id) as total_actions,
			COUNT(CASE WHEN ra.status = 'EXECUTED' AND ro.recovered = true THEN 1 END) as successful
		FROM recovery_actions ra
		JOIN recovery_workflows rw ON ra.workflow_id = rw.id
		JOIN payments p ON rw.payment_id = p.id
		LEFT JOIN recovery_outcomes ro ON ro.action_id = ra.id
		%s
		GROUP BY ra.action_type
	`, filterClause)

	actRows, err := le.pool.Query(ctx, actionQuery, args...)
	if err == nil {
		defer actRows.Close()
		for actRows.Next() {
			var act string
			var tot, succ int
			if err := actRows.Scan(&act, &tot, &succ); err == nil {
				rate := 0.0
				if tot > 0 {
					rate = float64(succ) / float64(tot)
				}
				stats.ActionStats[act] = ActionMetric{
					TotalCount:  tot,
					Successful:  succ,
					SuccessRate: rate,
				}
			}
		}
	}

	return stats, nil
}
