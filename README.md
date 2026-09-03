# ReviveOS

## Adaptive Payment Recovery & Revenue Protection Infrastructure

ReviveOS is an AI-assisted payment recovery orchestration layer that sits on top of existing payment infrastructure. It turns payment failures into an adaptive, auditable recovery workflow: Diagnose → Predict → Decide → Time → Execute → Verify → Learn.

---

## 1. The Problem

A conventional payment flow often treats a failed payment as a terminal event:

```text
Payment → Failed → Error → End
```

ReviveOS treats the failure as a decision point:

```text
Payment Failure
      ↓
Secure Event Ingestion
      ↓
Failure Classification
      ↓
Recovery Prediction + Fraud/Risk Assessment
      ↓
Customer Context / Recovery Memory
      ↓
Policy & Next-Best-Action Decision
      ↓
Optimal Timing / Cadence
      ↓
Voice / Email / Checkout / SMS / WhatsApp
      ↓
Customer Response
      ↓
Payment Reconciliation
      ↓
Audit Proof + Outcome Memory
      ↓
Better Future Decisions
```

The objective is not to "retry more." The objective is to choose the safest, most appropriate recovery action for the specific failure and customer context.

---

## 2. What Makes ReviveOS Different?

### Traditional recovery

```text
FAILED
  ↓
RETRY
```

### ReviveOS

```text
FAILED
  ↓
WHY DID IT FAIL?
  ↓
CAN THIS CUSTOMER RECOVER?
  ↓
IS THERE FRAUD / RISK?
  ↓
WHAT ACTION HAS THE HIGHEST EXPECTED VALUE?
  ↓
WHEN SHOULD WE ACT?
  ↓
WHICH CHANNEL SHOULD WE USE?
  ↓
EXECUTE WITH POLICY GUARDS
  ↓
DID THE PAYMENT ACTUALLY RECOVER?
  ↓
RECORD THE OUTCOME
  ↓
USE THE OUTCOME AS FUTURE CONTEXT
```

ReviveOS combines ML decisioning, customer recovery memory, policy controls, multi-channel orchestration, reconciliation, human escalation, PTP commitments, and cryptographic auditability into one recovery control plane.

---

## 3. Core Product Pillars

### Intelligence
- Deterministic payment failure classification across 14 failure codes
- Recovery probability scoring
- Customer recovery memory
- Customer/context-aware channel selection
- Random Forest fraud/risk scoring
- Expected-loss / risk-aware safeguards
- Next-best-action decisioning
- Dynamic recovery delay and cadence calculation

### Recovery Orchestration
- Razorpay webhook ingestion
- Smart checkout recovery
- Alternate payment rails: UPI, NetBanking, alternate cards
- Resend email dunning
- Twilio outbound voice
- Hinglish / Hindi speech interaction
- SMS / WhatsApp triggers
- Workflow state machine
- Automatic reconciliation after successful payment

### Revenue Recovery
- B2B receivables and aging buckets
- NET_15 / NET_30 / NET_60 / due-on-receipt workflows
- Dunning cadences
- Promise-to-Pay (PTP) commitments
- PTP honor / extension lifecycle
- Escalation to human inspection
- Audit dossier export

### Trust + Developer Infrastructure
- HMAC-SHA256 webhook verification
- Timestamp / anti-replay protection
- Payload invariant validation
- PII masking / tokenization
- Retry limits and circuit breakers
- SHA-256 audit hash chain
- Payment reconciliation
- Human override controls
- Go backend for a clean, high-performance service boundary
- TypeScript SDK
- One-command `npx reviveos init` integration

---

## 4. End-to-End Architecture

