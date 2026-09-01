package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/reviveos/utils/db"
)

func main() {
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Starting database wipe and 2,500 synthetic dataset seeding...")

	// 1. Wipe existing transactional and analytics data safely
	wipeQuery := `
		TRUNCATE TABLE audit_logs CASCADE;
		TRUNCATE TABLE audit_events CASCADE;
		TRUNCATE TABLE risk_assessments CASCADE;
		TRUNCATE TABLE recovery_outcomes CASCADE;
		TRUNCATE TABLE recovery_actions CASCADE;
		TRUNCATE TABLE ai_decisions CASCADE;
		TRUNCATE TABLE model_predictions CASCADE;
		TRUNCATE TABLE recovery_workflows CASCADE;
		TRUNCATE TABLE failure_events CASCADE;
		TRUNCATE TABLE payment_events CASCADE;
		TRUNCATE TABLE payments CASCADE;
		TRUNCATE TABLE subscriptions CASCADE;
		TRUNCATE TABLE customers CASCADE;
		TRUNCATE TABLE policies CASCADE;
		TRUNCATE TABLE api_keys CASCADE;
	`
	_, err = pool.Exec(ctx, wipeQuery)
	if err != nil {
		log.Fatalf("Failed to truncate tables: %v", err)
	}
	log.Println("Existing transactional tables truncated cleanly.")

	// 2. Reseed Merchants
	merchants := []struct {
		ID   string
		Name string
	}{
		{"00000000-0000-0000-0000-000000000001", "Acme Cloud Services"},
		{"00000000-0000-0000-0000-000000000002", "Zenith Health SaaS"},
		{"00000000-0000-0000-0000-000000000003", "FinPulse Technologies"},
	}

	for _, m := range merchants {
		_, err := pool.Exec(ctx, `
			INSERT INTO merchants (id, name, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
		`, m.ID, m.Name)
		if err != nil {
			log.Fatalf("Failed to seed merchant %s: %v", m.Name, err)
		}

		// Policy
		_, err = pool.Exec(ctx, `
			INSERT INTO policies (merchant_id, max_retries, max_contacts, max_recovery_window, confidence_threshold, amount_threshold)
			VALUES ($1, 3, 2, '7 days'::interval, 0.70, 50000)
			ON CONFLICT (merchant_id) DO NOTHING
		`, m.ID)
		if err != nil {
			log.Fatalf("Failed to seed policy for %s: %v", m.Name, err)
		}
	}
	log.Println("Merchants and policies seeded.")

	// 3. Reseed Users (Admin & Merchants)
	// admin@reviveos.io -> admin123
	// merchant@acme.com -> merchant123
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, merchant_id, email, password_hash, name, role)
		VALUES 
			('10000000-0000-0000-0000-000000000001', NULL, 'admin@reviveos.io', '2167d46816a7dc9fae5e6e66e746a5b2:fbf0f4e24ef5a4e320f305085e3cb289b4f2c050ec469f3796fcb1d283626e2e', 'System Administrator', 'ADMIN'),
			('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'merchant@acme.com', '2167d46816a7dc9fae5e6e66e746a5b2:c7e75525c56784865103a743419ea7bbde0490b4d45543c7b8d4b3dfba5bce68', 'Acme Merchant Owner', 'MERCHANT')
		ON CONFLICT (email) DO NOTHING
	`)
	if err != nil {
		log.Printf("Warning seeding users: %v", err)
	}

	// 4. Seed API Keys for Dev Platform
	testKeyHash := sha256.Sum256([]byte("rvo_test_acme_secret_key_12345"))
	liveKeyHash := sha256.Sum256([]byte("rvo_live_acme_prod_key_99999"))
	_, err = pool.Exec(ctx, `
		INSERT INTO api_keys (merchant_id, key_prefix, key_hash, name, mode, is_active)
		VALUES 
			('00000000-0000-0000-0000-000000000001', 'rvo_test_acme_', $1, 'Development Sandbox Key', 'test', true),
			('00000000-0000-0000-0000-000000000001', 'rvo_live_acme_', $2, 'Production Live Key', 'live', true)
	`, hex.EncodeToString(testKeyHash[:]), hex.EncodeToString(liveKeyHash[:]))
	if err != nil {
		log.Printf("Warning seeding api_keys: %v", err)
	}

	// 5. Generate 350 Realistic Customers
	r := rand.New(rand.NewSource(42)) // Deterministic seed for reproducible high quality
	firstNames := []string{"Aarav", "Aditi", "Rohan", "Priya", "Vikram", "Sneha", "Ananya", "Rahul", "Kavita", "Siddharth", "Meera", "Arjun", "Neha", "Rajesh", "Pooja", "Varun", "Tanvi", "Nikhil", "Ishita", "Gaurav"}
	lastNames := []string{"Sharma", "Verma", "Patel", "Mehta", "Chopra", "Gupta", "Nair", "Iyer", "Rao", "Joshi", "Singhania", "Reddy", "Deshmukh", "Kapoor", "Bhatia", "Saxena", "Mishra", "Banerjee", "Kulkarni", "Aggarwal"}
	domains := []string{"gmail.com", "outlook.com", "enterprise.in", "techcorp.io", "startup.co", "yahoo.com", "fintech.ai", "acme.com"}

	customerIDs := make([]string, 350)
	customerEmails := make([]string, 350)
	customerPhones := make([]string, 350)
	customerMerchants := make([]string, 350)
	customerOptOuts := make([]bool, 350)

	for i := 0; i < 350; i++ {
		custID := uuid.New().String()
		fn := firstNames[r.Intn(len(firstNames))]
		ln := lastNames[r.Intn(len(lastNames))]
		domain := domains[r.Intn(len(domains))]
		email := fmt.Sprintf("%s.%s%d@%s", fn, ln, r.Intn(900)+100, domain)
		phone := fmt.Sprintf("+91%d", 9800000000+r.Int63n(199999999))
		merchantID := merchants[r.Intn(len(merchants))].ID
		optOut := r.Float64() < 0.05 // 5% opt-out

		prefMethods := []string{"card", "upi"}
		if r.Float64() < 0.4 {
			prefMethods = append(prefMethods, "netbanking")
		}
		prefJSON, _ := json.Marshal(map[string]interface{}{
			"preferred_methods": prefMethods,
			"tier":              []string{"starter", "pro", "enterprise"}[r.Intn(3)],
		})

		_, err := pool.Exec(ctx, `
			INSERT INTO customers (id, merchant_id, email, phone, communication_opt_out, preferences, payment_profile, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, '{}'::jsonb, NOW() - ($7 || ' days')::interval, NOW())
		`, custID, merchantID, email, phone, optOut, prefJSON, fmt.Sprintf("%d", r.Intn(60)+10))
		if err != nil {
			log.Fatalf("Failed to seed customer: %v", err)
		}

		customerIDs[i] = custID
		customerEmails[i] = email
		customerPhones[i] = phone
		customerMerchants[i] = merchantID
		customerOptOuts[i] = optOut
	}
	log.Printf("Seeded %d customers.", len(customerIDs))

	// 6. Generate 450 Subscriptions
	subscriptionIDs := make([]string, 450)
	planAmounts := []float64{499, 999, 1499, 2499, 4999, 9999, 19999, 49999}
	for i := 0; i < 450; i++ {
		subID := uuid.New().String()
		custIdx := r.Intn(len(customerIDs))
		amt := planAmounts[r.Intn(len(planAmounts))]
		status := "ACTIVE"
		if r.Float64() < 0.15 {
			status = "PAST_DUE"
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO subscriptions (id, merchant_id, customer_id, amount, currency, status, billing_interval, next_billing_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'INR', $5, 'monthly', NOW() + ($6 || ' days')::interval, NOW() - ($7 || ' days')::interval, NOW())
		`, subID, customerMerchants[custIdx], customerIDs[custIdx], amt, status, r.Intn(28)+1, r.Intn(90)+10)
		if err != nil {
			log.Fatalf("Failed to seed subscription: %v", err)
		}
		subscriptionIDs[i] = subID
	}
	log.Printf("Seeded %d recurring subscriptions.", len(subscriptionIDs))

	// 7. Seed 2,500 Comprehensive Payment Scenarios
	log.Println("Generating 2,500 realistic payment scenarios...")

	type CaseType int
	const (
		CaseInsufficientFunds CaseType = iota
		CaseBankDowntime
		CaseAuthFailed
		CaseExpiredCard
		CaseLimitExceeded
		CaseSuspectedFraud
	)

	// Batch insertion structures
	totalCount := 2500
	recoveredCount := 0
	fraudBlockedCount := 0
	var totalRecoveredRev float64
	var totalAtRiskRev float64

	now := time.Now()

	for i := 0; i < totalCount; i++ {
		custIdx := r.Intn(len(customerIDs))
		merchantID := customerMerchants[custIdx]
		custID := customerIDs[custIdx]

		var subID *string
		if r.Float64() < 0.65 {
			s := subscriptionIDs[r.Intn(len(subscriptionIDs))]
			subID = &s
		}

		// Pick realistic timestamp in last 30 days
		daysAgo := r.Float64() * 30.0
		hoursAgo := daysAgo*24.0 + r.Float64()*12.0
		createdAt := now.Add(-time.Duration(hoursAgo) * time.Hour)

		// Pick case type based on distribution:
		// 35% Insufficient Funds, 25% Bank Downtime, 18% Auth Failed, 10% Expired Card, 6% Limit, 6% Fraud
		randVal := r.Float64()
		var cType CaseType
		if randVal < 0.35 {
			cType = CaseInsufficientFunds
		} else if randVal < 0.60 {
			cType = CaseBankDowntime
		} else if randVal < 0.78 {
			cType = CaseAuthFailed
		} else if randVal < 0.88 {
			cType = CaseExpiredCard
		} else if randVal < 0.94 {
			cType = CaseLimitExceeded
		} else {
			cType = CaseSuspectedFraud
		}

		// Configure details per case
		var (
			failureCode        string
			failureCategory    string
			method             string
			amount             float64
			recoveryProb       float64
			fraudProb          float64
			fraudLevel         string
			returnProb         float64
			returnLevel        string
			overallRisk        string
			expectedLoss       float64
			riskAction         string
			selectedAction     string
			actionType         string
			wfStatus           string
			isRecovered        bool
			delayHours         int
			decisionReason     string
			aiDiagnosis        string
		)

		// Payment method selection
		methodRoll := r.Float64()
		if methodRoll < 0.45 {
			method = "card"
		} else if methodRoll < 0.80 {
			method = "upi"
		} else if methodRoll < 0.92 {
			method = "netbanking"
		} else {
			method = "wallet"
		}

		// Amount selection
		if r.Float64() < 0.80 {
			amount = planAmounts[r.Intn(len(planAmounts))]
		} else {
			// Enterprise tier
			amount = float64(r.Intn(60)+25) * 1000.0 // 25k - 85k
		}
		totalAtRiskRev += amount

		switch cType {
		case CaseInsufficientFunds:
			failureCode = "INSUFFICIENT_FUNDS"
			failureCategory = "INSUFFICIENT_FUNDS"
			recoveryProb = 0.76 + (r.Float64()*0.14 - 0.05) // 0.71 - 0.85
			fraudProb = 0.02 + r.Float64()*0.05
			fraudLevel = "LOW"
			returnProb = 0.01 + r.Float64()*0.04
			returnLevel = "LOW"
			overallRisk = "LOW"
			expectedLoss = 0.0
			riskAction = "ALLOW"
			selectedAction = "DELAYED_RETRY"
			actionType = "DELAYED_RETRY"
			delayHours = 24
			decisionReason = "Customer balance replenish window identified; scheduled smart zero-touch token retry on payday window."
			aiDiagnosis = "Insufficient funds at recurring cycle. Customer has positive 4-month payment history."
			isRecovered = r.Float64() < 0.78

		case CaseBankDowntime:
			failureCode = []string{"BANK_UNAVAILABLE", "BANK_DOWNTIME", "GATEWAY_TIMEOUT"}[r.Intn(3)]
			failureCategory = "BANK_UNAVAILABLE"
			recoveryProb = 0.90 + r.Float64()*0.08 // 0.90 - 0.98
			fraudProb = 0.01 + r.Float64()*0.03
			fraudLevel = "LOW"
			returnProb = 0.01 + r.Float64()*0.02
			returnLevel = "LOW"
			overallRisk = "LOW"
			expectedLoss = 0.0
			riskAction = "ALLOW"
			selectedAction = "IMMEDIATE_RETRY"
			actionType = "IMMEDIATE_RETRY"
			delayHours = 1
			decisionReason = "Temporary issuing bank gateway downtime. Immediate retry scheduled with exponential jitter backoff."
			aiDiagnosis = "Issuer banking switch unreachable. Transient network drop."
			isRecovered = r.Float64() < 0.92

		case CaseAuthFailed:
			failureCode = []string{"AUTHENTICATION_FAILED", "3DS_TIMEOUT", "OTP_EXPIRED"}[r.Intn(3)]
			failureCategory = "AUTHENTICATION_FAILED"
			recoveryProb = 0.68 + (r.Float64()*0.12 - 0.04) // 0.64 - 0.76
			fraudProb = 0.06 + r.Float64()*0.09
			fraudLevel = "LOW"
			returnProb = 0.02 + r.Float64()*0.06
			returnLevel = "LOW"
			overallRisk = "LOW"
			expectedLoss = 0.0
			riskAction = "ALLOW"
			selectedAction = "PAYMENT_LINK"
			actionType = "PAYMENT_LINK"
			delayHours = 0
			decisionReason = "Customer dropped 3DS authentication. Dispatched smart 1-click payment link via Resend email & WhatsApp."
			aiDiagnosis = "3DS friction / OTP timeout. High customer intent detected."
			isRecovered = r.Float64() < 0.72

		case CaseExpiredCard:
			failureCode = []string{"EXPIRED_CARD", "MANDATE_LAPSED"}[r.Intn(2)]
			failureCategory = "EXPIRED_CARD"
			recoveryProb = 0.58 + r.Float64()*0.12 // 0.58 - 0.70
			fraudProb = 0.02 + r.Float64()*0.04
			fraudLevel = "LOW"
			returnProb = 0.02 + r.Float64()*0.05
			returnLevel = "LOW"
			overallRisk = "LOW"
			expectedLoss = 0.0
			riskAction = "ALLOW"
			selectedAction = "UPDATE_PAYMENT_METHOD"
			actionType = "UPDATE_PAYMENT_METHOD"
			delayHours = 0
			decisionReason = "Card expired or mandate lapsed. Generated secure payment method update link for customer."
			aiDiagnosis = "Card credentials expired. Mandate renewal required."
			isRecovered = r.Float64() < 0.64

		case CaseLimitExceeded:
			failureCode = []string{"LIMIT_EXCEEDED", "VELOCITY_LIMIT_EXCEEDED"}[r.Intn(2)]
			failureCategory = "LIMIT_EXCEEDED"
			recoveryProb = 0.72 + r.Float64()*0.13
			fraudProb = 0.05 + r.Float64()*0.08
			fraudLevel = "LOW"
			returnProb = 0.02 + r.Float64()*0.04
			returnLevel = "LOW"
			overallRisk = "LOW"
			expectedLoss = 0.0
			riskAction = "ALLOW"
			selectedAction = "PAYMENT_LINK"
			actionType = "PAYMENT_LINK"
			delayHours = 0
			decisionReason = "Daily banking velocity exceeded. Dispatched multi-rail payment link with UPI and Netbanking fallbacks."
			aiDiagnosis = "Single-day velocity cap reached on card rail."
			isRecovered = r.Float64() < 0.70

		case CaseSuspectedFraud:
			failureCode = []string{"SUSPECTED_FRAUD", "STOLEN_CARD", "RISK_ANOMALY_BLOCKED"}[r.Intn(3)]
			failureCategory = "FRAUD_SUSPECTED"
			recoveryProb = 0.08 + r.Float64()*0.07 // 0.08 - 0.15
			fraudProb = 0.75 + r.Float64()*0.23    // 0.75 - 0.98 (HIGH)
			fraudLevel = "HIGH"
			returnProb = 0.60 + r.Float64()*0.32   // 0.60 - 0.92 (HIGH)
			returnLevel = "HIGH"
			overallRisk = "HIGH"
			expectedLoss = amount * fraudProb
			riskAction = "BLOCK"
			selectedAction = "HALT"
			actionType = "HALT"
			delayHours = 0
			decisionReason = "HIGH FRAUD RISK: Random forest anomaly detector flagged high risk score. Retries permanently halted to prevent merchant chargeback."
			aiDiagnosis = "Stolen card / proxy mismatch / card testing velocity anomaly."
			isRecovered = false
			fraudBlockedCount++
		}

		// Check active in-flight cases for recent timestamps (last 2 hours)
		var pmtStatus string
		if cType == CaseSuspectedFraud {
			wfStatus = "HALTED"
			pmtStatus = "FAILED"
		} else if hoursAgo < 2.0 && r.Float64() < 0.4 {
			// In-flight active recovery workflow
			wfStatus = []string{"ANALYZING", "SCHEDULED", "EXECUTING", "VERIFYING"}[r.Intn(4)]
			pmtStatus = "FAILED"
			isRecovered = false
		} else if isRecovered {
			wfStatus = "RECOVERED"
			pmtStatus = "CAPTURED"
			recoveredCount++
			totalRecoveredRev += amount
		} else {
			wfStatus = "FAILED"
			pmtStatus = "FAILED"
		}

		// 1. Insert Payment
		paymentID := uuid.New().String()
		razorpayPmtID := fmt.Sprintf("pay_synth_%06d%s", i+100000, method[:2])
		_, err := pool.Exec(ctx, `
			INSERT INTO payments (id, merchant_id, customer_id, subscription_id, amount, currency, status, method, failure_code, razorpay_payment_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'INR', $6, $7, $8, $9, $10, $10 + INTERVAL '5 minutes')
		`, paymentID, merchantID, custID, subID, amount, pmtStatus, method, failureCode, razorpayPmtID, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert payment %d: %v", i, err)
		}

		// 2. Insert Failure Event
		rawErrJSON, _ := json.Marshal(map[string]interface{}{
			"error": map[string]interface{}{
				"code":        failureCode,
				"description": decisionReason,
				"source":      "gateway",
				"step":        "payment_authentication",
				"reason":      failureCode,
			},
		})
		_, err = pool.Exec(ctx, `
			INSERT INTO failure_events (payment_id, failure_code, raw_response, created_at)
			VALUES ($1, $2, $3, $4)
		`, paymentID, failureCode, rawErrJSON, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert failure_event %d: %v", i, err)
		}

		// 3. Insert Webhook Payment Event
		rawEvtJSON, _ := json.Marshal(map[string]interface{}{
			"event": "payment.failed",
			"payload": map[string]interface{}{
				"payment": map[string]interface{}{
					"entity": map[string]interface{}{
						"id":           razorpayPmtID,
						"amount":       amount * 100,
						"currency":     "INR",
						"status":       "failed",
						"method":       method,
						"error_code":   failureCode,
						"error_reason": failureCode,
					},
				},
			},
		})
		_, _ = pool.Exec(ctx, `
			INSERT INTO payment_events (razorpay_event_id, event_type, razorpay_payment_id, raw_payload, processed, created_at)
			VALUES ($1, 'payment.failed', $2, $3, true, $4)
		`, fmt.Sprintf("evt_synth_%06d", i+100000), razorpayPmtID, rawEvtJSON, createdAt)

		// 4. Insert Recovery Workflow
		workflowID := uuid.New().String()
		var scheduledAt *time.Time
		if selectedAction == "DELAYED_RETRY" {
			t := createdAt.Add(time.Duration(delayHours) * time.Hour)
			scheduledAt = &t
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO recovery_workflows (id, payment_id, merchant_id, status, recovery_probability, selected_action, scheduled_at, fraud_probability, return_probability, overall_risk, expected_loss, risk_action, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13 + INTERVAL '10 minutes')
		`, workflowID, paymentID, merchantID, wfStatus, recoveryProb, selectedAction, scheduledAt, fraudProb, returnProb, overallRisk, expectedLoss, riskAction, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert workflow %d: %v", i, err)
		}

		// 5. Insert Risk Assessment
		_, err = pool.Exec(ctx, `
			INSERT INTO risk_assessments (payment_id, workflow_id, merchant_id, event_type, fraud_probability, fraud_risk_level, return_probability, return_risk_level, overall_risk_level, expected_loss, recommended_action, reason, model_version, created_at)
			VALUES ($1, $2, $3, 'payment.failed', $4, $5, $6, $7, $8, $9, $10, $11, 'fraud-rf-v1.0', $12)
		`, paymentID, workflowID, merchantID, fraudProb, fraudLevel, returnProb, returnLevel, overallRisk, expectedLoss, riskAction, decisionReason, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert risk_assessment %d: %v", i, err)
		}

		// 6. Insert AI Decision
		conf := 0.94
		infMs := 120 + r.Intn(80)
		_, err = pool.Exec(ctx, `
			INSERT INTO ai_decisions (workflow_id, provider, model, diagnosis, recommended_action, recommended_delay_hours, confidence, recoverability, reasoning, inference_duration_ms, created_at)
			VALUES ($1, 'ollama', 'deepseek-r1:1.5b', $2, $3, $4, $5, $6, $7, $8, $9)
		`, workflowID, aiDiagnosis, selectedAction, delayHours, conf, recoveryProb, decisionReason, infMs, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert ai_decision %d: %v", i, err)
		}

		// 7. Insert Model Prediction
		featJSON, _ := json.Marshal(map[string]interface{}{
			"failure_code":       failureCode,
			"method":             method,
			"amount":             amount,
			"history_failures":   r.Intn(4),
			"history_successes":  r.Intn(8) + 2,
			"fraud_score":        fraudProb,
		})
		_, err = pool.Exec(ctx, `
			INSERT INTO model_predictions (workflow_id, payment_id, model_version, probability, failure_category, features_used, created_at)
			VALUES ($1, $2, 'logistic-v1', $3, $4, $5, $6)
		`, workflowID, paymentID, recoveryProb, failureCategory, featJSON, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert model_prediction %d: %v", i, err)
		}

		// 8. Insert Recovery Actions & Outcomes
		actionID := uuid.New().String()
		actionStatus := "EXECUTED"
		if wfStatus == "ANALYZING" || wfStatus == "SCHEDULED" {
			actionStatus = "PENDING"
		} else if wfStatus == "HALTED" {
			actionStatus = "FAILED"
		}

		executedAt := createdAt.Add(time.Duration(delayHours)*time.Hour + time.Minute*2)
		actionResult := "Payment recovered successfully"
		if !isRecovered {
			actionResult = "Retry failed after max attempts"
		}
		if cType == CaseSuspectedFraud {
			actionResult = "Action blocked by ML Fraud Guard"
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO recovery_actions (id, workflow_id, action_type, status, attempt, executed_at, result, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 1, $5, $6, $7, $7)
		`, actionID, workflowID, actionType, actionStatus, executedAt, actionResult, createdAt)
		if err != nil {
			log.Fatalf("Failed to insert recovery_action %d: %v", i, err)
		}

		// Recovery Outcome
		var recoveredAmt *float64
		if isRecovered {
			recoveredAmt = &amount
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO recovery_outcomes (action_id, payment_id, recovered, recovered_amount, time_to_recovery, created_at)
			VALUES ($1, $2, $3, $4, '15 minutes'::interval, $5)
		`, actionID, paymentID, isRecovered, recoveredAmt, executedAt)
		if err != nil {
			log.Fatalf("Failed to insert recovery_outcome %d: %v", i, err)
		}

		// 9. Insert Cryptographic Audit Trail
		prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
		payloadHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%f", paymentID, failureCode, amount))))
		eventHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", prevHash, "WORKFLOW_INITIATED", payloadHash))))

		_, _ = pool.Exec(ctx, `
			INSERT INTO audit_events (workflow_id, timestamp, actor, action, payload_hash, previous_event_hash, event_hash)
			VALUES ($1, $2, 'SYSTEM_RECOVERY_ENGINE', 'PAYMENT_FAILURE_INGESTED', $3, $4, $5)
		`, workflowID, createdAt, payloadHash, prevHash, eventHash)

		if (i+1)%500 == 0 {
			log.Printf("Progress: %d / %d records seeded...", i+1, totalCount)
		}
	}

	log.Println("==========================================================")
	log.Println("SYNTHETIC DATASET GENERATION COMPLETE")
	log.Printf("Total Payments Created:    %d\n", totalCount)
	log.Printf("Total at Risk Volume:      ₹%.2f\n", totalAtRiskRev)
	log.Printf("Successfully Recovered:    %d (%.1f%% Recovery Rate)\n", recoveredCount, float64(recoveredCount)/float64(totalCount)*100)
	log.Printf("Total Recovered Revenue:   ₹%.2f\n", totalRecoveredRev)
	log.Printf("Fraud Anomalies Blocked:   %d (₹%.2f fraud loss prevented)\n", fraudBlockedCount, totalAtRiskRev-totalRecoveredRev)
	log.Println("==========================================================")
}
