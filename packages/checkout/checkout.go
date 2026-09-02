package checkout

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// SessionStatus represents stage of checkout
type SessionStatus string

const (
	StatusActive              SessionStatus = "ACTIVE"
	StatusDroppedOff          SessionStatus = "DROPPED_OFF"
	StatusRecoveryDispatched  SessionStatus = "RECOVERY_DISPATCHED"
	StatusRecovered           SessionStatus = "RECOVERED"
	StatusExpired             SessionStatus = "EXPIRED"
)

// Session represents an e-commerce or checkout flow attempt
type Session struct {
	ID                   string        `json:"id"`
	MerchantID           string        `json:"merchant_id"`
	SessionToken         string        `json:"session_token"`
	CustomerName         string        `json:"customer_name"`
	CustomerEmail        string        `json:"customer_email"`
	CustomerPhone        string        `json:"customer_phone,omitempty"`
	CartAmount           float64       `json:"cart_amount"`
	Currency             string        `json:"currency"`
	CartItemsJSON        string        `json:"cart_items_json"`
	StepReached          string        `json:"step_reached"` // CART_LOADED, DETAILS_ENTERED, PAYMENT_STEP, 3DS_INITIATED, ABANDONED
	Status               SessionStatus `json:"status"`
	DropOffReason        string        `json:"drop_off_reason,omitempty"`
	RecoveryLink         string        `json:"recovery_link,omitempty"`
	RecoveryDispatchedAt *time.Time    `json:"recovery_dispatched_at,omitempty"`
	RecoveredAt          *time.Time    `json:"recovered_at,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

// GenerateSessionToken creates a cryptographically secure token
func GenerateSessionToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// IsDroppedOff determines if a session is idle past the threshold
func IsDroppedOff(lastUpdated time.Time, timeoutMinutes int) bool {
	if timeoutMinutes <= 0 {
		timeoutMinutes = 15
	}
	return time.Since(lastUpdated) > time.Duration(timeoutMinutes)*time.Minute
}

// GenerateRecoveryLink generates a 1-click restored cart checkout URL
func GenerateRecoveryLink(baseURL, sessionToken string) string {
	if baseURL == "" {
		baseURL = "https://reviveos.onrender.com"
	}
	return fmt.Sprintf("%s/checkout/restore?token=%s", baseURL, sessionToken)
}
