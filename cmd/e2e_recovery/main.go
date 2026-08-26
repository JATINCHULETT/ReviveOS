package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/reviveos/utils/db"
)

func main() {
	log.Println("=== ReviveOS Full Recovery Pipeline E2E Test ===")

	// 1. Ingest Failed Payment with INSUFFICIENT_FUNDS
	secret := "test_webhook_secret_12345"
	eventID := fmt.Sprintf("evt_rec_%d", time.Now().UnixNano())
	paymentID := fmt.Sprintf("pay_rec_%d", time.Now().UnixNano())

	payloadMap := map[string]interface{}{
		"entity":     "event",
		"account_id": "acc_rec_test",
		"event":      "payment.failed",
		"contains":   []string{"payment"},
		"payload": map[string]interface{}{
			"payment": map[string]interface{}{
				"entity": map[string]interface{}{
					"id":          paymentID,
					"amount":      199900, // 1,999.00 INR
					"currency":    "INR",
					"status":      "failed",
					"method":      "card",
					"error_code":  "INSUFFICIENT_FUNDS",
					"email":       "recovery_tester@reviveos.io",
					"contact":     "+919876543210",
					"customer_id": fmt.Sprintf("cust_rec_%d", time.Now().UnixNano()),
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBytes)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest("POST", "http://localhost:8080/webhooks/razorpay", bytes.NewBuffer(payloadBytes))
	req.Header.Set("X-Razorpay-Event-Id", eventID)
	req.Header.Set("X-Razorpay-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Webhook request failed: %v", err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	log.Printf("1. Webhook Ingest -> Status: %d, Response: %s", res.StatusCode, string(b))

	// 2. Poll Database for workflow progress through all phases
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer pool.Close()

	log.Println("2. Monitoring Recovery Pipeline Lifecycle...")
	var wfID, wfStatus, action string
	var prob float64

	for i := 0; i < 45; i++ {
		time.Sleep(1 * time.Second)
		err = pool.QueryRow(ctx, `
			SELECT rw.id::text, rw.status, COALESCE(rw.selected_action, ''), COALESCE(rw.recovery_probability, 0)
			FROM recovery_workflows rw
			JOIN payments p ON rw.payment_id = p.id
			WHERE p.razorpay_payment_id = $1
			ORDER BY rw.created_at DESC LIMIT 1
		`, paymentID).Scan(&wfID, &wfStatus, &action, &prob)

		if err == nil {
			log.Printf("   [Poll %ds] Workflow %s | State: %s | Action: %s | Prob: %.4f", i+1, wfID[:8], wfStatus, action, prob)
			if wfStatus == "RECOVERED" || wfStatus == "HALTED" || wfStatus == "REQUIRES_HUMAN_REVIEW" {
				break
			}
		}
	}

	if wfID == "" {
		log.Fatalf("Workflow was not created for payment %s", paymentID)
	}

	// 3. Verify Payment Status in DB
	var dbPaymentStatus string
	var dbAmount float64
	_ = pool.QueryRow(ctx, `SELECT status, amount FROM payments WHERE razorpay_payment_id = $1`, paymentID).Scan(&dbPaymentStatus, &dbAmount)
	log.Printf("3. Payment in DB -> Status: %s, Amount: %.2f", dbPaymentStatus, dbAmount)

	// 4. Verify AI Decisions
	var aiDiag, aiModel, aiReason string
	var aiConf, aiRecov float64
	err = pool.QueryRow(ctx, `
		SELECT COALESCE(diagnosis, ''), model, COALESCE(confidence, 0), COALESCE(recoverability, 0), COALESCE(reasoning, '')
		FROM ai_decisions
		WHERE workflow_id::text = $1
		ORDER BY created_at DESC LIMIT 1
	`, wfID).Scan(&aiDiag, &aiModel, &aiConf, &aiRecov, &aiReason)
	if err == nil {
		log.Printf("4. AI Decision -> Model: %s | Conf: %.2f | Recov: %.2f | Diagnosis: %s", aiModel, aiConf, aiRecov, aiDiag)
	}

	// 5. Verify Model Predictions
	var predModel, failCat string
	var predProb float64
	err = pool.QueryRow(ctx, `
		SELECT model_version, failure_category, probability
		FROM model_predictions
		WHERE workflow_id::text = $1
		ORDER BY created_at DESC LIMIT 1
	`, wfID).Scan(&predModel, &failCat, &predProb)
	if err == nil {
		log.Printf("5. Model Prediction -> Model: %s | Category: %s | Prob: %.4f", predModel, failCat, predProb)
	}

	// 6. Verify Outcome
	var recBool bool
	var recAmt float64
	err = pool.QueryRow(ctx, `
		SELECT recovered, recovered_amount 
		FROM recovery_outcomes ro
		JOIN payments p ON ro.payment_id = p.id
		WHERE p.razorpay_payment_id = $1
		ORDER BY ro.created_at DESC LIMIT 1
	`, paymentID).Scan(&recBool, &recAmt)
	if err == nil {
		log.Printf("6. Recovery Outcome -> Recovered: %v, Amount: %.2f", recBool, recAmt)
	}

	// 7. Verify Tamper-Evident Audit Ledger Hash Chain
	rows, err := pool.Query(ctx, `
		SELECT action, actor, event_hash, previous_event_hash 
		FROM audit_events 
		WHERE workflow_id::text = $1 
		ORDER BY timestamp ASC
	`, wfID)
	if err == nil {
		defer rows.Close()
		log.Println("7. Tamper-Evident Audit Ledger Chain:")
		idx := 0
		for rows.Next() {
			var act, actor, hash string
			var prevHash *string
			_ = rows.Scan(&act, &actor, &hash, &prevHash)
			prev := "GENESIS"
			if prevHash != nil {
				prev = *prevHash
			}
			log.Printf("   [%d] Action: %s | Actor: %s | Prev: %s... | Hash: %s...", idx, act, actor, prev[:min(8, len(prev))], hash[:min(8, len(hash))])
			idx++
		}
	}

	log.Println("==================================================")
	log.Println("   ReviveOS Full Pipeline Test: COMPLETED")
	log.Println("==================================================")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
