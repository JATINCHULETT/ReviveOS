DROP TABLE IF EXISTS model_predictions;
DROP TABLE IF EXISTS ai_decisions;
DROP TABLE IF EXISTS payment_events;
ALTER TABLE payments DROP COLUMN IF EXISTS razorpay_payment_id;
ALTER TABLE recovery_workflows DROP COLUMN IF EXISTS merchant_id;
