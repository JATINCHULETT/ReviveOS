package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	paymentprovider "github.com/reviveos/services/payment-provider"
	"github.com/reviveos/utils/db"
	"github.com/reviveos/worker/internal/executor"
	"github.com/reviveos/worker/internal/pipeline"
)

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

func main() {
	loadEnvFile(".env")
	loadEnvFile("../../.env")

	fmt.Println("====================================================================")
	fmt.Println("   ReviveOS — Recurring Subscription & Autopay Test Runner")
	fmt.Println("====================================================================")

	dbCtx, cancelDB := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelDB()

	pool, err := db.Connect(dbCtx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	localProvider := paymentprovider.NewLocalPaymentProvider(pool)
	pipe := pipeline.NewPipeline(pool)
	exec := executor.NewRecoveryExecutor(pool, localProvider)

	// 1. Setup Merchant and Subscriber
	stepCtx, cancelStep := context.WithTimeout(context.Background(), 15*time.Second)
	var merchantID string
	err = pool.QueryRow(stepCtx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text",
		fmt.Sprintf("Subscription Store %d", time.Now().Unix())).Scan(&merchantID)
	cancelStep()
	if err != nil {
		log.Fatalf("Failed to create merchant: %v", err)
	}

	stepCtx, cancelStep = context.WithTimeout(context.Background(), 15*time.Second)
	var customerID string
	custEmail := fmt.Sprintf("subscriber_%d@test.com", time.Now().Unix())
	err = pool.QueryRow(stepCtx, `
		INSERT INTO customers (merchant_id, email, phone, communication_opt_out)
		VALUES ($1, $2, '+919876543210', false)
		RETURNING id::text
	`, merchantID, custEmail).Scan(&customerID)
	cancelStep()
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}

	// 2. Setup Recurring Subscription
	stepCtx, cancelStep = context.WithTimeout(context.Background(), 15*time.Second)
	var subID string
	rzpSubID := fmt.Sprintf("sub_rzp_%d", time.Now().UnixNano())
	err = pool.QueryRow(stepCtx, `
		INSERT INTO subscriptions (
			merchant_id, customer_id, amount, currency, status, billing_interval, plan_id, razorpay_subscription_id, next_billing_at
		)
		VALUES ($1, $2, 1999.00, 'INR', 'ACTIVE', 'monthly', 'plan_pro_annual', $3, CURRENT_TIMESTAMP + INTERVAL '1 month')
		RETURNING id::text
	`, merchantID, customerID, rzpSubID).Scan(&subID)
	cancelStep()
	if err != nil {
		log.Fatalf("Failed to create subscription: %v", err)
	}

	fmt.Printf("[1/3] Created Subscription %s for %s (Status: ACTIVE, Plan: Pro Monthly @ ₹1,999)\n", subID, custEmail)

	// 3. CYCLE 1: Simulate Recurring Charge Failure (Transient Bank / Insufficient Funds)
	fmt.Println("\n[2/3] Simulating Billing Cycle 1: Autopay Charge Failed (INSUFFICIENT_FUNDS)...")
	stepCtx, cancelStep = context.WithTimeout(context.Background(), 15*time.Second)
	var pmt1ID string
	err = pool.QueryRow(stepCtx, `
		INSERT INTO payments (
			merchant_id, customer_id, subscription_id, amount, currency, status, method, failure_code
		)
		VALUES ($1, $2, $3, 1999.00, 'INR', 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerID, subID).Scan(&pmt1ID)
	_, _ = pool.Exec(stepCtx, "UPDATE subscriptions SET status = 'PAST_DUE' WHERE id::text = $1", subID)
	cancelStep()
	if err != nil {
		log.Fatalf("Failed to record payment failure: %v", err)
	}

	// Run Adaptive Recovery Analysis
	analyzeCtx, cancelAnalyze := context.WithTimeout(context.Background(), 45*time.Second)
	res1, err := pipe.AnalyzePayment(analyzeCtx, pmt1ID)
	cancelAnalyze()
	if err != nil {
		log.Fatalf("Pipeline analysis failed: %v", err)
	}
	fmt.Printf("   -> Diagnosis: %s | Probability: %.2f | Selected Action: %s\n",
		res1.FailureCategory, res1.Probability, res1.AIRecommendation.RecommendedAction)

	// Run Recovery Execution
	execCtx, cancelExec := context.WithTimeout(context.Background(), 20*time.Second)
	execRes1, err := exec.ExecuteWorkflow(execCtx, res1.WorkflowID)
	cancelExec()
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
	fmt.Printf("   -> Execution Result: Reconciliation=%s | ActionTaken=%s | Recovered=%v\n",
		execRes1.Reconciliation, execRes1.ActionTaken, execRes1.Recovered)

	// 4. CYCLE 2: Simulate Card Expiration on Recurring Autopay (EXPIRED_CARD)
	fmt.Println("\n[3/3] Simulating Billing Cycle 2: Autopay Mandate Failed (EXPIRED_CARD)...")
	stepCtx, cancelStep = context.WithTimeout(context.Background(), 15*time.Second)
	var pmt2ID string
	err = pool.QueryRow(stepCtx, `
		INSERT INTO payments (
			merchant_id, customer_id, subscription_id, amount, currency, status, method, failure_code
		)
		VALUES ($1, $2, $3, 1999.00, 'INR', 'FAILED', 'card', 'EXPIRED_CARD')
		RETURNING id::text
	`, merchantID, customerID, subID).Scan(&pmt2ID)
	cancelStep()
	if err != nil {
		log.Fatalf("Failed to record payment failure: %v", err)
	}

	analyzeCtx2, cancelAnalyze2 := context.WithTimeout(context.Background(), 45*time.Second)
	res2, err := pipe.AnalyzePayment(analyzeCtx2, pmt2ID)
	cancelAnalyze2()
	if err != nil {
		log.Fatalf("Pipeline analysis 2 failed: %v", err)
	}
	fmt.Printf("   -> Diagnosis: %s | Selected Action: %s\n", res2.FailureCategory, res2.AIRecommendation.RecommendedAction)

	execCtx2, cancelExec2 := context.WithTimeout(context.Background(), 20*time.Second)
	execRes2, err := exec.ExecuteWorkflow(execCtx2, res2.WorkflowID)
	cancelExec2()
	if err != nil {
		log.Fatalf("Execution 2 failed: %v", err)
	}
	fmt.Printf("   -> Execution Result: Reconciliation=%s | ActionTaken=%s | NotificationSent=%v\n",
		execRes2.Reconciliation, execRes2.ActionTaken, execRes2.NotificationSent)

	// Verify Final Subscription Status in DB
	stepCtx, cancelStep = context.WithTimeout(context.Background(), 10*time.Second)
	var finalSubStatus string
	_ = pool.QueryRow(stepCtx, "SELECT status FROM subscriptions WHERE id::text = $1", subID).Scan(&finalSubStatus)
	cancelStep()

	fmt.Println("\n====================================================================")
	fmt.Printf("   FINAL VERIFICATION: Subscription Status = %s\n", finalSubStatus)
	fmt.Printf("   AI Recovery Strategy: Zero-touch smart retry for transient errors,\n")
	fmt.Printf("   and automatic payment link dispatch for expired mandates.\n")
	fmt.Println("====================================================================")
}
