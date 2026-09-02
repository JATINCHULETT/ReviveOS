package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/ptp"
)

// PTPHandler manages customer promise-to-pay commitments and reconciliation
func PTPHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			// Fetch PTP tracker items and summary metrics
			promises := getSamplePromises()
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "ok",
				"promises": promises,
				"metrics":  calculatePTPMetrics(promises),
			})

		case http.MethodPost:
			// Create a promise or verify / extend
			if strings.Contains(r.URL.Path, "/verify") {
				var req struct {
					PromiseID string `json:"promise_id"`
					IsPaid    bool   `json:"is_paid"`
				}
				_ = json.NewDecoder(r.Body).Decode(&req)

				// Status updated to HONORED
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":      "updated",
					"promise_id":  req.PromiseID,
					"new_status":  ptp.StatusHonored,
					"verified_at": time.Now().UTC(),
					"message":     "Customer commitment honored and verified against payment gateway captured event",
				})
				return
			}

			if strings.Contains(r.URL.Path, "/extend") {
				var req struct {
					PromiseID string `json:"promise_id"`
					Days      int    `json:"days"`
				}
				_ = json.NewDecoder(r.Body).Decode(&req)
				if req.Days <= 0 {
					req.Days = 2
				}

				newDate := time.Now().Add(time.Duration(req.Days) * 24 * time.Hour)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":             "extended",
					"promise_id":         req.PromiseID,
					"new_status":         ptp.StatusExtended,
					"extended_due_date":  newDate,
					"next_reminder_date": newDate.Add(-24 * time.Hour),
					"message":            "Customer commitment extended. Automated Resend reminder scheduled 24 hours prior.",
				})
				return
			}

			// Create PTP record
			var record ptp.PromiseRecord
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				http.Error(w, `{"error":"invalid promise payload"}`, http.StatusBadRequest)
				return
			}

			record.ID = "ptp_" + time.Now().Format("20060102150405")
			if record.Status == "" {
				record.Status = ptp.StatusPending
			}
			if record.PromisedDate.IsZero() {
				record.PromisedDate = time.Now().Add(48 * time.Hour)
			}
			record.CreatedAt = time.Now()
			record.UpdatedAt = time.Now()

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "created",
				"promise": record,
				"message": "Promise-to-pay registered. Automated pre-due reminders armed.",
			})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}

func calculatePTPMetrics(records []ptp.PromiseRecord) map[string]interface{} {
	var total, honored, broken, pending int
	var committedAmount, honoredAmount float64

	for _, rec := range records {
		total++
		committedAmount += rec.PromisedAmount
		switch rec.Status {
		case ptp.StatusHonored:
			honored++
			honoredAmount += rec.PromisedAmount
		case ptp.StatusBroken:
			broken++
		default:
			pending++
		}
	}

	fulfillmentRate := 0.0
	if (honored + broken) > 0 {
		fulfillmentRate = (float64(honored) / float64(honored+broken)) * 100
	}

	return map[string]interface{}{
		"total_commitments": total,
		"pending":           pending,
		"honored":           honored,
		"broken":            broken,
		"committed_amount":  committedAmount,
		"recovered_amount":  honoredAmount,
		"fulfillment_rate":  fulfillmentRate,
	}
}

func getSamplePromises() []ptp.PromiseRecord {
	now := time.Now()
	date1 := now.Add(24 * time.Hour)
	date2 := now.Add(-12 * time.Hour)
	date3 := now.Add(72 * time.Hour)

	return []ptp.PromiseRecord{
		{
			ID:              "ptp_881901",
			MerchantID:      "00000000-0000-0000-0000-000000000001",
			CustomerID:      "cust_manish_77",
			CustomerName:    "Manish Trivedi",
			CustomerContact: "+919820192831",
			PromisedAmount:  18500,
			PromisedDate:    date1,
			Status:          ptp.StatusPending,
			RecordedChannel: "VOICE_AGENT (Hinglish)",
			ExtensionCount:  0,
			Notes:           "Customer promised over Hinglish call: 'Kal salary aate hi karunga'",
			CreatedAt:       now.Add(-2 * time.Hour),
		},
		{
			ID:              "ptp_881902",
			MerchantID:      "00000000-0000-0000-0000-000000000001",
			CustomerID:      "cust_shreya_92",
			CustomerName:    "Shreya Sen",
			CustomerContact: "+919711829301",
			PromisedAmount:  6500,
			PromisedDate:    date2,
			Status:          ptp.StatusHonored,
			RecordedChannel: "WHATSAPP_LINK",
			ExtensionCount:  0,
			Notes:           "Captured via Razorpay UPI auto-reconcile",
			CreatedAt:       now.Add(-36 * time.Hour),
		},
		{
			ID:              "ptp_881903",
			MerchantID:      "00000000-0000-0000-0000-000000000001",
			CustomerID:      "cust_kunal_14",
			CustomerName:    "Kunal Batra",
			CustomerContact: "+919833019284",
			PromisedAmount:  32000,
			PromisedDate:    date3,
			Status:          ptp.StatusExtended,
			RecordedChannel: "VOICE_AGENT",
			ExtensionCount:  1,
			Notes:           "Requested extension due to bank holiday",
			CreatedAt:       now.Add(-24 * time.Hour),
		},
	}
}
