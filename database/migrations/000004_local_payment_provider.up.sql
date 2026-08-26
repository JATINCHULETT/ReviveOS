-- Local Payment Provider Tables
CREATE TABLE IF NOT EXISTS local_provider_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id VARCHAR(255) NOT NULL UNIQUE,
    provider_payment_id VARCHAR(255) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    status VARCHAR(50) NOT NULL, -- FAILED, AUTHORIZED, CAPTURED, REFUNDED, PENDING
    method VARCHAR(50) NOT NULL DEFAULT 'card',
    failure_code VARCHAR(100),
    failure_reason TEXT,
    attempts_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_local_provider_payments_payment_id ON local_provider_payments(payment_id);
CREATE INDEX IF NOT EXISTS idx_local_provider_payments_provider_id ON local_provider_payments(provider_payment_id);

CREATE TABLE IF NOT EXISTS local_provider_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    local_payment_id UUID REFERENCES local_provider_payments(id) ON DELETE CASCADE,
    payment_id VARCHAR(255) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    status VARCHAR(50) NOT NULL, -- SUCCESS, FAILED, PENDING
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
