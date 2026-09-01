# @reviveos/razorpay

> **Official Node.js & TypeScript SDK for ReviveOS AI Payment Recovery**

Plugs directly into your existing Razorpay integration to diagnose failed transactions, prevent fraud, predict recovery likelihood, and autonomously orchestrate recovery without proxying normal traffic.

---

## Installation

```bash
npm install @reviveos/razorpay
```

---

## Quickstart

### 1. Initialize the SDK

```typescript
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS({
  apiKey: process.env.REVIVEOS_API_KEY,
  webhookSecret: process.env.RAZORPAY_WEBHOOK_SECRET,
  endpoint: process.env.REVIVEOS_API_URL || "https://api.reviveos.io",
});
```

### 2. Next.js App Router Webhook (`app/api/reviveos/webhook/route.ts`)

```typescript
import { NextResponse } from "next/server";
import { ReviveOS } from "@reviveos/razorpay";

const revive = new ReviveOS();

export async function POST(req: Request) {
  const rawBody = await req.text();
  const signature = req.headers.get("x-razorpay-signature") || "";

  // 1. Verify cryptographic signature & normalize to provider-agnostic event
  const event = revive.webhooks.verifyAndNormalize(rawBody, signature);

  // 2. Ingest into ReviveOS AI Recovery Engine
  const result = await revive.events.process(event);

  return NextResponse.json(result);
}
```

### 3. Synchronous Payment Failure Intelligence

```typescript
const diagnosis = await revive.payments.analyze({
  paymentId: "pay_L8k3n2...",
  amount: 4999.0,
  currency: "INR",
  failureCode: "BAD_REQUEST_ERROR",
  failureReason: "Payment failed due to temporary issuer downtime",
  customerId: "cust_9921",
});

console.log(diagnosis.decision); // "RECOVER"
console.log(diagnosis.nextBestAction); // "RETRY_LATER"
console.log(diagnosis.delaySeconds); // 120
console.log(diagnosis.fraudRisk); // { fraudProbability: 0.03, riskLevel: 'LOW' }
```

---

## Features

- **Zero Gateway Replacement**: Normal successful payments continue directly through Razorpay.
- **Unified Event Model**: Standardizes error codes and transaction lifecycles.
- **Built-in Mock Sandbox**: Set `REVIVEOS_MODE=mock` for zero-latency local development.
- **Replay & Idempotency Protection**: Safe under webhook duplication.
- **Lightweight**: Zero native ML dependencies inside the npm package.
