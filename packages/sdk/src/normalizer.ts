import { UnifiedPaymentEvent } from "./types";

/**
 * Normalize Razorpay webhook payload into a provider-independent ReviveOS UnifiedPaymentEvent.
 */
export function normalizeRazorpayEvent(
  rawBody: any,
  headers?: Record<string, string | string[] | undefined>
): UnifiedPaymentEvent {
  const payload = typeof rawBody === "string" ? JSON.parse(rawBody) : rawBody;

  const eventType = payload.event || "payment.failed";
  const paymentEntity = payload.payload?.payment?.entity || {};
  const orderEntity = payload.payload?.order?.entity || {};
  const refundEntity = payload.payload?.refund?.entity || {};

  const paymentId = paymentEntity.id || refundEntity.payment_id || "pay_unknown";
  const orderId = paymentEntity.order_id || orderEntity.id || undefined;
  
  // Extract customer info
  let email = paymentEntity.email;
  let phone = paymentEntity.contact;
  if (!email && paymentEntity.customer) {
    if (typeof paymentEntity.customer === "object") {
      email = paymentEntity.customer.email;
      phone = paymentEntity.customer.contact;
    }
  }

  // Parse amount from paise to major units (INR)
  const rawAmount = paymentEntity.amount ?? orderEntity.amount ?? refundEntity.amount ?? 0;
  const amount = typeof rawAmount === "number" ? rawAmount / 100.0 : parseFloat(rawAmount) / 100.0;

  // Extract failure details
  const failure =
    paymentEntity.error_code || paymentEntity.error_description || paymentEntity.error_reason
      ? {
          code: paymentEntity.error_code || "UNKNOWN_ERROR",
          reason: paymentEntity.error_reason || "",
          description: paymentEntity.error_description || "",
          step: paymentEntity.error_step || "",
          source: paymentEntity.error_source || "",
        }
      : undefined;

  // Extract Event ID from header or payload
  let eventId = "";
  if (headers) {
    const headerEventId = headers["x-razorpay-event-id"];
    if (typeof headerEventId === "string") {
      eventId = headerEventId;
    } else if (Array.isArray(headerEventId) && headerEventId.length > 0) {
      eventId = headerEventId[0];
    }
  }
  if (!eventId) {
    eventId = `evt_${paymentId}_${payload.created_at || Date.now()}`;
  }

  const timestamp = payload.created_at
    ? new Date(payload.created_at * 1000).toISOString()
    : new Date().toISOString();

  return {
    provider: "razorpay",
    eventType,
    eventId,
    idempotencyKey: eventId,
    paymentId,
    orderId,
    customerId: paymentEntity.customer_id || undefined,
    customerEmail: email || undefined,
    customerPhone: phone || undefined,
    amount,
    currency: paymentEntity.currency || "INR",
    paymentMethod: paymentEntity.method || undefined,
    bank: paymentEntity.bank || undefined,
    timestamp,
    failure,
    rawPayload: payload,
  };
}
