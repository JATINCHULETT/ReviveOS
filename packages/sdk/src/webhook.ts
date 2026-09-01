import * as crypto from "crypto";
import { normalizeRazorpayEvent } from "./normalizer";
import { ReviveOSSignatureError, ReviveOSValidationError } from "./errors";
import { UnifiedPaymentEvent, WebhookVerificationResult } from "./types";

export class RazorpayWebhookManager {
  constructor(private readonly webhookSecret?: string) {}

  /**
   * Cryptographically verify the incoming webhook signature from Razorpay.
   */
  public verifySignature(rawBody: string | Buffer, signature: string): boolean {
    if (!this.webhookSecret) {
      // In mock/sandbox development without secret, allow validation
      return true;
    }

    if (!signature) {
      return false;
    }

    try {
      const hmac = crypto.createHmac("sha256", this.webhookSecret);
      const bodyBuffer = typeof rawBody === "string" ? Buffer.from(rawBody, "utf8") : rawBody;
      hmac.update(bodyBuffer);
      const calculatedSignature = hmac.digest("hex");

      const sigBuffer = Buffer.from(signature.toLowerCase(), "utf8");
      const calcBuffer = Buffer.from(calculatedSignature.toLowerCase(), "utf8");

      if (sigBuffer.length !== calcBuffer.length) {
        return false;
      }

      return crypto.timingSafeEqual(sigBuffer, calcBuffer);
    } catch {
      return false;
    }
  }

  /**
   * Verify signature and normalize raw payload into UnifiedPaymentEvent.
   * Throws ReviveOSSignatureError if signature verification fails.
   */
  public verifyAndNormalize(
    rawBody: string | Buffer,
    signature?: string,
    headers?: Record<string, string | string[] | undefined>
  ): UnifiedPaymentEvent {
    if (this.webhookSecret && signature) {
      const isValid = this.verifySignature(rawBody, signature);
      if (!isValid) {
        throw new ReviveOSSignatureError("Razorpay webhook signature verification failed.");
      }
    }

    const bodyString = typeof rawBody === "string" ? rawBody : rawBody.toString("utf8");
    return normalizeRazorpayEvent(bodyString, headers);
  }

  /**
   * Safe non-throwing verification helper.
   */
  public safeVerifyAndNormalize(
    rawBody: string | Buffer,
    signature?: string,
    headers?: Record<string, string | string[] | undefined>
  ): WebhookVerificationResult {
    try {
      const event = this.verifyAndNormalize(rawBody, signature, headers);
      return { valid: true, event };
    } catch (err: any) {
      return { valid: false, error: err.message || "Invalid webhook payload" };
    }
  }
}
