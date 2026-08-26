package types

import "strings"

// FailureCategory represents a deterministic failure class
type FailureCategory string

const (
	NetworkError           FailureCategory = "NETWORK_ERROR"
	BankUnavailable        FailureCategory = "BANK_UNAVAILABLE"
	Timeout                FailureCategory = "TIMEOUT"
	InsufficientFunds      FailureCategory = "INSUFFICIENT_FUNDS"
	ExpiredCard            FailureCategory = "EXPIRED_CARD"
	AuthenticationFailed   FailureCategory = "AUTHENTICATION_FAILED"
	LimitExceeded          FailureCategory = "LIMIT_EXCEEDED"
	MandateFailed          FailureCategory = "MANDATE_FAILED"
	CustomerActionRequired FailureCategory = "CUSTOMER_ACTION_REQUIRED"
	UnknownError           FailureCategory = "UNKNOWN"
)

// Classifier deterministically maps raw gateway error codes/messages to a standardized category
type Classifier struct{}

func NewClassifier() *Classifier {
	return &Classifier{}
}

// Classify maps a raw error code or description to a standard FailureCategory.
func (c *Classifier) Classify(rawCode, rawMessage string) FailureCategory {
	code := strings.ToUpper(rawCode)
	msg := strings.ToUpper(rawMessage)

	// Simple deterministic matching
	// In reality, this would contain a large mapping of standard gateway codes
	// e.g. Razorpay error codes, Stripe error codes, etc.

	switch {
	case containsAny(code, msg, "INSUFFICIENT", "FUNDS", "LOW_BALANCE"):
		return InsufficientFunds
	case containsAny(code, msg, "EXPIRED", "CARD_EXPIRED"):
		return ExpiredCard
	case containsAny(code, msg, "AUTH_FAILED", "AUTHENTICATION", "3D_SECURE"):
		return AuthenticationFailed
	case containsAny(code, msg, "LIMIT", "QUOTA_EXCEEDED", "MAX_AMOUNT"):
		return LimitExceeded
	case containsAny(code, msg, "MANDATE", "RECURRING_FAILED"):
		return MandateFailed
	case containsAny(code, msg, "BANK_OFFLINE", "BANK_UNAVAILABLE", "ISSUER_DOWN"):
		return BankUnavailable
	case containsAny(code, msg, "TIMEOUT", "GATEWAY_TIMEOUT"):
		return Timeout
	case containsAny(code, msg, "NETWORK", "CONNECTION"):
		return NetworkError
	case containsAny(code, msg, "ACTION_REQUIRED", "DO_NOT_HONOR", "RESTRICTED"):
		return CustomerActionRequired
	}

	// Fallback for explicitly passed standard categories
	switch FailureCategory(code) {
	case NetworkError, BankUnavailable, Timeout, InsufficientFunds, ExpiredCard, AuthenticationFailed, LimitExceeded, MandateFailed, CustomerActionRequired:
		return FailureCategory(code)
	}

	return UnknownError
}

func containsAny(code, msg string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(code, kw) || strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}
