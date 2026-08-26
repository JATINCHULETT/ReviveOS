package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/reviveos/packages/types"
	aiprovider "github.com/reviveos/services/ai-provider"
	"github.com/reviveos/utils/db"
)

func TestPipeline_RealCustomerHistoryAndPredictions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 1. Create a Merchant
	var merchantID string
	err = pool.QueryRow(ctx, `
		INSERT INTO merchants (name)
		VALUES ($1)
		RETURNING id::text
	`, fmt.Sprintf("Merchant_%d", time.Now().UnixNano())).Scan(&merchantID)
	if err != nil {
		t.Fatalf("Failed to create test merchant: %v", err)
	}

	// 2. Create Customer A (Strong Success History: 4 CAPTURED, 0 FAILED)
	var customerAID string
	err = pool.QueryRow(ctx, `
		INSERT INTO customers (merchant_id, email, phone)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, merchantID, fmt.Sprintf("custA_%d@test.com", time.Now().UnixNano()), "+1111111111").Scan(&customerAID)
	if err != nil {
		t.Fatalf("Failed to create customer A: %v", err)
	}

	for i := 0; i < 4; i++ {
		_, err := pool.Exec(ctx, `
			INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method)
			VALUES ($1, $2, $3, 'INR', 'CAPTURED', 'card')
		`, merchantID, customerAID, 1500.00)
		if err != nil {
			t.Fatalf("Failed to insert historical payment for customer A: %v", err)
		}
	}

	// 3. Create Customer B (Poor History: 0 CAPTURED, 4 FAILED)
	var customerBID string
	err = pool.QueryRow(ctx, `
		INSERT INTO customers (merchant_id, email, phone)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, merchantID, fmt.Sprintf("custB_%d@test.com", time.Now().UnixNano()), "+2222222222").Scan(&customerBID)
	if err != nil {
		t.Fatalf("Failed to create customer B: %v", err)
	}

	for i := 0; i < 4; i++ {
		_, err := pool.Exec(ctx, `
			INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method, failure_code)
			VALUES ($1, $2, $3, 'INR', 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		`, merchantID, customerBID, 1500.00)
		if err != nil {
			t.Fatalf("Failed to insert historical payment for customer B: %v", err)
		}
	}

	// 4. Create Failed Payment for Customer A
	var paymentAID string
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method, failure_code)
		VALUES ($1, $2, 2000.00, 'INR', 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerAID).Scan(&paymentAID)
	if err != nil {
		t.Fatalf("Failed to create failed payment for customer A: %v", err)
	}

	// 5. Create Failed Payment for Customer B (Same failure code and amount)
	var paymentBID string
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method, failure_code)
		VALUES ($1, $2, 2000.00, 'INR', 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerBID).Scan(&paymentBID)
	if err != nil {
		t.Fatalf("Failed to create failed payment for customer B: %v", err)
	}

	// 6. Run Pipeline Analysis with deterministic fast AI fallback for unit verification
	p := NewPipelineWithAI(pool, nil)

	resA, err := p.AnalyzePayment(ctx, paymentAID)
	if err != nil {
		t.Fatalf("AnalyzePayment for Customer A failed: %v", err)
	}

	resB, err := p.AnalyzePayment(ctx, paymentBID)
	if err != nil {
		t.Fatalf("AnalyzePayment for Customer B failed: %v", err)
	}

	t.Logf("Customer A (Good History) -> Successes: %d, Failures: %d | Prob: %.4f | PredID: %s",
		resA.CustomerHistory.SuccessfulPayments, resA.CustomerHistory.FailedPayments, resA.Probability, resA.PredictionID)
	t.Logf("Customer B (Poor History) -> Successes: %d, Failures: %d | Prob: %.4f | PredID: %s",
		resB.CustomerHistory.SuccessfulPayments, resB.CustomerHistory.FailedPayments, resB.Probability, resB.PredictionID)

	// 7. Verify Real History was Queried
	if resA.CustomerHistory.SuccessfulPayments != 4 || resA.CustomerHistory.FailedPayments != 0 {
		t.Errorf("Customer A history mismatch: expected (4 success, 0 fail), got (%d, %d)",
			resA.CustomerHistory.SuccessfulPayments, resA.CustomerHistory.FailedPayments)
	}

	if resB.CustomerHistory.SuccessfulPayments != 0 || resB.CustomerHistory.FailedPayments != 4 {
		t.Errorf("Customer B history mismatch: expected (0 success, 4 fail), got (%d, %d)",
			resB.CustomerHistory.SuccessfulPayments, resB.CustomerHistory.FailedPayments)
	}

	// 8. Verify Classifier
	if resA.FailureCategory != types.InsufficientFunds {
		t.Errorf("Expected failure category INSUFFICIENT_FUNDS, got %s", resA.FailureCategory)
	}

	// 9. Verify Probability difference based on real features
	if resA.Probability <= resB.Probability {
		t.Errorf("Customer A with 4 previous successes should have higher probability than Customer B with 4 failures (A=%.4f, B=%.4f)",
			resA.Probability, resB.Probability)
	}

	// 10. Verify model_predictions record in PostgreSQL
	var (
		predWorkflowID string
		predPaymentID  string
		predModel      string
		predProb       float64
		predCat        string
		predFeatures   []byte
	)
	err = pool.QueryRow(ctx, `
		SELECT workflow_id::text, payment_id::text, model_version, probability::float8, failure_category, features_used::text
		FROM model_predictions
		WHERE id::text = $1
	`, resA.PredictionID).Scan(
		&predWorkflowID,
		&predPaymentID,
		&predModel,
		&predProb,
		&predCat,
		&predFeatures,
	)
	if err != nil {
		t.Fatalf("Failed to query persisted model_predictions row: %v", err)
	}

	if predPaymentID != paymentAID {
		t.Errorf("model_predictions payment_id mismatch: expected %s, got %s", paymentAID, predPaymentID)
	}
	if predModel != "logistic-v1" {
		t.Errorf("model_predictions model_version mismatch: expected logistic-v1, got %s", predModel)
	}
	if predCat != "INSUFFICIENT_FUNDS" {
		t.Errorf("model_predictions failure_category mismatch: expected INSUFFICIENT_FUNDS, got %s", predCat)
	}

	t.Logf("Persisted model_predictions row verified: ID=%s, Features=%s", resA.PredictionID, string(predFeatures))
}

