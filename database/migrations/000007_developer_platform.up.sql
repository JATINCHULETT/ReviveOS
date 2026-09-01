-- Migration 000007: Developer Platform, API Keys, Unified Events, and Audit Logs

-- 1. API Keys Table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(16) NOT NULL, -- e.g. 'rvo_test_' or 'rvo_live_'
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hex digest
    mode VARCHAR(10) NOT NULL DEFAULT 'test', -- 'test' or 'live'
    last_used_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_merchant ON api_keys(merchant_id);

-- 2. Audit Logs for Security & Traceability
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    actor_type VARCHAR(32) NOT NULL, -- 'API_KEY', 'WEBHOOK', 'USER', 'SYSTEM'
    actor_id VARCHAR(128),
    action VARCHAR(64) NOT NULL, -- e.g. 'EVENT_INGESTED', 'PAYMENT_ANALYZED', 'RECOVERY_EXECUTED'
    ip_address VARCHAR(45),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_merchant_created ON audit_logs(merchant_id, created_at DESC);

-- 3. Ensure default test API key for default merchants
-- Pre-seeded key: rvo_test_acme_secret_key_12345
-- SHA-256 of "rvo_test_acme_secret_key_12345" = 3f5d52b123ba4e67f7bbff2f2cb5efcf31f38e07297e68bc43f1ddbc762b3a7a
INSERT INTO api_keys (id, merchant_id, name, key_prefix, key_hash, mode)
VALUES (
    '30000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'Default Test Key',
    'rvo_test_',
    '3f5d52b123ba4e67f7bbff2f2cb5efcf31f38e07297e68bc43f1ddbc762b3a7a',
    'test'
)
ON CONFLICT (key_hash) DO NOTHING;
