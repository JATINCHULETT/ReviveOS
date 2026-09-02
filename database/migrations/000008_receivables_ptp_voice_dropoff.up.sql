-- Migration: 000008_receivables_ptp_voice_dropoff.up.sql
-- New Capabilities: B2B Receivables Chaser, Promise-to-Pay (PTP) Tracker, 
-- Hinglish AI Voice Recovery, and Checkout Drop-off Recovery

-- 1. B2B Invoices & Receivables
CREATE TABLE IF NOT EXISTS b2b_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    invoice_number VARCHAR(100) NOT NULL,
    buyer_company VARCHAR(255) NOT NULL,
    buyer_name VARCHAR(255) NOT NULL,
    buyer_email VARCHAR(255) NOT NULL,
    buyer_phone VARCHAR(50),
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    issue_date DATE NOT NULL,
    due_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'UNPAID', -- UNPAID, PARTIALLY_PAID, PAID, OVERDUE, WRITTEN_OFF
    credit_terms VARCHAR(50) DEFAULT 'NET_30', -- NET_15, NET_30, NET_60
    current_bucket VARCHAR(50) DEFAULT 'CURRENT', -- CURRENT, 1_30, 31_60, 61_90, 90_PLUS
    days_past_due INT DEFAULT 0,
    dunning_stage INT DEFAULT 0,
    last_contacted_at TIMESTAMPTZ,
    next_action_due TIMESTAMPTZ,
    payment_link_url TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(merchant_id, invoice_number)
);

CREATE TABLE IF NOT EXISTS b2b_dunning_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID REFERENCES b2b_invoices(id) ON DELETE CASCADE,
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    dunning_stage INT NOT NULL,
    channel VARCHAR(50) NOT NULL, -- EMAIL, SMS, VOICE, WHATSAPP
    status VARCHAR(50) DEFAULT 'SENT', -- SENT, DELIVERED, OPENED, FAILED
    recipient VARCHAR(255) NOT NULL,
    message_excerpt TEXT,
    external_message_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Promise to Pay (PTP) Tracker
CREATE TABLE IF NOT EXISTS promise_to_pay_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    customer_id VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255),
    customer_contact VARCHAR(255) NOT NULL,
    invoice_id UUID REFERENCES b2b_invoices(id) ON DELETE SET NULL,
    payment_event_id UUID REFERENCES payment_events(id) ON DELETE SET NULL,
    promised_amount NUMERIC(12, 2) NOT NULL,
    promised_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, HONORED, BROKEN, EXTENDED, CANCELLED
    recorded_channel VARCHAR(50) DEFAULT 'VOICE_AGENT', -- VOICE_AGENT, PORTAL, EMAIL, SMS
    extension_count INT DEFAULT 0,
    last_reminder_sent_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Hinglish AI Voice Recovery Calls
CREATE TABLE IF NOT EXISTS voice_recovery_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    customer_id VARCHAR(255),
    recipient_phone VARCHAR(50) NOT NULL,
    customer_name VARCHAR(255),
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    language VARCHAR(50) DEFAULT 'Hinglish',
    provider VARCHAR(50) DEFAULT 'local', -- local, twilio, exotel
    provider_call_sid VARCHAR(255),
    call_status VARCHAR(50) DEFAULT 'COMPLETED', -- INITIATED, RINGING, IN_PROGRESS, COMPLETED, FAILED, NO_ANSWER
    duration_seconds INT DEFAULT 0,
    hinglish_script TEXT,
    customer_response TEXT,
    intent_detected VARCHAR(100), -- PROMISE_TO_PAY, DISPUTE, REQUEST_LINK, ALREADY_PAID, CALL_LATER
    ptp_created_id UUID REFERENCES promise_to_pay_records(id) ON DELETE SET NULL,
    call_recording_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Checkout Drop-off Recovery
CREATE TABLE IF NOT EXISTS checkout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    customer_name VARCHAR(255),
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    cart_amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'INR',
    cart_items_json JSONB DEFAULT '[]'::jsonb,
    step_reached VARCHAR(100) DEFAULT 'CART_LOADED', -- CART_LOADED, DETAILS_ENTERED, PAYMENT_STEP, 3DS_INITIATED, ABANDONED
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, DROPPED_OFF, RECOVERY_DISPATCHED, RECOVERED, EXPIRED
    drop_off_reason VARCHAR(255),
    recovery_link TEXT,
    recovery_dispatched_at TIMESTAMPTZ,
    recovered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_b2b_invoices_merchant ON b2b_invoices(merchant_id, status);
CREATE INDEX IF NOT EXISTS idx_b2b_invoices_bucket ON b2b_invoices(current_bucket);
CREATE INDEX IF NOT EXISTS idx_ptp_records_status ON promise_to_pay_records(status, promised_date);
CREATE INDEX IF NOT EXISTS idx_voice_calls_merchant ON voice_recovery_calls(merchant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_checkout_sessions_status ON checkout_sessions(status, created_at DESC);
