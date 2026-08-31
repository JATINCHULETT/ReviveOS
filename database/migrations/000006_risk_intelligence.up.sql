CREATE TABLE IF NOT EXISTS risk_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID REFERENCES payments(id) ON DELETE CASCADE,
    workflow_id UUID REFERENCES recovery_workflows(id) ON DELETE SET NULL,
    merchant_id UUID REFERENCES merchants(id) ON DELETE SET NULL,
    event_type VARCHAR(64) NOT NULL,
    fraud_probability FLOAT NOT NULL DEFAULT 0.0,
    fraud_risk_level VARCHAR(32) NOT NULL DEFAULT 'LOW',
    return_probability FLOAT DEFAULT 0.0,
    return_risk_level VARCHAR(32) DEFAULT 'LOW',
    overall_risk_level VARCHAR(32) NOT NULL DEFAULT 'LOW',
    expected_loss NUMERIC(12, 2) DEFAULT 0.00,
    recommended_action VARCHAR(64) NOT NULL DEFAULT 'ALLOW',
    reason TEXT,
    model_version VARCHAR(64) DEFAULT 'fraud-rf-v1.0',
    raw_payload JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_risk_assessments_payment ON risk_assessments(payment_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_workflow ON risk_assessments(workflow_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_merchant ON risk_assessments(merchant_id);

ALTER TABLE recovery_workflows ADD COLUMN IF NOT EXISTS fraud_probability FLOAT DEFAULT 0.0;
ALTER TABLE recovery_workflows ADD COLUMN IF NOT EXISTS return_probability FLOAT DEFAULT 0.0;
ALTER TABLE recovery_workflows ADD COLUMN IF NOT EXISTS overall_risk VARCHAR(32) DEFAULT 'LOW';
ALTER TABLE recovery_workflows ADD COLUMN IF NOT EXISTS expected_loss NUMERIC(12, 2) DEFAULT 0.00;
ALTER TABLE recovery_workflows ADD COLUMN IF NOT EXISTS risk_action VARCHAR(64) DEFAULT 'ALLOW';