```text
                         MERCHANT APPLICATION
              Next.js / Express / Node.js / Mobile Apps
                                  │
                                  │
                    Razorpay Payment Webhooks / SDK
                                  │
                                  ▼
                    ┌──────────────────────────┐
                    │ 1. INGESTION & SECURITY  │
                    │                          │
                    │ HMAC verification        │
                    │ Replay protection        │
                    │ Timestamp validation     │
                    │ Payload invariants       │
                    │ PII masking/tokenization │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │ 2. AI / ML INTELLIGENCE  │
                    │                          │
                    │ Failure classifier       │
                    │ Recovery model           │
                    │ Fraud model              │
                    │ Customer recovery memory │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │ 3. POLICY + DECISION     │
                    │                          │
                    │ Next-best-action         │
                    │ Dynamic timing           │
                    │ Retry bounds             │
                    │ Circuit breakers         │
                    │ Risk / loss safeguards   │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                 ┌─────────────────────────────────┐
                 │ 4. RECOVERY ORCHESTRATOR        │
                 │                                 │
                 │ Twilio Voice / Hinglish         │
                 │ Resend Email                    │
                 │ SMS / WhatsApp                  │
                 │ Smart Checkout                  │
                 └────────────────┬────────────────┘
                                  │
                                  ▼
                    ┌──────────────────────────┐
                    │ 5. RESPONSE + STATE      │
                    │                          │
                    │ PTP / dispute / link     │
                    │ clicks / callbacks       │
                    │ escalation / human review│
                    └────────────┬─────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │ 6. RECONCILIATION        │
                    │                          │
                    │ payment.captured         │
                    │ outcome verification     │
                    │ recovery confirmation    │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │ 7. AUDIT + MEMORY        │
                    │                          │
                    │ SHA-256 audit chain      │
                    │ PTP ledger               │
                    │ Recovery memory          │
                    │ Workflow finality        │
                    └──────────────────────────┘
```

---

## 5. The 10-Stage Recovery State Machine

Every recovery workflow can be understood as a controlled state transition:

| # | Stage | Responsibility |
|---|---|---|
| 1 | `INGESTED` | Receive and authenticate the payment event |
| 2 | `IDENTIFIED` | Resolve customer context, velocity and communication constraints |
| 3 | `ML_EVALUATED` | Score recovery probability and relevant risk |
| 4 | `RISK_CHECKED` | Apply fraud/risk and expected-loss safeguards |
| 5 | `POLICY_DECIDED` | Enforce retry limits, circuit breakers and merchant policy |
| 6 | `STRATEGY_SELECTED` | Select action, channel and recovery cadence |
| 7 | `DISPATCHED` | Execute Voice, Email, SMS, WhatsApp or Checkout recovery |
| 8 | `CUSTOMER_RESPONDED` | Capture PTP, dispute, callback, link click or other response |
| 9 | `RECONCILED` | Verify the final payment outcome using gateway events |
| 10 | `FINALIZED / ESCALATED / PTP_LOGGED` | Seal the lifecycle and route exceptions appropriately |

The state machine prevents an AI recommendation from becoming an uncontrolled payment action.

---

## 6. Machine Learning

ReviveOS uses different mechanisms for different decisions instead of forcing every problem through an LLM.

### Recovery Prediction
- **Model**: Logistic Regression
- **Purpose**:
  - Estimate probability of successful recovery
  - Support next-best-action selection
  - Provide a compact, interpretable score
  - Feed policy/cadence decisions

Conceptually:

```text
Payment Context
     +
Failure Type
     +
Customer History
     +
Prior Recovery Outcomes
     +
Channel Context
     ↓
Recovery Probability
```

### Fraud / Risk Detection
- **Model**: Random Forest
- **Purpose**:
  - Calculate fraud probability / risk
  - Prevent aggressive recovery behavior when risk is high
  - Feed expected-loss and policy safeguards
  - Separate "recoverable failure" from "unsafe transaction"

```text
Transaction Signals
       ↓
Fraud Model
       ↓
Risk Score
       ↓
Policy Guard
       ↓
ALLOW / REVIEW / BLOCK
```

### Failure Intelligence
A deterministic classifier maps payment failures into actionable categories, including examples such as:
- Insufficient funds
- Issuer bank unavailable
- Authentication failure
- Network timeout
- Other supported gateway failure codes

Failure classification is deliberately separate from probabilistic recovery and fraud models.

### Why not use an LLM for everything?
LLMs are useful for reasoning and unstructured customer interaction, but payment execution needs deterministic controls.

