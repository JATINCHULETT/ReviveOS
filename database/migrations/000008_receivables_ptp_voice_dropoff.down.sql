-- Migration: 000008_receivables_ptp_voice_dropoff.down.sql
DROP TABLE IF EXISTS checkout_sessions;
DROP TABLE IF EXISTS voice_recovery_calls;
DROP TABLE IF EXISTS promise_to_pay_records;
DROP TABLE IF EXISTS b2b_dunning_logs;
DROP TABLE IF EXISTS b2b_invoices;
