package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	paymentprovider "github.com/reviveos/services/payment-provider"
	"github.com/reviveos/utils/outbox"
)

// RazorpayWebhookPayload models the incoming webhook JSON body from Razorpay.
type RazorpayWebhookPayload struct {
	Entity    string   `json:"entity"`
	AccountID string   `json:"account_id"`
	Event     string   `json:"event"`
	Contains  []string `json:"contains"`
	Payload   struct {
		Payment struct {
			Entity struct {
				ID               string `json:"id"`
				Entity           string `json:"entity"`
				Amount           int64  `json:"amount"` // in paise
				Currency         string `json:"currency"`
				Status           string `json:"status"`
				Method           string `json:"method"`
				Description      string                 `json:"description"`
				Email            string                 `json:"email"`
				Contact          string                 `json:"contact"`
				SubscriptionID   string                 `json:"subscription_id,omitempty"`
				InvoiceID        string                 `json:"invoice_id,omitempty"`
				Notes            map[string]interface{} `json:"notes,omitempty"`
				ErrorCode        string                 `json:"error_code"`
				ErrorDescription string                 `json:"error_description"`
				ErrorSource      string `json:"error_source"`
				ErrorStep        string `json:"error_step"`
				ErrorReason      string `json:"error_reason"`
				CreatedAt        int64  `json:"created_at"`
			} `json:"entity"`
		} `json:"payment"`
		PaymentLink struct {
			Entity struct {
				ID          string                 `json:"id"`
				Amount      int64                  `json:"amount"` // in paise
				Currency    string                 `json:"currency"`
				Status      string                 `json:"status"`
				Description string                 `json:"description"`
				ShortURL    string                 `json:"short_url"`
				ReferenceID string                 `json:"reference_id"`
				Customer    struct {
					Name    string `json:"name"`
					Email   string `json:"email"`
					Contact string `json:"contact"`
				} `json:"customer"`
				Notes     map[string]interface{} `json:"notes,omitempty"`
				CreatedAt int64                  `json:"created_at"`
			} `json:"entity"`
		} `json:"payment_link"`
		Subscription struct {
			Entity struct {
				ID         string `json:"id"`
				PlanID     string `json:"plan_id"`
				Status     string `json:"status"`
				CustomerID string `json:"customer_id"`
				CreatedAt  int64  `json:"created_at"`
			} `json:"entity"`
		} `json:"subscription"`
	} `json:"payload"`
	CreatedAt int64 `json:"created_at"`
}

