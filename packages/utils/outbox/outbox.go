package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID            string          `json:"id"`
	EventType     string          `json:"event_type"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	Attempts      int             `json:"attempts"`
	LastError     *string         `json:"last_error,omitempty"`
}

// InsertOutboxEvent records an event in the outbox within an existing transaction.
func InsertOutboxEvent(ctx context.Context, tx pgx.Tx, eventType, aggregateType, aggregateID string, payload interface{}) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	var eventID string
	query := `
		INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`
	err = tx.QueryRow(ctx, query, eventType, aggregateType, aggregateID, payloadBytes).Scan(&eventID)
	if err != nil {
		return "", fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return eventID, nil
}

// InsertOutboxEventPool records an event in the outbox using a standalone pool query.
func InsertOutboxEventPool(ctx context.Context, pool *pgxpool.Pool, eventType, aggregateType, aggregateID string, payload interface{}) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	var eventID string
	query := `
		INSERT INTO outbox_events (event_type, aggregate_type, aggregate_id, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`
	err = pool.QueryRow(ctx, query, eventType, aggregateType, aggregateID, payloadBytes).Scan(&eventID)
	if err != nil {
		return "", fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return eventID, nil
}

// Relay handles polling outbox_events and enqueuing them to Asynq / Redis.
type Relay struct {
	pool        *pgxpool.Pool
	asynqClient *asynq.Client
	batchSize   int
}

func NewRelay(pool *pgxpool.Pool, client *asynq.Client, batchSize int) *Relay {
	if batchSize <= 0 {
		batchSize = 20
	}
	return &Relay{
		pool:        pool,
		asynqClient: client,
		batchSize:   batchSize,
	}
}

// ProcessBatch locks and publishes a batch of unpublished outbox events using FOR UPDATE SKIP LOCKED.
func (r *Relay) ProcessBatch(ctx context.Context) (int, error) {
	if r.pool == nil || r.asynqClient == nil {
		return 0, errors.New("relay pool or asynq client is nil")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin relay tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// FOR UPDATE SKIP LOCKED ensures safe multi-instance concurrency
	query := `
		SELECT id::text, event_type, aggregate_type, aggregate_id::text, payload, attempts
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := tx.Query(ctx, query, r.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to select pending outbox events: %w", err)
	}

	type pendingEvent struct {
		id            string
		eventType     string
		aggregateType string
		aggregateID   string
		payload       []byte
		attempts      int
	}

	var events []pendingEvent
	for rows.Next() {
		var ev pendingEvent
		if err := rows.Scan(&ev.id, &ev.eventType, &ev.aggregateType, &ev.aggregateID, &ev.payload, &ev.attempts); err == nil {
			events = append(events, ev)
		}
	}
	rows.Close()

	if len(events) == 0 {
		return 0, nil
	}

	publishedCount := 0
	for _, ev := range events {
		// Idempotent Task ID derived directly from the outbox event ID
		taskID := fmt.Sprintf("outbox:%s", ev.id)

		task := asynq.NewTask(ev.eventType, ev.payload)
		opts := []asynq.Option{
			asynq.TaskID(taskID),
			asynq.MaxRetry(5),
			asynq.Retention(24 * time.Hour),
		}

		_, err := r.asynqClient.EnqueueContext(ctx, task, opts...)
		if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
			// Failed to enqueue: increment attempts and record error
			errMsg := err.Error()
			_, _ = tx.Exec(ctx, `
				UPDATE outbox_events
				SET attempts = attempts + 1, last_error = $1
				WHERE id::text = $2
			`, errMsg, ev.id)
			log.Printf("Relay: failed to enqueue outbox event %s (%s): %v", ev.id, ev.eventType, err)
			continue
		}

		// Marked as published on success or if task ID already exists (idempotent)
		_, err = tx.Exec(ctx, `
			UPDATE outbox_events
			SET published_at = CURRENT_TIMESTAMP, last_error = NULL
			WHERE id::text = $1
		`, ev.id)
		if err != nil {
			log.Printf("Relay: failed to mark outbox event %s as published: %v", ev.id, err)
			continue
		}

		publishedCount++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit relay tx: %w", err)
	}

	return publishedCount, nil
}

// Start runs the relay polling loop until ctx is cancelled
func (r *Relay) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := r.ProcessBatch(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("Relay loop error: %v", err)
			} else if count > 0 {
				log.Printf("Relay published %d outbox events to Asynq", count)
			}
		}
	}
}