The intended boundary is:

```text
Signals
  ↓
ML Models
  ↓
Structured Decision
  ↓
Policy Validation
  ↓
Execution
```

For voice interactions, AI/NLP can interpret the customer's response, while the workflow engine determines what state transition is permitted.

---

## 7. Recovery Data

ReviveOS recovery intelligence is designed around transaction-level recovery cases, not just raw payment rows.

A recovery case should preserve enough context to answer:
- What failed?
- Who was the customer?
- What happened previously?
- What action was attempted?
- When was it attempted?
- Through which channel?
- What did the customer do?
- Did the payment recover?
- How long did recovery take?
- What should the system learn from the outcome?

### Recommended logical schema

```text
recovery_case
├── case_id
├── transaction_id
├── customer_id
├── amount
├── currency
├── payment_method
├── failure_code
├── failure_category
├── timestamp
├── prior_attempt_count
├── prior_recovery_count
├── previous_successful_channel
├── selected_channel
├── selected_action
├── selected_delay
├── recovery_probability
├── risk_probability
├── actual_outcome
├── recovery_time
└── final_state
```

### Target variable
For recovery prediction:
- `recovered = 1` when the failed payment is successfully recovered within the defined recovery window; otherwise:
- `recovered = 0`

The exact recovery window and label definition should be documented alongside the dataset used for training/evaluation.

---

## 8. Fraud Detection Data

### 8A. Example Data

The following examples are illustrative synthetic records for documentation and local testing. They are not production customer data and must not be interpreted as measured model performance.

#### Recovery Dataset — Example Rows

| case_id | failure_code | amount_inr | payment_method | prior_attempts | prior_recoveries | preferred_channel | recovery_delay_min | recovered |
|---|---|---|---|---|---|---|---|---|
| REC-1001 | INSUFFICIENT_FUNDS | 4999 | CARD | 1 | 2 | UPI | 23 | 1 |
| REC-1002 | AUTHENTICATION_FAILURE | 14999 | CARD | 2 | 1 | CHECKOUT | 10 | 1 |
| REC-1003 | NETWORK_TIMEOUT | 2499 | UPI | 1 | 0 | EMAIL | 60 | 0 |
| REC-1004 | ISSUER_BANK_DOWN | 8999 | CARD | 1 | 3 | SMS | 120 | 1 |
| REC-1005 | INSUFFICIENT_FUNDS | 1299 | CARD | 3 | 0 | VOICE | 240 | 0 |
| REC-1006 | AUTHENTICATION_FAILURE | 19999 | CARD | 1 | 2 | UPI | 30 | 1 |

#### Recovery record interpretation
- `REC-1001`:
  - Failure: `INSUFFICIENT_FUNDS`
  - Amount: ₹4,999
  - History: 1 previous attempt, 2 previous successful recoveries
  - Candidate strategy: UPI recovery after 23 minutes
  - Outcome: `recovered = 1`

A real training dataset should contain substantially more rows and should document its provenance, feature engineering, label definition, split strategy and evaluation methodology.

#### Fraud Dataset — Example Rows

| transaction_id | amount_inr | payment_method | attempts_10m | account_age_days | prior_failures | velocity_score | fraud_label |
|---|---|---|---|---|---|---|---|
| TX-2001 | 4999 | CARD | 1 | 420 | 1 | 0.12 | 0 |
| TX-2002 | 89999 | CARD | 8 | 3 | 6 | 0.94 | 1 |
| TX-2003 | 2499 | UPI | 1 | 730 | 0 | 0.04 | 0 |
| TX-2004 | 45999 | CARD | 6 | 12 | 4 | 0.87 | 1 |
| TX-2005 | 7999 | NETBANKING | 1 | 210 | 1 | 0.18 | 0 |
| TX-2006 | 129999 | CARD | 10 | 1 | 8 | 0.98 | 1 |

Where:
- `fraud_label = 1` → synthetic fraud example
- `fraud_label = 0` → synthetic legitimate example

The fraud model should use these signals as inputs to a risk assessment. A model score is not proof of fraud; the policy layer determines whether to allow, review or block a workflow.

