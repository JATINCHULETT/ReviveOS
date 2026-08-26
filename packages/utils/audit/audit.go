package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditEvent struct {
	WorkflowID string
	Actor      string
	Action     string
	Metadata   map[string]interface{}
}

// OutcomeTracker provides audit ledger management and cryptographic verification.
type OutcomeTracker struct {
	pool *pgxpool.Pool
}

// NewOutcomeTracker initializes a new OutcomeTracker instance.
func NewOutcomeTracker(pool *pgxpool.Pool) *OutcomeTracker {
	return &OutcomeTracker{pool: pool}
}

// AppendAuditLog adds an entry to the tamper-evident hash-chained ledger.
func (ot *OutcomeTracker) AppendAuditLog(ctx context.Context, event AuditEvent) error {
	return AppendAuditLog(ctx, ot.pool, event)
}

// VerifyAuditChain cryptographically verifies the entire SHA-256 hash chain for a workflow.
func (ot *OutcomeTracker) VerifyAuditChain(ctx context.Context, workflowID string) (bool, error) {
	return VerifyChain(ctx, ot.pool, workflowID)
}

// normalizeMetadata serializes a metadata map to deterministic normalized JSON.
func normalizeMetadata(meta map[string]interface{}) string {
	if meta == nil || len(meta) == 0 {
		return "{}"
	}
	bytes, err := json.Marshal(meta)
	if err != nil {
		return "{}"
	}
	// Re-unmarshal and marshal to ensure canonical key ordering
	var canonical map[string]interface{}
	if err := json.Unmarshal(bytes, &canonical); err == nil {
		if canonBytes, err := json.Marshal(canonical); err == nil {
			return string(canonBytes)
		}
	}
	return string(bytes)
}

// AppendAuditLog adds an entry to the tamper-evident hash-chained ledger.
func AppendAuditLog(ctx context.Context, pool *pgxpool.Pool, event AuditEvent) error {
	if pool == nil || event.WorkflowID == "" {
		return nil
	}

	normalizedMetaStr := normalizeMetadata(event.Metadata)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch previous hash for the workflow
	var previousHash *string
	err = tx.QueryRow(ctx, `
		SELECT event_hash FROM audit_events 
		WHERE workflow_id::text = $1 
		ORDER BY timestamp DESC LIMIT 1
	`, event.WorkflowID).Scan(&previousHash)

	if err != nil && err.Error() != "no rows in result set" {
		// Non-fatal or first record
	}

	// Create payload hash: SHA256(workflow_id | actor | action | metadata)
	payloadStr := fmt.Sprintf("%s|%s|%s|%s", event.WorkflowID, event.Actor, event.Action, normalizedMetaStr)
	payloadHashBytes := sha256.Sum256([]byte(payloadStr))
	payloadHash := hex.EncodeToString(payloadHashBytes[:])

	// Create event hash: SHA256(previous_event_hash | payload_hash)
	var eventHashStr string
	if previousHash != nil {
		eventHashStr = fmt.Sprintf("%s|%s", *previousHash, payloadHash)
	} else {
		eventHashStr = payloadHash
	}
	eventHashBytes := sha256.Sum256([]byte(eventHashStr))
	eventHash := hex.EncodeToString(eventHashBytes[:])

	// Insert audit event
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_events (workflow_id, actor, action, payload_hash, previous_event_hash, event_hash, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
	`, event.WorkflowID, event.Actor, event.Action, payloadHash, previousHash, eventHash, normalizedMetaStr)
	if err != nil {
		return fmt.Errorf("failed to append audit log: %w", err)
	}

	return tx.Commit(ctx)
}

// VerifyChain checks all audit events for a workflow from genesis to latest, re-verifying hashes.
func VerifyChain(ctx context.Context, pool *pgxpool.Pool, workflowID string) (bool, error) {
	if pool == nil || workflowID == "" {
		return false, fmt.Errorf("invalid pool or empty workflowID")
	}

	rows, err := pool.Query(ctx, `
		SELECT actor, action, payload_hash, previous_event_hash, event_hash, metadata
		FROM audit_events
		WHERE workflow_id::text = $1
		ORDER BY timestamp ASC
	`, workflowID)
	if err != nil {
		return false, fmt.Errorf("failed to query audit chain: %w", err)
	}
	defer rows.Close()

	var expectedPrevHash *string
	eventIndex := 0

	for rows.Next() {
		var (
			actor            string
			action           string
			payloadHash      string
			prevHash         *string
			eventHash        string
			metadataBytes    []byte
		)

		if err := rows.Scan(&actor, &action, &payloadHash, &prevHash, &eventHash, &metadataBytes); err != nil {
			return false, fmt.Errorf("failed to scan audit event row: %w", err)
		}

		// 1. Check previous hash continuity
		if eventIndex == 0 {
			if prevHash != nil {
				return false, fmt.Errorf("genesis event should have nil previous_event_hash, got %s", *prevHash)
			}
		} else {
			if prevHash == nil || expectedPrevHash == nil || *prevHash != *expectedPrevHash {
				return false, fmt.Errorf("broken chain link at index %d: expected prev %v, got %v", eventIndex, expectedPrevHash, prevHash)
			}
		}

		// 2. Re-compute payload hash using canonical metadata
		var metaMap map[string]interface{}
		if len(metadataBytes) > 0 {
			_ = json.Unmarshal(metadataBytes, &metaMap)
		}
		normalizedMetaStr := normalizeMetadata(metaMap)

		payloadStr := fmt.Sprintf("%s|%s|%s|%s", workflowID, actor, action, normalizedMetaStr)
		computedPayloadHashBytes := sha256.Sum256([]byte(payloadStr))
		computedPayloadHash := hex.EncodeToString(computedPayloadHashBytes[:])

		if computedPayloadHash != payloadHash {
			return false, fmt.Errorf("payload hash mismatch at index %d: computed %s, stored %s", eventIndex, computedPayloadHash, payloadHash)
		}

		// 3. Re-compute event hash
		var eventHashStr string
		if prevHash != nil {
			eventHashStr = fmt.Sprintf("%s|%s", *prevHash, payloadHash)
		} else {
			eventHashStr = payloadHash
		}
		computedEventHashBytes := sha256.Sum256([]byte(eventHashStr))
		computedEventHash := hex.EncodeToString(computedEventHashBytes[:])

		if computedEventHash != eventHash {
			return false, fmt.Errorf("event hash mismatch at index %d: computed %s, stored %s", eventIndex, computedEventHash, eventHash)
		}

		currentEventHashCopy := eventHash
		expectedPrevHash = &currentEventHashCopy
		eventIndex++
	}

	return eventIndex > 0, nil
}
