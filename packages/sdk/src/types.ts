export type ReviveOSMode = "live" | "test" | "mock";

export interface ReviveOSFeatures {
  recovery?: boolean;
  fraud?: boolean;
  customerMemory?: boolean;
  refundRecovery?: boolean;
}

export interface ReviveOSOptions {
  apiKey?: string;
  webhookSecret?: string;
  endpoint?: string;
  mode?: ReviveOSMode;
  features?: ReviveOSFeatures;
  timeoutMs?: number;
  maxRetries?: number;
}

export interface UnifiedFailureDetail {
  code: string;
  reason?: string;
  description?: string;
  step?: string;
  source?: string;
}

export interface UnifiedPaymentEvent {
  provider: "razorpay" | "stripe" | "mock" | string;
  eventType: string; // "payment.failed", "payment.captured", "refund.created", etc.
  eventId: string;
  idempotencyKey?: string;
  paymentId: string;
  orderId?: string;
  customerId?: string;
  customerEmail?: string;
  customerPhone?: string;
  amount: number; // in major units, e.g. 4999.00
  currency: string;
  paymentMethod?: string;
  bank?: string;
  timestamp: string;
  failure?: UnifiedFailureDetail;
  metadata?: Record<string, any>;
  rawPayload?: any;
}

export interface CustomerHistorySummary {
  successful_payments: number;
  failed_payments: number;
}

export interface FraudRiskScore {
  fraudProbability: number;
  riskLevel: "LOW" | "MEDIUM" | "HIGH";
  expectedLoss?: number;
  overallRisk: string;
}

export interface PaymentAnalysisResult {
  paymentId: string;
  failureCategory: string;
  diagnosis: string;
  fraudRisk: FraudRiskScore;
  recoveryProbability: number;
  nextBestAction: "RETRY_NOW" | "RETRY_LATER" | "ALTERNATIVE_PAYMENT" | "PAYMENT_LINK" | "NO_ACTION" | "BLOCK" | string;
  action: string;
  delaySeconds: number;
  confidence: number;
  reason: string;
  customerHistory: CustomerHistorySummary;
  decision: "RECOVER" | "BLOCK" | "ESCALATE" | "NO_ACTION" | string;
  timestamp: string;
}

export interface RecoveryDecisionRequest {
  paymentId: string;
  failureType: string;
  fraudProbability: number;
  recoveryProbability: number;
  customerHistory?: Record<string, any>;
  paymentContext?: Record<string, any>;
}

export interface RecoveryDecisionResponse {
  decision: "RECOVER" | "BLOCK" | "ESCALATE" | "NO_ACTION";
  action: string;
  delaySeconds: number;
  confidence: number;
  reason: string;
  timestamp: string;
}

export interface RecoveryExecutionResult {
  status: "EXECUTED" | "FAILED" | "SCHEDULED";
  paymentId: string;
  action: string;
  resultStatus: string;
  paymentLinkUrl?: string;
  verified: boolean;
  executedAt: string;
}

export interface CustomerRecoveryProfile {
  customerId: string;
  email?: string;
  phone?: string;
  previousFailures: number;
  previousSuccessfulRecoveries: number;
  preferredPaymentMethods: string[];
  averageRecoveryTime: number; // in seconds
  recoveryProbability: number;
  communicationOptOut: boolean;
  lastActivityAt: string;
}

export interface WebhookVerificationResult {
  valid: boolean;
  event?: UnifiedPaymentEvent;
  error?: string;
}

export interface IngestEventResponse {
  status: "INGESTED" | "DUPLICATE_IGNORED";
  event_id: string;
  idempotency_key?: string;
  payment_id: string;
  internal_id?: string;
  timestamp: string;
  message?: string;
}