#### Example Customer Recovery Memory

```json
{
  "customer_id": "CUS-DEMO-001",
  "previous_failures": 4,
  "previous_recoveries": 3,
  "successful_channels": [
    "UPI_CHECKOUT",
    "EMAIL"
  ],
  "preferred_payment_method": "UPI",
  "last_successful_delay_minutes": 23,
  "ptp_honor_rate": 0.67
}
```

This demonstrates how historical recovery outcomes can become context for the next decision.

#### Example Model Input → Decision

```json
{
  "failure_code": "INSUFFICIENT_FUNDS",
  "amount_inr": 14999,
  "payment_method": "CARD",
  "prior_attempts": 1,
  "prior_recoveries": 3,
  "preferred_channel": "UPI_CHECKOUT",
  "fraud_probability": 0.021
}
```

Possible structured decision:

```json
{
  "recovery_probability": 0.728,
  "recommended_action": "SEND_RECOVERY_LINK",
  "channel": "UPI_CHECKOUT",
  "delay_seconds": 1380,
  "policy_status": "ALLOW"
}
```

These values are illustrative examples only. They are included to explain the data contract and decision flow, not to claim model performance.

Fraud detection should be treated as a risk guard, not simply another dashboard score.

#### Example logical feature groups

```text
Transaction
├── amount
├── payment method
├── transaction frequency
├── attempt velocity
└── failure pattern

Customer / Account Context
├── historical behavior
├── prior successful activity
├── prior failed activity
└── account-level signals

Risk Context
├── abnormal attempt patterns
├── unusual transaction behavior
└── model-derived risk score
```

#### Target
- `fraud = 1`
- `legitimate = 0`

Use a separate train/validation/test methodology and report the actual metrics produced by the repository's evaluation pipeline.

Do not interpret a high fraud probability as proof of fraud. It is a risk signal used by the policy layer.

---

## 9. Dataset & Model Reproducibility

This repository should keep data provenance explicit.

Recommended structure:

```text
data/
├── README.md
├── raw/              # source/synthetic raw data; do not commit secrets
├── processed/        # transformed model-ready datasets
├── recovery/         # recovery training/evaluation data
└── fraud/            # fraud training/evaluation data

models/
├── recovery/
└── fraud/

scripts/
├── prepare_data.*
├── train_recovery.*
├── train_fraud.*
└── evaluate.*
```

For every model, document:
- Dataset source
- Whether data is synthetic, anonymized, or production-derived
- Feature definitions
- Label definition
- Train/validation/test split
- Preprocessing
- Model type
- Hyperparameters
- Evaluation metrics
- Known limitations
- Model artifact/version

Never commit API keys, payment secrets, phone credentials, or personally identifiable production data.

---

## 10. Customer Recovery Memory

ReviveOS does not have to treat every failure independently.

The memory layer can retain recovery-relevant signals such as:

```text
Customer
  ↓
Previous failures
Previous successful recoveries
Successful channels
Successful timing windows
Prior PTP commitments
Communication preferences / opt-outs
  ↓
Current decision
```

Example:
- Previous recovery: Email + 23 min delay → successful
- Current failure: Same customer + similar failure
- Decision: Prior successful strategy becomes a candidate, subject to current risk and policy checks.

This is the foundation for customer-aware recovery rather than blind retry logic.

---

## 11. Multi-Channel Recovery

### Smart Checkout
When a card/payment rail fails, ReviveOS can generate a self-serve recovery experience:

```text
Original payment failed
        ↓
Recovery link
        ↓
UPI / NetBanking / alternate card
        ↓
Payment
        ↓
Gateway verification
```

Recovery links should be short-lived and protected against stale/duplicate attempts.

### Resend Email
Email is used for:
- Recovery links
- Payment reminders
- Dunning cadences
- PTP follow-up
- Workflow-triggered notifications

### Twilio Voice + Hinglish
ReviveOS supports a conversational recovery flow:

```text
Workflow
   ↓
Outbound call
   ↓
Hindi/Hinglish prompt
   ↓
Speech recognition
   ↓
Intent extraction
   ↓
Workflow state transition
```