func TestPipeline_RealOllamaAIDecisionPersistence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Check if local Ollama is reachable
	ollama := aiprovider.NewOllamaProvider("http://localhost:11434", "deepseek-r1:1.5b")
	available, err := ollama.CheckModelAvailable(ctx)
	if err != nil || !available {
		t.Skipf("Skipping Ollama test: deepseek-r1:1.5b not available (%v)", err)
	}

	// 1. Create merchant, customer, payment
	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ('AI_Test_Merchant') RETURNING id::text").Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, 'ai_test@example.com') RETURNING id::text", merchantID).Scan(&customerID)

	var paymentID string
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, 4999.00, 'FAILED', 'card', 'INSUFFICIENT_FUNDS')
		RETURNING id::text
	`, merchantID, customerID).Scan(&paymentID)
	if err != nil {
		t.Fatalf("Failed to insert payment: %v", err)
	}

	// 2. Run Pipeline with real Ollama provider
	p := NewPipelineWithAI(pool, ollama)
	res, err := p.AnalyzePayment(ctx, paymentID)
	if err != nil {
		t.Fatalf("AnalyzePayment with Ollama failed: %v", err)
	}

	if res.AIDecisionID == "" {
		t.Fatalf("Expected non-empty AIDecisionID")
	}

	// 3. Inspect ai_decisions row in PostgreSQL
	var (
		provider    string
		model       string
		promptHash  string
		rawResponse string
		action      string
		confidence  float64
		latencyMs   int
	)
	err = pool.QueryRow(ctx, `
		SELECT provider, model, prompt_hash, raw_response, recommended_action, confidence::float8, inference_duration_ms
		FROM ai_decisions
		WHERE id::text = $1
	`, res.AIDecisionID).Scan(
		&provider,
		&model,
		&promptHash,
		&rawResponse,
		&action,
		&confidence,
		&latencyMs,
	)
	if err != nil {
		t.Fatalf("Failed to query ai_decisions table row: %v", err)
	}

	if provider != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%s'", provider)
	}
	if model != "deepseek-r1:1.5b" {
		t.Errorf("Expected model 'deepseek-r1:1.5b', got '%s'", model)
	}
	if promptHash == "" {
		t.Errorf("Expected non-empty prompt_hash")
	}
	if rawResponse == "" {
		t.Errorf("Expected non-empty raw_response")
	}
	if action == "" {
		t.Errorf("Expected non-empty recommended_action")
	}
	if confidence <= 0 || confidence > 1.0 {
		t.Errorf("Expected confidence in (0, 1], got %f", confidence)
	}
	if latencyMs <= 0 {
		t.Errorf("Expected positive inference_duration_ms, got %d", latencyMs)
	}

	t.Logf("SUCCESS: Real Ollama decision verified in DB -> ID=%s | Action=%s | Conf=%.2f | Latency=%dms | Model=%s",
		res.AIDecisionID, action, confidence, latencyMs, model)
}

func TestPipeline_FallbackWhenOllamaUnavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Provider with unreachable URL
	unreachableAI := aiprovider.NewOllamaProvider("http://localhost:59998", "deepseek-r1:1.5b")

	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ('Fallback_Merchant') RETURNING id::text").Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, 'fallback@example.com') RETURNING id::text", merchantID).Scan(&customerID)

	var paymentID string
	_ = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, method, failure_code)
		VALUES ($1, $2, 3000.00, 'FAILED', 'card', 'TIMEOUT')
		RETURNING id::text
	`, merchantID, customerID).Scan(&paymentID)

	p := NewPipelineWithAI(pool, unreachableAI)
	res, err := p.AnalyzePayment(ctx, paymentID)
	if err != nil {
		t.Fatalf("AnalyzePayment failed unexpectedly on fallback: %v", err)
	}

	if res.AIDecisionID == "" {
		t.Fatalf("Expected AIDecisionID for fallback record")
	}

	var (
		provider string
		model    string
		action   string
	)
	err = pool.QueryRow(ctx, `
		SELECT provider, model, recommended_action
		FROM ai_decisions
		WHERE id::text = $1
	`, res.AIDecisionID).Scan(&provider, &model, &action)
	if err != nil {
		t.Fatalf("Failed to query fallback ai_decisions row: %v", err)
	}

	if provider != "deterministic-fallback" {
		t.Errorf("Expected provider 'deterministic-fallback', got '%s'", provider)
	}
	if action == "" {
		t.Errorf("Expected valid recommended_action from fallback")
	}

	t.Logf("SUCCESS: Fallback decision recorded in DB -> ID=%s | Provider=%s | Action=%s",
		res.AIDecisionID, provider, action)
}
