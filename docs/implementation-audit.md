# ReviveOS Implementation Audit

This audit evaluates the actual state of the codebase against the Master Build Specification. The previous "checkbox" development has left the repository in a state of superficial completeness, heavily reliant on mocks, stubbed endpoints, and frontend magic.

## 1. Database & Migrations
### Claimed status: `[x]`
### Actual status: **PARTIAL**
### Evidence in code: `000001_initial_schema.up.sql` exists and has the correct tables.
### Missing behavior: Constraints are basic. State isn't strongly enforced via DB triggers or strict state machines.
### Required fix: Verify foreign keys and add strict state enums if possible.
### Verification method: Run migrations against a fresh Postgres instance.

## 2. Transactional Outbox & Asynq
### Claimed status: `[x]`
### Actual status: **PARTIAL**
### Evidence in code: `apps/worker/internal/outbox/relay.go` exists. `main.go` registers `payment.failed`.
### Missing behavior: The outbox does NOT wrap `BEGIN TRANSACTION`. Asynq is only used for `payment.failed`. Crucial jobs like `recovery:analyze`, `recovery:plan`, `recovery:execute`, and `recovery:verify` **do not exist**. Delay is completely faked by setting a `SCHEDULED` status in DB without queuing a delayed Asynq task.
### Required fix: Implement real Asynq jobs (`recovery:execute`, `recovery:verify`). Wrap DB inserts and outbox creation in atomic transactions.
### Verification method: Unit test outbox crashes. E2E test delayed Asynq job execution.

## 3. Recovery State Machine & Idempotency
### Claimed status: `[x]`
### Actual status: **MISSING / FAKE**
### Evidence in code: `handler.go` transitions states manually with raw strings (e.g., `"REQUIRES_HUMAN_REVIEW"`). No idempotency checks (Redis locks) exist anywhere.
### Missing behavior: A duplicate `payment.failed` event will blindly create two concurrent recovery workflows. 
### Required fix: Implement an explicit FSM. Acquire Redis distributed locks before financial actions. Reconcile state before executing.
### Verification method: Send duplicate webhooks simultaneously; verify only one executes.

## 4. Simulator & Synthetic Generator
### Claimed status: `[x]`
### Actual status: **FAKE / HARDCODED**
### Evidence in code: `POST /simulator/failure` directly calls `IngestSimulatorFailure`. Fast-forward does not exist. 
### Missing behavior: It simulates ingestion but doesn't actually simulate execution, verification, or success because the execution workers don't exist.
### Required fix: Build the execution loop. Implement `/simulator/fast-forward` to manipulate Asynq job schedules or system time logic. Build a true 500-event synthetic generator for evaluation.
### Verification method: Trigger failure, fast-forward 24h, verify action executes.

## 5. Razorpay Adapter & Webhook
### Claimed status: `[x]`
### Actual status: **FAKE**
### Evidence in code: `webhook.go` literally calls `app.paymentService.IngestSimulatorFailure(event)`.
### Missing behavior: It bypasses real ingestion, doesn't handle idempotency, and pretends to be the simulator. It doesn't handle `payment.captured` for reconciliation.
### Required fix: Implement strict webhook processing, deduplication, signature validation, and map events cleanly into the single pipeline.
### Verification method: Razorpay integration test with real/mock webhook payloads.

## 6. AI Provider (Ollama/DeepSeek)
### Claimed status: `[x]`
### Actual status: **PARTIAL**
### Evidence in code: `ai-provider` exists and is used in `handler.go`. 
### Missing behavior: The prompt doesn't strictly enforce schema validation if the LLM hallucinates.
### Required fix: Ensure JSON schema strictness and deterministic fallback.
### Verification method: Unit test with corrupted JSON mock.

## 7. Policy Engine
### Claimed status: `[x]`
### Actual status: **PARTIAL**
### Evidence in code: `policy-engine` exists and has `ALLOW`/`BLOCK`.
### Missing behavior: Not fully evaluating dynamic attempt limits, limits aren't robustly tested against concurrent workers.
### Required fix: Expand policy rules, tie to Redis locks.
### Verification method: Test `customer says STOP` scenario.

## 8. State Reconciliation
### Claimed status: `[x]`
### Actual status: **MISSING**
### Evidence in code: Not found anywhere in `handler.go` or worker flows.
### Missing behavior: If a payment is captured manually before the retry, the system will blind-retry anyway.
### Required fix: Before executing an Asynq task, query the current payment state. If `CAPTURED`, stop the workflow immediately.
### Verification method: Simulate payment success, then execute worker.

## 9. Outcome Tracking, Evaluation & Dashboard
### Claimed status: `[x]`
### Actual status: **FAKE / HARDCODED**
### Evidence in code: `engine.go` uses `totalRisk * 0.8`. `page.tsx` uses `1842000`, `1273000`, `MOCK_DATA`.
### Missing behavior: The dashboard shows metrics that are heavily stubbed or hallucinated. Baseline vs ReviveOS evaluation is barely functional.
### Required fix: Remove all mock data. Dashboard must show exactly what is in PostgreSQL.
### Verification method: Empty database must show exactly 0. 500 synthetic events must show accurate computed lift.

## 10. Supabase Auth & Tenancy
### Claimed status: `[x]`
### Actual status: **MISSING ON BACKEND**
### Evidence in code: `api.go` and `main.go` expose endpoints without verifying JWTs.
### Missing behavior: Any user can query any endpoint and mutate any workflow.
### Required fix: Add JWT middleware. Scope all DB queries by `merchant_id`.
### Verification method: Integration test with a token belonging to Merchant B accessing Merchant A's data (must 403/404).