Example intents:
- `PROMISE_TO_PAY`
- `REQUEST_LINK`
- `CALL_LATER`
- `DISPUTE`

Example:
> *"Kal subah 11 baje tak payment kar dunga."*

becomes:
- `PROMISE_TO_PAY`
- `date = tomorrow`
- `time = 11:00`
- `amount = outstanding amount`

The resulting commitment enters the PTP lifecycle rather than remaining only as a call transcript.

---

## 12. Promise-to-Pay (PTP)

PTP bridges conversational recovery and operational collections.

```text
VOICE / CHAT
     ↓
PROMISE_TO_PAY
     ↓
PENDING
     ├──→ HONORED
     ├──→ EXTENDED
     └──→ ESCALATED
```

Operational actions can include:
- Verify Paid
- Extend commitment
- Trigger reminder
- Escalate for human inspection

Payment verification should reconcile against gateway payment events rather than relying only on a customer's statement.

---

## 13. B2B Receivables

ReviveOS extends beyond consumer card recovery into accounts receivable workflows.

### Aging buckets
- `CURRENT`
- `1–30`
- `31–60`
- `90+`

### Example B2B flow

```text
ERP / Invoice
     ↓
Credit-term reconciliation
     ↓
Aging classification
     ↓
Dunning strategy
     ↓
Payment link
     ↓
PTP / payment
     ↓
Reconciliation
     ↓
Audit dossier
```

Supported terms include: `NET_15`, `NET_30`, `NET_60`, `DUE_ON_RECEIPT`.

---

## 14. Human-in-the-Loop

Autonomy should not mean removing operators.

ReviveOS provides human controls for workflows that require judgment:

```text
AI / ML
   ↓
Recommended Strategy
   ↓
Policy / Risk Gate
   ↓
┌─────────────────────────────┐
│ Human Inspection (if needed)│
│                             │
│ Approve & Dispatch          │
│ Override Strategy           │
│ Extend PTP                  │
│ Escalate                    │
└─────────────────────────────┘
```

This creates a practical boundary between automation and operational control.

---

## 15. Auditability & Cryptographic Proof

A payment-recovery system needs more than logs.

ReviveOS records the workflow lifecycle and can seal the audit history with a SHA-256 hash chain.

Conceptually:

```text
Event 1
  ↓ hash
Event 2 + previous_hash
  ↓ hash
Event 3 + previous_hash
  ↓ hash
...
```

This makes tampering detectable because changing an earlier event changes the downstream hash chain.

The audit trail can cover:
- Webhook ingestion
- ML decisions
- Risk decisions
- Policy decisions
- Dispatch
- Customer responses
- PTP events
- Reconciliation
- Human overrides
- Final outcome

---

## 16. Payment Integration

ReviveOS is designed as a layer around existing payment infrastructure rather than a replacement payment gateway.

```text
Merchant
   ↓
Razorpay
   ↓
Payment event / webhook
   ↓
ReviveOS
   ↓
Recovery decision
   ↓
Merchant / payment flow
```

### Razorpay webhook integration
The ingestion boundary should validate:
- HMAC signature
- Timestamp / replay protection
- Payload invariants
- Event identity / idempotency requirements

Successful recovery is verified using subsequent payment gateway events such as `payment.captured`.

---

## 17. Developer Experience

The goal is to make integration feel like infrastructure, not a migration project.

### One-command initialization

```bash
npx reviveos init
```

The CLI is intended to:
- Detect the application framework
- Generate the required webhook route
- Install/configure the ReviveOS SDK
- Create integration configuration
- Provide the next steps for connecting payment events

### TypeScript SDK

```bash
npm install @reviveos/razorpay
```

The SDK provides a typed integration surface for synchronous diagnosis and webhook ingestion.

### Backend
ReviveOS uses Go (Golang) for its backend service layer.

This gives the platform a compact, strongly typed, high-performance integration boundary that is well suited to webhook processing, APIs, concurrent workflow execution, and payment infrastructure.

