package paymentprovider

import (
	"context"
	"time"
)

// PaymentStatus represents the authoritative state of a payment from a PaymentProvider.
type PaymentStatus struct {
	PaymentID         string    `json:"payment_id"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	Status            string    `json:"status"` // FAILED, AUTHORIZED, CAPTURED, REFUNDED, PENDING
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Method            string    `json:"method"`
	FailureCode       string    `json:"failure_code,omitempty"`
	FailureReason     string    `json:"failure_reason,omitempty"`
	Captured          bool      `json:"captured"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RetryResult represents the outcome of a provider retry attempt.
type RetryResult struct {
	AttemptID         string    `json:"attempt_id"`
	PaymentID         string    `json:"payment_id"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	Status            string    `json:"status"` // SUCCESS, FAILED, PENDING
	Amount            float64   `json:"amount"`
	PaymentLinkURL    string    `json:"payment_link_url,omitempty"`
	ErrorMessage      string    `json:"error_message,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// PaymentProvider defines the authoritative payment gateway interface.
type PaymentProvider interface {
	GetPayment(ctx context.Context, paymentID string) (*PaymentStatus, error)
	CreateRetryAttempt(ctx context.Context, paymentID string, amount float64) (*RetryResult, error)
	VerifyPayment(ctx context.Context, paymentID string) (*PaymentStatus, error)
}
