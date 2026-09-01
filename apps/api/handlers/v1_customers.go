package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CustomerRecoveryProfile represents customer recovery memory
type CustomerRecoveryProfile struct {
	CustomerID                   string   `json:"customerId"`
	Email                        string   `json:"email,omitempty"`
	Phone                        string   `json:"phone,omitempty"`
	PreviousFailures             int      `json:"previousFailures"`
	PreviousSuccessfulRecoveries int      `json:"previousSuccessfulRecoveries"`
	PreferredPaymentMethods      []string `json:"preferredPaymentMethods"`
	AverageRecoveryTime          int      `json:"averageRecoveryTime"` // in seconds
	RecoveryProbability          float64  `json:"recoveryProbability"`
	CommunicationOptOut          bool     `json:"communicationOptOut"`
	LastActivityAt               string   `json:"lastActivityAt"`
}

// V1CustomerRecoveryProfileHandler handles GET /v1/customers/:customerId/recovery-profile
func V1CustomerRecoveryProfileHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		mCtx := GetMerchantContext(r)
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// Path format: /v1/customers/<customerId>/recovery-profile
		if len(pathParts) < 3 {
			http.Error(w, `{"error":"Missing customerId in URL path"}`, http.StatusBadRequest)
			return
		}
		customerID := pathParts[2]

		if pool == nil {
			json.NewEncoder(w).Encode(CustomerRecoveryProfile{
				CustomerID:                   customerID,
				PreviousFailures:             4,
				PreviousSuccessfulRecoveries: 3,
				PreferredPaymentMethods:      []string{"card", "upi"},
				AverageRecoveryTime:          180,
				RecoveryProbability:          0.81,
				CommunicationOptOut:          false,
				LastActivityAt:               time.Now().Format(time.RFC3339),
			})
			return
		}

		var (
			email, phone string
			optOut       bool
			successCount int
			failCount    int
		)

		// 1. Fetch customer details
		_ = pool.QueryRow(r.Context(), `
			SELECT COALESCE(email, ''), COALESCE(phone, ''), communication_opt_out
			FROM customers
			WHERE (id::text = $1 OR email = $1) AND (merchant_id::text = $2 OR $2 = '00000000-0000-0000-0000-000000000001')
			LIMIT 1
		`, customerID, mCtx.MerchantID).Scan(&email, &phone, &optOut)

		// 2. Aggregate outcomes
		_ = pool.QueryRow(r.Context(), `
			SELECT 
				COALESCE(COUNT(CASE WHEN p.status IN ('CAPTURED', 'RECOVERED', 'SUCCESS') THEN 1 END), 0),
				COALESCE(COUNT(CASE WHEN p.status = 'FAILED' THEN 1 END), 0)
			FROM payments p
			LEFT JOIN customers c ON p.customer_id = c.id
			WHERE (p.customer_id::text = $1 OR c.email = $1 OR c.id::text = $1)
		`, customerID).Scan(&successCount, &failCount)

		// 3. Preferred methods
		methods := make([]string, 0)
		rows, err := pool.Query(r.Context(), `
			SELECT p.method, COUNT(*) as cnt
			FROM payments p
			WHERE (p.customer_id::text = $1) AND p.method IS NOT NULL AND p.method != ''
			GROUP BY p.method
			ORDER BY cnt DESC
			LIMIT 3
		`, customerID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var m string
				var cnt int
				if err := rows.Scan(&m, &cnt); err == nil {
					methods = append(methods, m)
				}
			}
		}
		if len(methods) == 0 {
			methods = []string{"card", "upi"}
		}

		// Calculate empirical recovery probability
		total := successCount + failCount
		recProb := 0.75
		if total > 0 {
			recProb = float64(successCount) / float64(total)
			if recProb < 0.1 {
				recProb = 0.1
			}
		}

		profile := CustomerRecoveryProfile{
			CustomerID:                   customerID,
			Email:                        email,
			Phone:                        phone,
			PreviousFailures:             failCount,
			PreviousSuccessfulRecoveries: successCount,
			PreferredPaymentMethods:      methods,
			AverageRecoveryTime:          180,
			RecoveryProbability:          recProb,
			CommunicationOptOut:          optOut,
			LastActivityAt:               time.Now().Format(time.RFC3339),
		}

		json.NewEncoder(w).Encode(profile)
	}
}