The Go backend also makes ReviveOS straightforward to deploy as an independent service alongside a merchant's existing stack rather than forcing merchants to rewrite their application.

---

## 18. Technology Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 15, App Router |
| UI | React 19, TypeScript |
| Styling | Vanilla CSS tokens |
| Backend | Go 1.22+ |
| HTTP | Chi / net/http |
| Database | PostgreSQL |
| PostgreSQL driver | pgx v5 |
| Recovery ML | Logistic Regression |
| Fraud ML | Random Forest |
| Failure intelligence | Deterministic classifier |
| AI reasoning | DeepSeek-R1 where applicable |
| Voice | Twilio Voice REST API |
| Speech input | Twilio `<Gather>`, hi-IN |
| Email | Resend |
| Payments | Razorpay webhook integration |
| Security | HMAC-SHA256, anti-replay, PII masking |
| Audit | SHA-256 hash chain |

---

## 19. Security Design

Security is part of the recovery pipeline, not an afterthought.

```text
Webhook
  ↓
Signature Verification
  ↓
Timestamp / Replay Guard
  ↓
Payload Validation
  ↓
PII Masking / Tokenization
  ↓
Processing
```

Additional controls include:
- Bounded retries
- Circuit breakers
- Risk safeguards
- Communication opt-out checks
- Human override paths
- Short-lived recovery links
- Idempotent workflow handling
- No secrets committed to source control

### Important principle
> **AI recommends. Policy validates. Deterministic infrastructure executes.**

---

## 20. Example Recovery Decision

```json
{
  "failure": "INSUFFICIENT_FUNDS",
  "recovery_probability": 0.728,
  "risk_probability": 0.021,
  "recommended_action": "SEND_RECOVERY_LINK",
  "channel": "UPI_CHECKOUT",
  "delay_seconds": 1380,
  "reason": "Customer history indicates strong recovery after a short delay through an alternate payment rail."
}
```

The exact output schema in production should follow the repository's implemented API contracts.

---

## 21. Example Workflow

```text
1. Razorpay emits payment failure
        ↓
2. ReviveOS verifies webhook
        ↓
3. Failure is classified
        ↓
4. Recovery model scores the case
        ↓
5. Fraud/risk model evaluates risk
        ↓
6. Customer memory is loaded
        ↓
7. Policy engine checks limits
        ↓
8. Next-best action is selected
        ↓
9. Cadence / timing is selected
        ↓
10. Recovery channel executes
        ↓
11. Customer responds
        ↓
12. PTP / checkout / follow-up workflow starts
        ↓
13. payment.captured arrives
        ↓
14. Recovery is reconciled
        ↓
15. Audit trail is sealed
        ↓
16. Outcome updates recovery memory
```

---

## 22. Failure → Recovery Examples

| Failure / Situation | Possible ReviveOS Response |
|---|---|
| Insufficient funds | Delay + alternate payment link |
| Authentication failure | Recovery checkout / alternate rail |
| Issuer/network issue | Cooldown before retry |
| High fraud probability | Risk guard / human review |
| Customer requests payment link | Send recovery checkout |
| Customer says "I'll pay tomorrow" | Create PTP commitment |
| Customer disputes amount | Escalate rather than retry |
| Repeated failed recovery | Apply retry bound / escalate |
| B2B invoice overdue | Aging-based dunning workflow |

The final action is always subject to the implemented policy and risk rules.

---

## 23. Repository Guide

```text
reviveos/
├── apps/
│   └── web/                   # Next.js dashboard & operations app
│
├── apps/api/                  # Go service entrypoints & handlers
│   ├── handlers/              # Webhooks, voice, receivables, PTP, analytics
│   └── main.go                # API server initialization
│
├── packages/
│   ├── checkout/              # Smart checkout generator
│   ├── cli/                   # npx reviveos CLI auto-configurator
│   ├── ptp/                   # Promise-to-Pay state machine
│   ├── receivables/           # B2B aging & dunning
│   ├── recovery/              # Decision engine & ML models
│   ├── risk/                  # Fraud detection & expected loss
│   ├── voice/                 # Twilio Hinglish voice & speech recognition
│   └── sdk/                   # TypeScript @reviveos/razorpay SDK
│
└── README.md
```

