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
				Description      string `json:"description"`
				Email            string `json:"email"`
				Contact          string `json:"contact"`
				ErrorCode        string `json:"error_code"`
				ErrorDescription string `json:"error_description"`
				ErrorSource      string `json:"error_source"`
				ErrorStep        string `json:"error_step"`
				ErrorReason      string `json:"error_reason"`
				CreatedAt        int64  `json:"created_at"`
			} `json:"entity"`
		} `json:"payment"`
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

		// 6. If payment.failed, ingest payment and publish recovery:analyze task to transactional outbox
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
			custEmail := paymentEntity.Email
			if custEmail == "" {
				custEmail = fmt.Sprintf("cust_%s@example.com", razorpayPaymentID)
			}
			custPhone := paymentEntity.Contact
			if custPhone == "" {
				custPhone = "+919999999999"
			}

			err = tx.QueryRow(ctx, `
				SELECT id::text FROM customers WHERE merchant_id = $1 AND (email = $2 OR phone = $3) LIMIT 1
			`, merchantID, custEmail, custPhone).Scan(&customerID)
			if err != nil {
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

			// Insert payment record
			var paymentUUID string
			err = tx.QueryRow(ctx, `
				INSERT INTO payments (
					merchant_id, customer_id, amount, currency, status, method, failure_code, razorpay_payment_id
				)
				VALUES ($1, $2, $3, $4, 'FAILED', $5, $6, $7)
				RETURNING id::text
			`, merchantID, customerID, amountFloat, currency, method, failureCode, razorpayPaymentID).Scan(&paymentUUID)
			if err != nil {
				log.Printf("[Webhook] Failed to insert payment: %v", err)
				http.Error(w, `{"error": "failed to insert payment"}`, http.StatusInternalServerError)
				return
			}

			// Insert outbox event for worker analysis
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
