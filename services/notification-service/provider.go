package notificationservice

import (
	"context"
	"time"
)

// NotificationRequest defines payload for sending payment recovery notifications.
type NotificationRequest struct {
	PaymentID     string  `json:"payment_id"`
	WorkflowID    string  `json:"workflow_id"`
	MerchantName  string  `json:"merchant_name"`
	CustomerEmail string  `json:"customer_email"`
	CustomerPhone string  `json:"customer_phone,omitempty"`
	CustomerName  string  `json:"customer_name,omitempty"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentLink   string  `json:"payment_link"`
	FailureReason string  `json:"failure_reason,omitempty"`
	ActionType    string  `json:"action_type"`
}

// NotificationResult represents the outcome of sending a recovery notification.
type NotificationResult struct {
	MessageID string    `json:"message_id"`
	Provider  string    `json:"provider"`
	Status    string    `json:"status"` // SENT, FAILED, SKIPPED
	Recipient string    `json:"recipient"`
	SentAt    time.Time `json:"sent_at"`
	Error     string    `json:"error,omitempty"`
}

// NotificationProvider defines the interface for delivering recovery notifications.
type NotificationProvider interface {
	SendRecoveryNotification(ctx context.Context, req NotificationRequest) (*NotificationResult, error)
}
