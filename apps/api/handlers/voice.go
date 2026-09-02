package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/voice"
)

// VoiceRecoveryHandler manages conversational Hinglish telephony calls and intent processing
func VoiceRecoveryHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			// List voice recovery call history & preview scripts
			if strings.HasSuffix(r.URL.Path, "/scripts/preview") {
				name := r.URL.Query().Get("customer_name")
				if name == "" {
					name = "Aman Gupta"
				}
				amount := 14999.0
				if amtStr := r.URL.Query().Get("amount"); amtStr != "" {
					var parsed float64
					if _, err := fmt.Sscanf(amtStr, "%f", &parsed); err == nil && parsed > 0 {
						amount = parsed
					}
				}
				payload := voice.CallPayload{
					CustomerName: name,
					Phone:        "+919876543210",
					Amount:       amount,
					Currency:     "INR",
					DueDate:      time.Now().Add(-3 * 24 * time.Hour),
				}
				script := voice.GenerateHinglishScript(payload)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":          "ok",
					"sample_payload":  payload,
					"hinglish_script": script,
					"supported_intents": []string{
						"PROMISE_TO_PAY ('kal karunga', 'somwar tak ho jayega')",
						"REQUEST_LINK ('WhatsApp pe link bhejo')",
						"DISPUTE ('galat amount hai')",
						"CALL_LATER ('abhi busy hoon')",
						"ALREADY_PAID ('pay kar diya')",
					},
					"provider_mode": getVoiceProvider(),
				})
				return
			}

			// List voice logs from database first
			var dbCalls []voice.CallResult
			if pool != nil {
				rows, qErr := pool.Query(r.Context(), `
					SELECT 
						COALESCE(provider_call_sid, id::text),
						provider,
						call_status,
						COALESCE(hinglish_script, ''),
						COALESCE(customer_response, ''),
						COALESCE(intent_detected, 'PROMISE_TO_PAY'),
						duration_seconds
					FROM voice_recovery_calls
					ORDER BY created_at DESC
					LIMIT 50
				`)
				if qErr == nil {
					defer rows.Close()
					for rows.Next() {
						var cr voice.CallResult
						var intentStr string
						if err := rows.Scan(
							&cr.CallSID,
							&cr.Provider,
							&cr.Status,
							&cr.HinglishScript,
							&cr.CustomerSpoken,
							&intentStr,
							&cr.DurationSec,
						); err == nil {
							cr.Intent = voice.Intent(intentStr)
							dbCalls = append(dbCalls, cr)
						}
					}
				}
			}

			// Merge database records with sample baseline
			sampleLogs := getSampleVoiceCalls()
			existingSids := make(map[string]bool)
			for _, c := range dbCalls {
				existingSids[c.CallSID] = true
			}
			for _, s := range sampleLogs {
				if !existingSids[s.CallSID] {
					dbCalls = append(dbCalls, s)
				}
			}

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"calls":  dbCalls,
				"total":  len(dbCalls),
			})

		case http.MethodPost:
			// Trigger a voice call or process spoken intent callback
			if strings.HasSuffix(r.URL.Path, "/webhook/incoming") {
				// Webhook from Twilio / Exotel speech-to-text
				var webhookReq struct {
					CallSID       string `json:"CallSid"`
					SpeechResult  string `json:"SpeechResult"`
					Digits        string `json:"Digits"`
					From          string `json:"From"`
				}
				_ = json.NewDecoder(r.Body).Decode(&webhookReq)

				speech := webhookReq.SpeechResult
				if speech == "" && webhookReq.Digits == "1" {
					speech = "Haan link bhejo"
				}

				intent, ptp := voice.ClassifyHinglishIntent(speech)

				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":         "received",
					"call_sid":       webhookReq.CallSID,
					"speech":         speech,
					"intent":         intent,
					"scheduled_date": ptp,
				})
				return
			}

			// Trigger call
			var callReq voice.CallPayload
			if err := json.NewDecoder(r.Body).Decode(&callReq); err != nil {
				http.Error(w, `{"error":"invalid voice call payload"}`, http.StatusBadRequest)
				return
			}

			if callReq.DueDate.IsZero() {
				callReq.DueDate = time.Now()
			}
			script := voice.GenerateHinglishScript(callReq)

			providerName := getVoiceProvider()
			var result *voice.CallResult
			var err error

			if providerName == "twilio" && os.Getenv("TWILIO_ACCOUNT_SID") != "" {
				twilioProvider := &voice.TwilioVoiceProvider{
					AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
					AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
					CallerID:   os.Getenv("TWILIO_VOICE_CALLER_ID"),
				}
				result, err = twilioProvider.InitiateCall(r.Context(), callReq, script)
			} else {
				localProvider := &voice.LocalSimulatorProvider{}
				result, err = localProvider.InitiateCall(r.Context(), callReq, script)
			}

			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"failed to initiate voice recovery call: %v"}`, err), http.StatusInternalServerError)
				return
			}
			result.Provider = providerName

			// Lodge call, customer transcript intent, and link into workflow audit trail
			if pool != nil {
				var ptpID *string
				// If customer committed a promise to pay date, lodge PTP record
				if result.Intent == voice.IntentPromiseToPay {
					var createdPtpID string
					ptpDate := time.Now().Add(24 * time.Hour)
					if result.PTPDate != nil {
						ptpDate = *result.PTPDate
					}
					_ = pool.QueryRow(r.Context(), `
						INSERT INTO promise_to_pay_records (
							customer_id, customer_name, customer_contact, promised_amount, promised_date, status, recorded_channel, notes
						) VALUES ($1, $2, $3, $4, $5, 'PENDING', 'VOICE_AGENT', $6)
						RETURNING id::text
					`, callReq.CustomerID, callReq.CustomerName, callReq.Phone, callReq.Amount, ptpDate, "Committed during Hinglish AI voice call").Scan(&createdPtpID)
					if createdPtpID != "" {
						ptpID = &createdPtpID
					}
				}

				// Lodge into voice_recovery_calls
				_, _ = pool.Exec(r.Context(), `
					INSERT INTO voice_recovery_calls (
						recipient_phone, customer_name, amount, currency, language, provider, provider_call_sid, call_status,
						duration_seconds, hinglish_script, customer_response, intent_detected, ptp_created_id
					) VALUES ($1, $2, $3, $4, 'Hinglish', $5, $6, $7, $8, $9, $10, $11, $12)
				`, callReq.Phone, callReq.CustomerName, callReq.Amount, callReq.Currency, providerName, result.CallSID,
					result.Status, result.DurationSec, script, result.CustomerSpoken, string(result.Intent), ptpID)

				// Lodge into recovery_workflows audit log if workflow_id provided or customer exists
				wfID := strings.TrimSpace(callReq.WorkflowID)
				// Check if wfID is a valid Postgres UUID (36 chars with dashes)
				isValidUUID := len(wfID) == 36 && strings.Count(wfID, "-") == 4
				if wfID == "" {
					_ = pool.QueryRow(r.Context(), `
						SELECT rw.id::text 
						FROM recovery_workflows rw
						JOIN payments p ON rw.payment_id = p.id
						JOIN customers c ON p.customer_id = c.id
						WHERE c.phone = $1 OR c.email = $2
						ORDER BY rw.created_at DESC LIMIT 1
					`, callReq.Phone, callReq.CustomerEmail).Scan(&wfID)
					if wfID != "" {
						isValidUUID = true
					}
				}

				if isValidUUID && wfID != "" {
					// Add action and audit trail
					_, _ = pool.Exec(r.Context(), `
						INSERT INTO recovery_actions (workflow_id, action_type, status, attempt, result, executed_at)
						VALUES ($1, 'VOICE_RECOVERY_CALL', 'EXECUTED', 1, $2, CURRENT_TIMESTAMP)
					`, wfID, fmt.Sprintf("Hinglish Call %s: Customer intent: %s. %s", result.CallSID, result.Intent, result.CustomerSpoken))

					// Record audit event
					metaBytes, _ := json.Marshal(map[string]interface{}{
						"phone":           callReq.Phone,
						"customer_name":   callReq.CustomerName,
						"call_sid":        result.CallSID,
						"spoken_response": result.CustomerSpoken,
						"intent":          result.Intent,
						"provider":        providerName,
					})
					_, _ = pool.Exec(r.Context(), `
						INSERT INTO audit_events (workflow_id, actor, action, metadata, timestamp)
						VALUES ($1, 'system:voice_agent', 'VOICE_CALL_DISPATCHED', $2, CURRENT_TIMESTAMP)
					`, wfID, metaBytes)
				}
			}

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":         "initiated",
				"call_result":    result,
				"provider":       providerName,
				"telephony_info": getTelephonyConfigSummary(),
			})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}

func getVoiceProvider() string {
	prov := strings.ToLower(os.Getenv("VOICE_PROVIDER"))
	if prov == "twilio" {
		return "twilio"
	}
	return "local"
}

func getTelephonyConfigSummary() map[string]interface{} {
	prov := getVoiceProvider()
	return map[string]interface{}{
		"active_provider":       prov,
		"twilio_configured":     os.Getenv("TWILIO_ACCOUNT_SID") != "",
		"resend_email_fallback": os.Getenv("RESEND_API_KEY") != "",
	}
}

func getSampleVoiceCalls() []voice.CallResult {
	now := time.Now()
	ptpDate1 := now.Add(24 * time.Hour)
	ptpDate2 := now.Add(48 * time.Hour)

	return []voice.CallResult{
		{
			CallSID:        "call_twilio_9021831",
			Provider:       "twilio",
			Status:         "COMPLETED",
			HinglishScript: "Namaste Rajesh ji! Main ReviveOS Payments Desk se baat kar raha hoon...",
			CustomerSpoken: "Haan bhai kal subah 11 baje tak payment kar dunga.",
			Intent:         voice.IntentPromiseToPay,
			PTPDate:        &ptpDate1,
			DurationSec:    45,
		},
		{
			CallSID:        "call_twilio_4810924",
			Provider:       "twilio",
			Status:         "COMPLETED",
			HinglishScript: "Namaste Priya ji! Aapka ₹4,999 ka autopay payment decline hua hai...",
			CustomerSpoken: "Aap mujhe WhatsApp pe UPI link bhej dijiye abhi kar deti hoon.",
			Intent:         voice.IntentRequestLink,
			PTPDate:        nil,
			DurationSec:    32,
		},
		{
			CallSID:        "call_local_3310892",
			Provider:       "local",
			Status:         "COMPLETED",
			HinglishScript: "Namaste Suresh ji! Aapka monthly subscription charge pending hai...",
			CustomerSpoken: "Salary aane me 2 din lagenge, tab karta hoon.",
			Intent:         voice.IntentPromiseToPay,
			PTPDate:        &ptpDate2,
			DurationSec:    51,
		},
	}
}
