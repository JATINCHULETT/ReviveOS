package notificationservice

import (
	"context"
	"fmt"
	"log"
	"time"
)

// LogOnlyProvider logs notifications to stdout without invoking external APIs (ideal for local development/testing).
type LogOnlyProvider struct{}

func NewLogOnlyProvider() *LogOnlyProvider {
	return &LogOnlyProvider{}
}

func (l *LogOnlyProvider) SendRecoveryNotification(ctx context.Context, req NotificationRequest) (*NotificationResult, error) {
	log.Printf("[NotificationService] DISPATCH_NOTIFICATION: Provider=LOG Recipient=%s Amount=%.2f Link=%s Action=%s",
		req.CustomerEmail, req.Amount, req.PaymentLink, req.ActionType)

	return &NotificationResult{
		MessageID: fmt.Sprintf("log_msg_%d", time.Now().UnixNano()),
		Provider:  "log",
		Status:    "SENT",
		Recipient: req.CustomerEmail,
		SentAt:    time.Now().UTC(),
	}, nil
}
