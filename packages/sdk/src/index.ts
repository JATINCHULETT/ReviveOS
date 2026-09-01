import { ReviveOSHttpClient } from "./client";
import { RazorpayWebhookManager } from "./webhook";
import { MockReviveOSEngine } from "./mock";
import {
  ReviveOSOptions,
  UnifiedPaymentEvent,
  PaymentAnalysisResult,
  RecoveryDecisionRequest,
  RecoveryDecisionResponse,
  RecoveryExecutionResult,
  CustomerRecoveryProfile,
  IngestEventResponse,
} from "./types";

export * from "./types";
export * from "./errors";
export * from "./normalizer";
export * from "./webhook";
export * from "./mock";

export class ReviveOS {
  public readonly webhooks: RazorpayWebhookManager;
  public readonly mock = MockReviveOSEngine;
  private readonly http: ReviveOSHttpClient;
  private readonly isMockMode: boolean;

  constructor(options: ReviveOSOptions = {}) {
    const mode = options.mode || (process.env.REVIVEOS_MODE as any) || "live";
    this.isMockMode = mode === "mock";

    const endpoint =
      options.endpoint ||
      process.env.REVIVEOS_API_URL ||
      "http://localhost:8080";

    const apiKey = options.apiKey || process.env.REVIVEOS_API_KEY;
    const webhookSecret = options.webhookSecret || process.env.RAZORPAY_WEBHOOK_SECRET;

    this.webhooks = new RazorpayWebhookManager(webhookSecret);
    this.http = new ReviveOSHttpClient({
      endpoint,
      apiKey,
      timeoutMs: options.timeoutMs,
      maxRetries: options.maxRetries,
    });
  }

  /**
   * Payment Event Ingestion methods.
   */
  public readonly events = {
    /**
     * Ingest a normalized payment event into the ReviveOS pipeline.
     */
    process: async (event: UnifiedPaymentEvent): Promise<IngestEventResponse> => {
      if (this.isMockMode) {
        return {
          status: "INGESTED",
          event_id: event.eventId,
          payment_id: event.paymentId,
          timestamp: new Date().toISOString(),
          message: "Mock payment event ingested successfully.",
        };
      }
      return this.http.request<IngestEventResponse>("/v1/events", "POST", event);
    },
  };

  /**
   * Real-time Payment Intelligence & Diagnosis methods.
   */
  public readonly payments = {
    /**
     * Run synchronous diagnosis, fraud scoring, and recovery prediction.
     */
    analyze: async (params: {
      paymentId: string;
      orderId?: string;
      customerId?: string;
      customerEmail?: string;
      amount: number;
      currency?: string;
      paymentMethod?: string;
      bank?: string;
      failureCode: string;
      failureReason?: string;
      attemptNumber?: number;
    }): Promise<PaymentAnalysisResult> => {
      if (this.isMockMode) {
        const dummyEvent: UnifiedPaymentEvent = {
          provider: "mock",
          eventType: "payment.failed",
          eventId: `evt_mock_${params.paymentId}`,
          paymentId: params.paymentId,
          customerId: params.customerId,
          amount: params.amount,
          currency: params.currency || "INR",
          timestamp: new Date().toISOString(),
          failure: {
            code: params.failureCode,
            reason: params.failureReason,
          },
        };
        return MockReviveOSEngine.analyzeMockPayment(dummyEvent);
      }
      return this.http.request<PaymentAnalysisResult>("/v1/payments/analyze", "POST", params);
    },

    /**
     * Fetch payment status and recovery history from ReviveOS.
     */
    get: async (paymentId: string): Promise<any> => {
      if (this.isMockMode) {
        return {
          id: paymentId,
          status: "FAILED",
          amount: 4999.0,
          currency: "INR",
          mock: true,
        };
      }
      return this.http.request<any>(`/v1/payments/${paymentId}`, "GET");
    },
  };

  /**
   * Recovery Engine methods.
   */
  public readonly recovery = {
    /**
     * Evaluate deterministic recovery decision against merchant policies.
     */
    decision: async (params: RecoveryDecisionRequest): Promise<RecoveryDecisionResponse> => {
      if (this.isMockMode) {
        return {
          decision: "RECOVER",
          action: "RETRY",
          delaySeconds: 120,
          confidence: 0.94,
          reason: "High recovery probability and low fraud risk (Simulated)",
          timestamp: new Date().toISOString(),
        };
      }
      return this.http.request<RecoveryDecisionResponse>("/v1/recovery/decision", "POST", params);
    },

    /**
     * Trigger payment recovery execution via gateway adapter.
     */
    execute: async (params: {
      paymentId: string;
      amount?: number;
      action: "RETRY" | "PAYMENT_LINK" | "ALTERNATIVE_PAYMENT";
      idempotencyKey?: string;
    }): Promise<RecoveryExecutionResult> => {
      if (this.isMockMode) {
        return {
          status: "EXECUTED",
          paymentId: params.paymentId,
          action: params.action,
          resultStatus: "CAPTURED",
          verified: true,
          executedAt: new Date().toISOString(),
        };
      }
      return this.http.request<RecoveryExecutionResult>("/v1/recovery/execute", "POST", params);
    },
  };

  /**
   * Customer Recovery Memory methods.
   */
  public readonly customers = {
    /**
     * Retrieve customer recovery profile and historical payment performance.
     */
    getRecoveryProfile: async (customerId: string): Promise<CustomerRecoveryProfile> => {
      if (this.isMockMode) {
        return MockReviveOSEngine.getMockCustomerProfile(customerId);
      }
      return this.http.request<CustomerRecoveryProfile>(
        `/v1/customers/${customerId}/recovery-profile`,
        "GET"
      );
    },
  };
}

export default ReviveOS;
