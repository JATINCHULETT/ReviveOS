export class ReviveOSError extends Error {
  constructor(message: string, public readonly code: string = "REVIVEOS_ERROR") {
    super(message);
    this.name = "ReviveOSError";
    Object.setPrototypeOf(this, new.target.prototype);
  }
}

export class ReviveOSAuthError extends ReviveOSError {
  constructor(message: string = "Invalid or missing ReviveOS API key") {
    super(message, "REVIVEOS_AUTH_ERROR");
    this.name = "ReviveOSAuthError";
  }
}

export class ReviveOSSignatureError extends ReviveOSError {
  constructor(message: string = "Invalid webhook signature from payment gateway") {
    super(message, "REVIVEOS_SIGNATURE_ERROR");
    this.name = "ReviveOSSignatureError";
  }
}

export class ReviveOSValidationError extends ReviveOSError {
  constructor(message: string) {
    super(message, "REVIVEOS_VALIDATION_ERROR");
    this.name = "ReviveOSValidationError";
  }
}

export class ReviveOSApiError extends ReviveOSError {
  constructor(
    message: string,
    public readonly statusCode: number,
    public readonly responseBody?: any
  ) {
    super(`ReviveOS API Error (${statusCode}): ${message}`, "REVIVEOS_API_ERROR");
    this.name = "ReviveOSApiError";
  }
}
