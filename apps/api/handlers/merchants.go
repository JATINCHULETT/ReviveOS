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
	notificationservice "github.com/reviveos/services/notification-service"
	paymentprovider "github.com/reviveos/services/payment-provider"
	"github.com/reviveos/utils/audit"
	"github.com/reviveos/utils/outbox"
)

type MerchantSummary struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	TotalCustomers      int       `json:"total_customers"`
	ActiveSubscriptions int       `json:"active_subscriptions"`
	FailedPayments      int       `json:"failed_payments"`
	RecoveredPayments   int       `json:"recovered_payments"`
	RecoveryRate        float64   `json:"recovery_rate"`
	CreatedAt           time.Time `json:"created_at"`
}

type MerchantDashboardData struct {
	Merchant struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"merchant"`
	Metrics struct {
		ActiveSubscriptions int     `json:"active_subscriptions"`
		TotalSubscriptions  int     `json:"total_subscriptions"`
		TotalAtRiskRevenue  float64 `json:"total_at_risk_revenue"`
		TotalRecovered      float64 `json:"total_recovered_revenue"`
		RecoveryRate        float64 `json:"recovery_rate"`
		PendingRecoveries   int     `json:"pending_recoveries"`
	} `json:"metrics"`
	Subscriptions []SubscriptionItem `json:"subscriptions"`
	Customers     []CustomerRecoveryItem `json:"customers"`
	SandboxLinks  []SandboxLinkItem  `json:"sandbox_links"`
}

type SubscriptionItem struct {
	ID                     string    `json:"id"`
	CustomerID             string    `json:"customer_id"`
	CustomerEmail          string    `json:"customer_email"`
	Amount                 float64   `json:"amount"`
	Currency               string    `json:"currency"`
	Status                 string    `json:"status"` // ACTIVE, PAST_DUE, RECOVERING, CANCELED
	BillingInterval        string    `json:"billing_interval"`
	PlanID                 string    `json:"plan_id"`
	RazorpaySubscriptionID string    `json:"razorpay_subscription_id,omitempty"`
	PaymentLinkURL         string    `json:"payment_link_url,omitempty"`
	NextBillingAt          *time.Time `json:"next_billing_at,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}

type CustomerRecoveryItem struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	OptOut         bool      `json:"communication_opt_out"`
	TotalPayments  int       `json:"total_payments"`
	FailedCount    int       `json:"failed_count"`
	RecoveredCount int       `json:"recovered_count"`
	LastStatus     string    `json:"last_status"`
	LastAction     string    `json:"last_action"`
	LastSeen       time.Time `json:"last_seen"`
}

