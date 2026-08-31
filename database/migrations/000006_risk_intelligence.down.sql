DROP TABLE IF EXISTS risk_assessments CASCADE;
ALTER TABLE recovery_workflows DROP COLUMN IF EXISTS fraud_probability;
ALTER TABLE recovery_workflows DROP COLUMN IF EXISTS return_probability;
ALTER TABLE recovery_workflows DROP COLUMN IF EXISTS overall_risk;
ALTER TABLE recovery_workflows DROP COLUMN IF EXISTS expected_loss;
ALTER TABLE recovery_workflows DROP COLUMN IF EXISTS risk_action;
