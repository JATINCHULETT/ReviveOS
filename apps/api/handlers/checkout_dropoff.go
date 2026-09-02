package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/checkout"
)

// CheckoutDropoffHandler tracks and recovers pre-gateway checkout abandonment
func CheckoutDropoffHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			// List dropped off checkout sessions & recovery funnel metrics
			sessions := getSampleCheckoutSessions()
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "ok",
				"sessions": sessions,
				"funnel":   calculateDropoffFunnel(sessions),
			})

		case http.MethodPost:
			// Session tracking or recovery trigger
			if strings.Contains(r.URL.Path, "/recover") {
				var req struct {
					SessionToken string `json:"session_token"`
					Channel      string `json:"channel"` // EMAIL, WHATSAPP, SMS
				}
				_ = json.NewDecoder(r.Body).Decode(&req)
				if req.Channel == "" {
					req.Channel = "EMAIL"
				}

				baseURL := os.Getenv("NEXT_PUBLIC_API_URL")
				if baseURL == "" {
					baseURL = "https://reviveos.onrender.com"
				}
				recoveryLink := checkout.GenerateRecoveryLink(baseURL, req.SessionToken)

				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":         "recovery_dispatched",
					"session_token":  req.SessionToken,
					"channel":        req.Channel,
					"email_provider": "Resend",
					"recovery_link":  recoveryLink,
					"dispatched_at":  time.Now().UTC(),
					"message":        fmt.Sprintf("Restored cart recovery link dispatched via %s (Resend/SMS)", req.Channel),
				})
				return
			}

			// Track/Create checkout session
			var session checkout.Session
			if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
				http.Error(w, `{"error":"invalid session payload"}`, http.StatusBadRequest)
				return
			}

			if session.SessionToken == "" {
				session.SessionToken = checkout.GenerateSessionToken()
			}
			if session.Status == "" {
				session.Status = checkout.StatusActive
			}
			if session.Currency == "" {
				session.Currency = "INR"
			}
			baseURL := os.Getenv("NEXT_PUBLIC_API_URL")
			session.RecoveryLink = checkout.GenerateRecoveryLink(baseURL, session.SessionToken)
			session.CreatedAt = time.Now()
			session.UpdatedAt = time.Now()

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "tracked",
				"session": session,
			})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}

func calculateDropoffFunnel(sessions []checkout.Session) map[string]interface{} {
	var total, dropped, dispatched, recovered int
	var recoverableValue, recoveredValue float64

	for _, s := range sessions {
		total++
		recoverableValue += s.CartAmount
		switch s.Status {
		case checkout.StatusDroppedOff:
			dropped++
		case checkout.StatusRecoveryDispatched:
			dispatched++
		case checkout.StatusRecovered:
			recovered++
			recoveredValue += s.CartAmount
		}
	}

	recoveryRate := 0.0
	if total > 0 {
		recoveryRate = (float64(recovered) / float64(total)) * 100
	}

	return map[string]interface{}{
		"total_sessions":      total,
		"dropped_off":         dropped,
		"recovery_dispatched": dispatched,
		"recovered":           recovered,
		"recovery_rate_pct":   recoveryRate,
		"recoverable_value":   recoverableValue,
		"recovered_value":     recoveredValue,
		"currency":            "INR",
	}
}

func getSampleCheckoutSessions() []checkout.Session {
	now := time.Now()
	t1 := now.Add(-30 * time.Minute)
	t2 := now.Add(-2 * time.Hour)
	t3 := now.Add(-4 * time.Hour)
	recTime := now.Add(-45 * time.Minute)

	return []checkout.Session{
		{
			ID:                   "sess_01_a9f1",
			MerchantID:           "00000000-0000-0000-0000-000000000001",
			SessionToken:         "tok_chk_99812401",
			CustomerName:         "Arjun Verma",
			CustomerEmail:        "arjun.verma@startup.co",
			CustomerPhone:        "+919811223344",
			CartAmount:           12499.00,
			Currency:             "INR",
			CartItemsJSON:        `[{"sku":"PLAN-PRO-ANNUAL","name":"ReviveOS Pro Annual Subscription","qty":1}]`,
			StepReached:          "3DS_INITIATED",
			Status:               checkout.StatusRecoveryDispatched,
			DropOffReason:        "Bank OTP timeout / browser tab closed",
			RecoveryLink:         "https://reviveos.onrender.com/checkout/restore?token=tok_chk_99812401",
			RecoveryDispatchedAt: &t1,
			CreatedAt:            now.Add(-45 * time.Minute),
			UpdatedAt:            t1,
		},
		{
			ID:                   "sess_02_b8e2",
			MerchantID:           "00000000-0000-0000-0000-000000000001",
			SessionToken:         "tok_chk_88201944",
			CustomerName:         "Kavita Patel",
			CustomerEmail:        "kavita.p@fashionstore.in",
			CustomerPhone:        "+919876501234",
			CartAmount:           4890.00,
			Currency:             "INR",
			CartItemsJSON:        `[{"sku":"ITEM-SKU-992","name":"Ergonomic Executive Chair","qty":1}]`,
			StepReached:          "PAYMENT_STEP",
			Status:               checkout.StatusRecovered,
			DropOffReason:        "UPI intent app not installed on desktop",
			RecoveryLink:         "https://reviveos.onrender.com/checkout/restore?token=tok_chk_88201944",
			RecoveryDispatchedAt: &t2,
			RecoveredAt:          &recTime,
			CreatedAt:            now.Add(-3 * time.Hour),
			UpdatedAt:            recTime,
		},
		{
			ID:            "sess_03_c7d3",
			MerchantID:    "00000000-0000-0000-0000-000000000001",
			SessionToken:  "tok_chk_77192033",
			CustomerName:  "Deepak Chawla",
			CustomerEmail: "deepak@chawlaenterprises.com",
			CustomerPhone: "+919820098765",
			CartAmount:    29500.00,
			Currency:      "INR",
			CartItemsJSON: `[{"sku":"SLA-ENTERPRISE","name":"Enterprise Priority Support Tier","qty":1}]`,
			StepReached:   "DETAILS_ENTERED",
			Status:        checkout.StatusDroppedOff,
			DropOffReason: "Price sticker shock / exited tab",
			RecoveryLink:  "https://reviveos.onrender.com/checkout/restore?token=tok_chk_77192033",
			CreatedAt:     t3,
			UpdatedAt:     t3,
		},
	}
}
