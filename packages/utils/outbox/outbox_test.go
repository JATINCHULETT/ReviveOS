package outbox

import (
	"context"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
	"github.com/reviveos/utils/queue"
)

func TestOutboxWriter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Skipf("Skipping: DB not available (%v)", err)
	}
	defer pool.Close()

	payload := map[string]string{
		"action": "test_insert",
	}

	var dummyID string
	_ = pool.QueryRow(ctx, "SELECT uuid_generate_v4()::text").Scan(&dummyID)

	eventID, err := InsertOutboxEventPool(ctx, pool, "test:event", "dummy", dummyID, payload)
	if err != nil {
		t.Fatalf("InsertOutboxEventPool failed: %v", err)
	}

	if eventID == "" {
		t.Fatalf("expected non-empty eventID")
	}

	// Verify row in DB
	var eventType, aggType string
	err = pool.QueryRow(ctx, "SELECT event_type, aggregate_type FROM outbox_events WHERE id::text = $1", eventID).Scan(&eventType, &aggType)
	if err != nil {
		t.Fatalf("failed to query inserted outbox event: %v", err)
	}

	if eventType != "test:event" || aggType != "dummy" {
		t.Errorf("unexpected event_type=%s aggType=%s", eventType, aggType)
	}
}

func TestRelayProcessBatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Skipf("Skipping: DB not available (%v)", err)
	}
	defer pool.Close()

	if err := queue.PingRedis(ctx); err != nil {
		t.Skipf("Skipping: Redis not available (%v)", err)
	}

	client := queue.NewClient()
	defer client.Close()

	relay := NewRelay(pool, client, 10)
	_, err = relay.ProcessBatch(ctx)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
}
