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
				payload := voice.CallPayload{
					CustomerName: "Aman Gupta",
					Phone:        "+919876543210",
					Amount:       14999,
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

			// List voice logs
			logs := getSampleVoiceCalls()
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "ok",
				"calls":  logs,
				"total":  len(logs),
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