---

## 24. Evaluation & Metrics

Model quality should be measured independently from product-level recovery outcomes.

### Recovery model
Report, when available:
- Accuracy
- Precision
- Recall
- F1
- ROC-AUC
- Calibration / probability quality
- Confusion matrix

### Fraud model
Report:
- Precision
- Recall
- F1
- ROC-AUC
- False-positive rate
- False-negative rate
- Confusion matrix

### Product-level metrics
The most important business measurements are:
- Recovery Rate
- Recovered Amount
- Revenue Protected
- Time to Recovery
- Attempts per Recovered Payment
- Channel Recovery Rate
- PTP Honor Rate
- Escalation Rate
- Fraud Loss Avoided

Always distinguish model metrics from business metrics. Do not claim production lift unless it has been measured against a real baseline.

---

## 24A. Impact & Results

*Note: The numbers below are illustrative benchmark/demo values unless the repository contains a reproducible experiment that generates the same results.*

### Example Baseline vs ReviveOS

| Metric | Baseline | ReviveOS | Change |
|---|---|---|---|
| Payment recovery rate | 34% | 58% | +24 pp |
| Average recovery attempts | 3.1 | 1.8 | -42% |
| Time to recovery | 4.2 h | 2.6 h | -38% |
| Unnecessary retry rate | 27% | 11% | -59% |
| Customer-contact coverage | 1 channel | 3 channels | +200% |
| Workflows with complete audit trail | 62% | 100% | +38 pp |

### What improved?

```text
                    BASELINE          REVIVEOS
                    ────────          ────────

Recovery rate        34%      →         58%
Attempts             3.1      →         1.8
Recovery time        4.2h     →         2.6h
Unnecessary retry    27%      →         11%
Audit coverage       62%      →        100%
```

The intended product-level improvement is not simply "more retries." ReviveOS aims to improve recovery by making the workflow:

```text
BLIND RETRY
    ↓
ADAPTIVE DECISION
    ↓
RIGHT ACTION
    ↓
RIGHT TIME
    ↓
RIGHT CHANNEL
    ↓
VERIFIED OUTCOME
```

---

## 25. Engineering Decisions

- **Why Go?**: The core backend is written in Go to provide a strongly typed, efficient service boundary for webhook ingestion, concurrent workflow execution, APIs and payment infrastructure integration.
- **Why PostgreSQL?**: Recovery workflows require durable relational state, transactional updates, reconciliation and audit records.
- **Why separate ML models?**: Recovery probability and fraud probability represent different objectives and should not be conflated.
- **Why a deterministic failure classifier?**: Payment failure codes are structured signals. A deterministic mapping is easier to test and reason about than using an LLM for a known finite classification problem.
- **Why policy after ML?**: A high recovery score must never bypass merchant limits, risk controls, opt-outs or retry bounds.
- **Why reconciliation?**: A recovery message or checkout click is not proof of payment. The gateway's resulting payment event is the source used to verify the outcome.
- **Why human-in-the-loop?**: Some disputes, high-risk transactions and exceptional workflows require operational judgment.

---

## 26. Technical Challenges

- **Webhook reliability**: Payment webhooks can be duplicated, delayed or replayed. ReviveOS addresses this with signature verification, timestamp checks, replay protection and idempotent workflow design.
- **AI-to-execution boundary**: An AI recommendation should not directly execute a financial action.
  ```text
  AI / ML → Structured recommendation → Policy engine → Risk guard → Deterministic execution
  ```
- **Recovery loops**: Repeated retries can increase customer friction and financial/operational loss. ReviveOS therefore uses retry bounds, dynamic cadence and circuit breakers.
- **Reconciliation**: The system must distinguish: `Message sent ≠ Customer clicked ≠ Customer promised ≠ Payment captured`. Only the appropriate gateway event should finalize a successful payment recovery.

---

## 27. Demo Flow

The strongest end-to-end demonstration is:

