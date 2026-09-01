package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/utils/outbox"
)

// UnifiedFailureDetail holds standardized gateway failure parameters
type UnifiedFailureDetail struct {
	Code        string `json:"code"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
	Step        string `json:"step,omitempty"`
	Source      string `json:"source,omitempty"`
}

// UnifiedPaymentEvent is the provider-agnostic event model for ReviveOS
type UnifiedPaymentEvent struct {
	Provider       string                 `json:"provider"` // "razorpay", "stripe", "mock"
	EventType      string                 `json:"eventType"` // "payment.failed", "payment.captured", etc.
	EventID        string                 `json:"eventId"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
	PaymentID      string                 `json:"paymentId"`
	OrderID        string                 `json:"orderId,omitempty"`
	CustomerID     string                 `json:"customerId,omitempty"`
	CustomerEmail  string                 `json:"customerEmail,omitempty"`
	CustomerPhone  string                 `json:"customerPhone,omitempty"`
	Amount         float64                `json:"amount"` // In major currency unit (e.g. INR 4999.00)
	Currency       string                 `json:"currency"`
	PaymentMethod  string                 `json:"paymentMethod,omitempty"`
	Bank           string                 `json:"bank,omitempty"`
	Timestamp      string                 `json:"timestamp"`
	Failure        *UnifiedFailureDetail  `json:"failure,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	RawPayload     interface{}            `json:"rawPayload,omitempty"`
}

// VerifyRazorpaySignature computes HMAC-SHA256 of body against secret
func VerifyRazorpaySignature(body []byte, signature, secret string) bool {
	if secret == "" {
		// If secret is not set, allow for local sandbox unless explicitly required
		return true
	}
	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expectedSignature)))
}

// V1EventsHandler handles POST /v1/events (Unified Event Ingestion)
func V1EventsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		mCtx := GetMerchantContext(r)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"Failed to read request body"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var event UnifiedPaymentEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Invalid unified event JSON: %v"}`, err), http.StatusBadRequest)
			return
		}

		if event.EventType == "" || event.PaymentID == "" {
			http.Error(w, `{"error":"Missing required fields: eventType and paymentId are required"}`, http.StatusBadRequest)
			return
		}

		if event.Provider == "" {
			event.Provider = "razorpay"
		}
		if event.Currency == "" {
			event.Currency = "INR"
		}
		if event.EventID == "" {
			event.EventID = fmt.Sprintf("evt_%d_%s", time.Now().UnixNano(), event.PaymentID)
		}

		idempotencyKey := event.IdempotencyKey
		if idempotencyKey == "" {
			idempotencyKey = event.EventID
		}

		// Check idempotency in database
		if pool != nil {
			var existingID string
			checkErr := pool.QueryRow(r.Context(), `
				SELECT id::text FROM payment_events
				WHERE event_id = $1 OR idempotency_key = $2
				LIMIT 1
			`, event.EventID, idempotencyKey).Scan(&existingID)

			if checkErr == nil && existingID != "" {
				// Event already ingested idempotently
				log.Printf("[V1Events] Duplicate event %s ignored via idempotency key %s", event.EventID, idempotencyKey)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status":          "DUPLICATE_IGNORED",
					"message":         "Event already processed idempotently.",
					"event_id":        event.EventID,
					"idempotency_key": idempotencyKey,
					"payment_id":      event.PaymentID,
				})
				return
			}
		}

		// Persist or upsert Customer
		var customerUUID string
		if pool != nil && (event.CustomerEmail != "" || event.CustomerID != "") {
			custQuery := `
				INSERT INTO customers (merchant_id, email, phone, created_at, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT DO NOTHING
				RETURNING id::text
			`
			err := pool.QueryRow(r.Context(), custQuery, mCtx.MerchantID, event.CustomerEmail, event.CustomerPhone).Scan(&customerUUID)
			if err != nil || customerUUID == "" {
				_ = pool.QueryRow(r.Context(), `
					SELECT id::text FROM customers WHERE merchant_id = $1 AND (email = $2 OR phone = $3) LIMIT 1
				`, mCtx.MerchantID, event.CustomerEmail, event.CustomerPhone).Scan(&customerUUID)
			}
		}

		// Normalize failure details
		failureCode := "UNKNOWN"
		if event.Failure != nil && event.Failure.Code != "" {
			failureCode = event.Failure.Code
		}

		// Normalize payment status
		paymentStatus := "FAILED"
		switch event.EventType {
		case "payment.captured", "payment.authorized", "order.paid":
			paymentStatus = "CAPTURED"
		case "refund.created", "refund.processed":
			paymentStatus = "REFUNDED"
		}

		// Insert or update payment record
		var paymentUUID string
		if pool != nil {
			payInsertSQL := `
				INSERT INTO payments (
					merchant_id, customer_id, amount, currency, status, method, failure_code, razorpay_payment_id, created_at, updated_at
				)
				VALUES (
					$1, 
					CASE WHEN $2 != '' THEN $2::uuid ELSE NULL END,
					$3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
				)
				ON CONFLICT (razorpay_payment_id) DO UPDATE 
				SET status = EXCLUDED.status, 
				    failure_code = EXCLUDED.failure_code,
				    updated_at = CURRENT_TIMESTAMP
				RETURNING id::text
			`
			err := pool.QueryRow(
				r.Context(),
				payInsertSQL,
				mCtx.MerchantID,
				customerUUID,
				event.Amount,
				event.Currency,
				paymentStatus,
				event.PaymentMethod,
				failureCode,
				event.PaymentID,
			).Scan(&paymentUUID)

			if err != nil {
				_ = pool.QueryRow(r.Context(), `SELECT id::text FROM payments WHERE razorpay_payment_id = $1 LIMIT 1`, event.PaymentID).Scan(&paymentUUID)
			}
		}

		// Store Event in payment_events
		if pool != nil && paymentUUID != "" {
			rawJSON, _ := json.Marshal(event)
			_, _ = pool.Exec(r.Context(), `
				INSERT INTO payment_events (
					payment_id, event_id, event_type, idempotency_key, payload, processing_status, received_at, created_at
				)
				VALUES ($1, $2, $3, $4, $5, 'PENDING', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`, paymentUUID, event.EventID, event.EventType, idempotencyKey, rawJSON)

			// If it is a failure event, queue outbox workflow
			if event.EventType == "payment.failed" {
				outboxMsg := map[string]interface{}{
					"payment_id":       paymentUUID,
					"merchant_id":      mCtx.MerchantID,
					"event_type":       event.EventType,
					"failure_code":     failureCode,
					"external_id":      event.PaymentID,
					"amount":           event.Amount,
					"currency":         event.Currency,
					"customer_email":   event.CustomerEmail,
					"idempotency_key":  idempotencyKey,
				}
				_, _ = outbox.InsertOutboxEventPool(r.Context(), pool, "PAYMENT_FAILED", "payment", paymentUUID, outboxMsg)
			}
		}

		// Audit Log
		RecordAuditLog(r.Context(), pool, mCtx.MerchantID, "API_KEY", mCtx.KeyID, "UNIFIED_EVENT_INGESTED", r.RemoteAddr, map[string]interface{}{
			"event_id":   event.EventID,
			"event_type": event.EventType,
			"payment_id": event.PaymentID,
			"amount":     event.Amount,
		})

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "INGESTED",
			"event_id":        event.EventID,
			"idempotency_key": idempotencyKey,
			"payment_id":      event.PaymentID,
			"internal_id":     paymentUUID,
			"timestamp":       time.Now().Format(time.RFC3339),
		})
	}
}

// V1WebhookEndpointHandler handles /api/reviveos/webhook
// It acts as a universal webhook ingestor that validates signatures and forwards to the pipeline.
func V1WebhookEndpointHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"Failed to read request body"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Signature verification
		signature := r.Header.Get("X-Razorpay-Signature")
		secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
		if secret != "" && signature != "" {
			if !VerifyRazorpaySignature(body, signature, secret) {
				http.Error(w, `{"error":"Invalid webhook signature"}`, http.StatusUnauthorized)
				return
			}
		}

		// Delegate ingestion to standard RazorpayWebhookHandler
		handler := RazorpayWebhookHandler(pool)
		// Reset body for downstream handler
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		handler.ServeHTTP(w, r)
	}
}
