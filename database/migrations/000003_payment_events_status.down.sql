-- Rollback migration 000003: Remove processing lifecycle columns from payment_events

DROP INDEX IF EXISTS idx_payment_events_received_at;
DROP INDEX IF EXISTS idx_payment_events_processing_status;

ALTER TABLE payment_events DROP COLUMN IF EXISTS processed_at;
ALTER TABLE payment_events DROP COLUMN IF EXISTS received_at;
ALTER TABLE payment_events DROP COLUMN IF EXISTS processing_status;