```text
Razorpay payment failure
        ↓
Webhook verified
        ↓
Failure classified
        ↓
Recovery + fraud scored
        ↓
Customer memory loaded
        ↓
Next-best action selected
        ↓
Recovery link / Voice call
        ↓
Customer response
        ↓
PTP or checkout recovery
        ↓
Payment captured
        ↓
Reconciliation
        ↓
Audit proof
        ↓
Outcome stored in memory
```

---

## 28. What ReviveOS Is NOT

ReviveOS is not:
- A generic chatbot
- An LLM connected directly to payment execution
- A simple retry scheduler
- A replacement payment gateway
- A dashboard-only analytics product
- A fraud model without an operational workflow

It is a payment recovery orchestration and revenue protection layer.

---

## 29. Current Prototype / Production Boundaries

Before production deployment, validate:
- Real merchant traffic
- Real-world model drift
- Payment-provider-specific edge cases
- Rate limits
- Consent and communication requirements
- Data retention requirements
- Secrets management
- High availability
- Queue durability
- Disaster recovery
- Compliance requirements
- Human escalation SLAs
- Model monitoring and retraining

---

## 30. Quick Start

```bash
# Clone
git clone https://github.com/JATINCHULETT/ReviveOS.git
cd ReviveOS

# Install web dependencies
cd apps/web
npm install

# Run the dashboard
npm run dev
```

Go backend:

```bash
cd apps/api
go run main.go
```

One-command merchant integration:

```bash
npx reviveos init
```

---

## 30A. How to Use ReviveOS Locally

### Prerequisites
- Git
- Node.js 20+
- npm
- Go 1.22+
- PostgreSQL
- Modern web browser

### 1. Install & Configure

```bash
cd apps/web
npm install
```

Configure `.env`:

```env
DATABASE_URL=postgres://user:password@localhost:5432/reviveos
RAZORPAY_KEY_ID=rzp_test_xxx
RAZORPAY_KEY_SECRET=xxx
RAZORPAY_WEBHOOK_SECRET=xxx
TWILIO_ACCOUNT_SID=ACxxx
TWILIO_AUTH_TOKEN=xxx
TWILIO_PHONE_NUMBER=+1xxx
RESEND_API_KEY=re_xxx
```

### 2. Start Go Backend

```bash
cd apps/api
go run main.go
```

### 3. Start Next.js App

```bash
cd apps/web
npm run dev
```

Open `http://localhost:3000`.

---

## 31. Configuration

```env
# Application
PORT=8080
DATABASE_URL=postgres://...

# Razorpay
RAZORPAY_KEY_ID=rzp_test_...
RAZORPAY_KEY_SECRET=...
RAZORPAY_WEBHOOK_SECRET=...

# Twilio
TWILIO_ACCOUNT_SID=AC...
TWILIO_AUTH_TOKEN=...
TWILIO_PHONE_NUMBER=+1...

# Resend
RESEND_API_KEY=re_...
```

---

## 32. Philosophy

ReviveOS is built around one principle:

> **A failed payment is not a binary outcome. It is a decision opportunity.**

```text
Why did it fail?
       ↓
Can it be recovered?
       ↓
Is recovery safe?
       ↓
What should we do?
       ↓
When should we do it?
       ↓
How should we communicate?
       ↓
Did it work?
       ↓
What did we learn?
```

That is the ReviveOS loop.

---

## 33. Summary

ReviveOS = Adaptive Recovery + Revenue Protection + Operational Control

```text
                    REVIVEOS
                       │
       ┌───────────────┼────────────────┐
       ↓               ↓                ↓
    RECOVER         PROTECT          COLLECT
       │               │                │
    Checkout        Fraud            B2B AR
    Voice           Risk             PTP
    Email           Guards           Dunning
    WhatsApp        Policies         Escalation
       │               │                │
       └───────────────┼────────────────┘
                       ↓
                 VERIFICATION
                       ↓
                  AUDIT + MEMORY
```

**Core loop**: Diagnose → Predict → Decide → Time → Execute → Verify → Learn

> **Don't just tell a merchant that a payment failed. Tell them what should happen next — and safely orchestrate it.**
