package generator

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/reviveos/schemas"
)

// Generator produces synthetic payment failure events.
type Generator struct {
	rand *rand.Rand
}

func New() *Generator {
	return &Generator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateBatch creates a batch of synthetic events according to the specified distribution.
func (g *Generator) GenerateBatch(count int) []schemas.PaymentFailureEvent {
	var events []schemas.PaymentFailureEvent
	for i := 0; i < count; i++ {
		events = append(events, g.GenerateSingle(i+1))
	}
	return events
}

// GenerateSingle creates a single realistic failure event.
func (g *Generator) GenerateSingle(index int) schemas.PaymentFailureEvent {
	// Probability distribution from specification:
	// transient: 20%
	// insufficient_funds: 20%
	// expired/payment_method: 15%
	// authentication: 10%
	// mandate: 10%
	// network: 10%
	// customer_action: 10%
	// ambiguous: 5%

	r := g.rand.Float64()
	var failureCode string
	var groundTruth schemas.SyntheticGroundTruth

	switch {
	case r < 0.20:
		// Transient
		failureCode = "TRANSIENT_ERROR"
		groundTruth.WouldRecoverNaturally = g.rand.Float64() < 0.30
		groundTruth.WouldRecoverAfterRetry = true
	case r < 0.40:
		// Insufficient Funds
		failureCode = "INSUFFICIENT_FUNDS"
		groundTruth.WouldRecoverAfterRetry = g.rand.Float64() < 0.60
		groundTruth.WouldRecoverAfterCustomer = true
		groundTruth.OptimalRetryTimingOffsetHs = 24 // salary hits next day maybe
	case r < 0.55:
		// Expired / Method
		failureCode = "EXPIRED_CARD"
		groundTruth.WouldRecoverAfterCustomer = true
	case r < 0.65:
		// Authentication
		failureCode = "AUTHENTICATION_FAILED"
		groundTruth.WouldRecoverAfterCustomer = true
	case r < 0.75:
		// Mandate
		failureCode = "MANDATE_FAILED"
		groundTruth.WouldRecoverAfterCustomer = true
	case r < 0.85:
		// Network
		failureCode = "NETWORK_ERROR"
		groundTruth.WouldRecoverNaturally = g.rand.Float64() < 0.50
		groundTruth.WouldRecoverAfterRetry = true
	case r < 0.95:
		// Customer Action Required
		failureCode = "CUSTOMER_ACTION_REQUIRED"
		groundTruth.WouldRecoverAfterCustomer = true
	default:
		// Ambiguous
		failureCode = "UNKNOWN_ERROR"
		groundTruth.WouldNeverRecover = g.rand.Float64() < 0.50
	}

	if !groundTruth.WouldRecoverNaturally && !groundTruth.WouldRecoverAfterRetry && !groundTruth.WouldRecoverAfterCustomer {
		groundTruth.WouldNeverRecover = true
	}

	amount := float64(1000 + g.rand.Intn(9000))
	method := "card"
	if g.rand.Float64() > 0.6 {
		method = "upi"
	}

	event := schemas.PaymentFailureEvent{
		PaymentID:      fmt.Sprintf("pay_%d_%d", time.Now().Unix(), index),
		CustomerID:     fmt.Sprintf("cust_%d", g.rand.Intn(1000)),
		SubscriptionID: fmt.Sprintf("sub_%d", g.rand.Intn(100)),
		Amount:         amount,
		Currency:       "INR",
		PaymentMethod:  method,
		FailureCode:    failureCode,
		AttemptNumber:  1,
		Timestamp:      time.Now(),
		CustomerHistory: schemas.CustomerHistory{
			SuccessfulPayments: g.rand.Intn(12),
			FailedPayments:     g.rand.Intn(3),
		},
		SimulatedGroundTruth: &groundTruth,
	}

	return event
}
