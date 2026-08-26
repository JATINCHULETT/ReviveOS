-- ReviveOS Rebuild Migration
-- Adds missing tables: payment_events, ai_decisions, model_predictions
-- Adds razorpay_payment_id to payments for external ID mapping

-- Raw webhook event store with deduplication
CREATE TABLE IF NOT EXISTS payment_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    razorpay_event_id VARCHAR(255) UNIQUE, -- X-Razorpay-Event-Id for dedup
    event_type VARCHAR(100) NOT NULL,       -- payment.failed, payment.captured, etc.
    razorpay_payment_id VARCHAR(255),       -- External Razorpay payment ID (pay_xxx)
    raw_payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_events_razorpay_event_id ON payment_events(razorpay_event_id);
CREATE INDEX IF NOT EXISTS idx_payment_events_processed ON payment_events(processed) WHERE processed = false;

-- AI decision persistence
CREATE TABLE IF NOT EXISTS ai_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES recovery_workflows(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,          -- 'ollama', 'openai', etc.
    model VARCHAR(100) NOT NULL,            -- 'deepseek-r1:1.5b'
    prompt_hash VARCHAR(64),
    raw_response TEXT,
    diagnosis TEXT,
    recommended_action VARCHAR(100),
    recommended_delay_hours INT,
    confidence NUMERIC(4,3),
    recoverability NUMERIC(4,3),
    reasoning TEXT,
    inference_duration_ms INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ai_decisions_workflow_id ON ai_decisions(workflow_id);

-- Model prediction persistence
CREATE TABLE IF NOT EXISTS model_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id UUID NOT NULL REFERENCES recovery_workflows(id) ON DELETE CASCADE,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    model_version VARCHAR(50) NOT NULL,     -- 'logistic-v1'
    probability NUMERIC(5,4) NOT NULL,
    failure_category VARCHAR(100),
    features_used JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_model_predictions_workflow_id ON model_predictions(workflow_id);

-- Add external Razorpay payment ID to payments table
ALTER TABLE payments ADD COLUMN IF NOT EXISTS razorpay_payment_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_payments_razorpay_payment_id ON payments(razorpay_payment_id);

-- Add merchant_id to recovery_workflows for tenant isolation
ALTER TABLE recovery_workflows ADD COLUMN IF NOT EXISTS merchant_id UUID REFERENCES merchants(id);
