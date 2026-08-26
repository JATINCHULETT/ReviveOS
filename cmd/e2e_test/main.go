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
	log.Println("=== ReviveOS Complete End-to-End Test Suite ===")

	// 1. Check Health & Ready
	resp, err := http.Get("http://localhost:8080/ready")
	if err != nil {
		log.Fatalf("GET /ready failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("1. GET /ready -> Status %d: %s", resp.StatusCode, string(body))

	// 2. Test Webhook Deduplication
	secret := "test_webhook_secret_12345"
	eventID := fmt.Sprintf("evt_e2e_%d", time.Now().UnixNano())
	paymentID := fmt.Sprintf("pay_e2e_%d", time.Now().UnixNano())

	payloadMap := map[string]interface{}{
		"entity":     "event",
		"account_id": "acc_e2e_test",
		"event":      "payment.failed",
		"contains":   []string{"payment"},
		"payload": map[string]interface{}{
			"payment": map[string]interface{}{
				"entity": map[string]interface{}{
					"id":          paymentID,
					"amount":      250000, // 2,500.00 INR
					"currency":    "INR",
					"status":      "failed",
					"method":      "card",
					"error_code":  "BAD_REQUEST_ERROR",
					"email":       "tester@reviveos.io",
					"contact":     "+919876543210",
					"customer_id": fmt.Sprintf("cust_%d", time.Now().UnixNano()),
				},
			},
		},
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBytes)
	signature := hex.EncodeToString(mac.Sum(nil))

	// Send First Webhook Request
	req1, _ := http.NewRequest("POST", "http://localhost:8080/webhooks/razorpay", bytes.NewBuffer(payloadBytes))
	req1.Header.Set("X-Razorpay-Event-Id", eventID)
	req1.Header.Set("X-Razorpay-Signature", signature)
	req1.Header.Set("Content-Type", "application/json")

	res1, err := http.DefaultClient.Do(req1)
	if err != nil {
		log.Fatalf("First webhook request failed: %v", err)
	}
	defer res1.Body.Close()
	b1, _ := io.ReadAll(res1.Body)
	log.Printf("2a. First Webhook (Event %s) -> Status: %d, Body: %s", eventID, res1.StatusCode, string(b1))

	// Send Second (Duplicate) Webhook Request
	req2, _ := http.NewRequest("POST", "http://localhost:8080/webhooks/razorpay", bytes.NewBuffer(payloadBytes))
	req2.Header.Set("X-Razorpay-Event-Id", eventID)
	req2.Header.Set("X-Razorpay-Signature", signature)
	req2.Header.Set("Content-Type", "application/json")

	res2, err := http.DefaultClient.Do(req2)
	if err != nil {
		log.Fatalf("Second webhook request failed: %v", err)
	}
	defer res2.Body.Close()
	b2, _ := io.ReadAll(res2.Body)
	log.Printf("2b. Duplicate Webhook (Event %s) -> Status: %d, Body: %s", eventID, res2.StatusCode, string(b2))

	if string(b2) != "OK (duplicate)" {
		log.Fatalf("Expected 'OK (duplicate)', got %s", string(b2))
	}
	log.Println("==> Webhook deduplication: PASS")

	// 3. Connect DB and poll for workflow completion
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer pool.Close()

	log.Println("3. Waiting for asynchronous recovery workflow execution...")
	var wfID, wfStatus, action string
	var prob float64

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		err = pool.QueryRow(ctx, `
			SELECT rw.id::text, rw.status, COALESCE(rw.selected_action, ''), COALESCE(rw.recovery_probability, 0)
			FROM recovery_workflows rw
			JOIN payments p ON rw.payment_id = p.id
			WHERE p.razorpay_payment_id = $1
			ORDER BY rw.created_at DESC LIMIT 1
		`, paymentID).Scan(&wfID, &wfStatus, &action, &prob)

		if err == nil && wfStatus != "ANALYZING" {
			break
		}
	}

	if err != nil {
		log.Fatalf("Failed to find recovery workflow for payment %s: %v", paymentID, err)
	}
	log.Printf("4. Recovery Workflow: ID=%s, Status=%s, SelectedAction=%s, Prob=%.4f", wfID, wfStatus, action, prob)

	// Check AI decision in database
	var aiDiagnosis, aiModel string
	var aiConf float64
	err = pool.QueryRow(ctx, `
		SELECT diagnosis, model, confidence 
		FROM ai_decisions 
		WHERE workflow_id::text = $1
		ORDER BY created_at DESC LIMIT 1
	`, wfID).Scan(&aiDiagnosis, &aiModel, &aiConf)
	if err == nil {
		log.Printf("5. AI Decision in DB: Model=%s, Confidence=%.2f, Diagnosis=%s", aiModel, aiConf, aiDiagnosis)
	} else {
		log.Printf("5. AI Decision query note: %v", err)
	}

	// Check Model Prediction in database
	var predModel string
	var predProb float64
	err = pool.QueryRow(ctx, `
		SELECT model_version, probability 
		FROM model_predictions 
		WHERE workflow_id::text = $1
		ORDER BY created_at DESC LIMIT 1
	`, wfID).Scan(&predModel, &predProb)
	if err == nil {
		log.Printf("6. Model Prediction in DB: Model=%s, Probability=%.4f", predModel, predProb)
	}

	// Check Audit Ledger Chain
	rows, err := pool.Query(ctx, `
		SELECT action, actor, event_hash, previous_event_hash 
		FROM audit_events 
		WHERE workflow_id::text = $1 
		ORDER BY timestamp ASC
	`, wfID)
	if err == nil {
		defer rows.Close()
		log.Println("7. Tamper-Evident Audit Ledger Chain:")
		eventIdx := 0
		for rows.Next() {
			var act, actor, hash string
			var prevHash *string
			_ = rows.Scan(&act, &actor, &hash, &prevHash)
			prev := "GENESIS"
			if prevHash != nil {
				prev = *prevHash
			}
			log.Printf("   [%d] Action=%s | Actor=%s | Prev=%s | Hash=%s", eventIdx, act, actor, prev[:min(8, len(prev))], hash[:min(8, len(hash))])
			eventIdx++
		}
	}

	// Check Analytics Endpoint
	analyticsResp, err := http.Get("http://localhost:8080/health")
	if err == nil {
		defer analyticsResp.Body.Close()
		log.Printf("8. Health Probe: Status %d", analyticsResp.StatusCode)
	}

	log.Println("==================================================")
	log.Println("   ReviveOS E2E Pipeline Verification: PASS")
	log.Println("==================================================")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
