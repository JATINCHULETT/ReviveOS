import {
  PaymentAnalysisResult,
  CustomerRecoveryProfile,
  UnifiedPaymentEvent,
} from "./types";

export class MockReviveOSEngine {
  public static createMockFailurePayload(options?: {
    paymentId?: string;
    amount?: number;
    failureCode?: string;
    email?: string;
  }) {
    const paymentId = options?.paymentId || `pay_mock_${Date.now()}`;
    const amount = options?.amount ? options.amount * 100 : 499900; // in paise
    const failureCode = options?.failureCode || "BAD_REQUEST_ERROR";
    const email = options?.email || "customer@example.com";

    return {
      entity: "event",
      account_id: "acc_mock_reviveos_01",
      event: "payment.failed",
      contains: ["payment"],
      payload: {
        payment: {
          entity: {
            id: paymentId,
            entity: "payment",
            amount,
            currency: "INR",
            status: "failed",
            method: "card",
            description: "Mock subscription charge",
            email,
            contact: "+919876543210",
            error_code: failureCode,
            error_description: "Payment failed due to temporary issuer downtime",
            error_source: "bank",
            error_step: "payment_authentication",
            error_reason: "payment_failed",
            created_at: Math.floor(Date.now() / 1000),
          },
        },
      },
      created_at: Math.floor(Date.now() / 1000),
    };
  }

  public static analyzeMockPayment(event: UnifiedPaymentEvent): PaymentAnalysisResult {
    const isHighFraud = event.amount > 80000;
    const isBankOutage =
      event.failure?.code.includes("BANK") ||
      event.failure?.description?.includes("downtime") ||
      event.failure?.description?.includes("issuer");

    const fraudProbability = isHighFraud ? 0.78 : 0.04;
    const fraudRiskLevel = isHighFraud ? "HIGH" : "LOW";
    const recoveryProbability = isHighFraud ? 0.12 : isBankOutage ? 0.94 : 0.82;
    const action = isHighFraud ? "BLOCK" : isBankOutage ? "RETRY_LATER" : "RETRY_NOW";
    const decision = isHighFraud ? "BLOCK" : "RECOVER";
    const delaySeconds = isBankOutage ? 120 : 0;

    return {
      paymentId: event.paymentId,
      failureCategory: isBankOutage ? "BANK_UNAVAILABLE" : "INSUFFICIENT_FUNDS",
      diagnosis: isHighFraud
        ? "High anomaly risk score flagged by Revenue Risk Engine"
        : "Temporary bank downtime detected with high recovery probability",
      fraudRisk: {
        fraudProbability,
        riskLevel: fraudRiskLevel,
        expectedLoss: isHighFraud ? event.amount : 0,
        overallRisk: fraudRiskLevel,
      },
      recoveryProbability,
      nextBestAction: action,
      action,
      delaySeconds,
      confidence: 0.93,
      reason: isHighFraud
        ? "Blocked due to risk policy thresholds"
        : "Deterministic recovery strategy scheduled",
      customerHistory: {
        successful_payments: 4,
        failed_payments: 1,
      },
      decision,
      timestamp: new Date().toISOString(),
    };
  }

  public static getMockCustomerProfile(customerId: string): CustomerRecoveryProfile {
    return {
      customerId,
      email: "alex.merchant@example.com",
      phone: "+919876543210",
      previousFailures: 3,
      previousSuccessfulRecoveries: 3,
      preferredPaymentMethods: ["card", "upi"],
      averageRecoveryTime: 180,
      recoveryProbability: 0.88,
      communicationOptOut: false,
      lastActivityAt: new Date().toISOString(),
    };
  }
}
