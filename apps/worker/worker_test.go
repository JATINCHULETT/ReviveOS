package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/reviveos/schemas"
	"github.com/reviveos/utils/db"
	"github.com/reviveos/utils/outbox"
	"github.com/reviveos/utils/queue"
	"github.com/reviveos/worker/internal/tasks"
)

func TestOutboxToAsynqWorkerPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 1. Connect to Database
	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 2. Connect to Redis
	if err := queue.PingRedis(ctx); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	asynqClient := queue.NewClient()
	defer asynqClient.Close()

	// Unique test identifier and unique task type so background workers don't intercept
	uniqueType := fmt.Sprintf("test:pipeline:%d", time.Now().UnixNano())
	uniqueTestID := fmt.Sprintf("test_msg_%d", time.Now().UnixNano())
	payload := map[string]interface{}{
		"test_id":   uniqueTestID,
		"message":   "Database -> Outbox -> Relay -> Redis -> Asynq -> Worker verification",
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	}

	// 3. Insert Outbox Event in Database
	var aggregateID string
	err = pool.QueryRow(ctx, "SELECT id::text FROM merchants LIMIT 1").Scan(&aggregateID)
	if err != nil {
		err = pool.QueryRow(ctx, "SELECT uuid_generate_v4()::text").Scan(&aggregateID)
		if err != nil {
			t.Fatalf("Failed to generate aggregate ID: %v", err)
		}
	}

	outboxEventID, err := outbox.InsertOutboxEventPool(ctx, pool, uniqueType, "merchant", aggregateID, payload)
	if err != nil {
		t.Fatalf("Failed to insert outbox event: %v", err)
	}
	t.Logf("1. Inserted outbox event ID=%s (type=%s, aggregate=%s)", outboxEventID, uniqueType, aggregateID)

	// Verify it's unpublished
	var publishedAt *time.Time
	err = pool.QueryRow(ctx, "SELECT published_at FROM outbox_events WHERE id::text = $1", outboxEventID).Scan(&publishedAt)
	if err != nil || publishedAt != nil {
		t.Fatalf("Expected published_at to be NULL, got %v (err: %v)", publishedAt, err)
	}

	// 4. Run Relay ProcessBatch
	relay := outbox.NewRelay(pool, asynqClient, 10)
	publishedCount, err := relay.ProcessBatch(ctx)
	if err != nil {
		t.Fatalf("Relay ProcessBatch failed: %v", err)
	}
	if publishedCount < 1 {
		t.Fatalf("Expected at least 1 published event, got %d", publishedCount)
	}
	t.Logf("2. Relay processed and published %d event(s) to Redis/Asynq", publishedCount)

	// Verify database record now has published_at timestamp
	err = pool.QueryRow(ctx, "SELECT published_at FROM outbox_events WHERE id::text = $1", outboxEventID).Scan(&publishedAt)
	if err != nil || publishedAt == nil {
		t.Fatalf("Expected published_at to be populated, got NULL (err: %v)", err)
	}
	t.Logf("3. Database outbox record updated: published_at=%s", publishedAt.Format(time.RFC3339))

	// 5. Start Worker to handle task from Redis/Asynq
	var (
		mu          sync.Mutex
		taskHandled = false
		receivedMsg string
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(uniqueType, func(ctx context.Context, task *asynq.Task) error {
		mu.Lock()
		defer mu.Unlock()
		taskHandled = true
		receivedMsg = string(task.Payload())
		t.Logf("4. Asynq worker received task ID=%s, Type=%s, Payload=%s", task.ResultWriter().TaskID(), task.Type(), receivedMsg)
		return nil
	})

	srv := asynq.NewServer(
		queue.GetRedisOpt(),
		asynq.Config{
			Concurrency: 2,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	// Start worker in background
	go func() {
		_ = srv.Run(mux)
	}()
	defer srv.Shutdown()

	// 6. Wait for task to be handled
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		done := taskHandled
		mu.Unlock()
		if done {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	mu.Lock()
	done := taskHandled
	mu.Unlock()

	if !done {
		t.Fatalf("FAIL: Asynq worker did not process task within deadline")
	}

	t.Logf("5. SUCCESS: Pipeline verified (DB Outbox -> Relay -> Redis -> Asynq -> Worker)")
}

func TestWorker_AnalyzeTaskExecutionEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := queue.PingRedis(ctx); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}

	asynqClient := queue.NewClient()
	defer asynqClient.Close()

	// 1. Create merchant, customer, and failed payment
	var merchantID string
	_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ($1) RETURNING id::text", fmt.Sprintf("E2E_M_%d", time.Now().UnixNano())).Scan(&merchantID)

	var customerID string
	_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email) VALUES ($1, $2) RETURNING id::text", merchantID, fmt.Sprintf("e2e_%d@test.com", time.Now().UnixNano())).Scan(&customerID)

	// Insert 2 prior successful payments
	for i := 0; i < 2; i++ {
		_, _ = pool.Exec(ctx, "INSERT INTO payments (merchant_id, customer_id, amount, status) VALUES ($1, $2, 1000, 'CAPTURED')", merchantID, customerID)
	}

	// Insert 1 failed payment
	var paymentID string
	err = pool.QueryRow(ctx, `
		INSERT INTO payments (merchant_id, customer_id, amount, status, failure_code)
		VALUES ($1, $2, 2500, 'FAILED', 'BANK_UNAVAILABLE')
		RETURNING id::text
	`, merchantID, customerID).Scan(&paymentID)
	if err != nil {
		t.Fatalf("Failed to insert payment: %v", err)
	}

	// 2. Queue recovery:analyze task via Outbox
	analyzePayload := schemas.RecoveryAnalyzePayload{PaymentID: paymentID}
	outboxID, err := outbox.InsertOutboxEventPool(ctx, pool, schemas.TaskRecoveryAnalyze, "payment", paymentID, analyzePayload)
	if err != nil {
		t.Fatalf("Failed to insert outbox event: %v", err)
	}

	// 3. Relay publishes to Asynq
	relay := outbox.NewRelay(pool, asynqClient, 10)
	count, err := relay.ProcessBatch(ctx)
	if err != nil || count < 1 {
		t.Fatalf("Relay failed to publish task: count=%d, err=%v", count, err)
	}

	// 4. Start Worker with real TaskHandler
	mux := asynq.NewServeMux()
	handler := tasks.NewTaskHandler(pool)
	handler.RegisterRoutes(mux)

	srv := asynq.NewServer(
		queue.GetRedisOpt(),
		asynq.Config{
			Concurrency: 2,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	go func() {
		_ = srv.Run(mux)
	}()
	defer srv.Shutdown()

	// 5. Poll database for both model_predictions and ai_decisions records created by worker
	var predID string
	var predProb float64
	var predCat string
	var aiDecisionID string
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		errPred := pool.QueryRow(ctx, `
			SELECT id::text, probability::float8, failure_category
			FROM model_predictions
			WHERE payment_id::text = $1
			ORDER BY created_at DESC LIMIT 1
		`, paymentID).Scan(&predID, &predProb, &predCat)

		errAI := pool.QueryRow(ctx, `
			SELECT ad.id::text
			FROM ai_decisions ad
			JOIN recovery_workflows rw ON ad.workflow_id = rw.id
			WHERE rw.payment_id::text = $1
			ORDER BY ad.created_at DESC LIMIT 1
		`, paymentID).Scan(&aiDecisionID)

		if errPred == nil && predID != "" && errAI == nil && aiDecisionID != "" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if predID == "" {
		t.Fatalf("FAIL: model_predictions was not populated by worker for payment %s", paymentID)
	}
	if aiDecisionID == "" {
		t.Fatalf("FAIL: ai_decisions was not populated by worker for payment %s", paymentID)
	}

	t.Logf("SUCCESS: Worker processed recovery:analyze -> PredID=%s | AIDecisionID=%s | Category=%s | Prob=%.4f (OutboxID=%s)",
		predID, aiDecisionID, predCat, predProb, outboxID)
}
