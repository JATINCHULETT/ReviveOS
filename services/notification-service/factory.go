package notificationservice

import (
	"os"
	"strings"
)

// NewNotificationProvider creates a configured NotificationProvider based on environment variables.
func NewNotificationProvider(providerType string) NotificationProvider {
	if providerType == "" {
		providerType = os.Getenv("NOTIFICATION_PROVIDER")
	}

	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case "resend":
		return NewResendProvider("", "")
	case "log", "mock", "local":
		return NewLogOnlyProvider()
	default:
		// If RESEND_API_KEY is present, default to Resend, otherwise safe Log provider
		if os.Getenv("RESEND_API_KEY") != "" {
			return NewResendProvider("", "")
		}
		return NewLogOnlyProvider()
	}
}
