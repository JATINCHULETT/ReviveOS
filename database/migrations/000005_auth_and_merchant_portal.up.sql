-- Users table for Authentication and Role Separation
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'MERCHANT', -- 'ADMIN' or 'MERCHANT'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ensure Subscriptions table has extended fields for Razorpay recurring & sandbox
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS plan_id VARCHAR(100);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS razorpay_subscription_id VARCHAR(100);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS payment_link_url VARCHAR(500);
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS billing_interval VARCHAR(50) DEFAULT 'monthly';

-- Seed default merchants if not exist
INSERT INTO merchants (id, name)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'Acme Cloud Services'),
    ('00000000-0000-0000-0000-000000000002', 'Zenith Health SaaS')
ON CONFLICT (id) DO NOTHING;

-- Seed default admin and merchant accounts
-- Passwords: 
-- admin@reviveos.io -> admin123
-- merchant@acme.com -> merchant123
INSERT INTO users (id, merchant_id, email, password_hash, name, role)
VALUES 
    ('10000000-0000-0000-0000-000000000001', NULL, 'admin@reviveos.io', '2167d46816a7dc9fae5e6e66e746a5b2:fbf0f4e24ef5a4e320f305085e3cb289b4f2c050ec469f3796fcb1d283626e2e', 'System Administrator', 'ADMIN'),
    ('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'merchant@acme.com', '2167d46816a7dc9fae5e6e66e746a5b2:c7e75525c56784865103a743419ea7bbde0490b4d45543c7b8d4b3dfba5bce68', 'Acme Merchant Owner', 'MERCHANT')
ON CONFLICT (email) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_merchant_id ON users(merchant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_merchant_id ON subscriptions(merchant_id);