// RazorpayWebhookHandler processes incoming webhooks from Razorpay with signature verification and persistent deduplication.
func RazorpayWebhookHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// 1. Read Headers
		eventID := r.Header.Get("X-Razorpay-Event-Id")
		signature := r.Header.Get("X-Razorpay-Signature")

		webhookSecret := strings.TrimSpace(os.Getenv("RAZORPAY_WEBHOOK_SECRET"))
		if webhookSecret == "" {
			webhookSecret = "test_webhook_secret_12345" // Safe development default
		}

		// 2. Read Request Body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 3. Verify Signature
		if signature == "" || !paymentprovider.VerifyWebhookSignature(bodyBytes, signature, webhookSecret) {
			log.Printf("[Webhook] Rejected: Invalid X-Razorpay-Signature")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid webhook signature"}`))
			return
		}

		// 4. Parse Payload
		var payload RazorpayWebhookPayload
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			log.Printf("[Webhook] JSON parse error: %v", err)
			http.Error(w, `{"error": "malformed webhook payload"}`, http.StatusBadRequest)
			return
		}

		if eventID == "" {
			eventID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
		}

		eventType := payload.Event
		if eventType == "" {
			eventType = "unknown"
		}

		razorpayPaymentID := payload.Payload.Payment.Entity.ID
		if razorpayPaymentID == "" {
			razorpayPaymentID = payload.Payload.PaymentLink.Entity.ID
		}

		ctx := r.Context()

		// 5. Persistent Deduplication via PostgreSQL
		var insertedEventID string
		err = pool.QueryRow(ctx, `
			INSERT INTO payment_events (
				razorpay_event_id, event_type, razorpay_payment_id, raw_payload, processing_status, received_at
			)
			VALUES ($1, $2, $3, $4, 'PENDING', CURRENT_TIMESTAMP)
			ON CONFLICT (razorpay_event_id) DO NOTHING
			RETURNING id::text
		`, eventID, eventType, razorpayPaymentID, bodyBytes).Scan(&insertedEventID)

		if err != nil || insertedEventID == "" {
			// Duplicate event detected: Deduplicate idempotently without spawning second workflow
			log.Printf("[Webhook] EVENT_DEDUPLICATED: Event ID %s was already received and processed", eventID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status": "duplicate_ignored", "event_id": "%s"}`, eventID)))
			return
		}

		log.Printf("[Webhook] WEBHOOK_RECEIVED: Event ID=%s, Type=%s, PaymentID=%s", eventID, eventType, razorpayPaymentID)

		// 6. Handle Payment Link Created (from Razorpay Dashboard or API)
		if strings.EqualFold(eventType, "payment_link.created") {
			plink := payload.Payload.PaymentLink.Entity
			if plink.ID != "" {
				amountFloat := float64(plink.Amount) / 100.0
				custEmail := strings.TrimSpace(plink.Customer.Email)
				custPhone := strings.TrimSpace(plink.Customer.Contact)
				if plink.Notes != nil {
					if cEmail, ok := plink.Notes["customer_email"].(string); ok && cEmail != "" {
						custEmail = strings.TrimSpace(cEmail)
					}
				}

				tx, err := pool.Begin(ctx)
				if err == nil {
					defer tx.Rollback(ctx)

					var merchantID string
					_ = tx.QueryRow(ctx, "SELECT id::text FROM merchants ORDER BY created_at ASC LIMIT 1").Scan(&merchantID)
					if merchantID == "" {
						_ = tx.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ('Default Merchant') RETURNING id::text").Scan(&merchantID)
					}

					var customerID string
					if custEmail != "" {
						_ = tx.QueryRow(ctx, "SELECT id::text FROM customers WHERE merchant_id = $1 AND email = $2 LIMIT 1", merchantID, custEmail).Scan(&customerID)
					}
					if customerID == "" {
						if custEmail == "" {
							custEmail = fmt.Sprintf("cust_%s@revive-os.me", plink.ID)
						}
						_ = tx.QueryRow(ctx, "INSERT INTO customers (merchant_id, email, phone) VALUES ($1, $2, $3) RETURNING id::text", merchantID, custEmail, custPhone).Scan(&customerID)
					}

					var existingPayID string
					_ = tx.QueryRow(ctx, "SELECT id::text FROM payments WHERE razorpay_payment_id = $1 LIMIT 1", plink.ID).Scan(&existingPayID)
					if existingPayID == "" {
						_, _ = tx.Exec(ctx, `
							INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method, razorpay_payment_id)
							VALUES ($1, $2, $3, $4, 'PENDING', 'payment_link', $5)
						`, merchantID, customerID, amountFloat, plink.Currency, plink.ID)
					}

					_, _ = tx.Exec(ctx, "UPDATE payment_events SET processing_status = 'PROCESSED', processed = true, processed_at = CURRENT_TIMESTAMP WHERE id = $1", insertedEventID)
					_ = tx.Commit(ctx)
					log.Printf("[Webhook] PAYMENT_LINK_CREATED: Ingested link %s for customer %s (%.2f %s)", plink.ID, custEmail, amountFloat, plink.Currency)
				}
			}
		}

		// 7. Handle Payment Success / Captured / Payment Link Paid
		if strings.EqualFold(eventType, "payment.captured") || strings.EqualFold(eventType, "payment_link.paid") || strings.EqualFold(eventType, "order.paid") {
			paymentEntity := payload.Payload.Payment.Entity
			plink := payload.Payload.PaymentLink.Entity

			amountFloat := float64(paymentEntity.Amount) / 100.0
			if amountFloat == 0 {
				amountFloat = float64(plink.Amount) / 100.0
			}

			var searchRefs []string
			if paymentEntity.ID != "" {
				searchRefs = append(searchRefs, paymentEntity.ID)
			}
			if paymentEntity.InvoiceID != "" {
				searchRefs = append(searchRefs, paymentEntity.InvoiceID)
			}
			if plink.ID != "" {
				searchRefs = append(searchRefs, plink.ID)
			}
			if paymentEntity.Notes != nil {
				if pid, ok := paymentEntity.Notes["payment_id"].(string); ok && pid != "" {
					searchRefs = append(searchRefs, pid)
				}
			}
			if strings.HasPrefix(paymentEntity.Description, "Payment recovery for ") {
				searchRefs = append(searchRefs, strings.TrimSpace(strings.TrimPrefix(paymentEntity.Description, "Payment recovery for ")))
			}

			tx, err := pool.Begin(ctx)
			if err == nil {
				defer tx.Rollback(ctx)

				var paymentUUID string
				for _, ref := range searchRefs {
					if ref == "" {
						continue
					}
					_ = tx.QueryRow(ctx, "SELECT id::text FROM payments WHERE id::text = $1 OR razorpay_payment_id = $1 LIMIT 1", ref).Scan(&paymentUUID)
					if paymentUUID != "" {
						break
					}
				}

				if paymentUUID != "" {
					_, _ = tx.Exec(ctx, "UPDATE payments SET status = 'CAPTURED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1", paymentUUID)
					_, _ = tx.Exec(ctx, "UPDATE recovery_workflows SET status = 'RECOVERED', updated_at = CURRENT_TIMESTAMP WHERE payment_id::text = $1", paymentUUID)
					_, _ = tx.Exec(ctx, `
						INSERT INTO recovery_outcomes (payment_id, recovered, recovered_amount, created_at)
						VALUES ($1, true, $2, CURRENT_TIMESTAMP)
					`, paymentUUID, amountFloat)
					log.Printf("[Webhook] PAYMENT_RECOVERED: Payment %s marked CAPTURED and workflow RECOVERED (%.2f)", paymentUUID, amountFloat)
				}

				_, _ = tx.Exec(ctx, "UPDATE payment_events SET processing_status = 'PROCESSED', processed = true, processed_at = CURRENT_TIMESTAMP WHERE id = $1", insertedEventID)
				_ = tx.Commit(ctx)
			}
		}

		// 8. Handle Payment Link Cancelled / Expired
		if strings.EqualFold(eventType, "payment_link.cancelled") || strings.EqualFold(eventType, "payment_link.expired") {
			plink := payload.Payload.PaymentLink.Entity
			if plink.ID != "" {
				tx, err := pool.Begin(ctx)
				if err == nil {
					defer tx.Rollback(ctx)
					var paymentUUID string
					_ = tx.QueryRow(ctx, "SELECT id::text FROM payments WHERE razorpay_payment_id = $1 LIMIT 1", plink.ID).Scan(&paymentUUID)
					if paymentUUID != "" {
						_, _ = tx.Exec(ctx, "UPDATE payments SET status = 'CANCELLED', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1", paymentUUID)
						_, _ = tx.Exec(ctx, "UPDATE recovery_workflows SET status = 'HALTED', updated_at = CURRENT_TIMESTAMP WHERE payment_id::text = $1", paymentUUID)
					}
					_, _ = tx.Exec(ctx, "UPDATE payment_events SET processing_status = 'PROCESSED', processed = true, processed_at = CURRENT_TIMESTAMP WHERE id = $1", insertedEventID)
					_ = tx.Commit(ctx)
				}
			}
		}

		// 9. If payment.failed, ingest payment and publish recovery:analyze task to transactional outbox
		if strings.EqualFold(eventType, "payment.failed") && razorpayPaymentID != "" {
			paymentEntity := payload.Payload.Payment.Entity
			amountFloat := float64(paymentEntity.Amount) / 100.0 // paise to INR
			if amountFloat == 0 {
				amountFloat = 1000.00
			}
			currency := paymentEntity.Currency
			if currency == "" {
				currency = "INR"
			}
			method := paymentEntity.Method
			if method == "" {
				method = "card"
			}
			failureCode := paymentEntity.ErrorCode
			if failureCode == "" {
				failureCode = paymentEntity.ErrorReason
			}
			if failureCode == "" {
				failureCode = "GATEWAY_ERROR"
			}

			// Ingest within transaction
			tx, err := pool.Begin(ctx)
			if err != nil {
				log.Printf("[Webhook] Failed to start tx: %v", err)
				http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
				return
			}
			defer tx.Rollback(ctx)

			// Get or create merchant
			var merchantID string
			err = tx.QueryRow(ctx, "SELECT id::text FROM merchants ORDER BY created_at ASC LIMIT 1").Scan(&merchantID)
			if err != nil {
				err = tx.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ('Default Merchant') RETURNING id::text").Scan(&merchantID)
				if err != nil {
					log.Printf("[Webhook] Failed to get/create merchant: %v", err)
					http.Error(w, `{"error": "failed to locate merchant"}`, http.StatusInternalServerError)
					return
				}
			}

			// Get or create customer
			var customerID string
			custEmail := strings.TrimSpace(paymentEntity.Email)
			custPhone := strings.TrimSpace(paymentEntity.Contact)
			description := strings.TrimSpace(paymentEntity.Description)
			plink := payload.Payload.PaymentLink.Entity

			// 1. Fallback to Payment Link customer email/phone if payment entity has void@ placeholder or empty
			if strings.EqualFold(custEmail, "void@razorpay.com") || custEmail == "" || strings.HasPrefix(strings.ToLower(custEmail), "void@") {
				if plink.Customer.Email != "" && !strings.HasPrefix(strings.ToLower(plink.Customer.Email), "void@") {
					custEmail = strings.TrimSpace(plink.Customer.Email)
				}
				if custPhone == "" && plink.Customer.Contact != "" {
					custPhone = strings.TrimSpace(plink.Customer.Contact)
				}
			}

			// 2. If this payment is an attempt on a ReviveOS recovery payment link or existing link,
			// check notes, description, or payment link ID to resolve the registered customer
			var origPaymentRef string
			if paymentEntity.Notes != nil {
				if pid, ok := paymentEntity.Notes["payment_id"].(string); ok && pid != "" {
					origPaymentRef = strings.TrimSpace(pid)
				}
				if cEmail, ok := paymentEntity.Notes["customer_email"].(string); ok && cEmail != "" {
					custEmail = strings.TrimSpace(cEmail)
				}
			}
			if origPaymentRef == "" && strings.HasPrefix(description, "Payment recovery for ") {
				origPaymentRef = strings.TrimSpace(strings.TrimPrefix(description, "Payment recovery for "))
			}
			if origPaymentRef == "" && plink.ID != "" {
				origPaymentRef = plink.ID
			}
			if origPaymentRef == "" && paymentEntity.InvoiceID != "" {
				origPaymentRef = paymentEntity.InvoiceID
			}

			if origPaymentRef != "" {
				var origCustID, origEmail, origPhone string
				lookupOrigErr := tx.QueryRow(ctx, `
					SELECT p.customer_id::text, COALESCE(c.email, ''), COALESCE(c.phone, '')
					FROM payments p
					JOIN customers c ON p.customer_id = c.id
					WHERE p.id::text = $1 OR p.razorpay_payment_id = $1
					LIMIT 1
				`, origPaymentRef).Scan(&origCustID, &origEmail, &origPhone)
				if lookupOrigErr == nil && origCustID != "" {
					customerID = origCustID
					if custEmail == "" || strings.EqualFold(custEmail, "void@razorpay.com") || strings.HasPrefix(strings.ToLower(custEmail), "void@") {
						custEmail = origEmail
					}
					if custPhone == "" {
						custPhone = origPhone
					}
					log.Printf("[Webhook] Successfully attributed failure to registered customer %s (%s) for ref: %s", customerID, custEmail, origPaymentRef)
				}
			}

			// 3. If not already resolved from recovery reference, query by email or phone
			if customerID == "" {
				// Clean void@ placeholders
				if strings.EqualFold(custEmail, "void@razorpay.com") || strings.HasPrefix(strings.ToLower(custEmail), "void@") {
					custEmail = ""
				}

				var lookupErr error
				if custEmail != "" {
					lookupErr = tx.QueryRow(ctx, `
						SELECT id::text FROM customers WHERE merchant_id = $1 AND email = $2 LIMIT 1
					`, merchantID, custEmail).Scan(&customerID)
				} else if custPhone != "" {
					lookupErr = tx.QueryRow(ctx, `
						SELECT id::text FROM customers WHERE merchant_id = $1 AND phone = $2 LIMIT 1
					`, merchantID, custPhone).Scan(&customerID)
				} else {
					lookupErr = fmt.Errorf("no email or phone provided")
				}

				if lookupErr != nil || customerID == "" {
					if custEmail == "" {
						custEmail = fmt.Sprintf("cust_%s@revive-os.me", razorpayPaymentID)
					}
					err = tx.QueryRow(ctx, `
						INSERT INTO customers (merchant_id, email, phone)
						VALUES ($1, $2, $3)
						RETURNING id::text
					`, merchantID, custEmail, custPhone).Scan(&customerID)
					if err != nil {
						log.Printf("[Webhook] Failed to insert customer: %v", err)
						http.Error(w, `{"error": "failed to insert customer"}`, http.StatusInternalServerError)
						return
					}
				}
			}

			// Check for recurring subscription context
			rzpSubID := paymentEntity.SubscriptionID
			if rzpSubID == "" {
				rzpSubID = payload.Payload.Subscription.Entity.ID
			}

			var subscriptionUUID *string
			if rzpSubID != "" {
				var subID string
				err := tx.QueryRow(ctx, "SELECT id::text FROM subscriptions WHERE razorpay_subscription_id = $1 LIMIT 1", rzpSubID).Scan(&subID)
				if err != nil {
					_ = tx.QueryRow(ctx, `
						INSERT INTO subscriptions (merchant_id, customer_id, amount, currency, status, razorpay_subscription_id)
						VALUES ($1, $2, $3, $4, 'PAST_DUE', $5)
						RETURNING id::text
					`, merchantID, customerID, amountFloat, currency, rzpSubID).Scan(&subID)
				} else {
					_, _ = tx.Exec(ctx, "UPDATE subscriptions SET status = 'PAST_DUE', updated_at = CURRENT_TIMESTAMP WHERE id::text = $1", subID)
				}
				if subID != "" {
					subscriptionUUID = &subID
				}
			}

			// Insert or update payment record
			var paymentUUID string
			// Check if a payment with this razorpay payment id or payment link ID already exists
			var existingPayUUID string
			_ = tx.QueryRow(ctx, "SELECT id::text FROM payments WHERE razorpay_payment_id = $1 OR (razorpay_payment_id = $2 AND $2 != '') LIMIT 1", razorpayPaymentID, plink.ID).Scan(&existingPayUUID)

			if existingPayUUID != "" {
				paymentUUID = existingPayUUID
				_, err = tx.Exec(ctx, `
					UPDATE payments
					SET customer_id = $1,
					    amount = $2,
					    currency = $3,
					    status = 'FAILED',
					    method = $4,
					    failure_code = $5,
					    razorpay_payment_id = $6,
					    updated_at = CURRENT_TIMESTAMP
					WHERE id::text = $7
				`, customerID, amountFloat, currency, method, failureCode, razorpayPaymentID, paymentUUID)
				if err != nil {
					log.Printf("[Webhook] Failed to update existing payment: %v", err)
				}
			} else {
				err = tx.QueryRow(ctx, `
					INSERT INTO payments (
						merchant_id, customer_id, subscription_id, amount, currency, status, method, failure_code, razorpay_payment_id
					)
					VALUES ($1, $2, $3, $4, $5, 'FAILED', $6, $7, $8)
					RETURNING id::text
				`, merchantID, customerID, subscriptionUUID, amountFloat, currency, method, failureCode, razorpayPaymentID).Scan(&paymentUUID)
				if err != nil {
					log.Printf("[Webhook] Failed to insert payment: %v", err)
					http.Error(w, `{"error": "failed to insert payment"}`, http.StatusInternalServerError)
					return
				}
			}

			// 4. Ensure immediate Recovery Workflow creation so it appears in the dashboard instantly
			var workflowUUID string
			_ = tx.QueryRow(ctx, `
				INSERT INTO recovery_workflows (payment_id, merchant_id, status, selected_action, recovery_probability, created_at, updated_at)
				VALUES ($1, $2, 'SCHEDULED', 'DELAYED_RETRY', 0.65, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT (payment_id) DO UPDATE SET status = 'SCHEDULED', updated_at = CURRENT_TIMESTAMP
				RETURNING id::text
			`, paymentUUID, merchantID).Scan(&workflowUUID)

			// Insert outbox event for worker background analysis
			analyzePayload := map[string]interface{}{
				"payment_id": paymentUUID,
			}
			_, err = outbox.InsertOutboxEvent(ctx, tx, "recovery:analyze", "payment", paymentUUID, analyzePayload)
			if err != nil {
				log.Printf("[Webhook] Failed to insert outbox event: %v", err)
				http.Error(w, `{"error": "failed to queue outbox event"}`, http.StatusInternalServerError)
				return
			}

			// Mark payment_events as processed
			_, err = tx.Exec(ctx, `
				UPDATE payment_events
				SET processing_status = 'PROCESSED',
				    processed = true,
				    processed_at = CURRENT_TIMESTAMP
				WHERE id = $1
			`, insertedEventID)
			if err != nil {
				log.Printf("[Webhook] Failed to update payment_events status: %v", err)
			}

			if err := tx.Commit(ctx); err != nil {
				log.Printf("[Webhook] Failed to commit webhook ingestion: %v", err)
				http.Error(w, `{"error": "failed to commit transaction"}`, http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status": "received", "event_id": "%s"}`, eventID)))
	}
}
