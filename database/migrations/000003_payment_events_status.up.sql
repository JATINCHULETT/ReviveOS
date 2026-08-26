-- Migration 000003: Extend payment_events with processing lifecycle columns
-- Adds received_at, processing_status, and processed_at to match
-- the required webhook event storage contract.
-- Does NOT remove existing 'processed' boolean to avoid breaking running code.

-- 1. Add processing_status column (richer than the boolean 'processed')
--    Values: PENDING, PROCESSING, PROCESSED, FAILED, SKIPPED
ALTER TABLE payment_events
    ADD COLUMN IF NOT EXISTS processing_status VARCHAR(50) NOT NULL DEFAULT 'PENDING';

-- 2. Add received_at timestamp (when the webhook was physically received)
ALTER TABLE payment_events
    ADD COLUMN IF NOT EXISTS received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- 3. Add processed_at timestamp (when processing completed)
ALTER TABLE payment_events
    ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE;

-- 4. Backfill processing_status from existing 'processed' boolean
UPDATE payment_events
SET processing_status = CASE WHEN processed = true THEN 'PROCESSED' ELSE 'PENDING' END
WHERE processing_status = 'PENDING' AND processed = true;

-- 5. Backfill processed_at from created_at for already-processed rows
UPDATE payment_events
SET processed_at = created_at
WHERE processed = true AND processed_at IS NULL;

-- 6. Backfill received_at from created_at for existing rows
UPDATE payment_events
SET received_at = created_at
WHERE received_at IS NULL;

-- 7. Index on processing_status for efficient queue polling
CREATE INDEX IF NOT EXISTS idx_payment_events_processing_status
    ON payment_events(processing_status)
    WHERE processing_status IN ('PENDING', 'PROCESSING');

-- 8. Index on received_at for chronological queries
CREATE INDEX IF NOT EXISTS idx_payment_events_received_at
    ON payment_events(received_at);
