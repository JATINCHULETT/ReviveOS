DROP TABLE IF EXISTS experiment_results;
DROP TABLE IF EXISTS experiments;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS recovery_outcomes;
DROP TABLE IF EXISTS recovery_actions;
DROP TABLE IF EXISTS recovery_workflows;
DROP TABLE IF EXISTS failure_events;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS merchants;

DROP EXTENSION IF EXISTS "uuid-ossp";
