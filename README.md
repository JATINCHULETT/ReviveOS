# ReviveOS: Autonomous AI Payment Recovery & Revenue Protection

> **ReviveOS is an AI payment recovery and revenue protection layer that plugs into your existing payment infrastructure with one command.**

```text
Merchant Application
        │
        ├───────────────► Razorpay (Normal Successful Traffic)
        │
        │ payment events / failures
        ▼
   ReviveOS SDK (@reviveos/razorpay)
        │
        ▼
   ReviveOS Core API
        │
        ├── Failure Intelligence (Deterministic Classifier)
        ├── Fraud Detection (Random Forest ML)
        ├── Recovery Prediction (Logistic Regression)
        ├── Customer Recovery Memory (Historical Calibration)
        ├── Next Best Action & Optimal Timing (Decision Engine)
        └── Recovery Orchestration (Local & Razorpay Adapters)
        │
        ▼
   Recovery Decision ──► Provider Retry / Interactive Smart Link
```

---

## ⚡ 1-Command Integration (Under 60 Seconds)

In your existing Node.js / Next.js / Express Razorpay application:

```bash
# 1. Install ReviveOS Razorpay SDK
npm install @reviveos/razorpay

# 2. Run the Auto-Configurator
npx reviveos init
```

The CLI automatically:
1. Detects your project framework (Next.js App Router, Pages Router, Express, Fastify).
2. Generates `.env.reviveos` configuration template.
3. Generates a secure, production-ready webhook handler (`app/api/reviveos/webhook/route.ts` or `src/api/reviveos-webhook.ts`).
4. Configures HMAC-SHA256 signature verification, replay protection, and failure intelligence.

---

## 💻 Minimal Webhook Handler Example

### Next.js App Router (`app/api/reviveos/webhook/route.ts`)
```typescript
import { NextResponse } from "next/server";
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
});

export async function POST(req: Request) {
  const rawBody = await req.text();
  const signature = req.headers.get("x-razorpay-signature") || "";

  // 1. Verify cryptographic HMAC signature & normalize event
  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);

  // 2. Ingest into ReviveOS AI pipeline
  const result = await revive.events.process(event);

  return NextResponse.json(result);
}
```

---

## 🧠 Synchronous Payment Failure Diagnosis

```typescript
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS();

const diagnosis = await revive.payments.analyze({
  paymentId: "pay_xyz_123",
  amount: 4999.0,
  currency: "INR",
  failureCode: "BAD_REQUEST_ERROR",
  failureReason: "Payment failed due to temporary issuer bank downtime",
  customerId: "cust_alex_123",
  paymentMethod: "card",
});

console.log(diagnosis.decision);           // "RECOVER"
console.log(diagnosis.nextBestAction);     // "RETRY_LATER"
console.log(diagnosis.delaySeconds);       // 120
console.log(diagnosis.fraudRisk);          // { fraudProbability: 0.04, riskLevel: 'LOW' }
console.log(diagnosis.recoveryProbability);// 0.94
```

---

## 🔐 Environment Variables

| Variable | Required | Description |
|---|---|---|
| `REVIVEOS_API_KEY` | Yes | Merchant API key (`rvo_test_...` or `rvo_live_...`) |
| `RAZORPAY_WEBHOOK_SECRET` | Optional | Webhook secret for cryptographic HMAC verification |
| `REVIVEOS_API_URL` | Optional | Core API endpoint (Default: `https://api.reviveos.io` or `http://localhost:8080`) |
| `REVIVEOS_MODE` | Optional | Set to `mock` for zero-latency offline sandbox simulation |

---

## 🏗️ Architecture & Monorepo Structure

```text
ReviveOS/
├── apps/
│   ├── api/             # Go High-Performance REST & Ingestion Engine (:8080)
│   ├── web/             # Next.js Analytics, Decision Hub & Developer Portal (:3000)
│   └── worker/          # Asynchronous 6-Stage Autonomous Pipeline Worker
├── packages/
│   ├── sdk/             # @reviveos/razorpay TypeScript SDK
│   ├── cli/             # reviveos CLI (npx reviveos init)
│   ├── recovery/        # Logistic Recovery Probability & Customer Memory Model
│   ├── risk/            # Go ML Risk Client
│   ├── types/           # Deterministic Failure Classifier & Lifecycle Enums
│   ├── schemas/         # Unified Event & Task Schemas
│   └── utils/           # Outbox transactional relay
├── services/
│   ├── ml_service/      # Python FastAPI Revenue Risk Engine (:8000)
│   ├── ai-provider/     # LLM Strategy Engine (Ollama / NVIDIA NIM with Safe Fallback)
│   ├── policy-engine/   # Deterministic Guardrails & Merchant Safety Policies
│   └── payment-provider/# Razorpay & Local Sandbox Adapters
└── database/
    └── migrations/      # Version-controlled PostgreSQL Schema (000001 - 000007)
```

---

## 🚀 Running Locally

```bash
# 1. Start ML Risk Service
python services/ml_service/main.py

# 2. Start Core Go API
go run ./apps/api

# 3. Start Background Worker
go run ./apps/worker

# 4. Start Next.js Dashboard
cd apps/web && npm run dev
```
