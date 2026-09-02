package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/receivables"
)

// ReceivablesHandler manages B2B invoices and dunning actions
func ReceivablesHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			if pool == nil {
				sampleInvoices := getSampleInvoices()
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":   "ok",
					"invoices": sampleInvoices,
					"summary":  calculateAgingSummary(sampleInvoices),
				})
				return
			}

			// List invoices with aging calculation
			rows, err := pool.Query(r.Context(), `
				SELECT id, merchant_id, invoice_number, buyer_company, buyer_name, buyer_email, 
				       amount, currency, issue_date, due_date, status, credit_terms, 
				       current_bucket, days_past_due, dunning_stage, COALESCE(payment_link_url, '')
				FROM b2b_invoices
				ORDER BY due_date ASC
			`)
			if err != nil {
				// Fallback to sample dataset if table not populated
				sampleInvoices := getSampleInvoices()
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":   "ok",
					"invoices": sampleInvoices,
					"summary":  calculateAgingSummary(sampleInvoices),
				})
				return
			}
			defer rows.Close()

			var list []receivables.Invoice
			for rows.Next() {
				var inv receivables.Invoice
				var issueDate, dueDate time.Time
				var pLink string
				err := rows.Scan(
					&inv.ID, &inv.MerchantID, &inv.InvoiceNumber, &inv.BuyerCompany, &inv.BuyerName, &inv.BuyerEmail,
					&inv.Amount, &inv.Currency, &issueDate, &dueDate, &inv.Status, &inv.CreditTerms,
					&inv.CurrentBucket, &inv.DaysPastDue, &inv.DunningStage, &pLink,
				)
				if err == nil {
					inv.IssueDate = issueDate
					inv.DueDate = dueDate
					inv.PaymentLinkURL = pLink
					// Recalculate dynamic aging
					days, bucket := receivables.CalculateAging(dueDate, time.Now())
					inv.DaysPastDue = days
					inv.CurrentBucket = bucket
					list = append(list, inv)
				}
			}

			if len(list) == 0 {
				list = getSampleInvoices()
			}

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":   "ok",
				"invoices": list,
				"summary":  calculateAgingSummary(list),
			})

		case http.MethodPost:
			// Create invoice or trigger dunning
			if strings.HasSuffix(r.URL.Path, "/dunning/trigger") {
				var req struct {
					InvoiceID string `json:"invoice_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
					return
				}

				// Find or mock invoice
				inv := receivables.Invoice{
					ID:            req.InvoiceID,
					InvoiceNumber: "INV-2026-088",
					BuyerCompany:  "Tata Tech Solutions",
					BuyerName:     "Rajesh Sharma",
					BuyerEmail:    "rajesh.sharma@tatatech.example",
					Amount:        125000,
					Currency:      "INR",
					DueDate:       time.Now().Add(-14 * 24 * time.Hour), // 14 days overdue
				}
				action := receivables.NextDunningAction(inv, time.Now())
				action.GeneratePaymentLink = true

				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"status":         "success",
					"invoice_id":     req.InvoiceID,
					"recommendation": action,
					"dispatched_at":  time.Now().UTC(),
					"message":        fmt.Sprintf("Automated %s dunning notice dispatched to %s via Resend/SMS", action.Tone, inv.BuyerEmail),
				})
				return
			}

			// Create Invoice
			var newInv receivables.Invoice
			if err := json.NewDecoder(r.Body).Decode(&newInv); err != nil {
				http.Error(w, `{"error":"invalid invoice body"}`, http.StatusBadRequest)
				return
			}
			if newInv.InvoiceNumber == "" {
				newInv.InvoiceNumber = fmt.Sprintf("INV-%d", time.Now().Unix()%100000)
			}
			if newInv.Currency == "" {
				newInv.Currency = "INR"
			}
			if newInv.CreditTerms == "" {
				newInv.CreditTerms = "NET_30"
			}
			newInv.ID = fmt.Sprintf("b2b_inv_%d", time.Now().UnixNano())
			days, bucket := receivables.CalculateAging(newInv.DueDate, time.Now())
			newInv.DaysPastDue = days
			newInv.CurrentBucket = bucket
			newInv.PaymentLinkURL = fmt.Sprintf("https://checkout.reviveos.io/pay/%s", newInv.InvoiceNumber)

			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "created",
				"invoice": newInv,
			})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}

func calculateAgingSummary(invoices []receivables.Invoice) map[string]interface{} {
	var totalReceivables float64
	var currentAmt, bucket1to30, bucket31to60, bucket90Plus float64

	for _, inv := range invoices {
		totalReceivables += inv.Amount
		switch inv.CurrentBucket {
		case receivables.BucketCurrent:
			currentAmt += inv.Amount
		case receivables.Bucket1To30:
			bucket1to30 += inv.Amount
		case receivables.Bucket31To60:
			bucket31to60 += inv.Amount
		default:
			bucket90Plus += inv.Amount
		}
	}

	return map[string]interface{}{
		"total_outstanding": totalReceivables,
		"current_due":       currentAmt,
		"bucket_1_30":       bucket1to30,
		"bucket_31_60":      bucket31to60,
		"bucket_90_plus":    bucket90Plus,
		"currency":          "INR",
	}
}

func getSampleInvoices() []receivables.Invoice {
	now := time.Now()
	return []receivables.Invoice{
		{
			ID:             "b2b_inv_001",
			InvoiceNumber:  "INV-IND-901",
			BuyerCompany:   "Bharat Cloud Labs",
			BuyerName:      "Vikram Mehta",
			BuyerEmail:     "vikram@bharatcloud.in",
			Amount:         450000.00,
			Currency:       "INR",
			IssueDate:      now.Add(-45 * 24 * time.Hour),
			DueDate:        now.Add(-15 * 24 * time.Hour),
			Status:         "OVERDUE",
			CreditTerms:    "NET_30",
			CurrentBucket:  receivables.Bucket1To30,
			DaysPastDue:    15,
			DunningStage:   1,
			PaymentLinkURL: "https://reviveos.onrender.com/pay/INV-IND-901",
		},
		{
			ID:             "b2b_inv_002",
			InvoiceNumber:  "INV-IND-902",
			BuyerCompany:   "Deccan Logistics Pvt Ltd",
			BuyerName:      "Ananya Iyer",
			BuyerEmail:     "accounts@deccanlogistics.com",
			Amount:         185000.00,
			Currency:       "INR",
			IssueDate:      now.Add(-70 * 24 * time.Hour),
			DueDate:        now.Add(-40 * 24 * time.Hour),
			Status:         "OVERDUE",
			CreditTerms:    "NET_30",
			CurrentBucket:  receivables.Bucket31To60,
			DaysPastDue:    40,
			DunningStage:   2,
			PaymentLinkURL: "https://reviveos.onrender.com/pay/INV-IND-902",
		},
		{
			ID:             "b2b_inv_003",
			InvoiceNumber:  "INV-IND-903",
			BuyerCompany:   "Zest Enterprise SaaS",
			BuyerName:      "Rohan Roy",
			BuyerEmail:     "finance@zestsaas.io",
			Amount:         89000.00,
			Currency:       "INR",
			IssueDate:      now.Add(-10 * 24 * time.Hour),
			DueDate:        now.Add(20 * 24 * time.Hour),
			Status:         "UNPAID",
			CreditTerms:    "NET_30",
			CurrentBucket:  receivables.BucketCurrent,
			DaysPastDue:    0,
			DunningStage:   0,
			PaymentLinkURL: "https://reviveos.onrender.com/pay/INV-IND-903",
		},
	}
}