type SandboxLinkItem struct {
	PaymentID      string    `json:"payment_id"`
	CustomerEmail  string    `json:"customer_email"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	PaymentLinkURL string    `json:"payment_link_url"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// MerchantsListHandler handles GET /merchants (Admin view) and POST /merchants
func MerchantsListHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method == http.MethodPost {
			var body struct {
				Name                string  `json:"name"`
				MaxRetries          int     `json:"max_retries"`
				MaxContacts         int     `json:"max_contacts"`
				ConfidenceThreshold float64 `json:"confidence_threshold"`
				AmountThreshold     float64 `json:"amount_threshold"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Name) == "" {
				http.Error(w, `{"error":"Valid merchant name is required"}`, http.StatusBadRequest)
				return
			}

			tx, err := pool.Begin(ctx)
			if err != nil {
				http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
				return
			}
			defer tx.Rollback(ctx)

			var merchantID string
			err = tx.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", body.Name).Scan(&merchantID)
			if err != nil {
				http.Error(w, `{"error":"Failed to create merchant"}`, http.StatusInternalServerError)
				return
			}

			maxRetries := 3
			if body.MaxRetries > 0 {
				maxRetries = body.MaxRetries
			}
			maxContacts := 2
			if body.MaxContacts > 0 {
				maxContacts = body.MaxContacts
			}
			confThresh := 0.70
			if body.ConfidenceThreshold > 0 {
				confThresh = body.ConfidenceThreshold
			}
			amtThresh := 50000.0
			if body.AmountThreshold > 0 {
				amtThresh = body.AmountThreshold
			}

			_, _ = tx.Exec(ctx, `
				INSERT INTO policies (merchant_id, max_retries, max_contacts, confidence_threshold, amount_threshold)
				VALUES ($1, $2, $3, $4, $5)
			`, merchantID, maxRetries, maxContacts, confThresh, amtThresh)

			if err := tx.Commit(ctx); err != nil {
				http.Error(w, `{"error":"Failed to commit transaction"}`, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   merchantID,
				"name": body.Name,
			})
			return
		}

		// GET /merchants: List all merchants with aggregated recovery statistics
		query := `
			SELECT 
				m.id::text,
				m.name,
				m.created_at,
				(SELECT COUNT(*) FROM customers c WHERE c.merchant_id = m.id) as total_customers,
				(SELECT COUNT(*) FROM subscriptions s WHERE s.merchant_id = m.id AND s.status = 'ACTIVE') as active_subs,
				(SELECT COUNT(*) FROM payments p WHERE p.merchant_id = m.id AND p.status = 'FAILED') as failed_pmts,
				(SELECT COUNT(*) FROM payments p JOIN recovery_outcomes ro ON ro.payment_id = p.id WHERE p.merchant_id = m.id AND ro.recovered = true) as rec_pmts
			FROM merchants m
			ORDER BY m.created_at ASC
		`

		rows, err := pool.Query(ctx, query)
		if err != nil {
			log.Printf("[Merchants] Query error: %v", err)
			http.Error(w, `{"error":"Failed to query merchants"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		merchants := make([]MerchantSummary, 0)
		for rows.Next() {
			var m MerchantSummary
			if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt, &m.TotalCustomers, &m.ActiveSubscriptions, &m.FailedPayments, &m.RecoveredPayments); err != nil {
				continue
			}
			tot := m.FailedPayments + m.RecoveredPayments
			if tot > 0 {
				m.RecoveryRate = (float64(m.RecoveredPayments) / float64(tot)) * 100.0
			}
			merchants = append(merchants, m)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(merchants)
	}
}

// MerchantDashboardHandler returns detailed metrics, subscriptions, and recovery workflows scoped to a merchant.
func MerchantDashboardHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		merchantID := r.URL.Query().Get("merchant_id")

		// If no merchant_id param, get first merchant or default
		if merchantID == "" {
			_ = pool.QueryRow(ctx, "SELECT id::text FROM merchants ORDER BY created_at ASC LIMIT 1").Scan(&merchantID)
		}

		if merchantID == "" {
			http.Error(w, `{"error":"No merchant found"}`, http.StatusNotFound)
			return
		}

		var data MerchantDashboardData
		data.Subscriptions = make([]SubscriptionItem, 0)
		data.Customers = make([]CustomerRecoveryItem, 0)
		data.SandboxLinks = make([]SandboxLinkItem, 0)

		// 1. Fetch Merchant Details
		err := pool.QueryRow(ctx, "SELECT id::text, name, created_at FROM merchants WHERE id::text = $1", merchantID).
			Scan(&data.Merchant.ID, &data.Merchant.Name, &data.Merchant.CreatedAt)
		if err != nil {
			http.Error(w, `{"error":"Merchant not found"}`, http.StatusNotFound)
			return
		}

		// 2. Fetch Subscriptions
		subQuery := `
			SELECT 
				s.id::text,
				s.customer_id::text,
				COALESCE(c.email, 'customer@example.com'),
				s.amount::float8,
				s.currency,
				s.status,
				COALESCE(s.billing_interval, 'monthly'),
				COALESCE(s.plan_id, 'plan_pro_tier'),
				COALESCE(s.razorpay_subscription_id, ''),
				COALESCE(s.payment_link_url, ''),
				s.next_billing_at,
				s.created_at
			FROM subscriptions s
			LEFT JOIN customers c ON s.customer_id = c.id
			WHERE s.merchant_id::text = $1
			ORDER BY s.created_at DESC
			LIMIT 50
		`
		subRows, err := pool.Query(ctx, subQuery, merchantID)
		if err == nil {
			defer subRows.Close()
			for subRows.Next() {
				var item SubscriptionItem
				if err := subRows.Scan(
					&item.ID, &item.CustomerID, &item.CustomerEmail, &item.Amount, &item.Currency,
					&item.Status, &item.BillingInterval, &item.PlanID, &item.RazorpaySubscriptionID,
					&item.PaymentLinkURL, &item.NextBillingAt, &item.CreatedAt,
				); err == nil {
					data.Subscriptions = append(data.Subscriptions, item)
					if item.Status == "ACTIVE" {
						data.Metrics.ActiveSubscriptions++
					}
					data.Metrics.TotalSubscriptions++
				}
			}
		}

		// 3. Fetch Customers & Recovery Status
		custQuery := `
			SELECT 
				c.id::text,
				COALESCE(c.email, 'N/A'),
				COALESCE(c.phone, 'N/A'),
				c.communication_opt_out,
				(SELECT COUNT(*) FROM payments p WHERE p.customer_id = c.id) as tot_pmts,
				(SELECT COUNT(*) FROM payments p WHERE p.customer_id = c.id AND p.status = 'FAILED') as failed_pmts,
				(SELECT COUNT(*) FROM payments p JOIN recovery_outcomes ro ON ro.payment_id = p.id WHERE p.customer_id = c.id AND ro.recovered = true) as rec_pmts,
				COALESCE((SELECT p.status FROM payments p WHERE p.customer_id = c.id ORDER BY p.created_at DESC LIMIT 1), 'NEW') as last_status,
				COALESCE((SELECT rw.selected_action FROM payments p JOIN recovery_workflows rw ON rw.payment_id = p.id WHERE p.customer_id = c.id ORDER BY p.created_at DESC LIMIT 1), 'N/A') as last_action,
				c.updated_at
			FROM customers c
			WHERE c.merchant_id::text = $1
			ORDER BY c.created_at DESC
			LIMIT 50
		`
		custRows, err := pool.Query(ctx, custQuery, merchantID)
		if err == nil {
			defer custRows.Close()
			for custRows.Next() {
				var item CustomerRecoveryItem
				if err := custRows.Scan(
					&item.ID, &item.Email, &item.Phone, &item.OptOut, &item.TotalPayments,
					&item.FailedCount, &item.RecoveredCount, &item.LastStatus, &item.LastAction, &item.LastSeen,
				); err == nil {
					data.Customers = append(data.Customers, item)
				}
			}
		}

		// 4. Fetch Aggregate Metrics
		_ = pool.QueryRow(ctx, `
			SELECT 
				COALESCE(SUM(p.amount::float8), 0),
				COALESCE(COUNT(CASE WHEN rw.status NOT IN ('RECOVERED', 'FAILED', 'HALTED') THEN 1 END), 0)
			FROM payments p
			LEFT JOIN recovery_workflows rw ON rw.payment_id = p.id
			WHERE p.merchant_id::text = $1 AND p.status = 'FAILED'
		`, merchantID).Scan(&data.Metrics.TotalAtRiskRevenue, &data.Metrics.PendingRecoveries)

		_ = pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(ro.recovered_amount::float8), 0)
			FROM payments p
			JOIN recovery_outcomes ro ON ro.payment_id = p.id
			WHERE p.merchant_id::text = $1 AND ro.recovered = true
		`, merchantID).Scan(&data.Metrics.TotalRecovered)

		totalVolume := data.Metrics.TotalAtRiskRevenue + data.Metrics.TotalRecovered
		if totalVolume > 0 {
			data.Metrics.RecoveryRate = (data.Metrics.TotalRecovered / totalVolume) * 100.0
		}

		// 5. Sandbox Payment Links (Latest 20)
		sbQuery := `
			SELECT 
				p.id::text,
				COALESCE(c.email, 'customer@sandbox.io'),
				p.amount::float8,
				p.currency,
				COALESCE(p.razorpay_payment_id, p.id::text),
				p.status,
				p.created_at
			FROM payments p
			LEFT JOIN customers c ON p.customer_id = c.id
			WHERE p.merchant_id::text = $1
			ORDER BY p.created_at DESC
			LIMIT 20
		`
		sbRows, err := pool.Query(ctx, sbQuery, merchantID)
		if err == nil {
			defer sbRows.Close()
			for sbRows.Next() {
				var pID, cEmail, curr, rzpID, stat string
				var amt float64
				var cAt time.Time
				if err := sbRows.Scan(&pID, &cEmail, &amt, &curr, &rzpID, &stat, &cAt); err == nil {
					linkURL := fmt.Sprintf("https://checkout.reviveos.io/pay/%s", pID)
					data.SandboxLinks = append(data.SandboxLinks, SandboxLinkItem{
						PaymentID:      pID,
						CustomerEmail:  cEmail,
						Amount:         amt,
						Currency:       curr,
						PaymentLinkURL: linkURL,
						Status:         stat,
						CreatedAt:      cAt,
					})
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// MerchantCreateSubscriptionHandler creates a subscription manually for a merchant.
func MerchantCreateSubscriptionHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			MerchantID      string  `json:"merchant_id"`
			CustomerEmail   string  `json:"customer_email"`
			CustomerPhone   string  `json:"customer_phone"`
			Amount          float64 `json:"amount"`
			PlanID          string  `json:"plan_id"`
			BillingInterval string  `json:"billing_interval"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid payload"}`, http.StatusBadRequest)
			return
		}

		if req.MerchantID == "" || req.Amount <= 0 {
			http.Error(w, `{"error":"Merchant ID and positive amount are required"}`, http.StatusBadRequest)
			return
		}

		if req.CustomerEmail == "" {
			req.CustomerEmail = fmt.Sprintf("customer_%d@example.com", time.Now().Unix())
		}
		if req.PlanID == "" {
			req.PlanID = "plan_pro_monthly"
		}
		if req.BillingInterval == "" {
			req.BillingInterval = "monthly"
		}

		ctx := r.Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(ctx)

		// Get or create customer
		var customerID string
		err = tx.QueryRow(ctx, `
			SELECT id::text FROM customers WHERE merchant_id::text = $1 AND email = $2 LIMIT 1
		`, req.MerchantID, req.CustomerEmail).Scan(&customerID)
		if err != nil {
			err = tx.QueryRow(ctx, `
				INSERT INTO customers (merchant_id, email, phone)
				VALUES ($1, $2, $3)
				RETURNING id::text
			`, req.MerchantID, req.CustomerEmail, req.CustomerPhone).Scan(&customerID)
			if err != nil {
				http.Error(w, `{"error":"Failed to create customer"}`, http.StatusInternalServerError)
				return
			}
		}

		nextBilling := time.Now().AddDate(0, 1, 0)
		paymentLink := fmt.Sprintf("https://checkout.reviveos.io/sub/%s_%d", customerID, time.Now().Unix())

		var subID string
		err = tx.QueryRow(ctx, `
			INSERT INTO subscriptions (
				merchant_id, customer_id, amount, currency, status, billing_interval, plan_id, payment_link_url, next_billing_at
			)
			VALUES ($1, $2, $3, 'INR', 'ACTIVE', $4, $5, $6, $7)
			RETURNING id::text
		`, req.MerchantID, customerID, req.Amount, req.BillingInterval, req.PlanID, paymentLink, nextBilling).Scan(&subID)
		if err != nil {
			http.Error(w, `{"error":"Failed to create subscription"}`, http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(ctx); err != nil {
			http.Error(w, `{"error":"Failed to commit subscription"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"subscription_id":  subID,
			"customer_id":      customerID,
			"payment_link_url": paymentLink,
			"status":           "ACTIVE",
			"next_billing_at":  nextBilling,
		})
	}
}

// MerchantSandboxPaymentLinkHandler creates a test payment link / sandbox checkout instance.
func MerchantSandboxPaymentLinkHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			MerchantID    string  `json:"merchant_id"`
			CustomerEmail string  `json:"customer_email"`
			CustomerPhone string  `json:"customer_phone"`
			Amount        float64 `json:"amount"`
			Description   string  `json:"description"`
			TriggerFail   bool    `json:"trigger_failure_immediately"`
			FailureCode   string  `json:"failure_code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid payload"}`, http.StatusBadRequest)
			return
		}

		if req.MerchantID == "" || req.Amount <= 0 {
			http.Error(w, `{"error":"Merchant ID and positive amount are required"}`, http.StatusBadRequest)
			return
		}
		if req.CustomerEmail == "" {
			req.CustomerEmail = fmt.Sprintf("sandbox_%d@example.com", time.Now().Unix())
		}
		if req.FailureCode == "" {
			req.FailureCode = "INSUFFICIENT_FUNDS"
		}

		ctx := r.Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(ctx)

		var customerID string
		err = tx.QueryRow(ctx, `
			SELECT id::text FROM customers WHERE merchant_id::text = $1 AND email = $2 LIMIT 1
		`, req.MerchantID, req.CustomerEmail).Scan(&customerID)
		if err != nil {
			err = tx.QueryRow(ctx, `
				INSERT INTO customers (merchant_id, email, phone)
				VALUES ($1, $2, $3)
				RETURNING id::text
			`, req.MerchantID, req.CustomerEmail, req.CustomerPhone).Scan(&customerID)
			if err != nil {
				http.Error(w, `{"error":"Failed to create customer"}`, http.StatusInternalServerError)
				return
			}
		}

		paymentStatus := "PENDING"
		if req.TriggerFail {
			paymentStatus = "FAILED"
		}

		provider, _ := paymentprovider.NewPaymentProvider("", pool)
		retryRes, _ := provider.CreateRetryAttemptWithCustomer(
			ctx,
			fmt.Sprintf("sb_%d", time.Now().UnixNano()),
			req.Amount,
			req.CustomerEmail,
			req.CustomerPhone,
			"",
		)

		linkURL := fmt.Sprintf("https://checkout.reviveos.io/pay/sb_%d", time.Now().UnixNano())
		if retryRes != nil && retryRes.PaymentLinkURL != "" {
			linkURL = retryRes.PaymentLinkURL
		}

		var paymentID string
		err = tx.QueryRow(ctx, `
			INSERT INTO payments (
				merchant_id, customer_id, amount, currency, status, method, failure_code
			)
			VALUES ($1, $2, $3, 'INR', $4, 'card', $5)
			RETURNING id::text
		`, req.MerchantID, customerID, req.Amount, paymentStatus, req.FailureCode).Scan(&paymentID)
		if err != nil {
			http.Error(w, `{"error":"Failed to create payment"}`, http.StatusInternalServerError)
			return
		}

		var workflowID string
		if req.TriggerFail {
			// Immediately insert recovery_workflows row so it shows up in /workflows
			_ = tx.QueryRow(ctx, `
				INSERT INTO recovery_workflows (
					payment_id, merchant_id, status, recovery_probability, selected_action, created_at, updated_at
				)
				VALUES ($1, $2, 'SCHEDULED', 0.78, 'PAYMENT_LINK', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				RETURNING id::text
			`, paymentID, req.MerchantID).Scan(&workflowID)

			// Record action in recovery_actions
			if workflowID != "" {
				_, _ = tx.Exec(ctx, `
					INSERT INTO recovery_actions (workflow_id, action_type, status, attempt, result, executed_at)
					VALUES ($1, 'PAYMENT_LINK', 'EXECUTED', 1, $2, CURRENT_TIMESTAMP)
				`, workflowID, linkURL)

				_ = audit.AppendAuditLog(ctx, pool, audit.AuditEvent{
					WorkflowID: workflowID,
					Actor:      "sandbox:recovery",
					Action:     "PAYMENT_LINK_DISPATCHED",
					Metadata: map[string]interface{}{
						"payment_id": paymentID,
						"link_url":   linkURL,
						"email":      req.CustomerEmail,
					},
				})
			}

			// Enqueue recovery analyze task into outbox
			analyzePayload := map[string]interface{}{"payment_id": paymentID}
			_, _ = outbox.InsertOutboxEvent(ctx, tx, "recovery:analyze", "payment", paymentID, analyzePayload)
		}

		if err := tx.Commit(ctx); err != nil {
			http.Error(w, `{"error":"Failed to commit transaction"}`, http.StatusInternalServerError)
			return
		}

		// Dispatch live email via Resend API if email is provided
		if req.TriggerFail && req.CustomerEmail != "" {
			notifProv := notificationservice.NewNotificationProvider("")
			if notifProv != nil {
				go func() {
					bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					_, _ = notifProv.SendRecoveryNotification(bgCtx, notificationservice.NotificationRequest{
						PaymentID:     paymentID,
						WorkflowID:    workflowID,
						MerchantName:  "ReviveOS Merchant",
						CustomerEmail: req.CustomerEmail,
						CustomerPhone: req.CustomerPhone,
						Amount:        req.Amount,
						Currency:      "INR",
						PaymentLink:   linkURL,
						FailureReason: req.FailureCode,
						ActionType:    "PAYMENT_LINK",
					})
				}()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"payment_id":       paymentID,
			"customer_id":      customerID,
			"workflow_id":      workflowID,
			"amount":           req.Amount,
			"payment_link_url": linkURL,
			"status":           paymentStatus,
			"recovery_queued":  req.TriggerFail,
		})
	}
}
