const test = require("node:test");
const assert = require("node:assert");
const crypto = require("node:crypto");

// Import from compiled or source files
const { normalizeRazorpayEvent } = require("../dist/normalizer");
const { RazorpayWebhookManager } = require("../dist/webhook");
const { MockReviveOSEngine } = require("../dist/mock");
const { ReviveOS } = require("../dist/index");
const { ReviveOSSignatureError } = require("../dist/errors");

test("SDK: Event Normalization converts Razorpay payload to Unified format", () => {
  const rawPayload = {
    entity: "event",
    account_id: "acc_test",
    event: "payment.failed",
    payload: {
      payment: {
        entity: {
          id: "pay_test_normalize_123",
          amount: 499900, // paise
          currency: "INR",
          status: "failed",
          method: "card",
          email: "customer@example.com",
          contact: "+919876543210",
          error_code: "BAD_REQUEST_ERROR",
          error_description: "Payment failed due to low balance",
          error_reason: "insufficient_funds",
          created_at: 1700000000,
        },
      },
    },
    created_at: 1700000000,
  };

  const headers = { "x-razorpay-event-id": "evt_test_rzp_99" };
  const event = normalizeRazorpayEvent(rawPayload, headers);

  assert.strictEqual(event.provider, "razorpay");
  assert.strictEqual(event.eventType, "payment.failed");
  assert.strictEqual(event.paymentId, "pay_test_normalize_123");
  assert.strictEqual(event.amount, 4999.0); // converted to INR
  assert.strictEqual(event.currency, "INR");
  assert.strictEqual(event.customerEmail, "customer@example.com");
  assert.strictEqual(event.customerPhone, "+919876543210");
  assert.strictEqual(event.failure?.code, "BAD_REQUEST_ERROR");
  assert.strictEqual(event.failure?.reason, "insufficient_funds");
  assert.strictEqual(event.eventId, "evt_test_rzp_99");
});

test("SDK: Webhook HMAC-SHA256 Signature Verification", () => {
  const secret = "whsec_reviveos_super_secret";
  const body = JSON.stringify({ event: "payment.failed", id: "pay_123" });

  const validSignature = crypto.createHmac("sha256", secret).update(body).digest("hex");
  const invalidSignature = "deadbeef12345678";

  const webhook = new RazorpayWebhookManager(secret);

  assert.strictEqual(webhook.verifySignature(body, validSignature), true);
  assert.strictEqual(webhook.verifySignature(body, invalidSignature), false);

  // verifyAndNormalize success
  const event = webhook.verifyAndNormalize(body, validSignature);
  assert.ok(event);

  // verifyAndNormalize rejection throws ReviveOSSignatureError
  assert.throws(
    () => webhook.verifyAndNormalize(body, invalidSignature),
    ReviveOSSignatureError
  );
});

test("SDK: ReviveOS Mock Mode End-to-End Execution", async () => {
  const revive = new ReviveOS({
    apiKey: "rvo_test_mock_key",
    mode: "mock",
  });

  // 1. Process Event
  const rawMockPayload = MockReviveOSEngine.createMockFailurePayload({
    paymentId: "pay_mock_test_1",
    amount: 1500,
    failureCode: "BANK_OFFLINE",
  });
  const normalized = normalizeRazorpayEvent(rawMockPayload);
  const ingestRes = await revive.events.process(normalized);
  assert.strictEqual(ingestRes.status, "INGESTED");

  // 2. Direct Synchronous Analyze
  const analysis = await revive.payments.analyze({
    paymentId: "pay_mock_test_1",
    amount: 1500,
    failureCode: "BANK_OFFLINE",
    failureReason: "Bank gateway downtime",
  });

  assert.strictEqual(analysis.paymentId, "pay_mock_test_1");
  assert.strictEqual(analysis.decision, "RECOVER");
  assert.strictEqual(analysis.action, "RETRY_LATER");
  assert.strictEqual(analysis.delaySeconds, 120);
  assert.ok(analysis.recoveryProbability > 0.8);
  assert.strictEqual(analysis.fraudRisk.riskLevel, "LOW");

  // 3. Customer Recovery Profile
  const profile = await revive.customers.getRecoveryProfile("cust_123");
  assert.strictEqual(profile.customerId, "cust_123");
  assert.strictEqual(profile.previousSuccessfulRecoveries, 3);
  assert.ok(profile.recoveryProbability > 0.8);

  // 4. Decision & Execution
  const decision = await revive.recovery.decision({
    paymentId: "pay_mock_test_1",
    failureType: "BANK_OFFLINE",
    fraudProbability: 0.03,
    recoveryProbability: 0.94,
  });
  assert.strictEqual(decision.decision, "RECOVER");

  const execution = await revive.recovery.execute({
    paymentId: "pay_mock_test_1",
    action: "RETRY",
  });
  assert.strictEqual(execution.status, "EXECUTED");
  assert.strictEqual(execution.verified, true);
});
