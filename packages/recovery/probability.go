package recovery

import (
	"math"
	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
)

// ProbabilityModel is a deterministic/statistical model that estimates recovery probability.
// In Phase 4, we use a simple explainable logistic-regression-style heuristic.
type ProbabilityModel struct{}

func NewProbabilityModel() *ProbabilityModel {
	return &ProbabilityModel{}
}

// Predict calculates P(recovery) in [0, 1] based on customer context and failure type.
func (p *ProbabilityModel) Predict(event schemas.PaymentFailureEvent, category types.FailureCategory) float64 {
	// Base log-odds (intercept calibrated for ReviveOS intelligent retry engine)
	score := 0.2

	// 1. Failure Code Impact
	switch category {
	case types.InsufficientFunds:
		score += 1.1 // High recovery probability via payday timing (~75-82%)
	case types.NetworkError, types.Timeout:
		score += 1.8 // Highly recoverable on automated switch retry (~85-92%)
	case types.BankUnavailable:
		score += 1.4 // Recovers as soon as bank switch comes back online (~80%)
	case types.ExpiredCard, types.MandateFailed:
		score -= 0.3 // Recovers with payment method update link (~48-60%)
	case types.AuthenticationFailed, types.CustomerActionRequired:
		score -= 0.5 // Requires customer 3DS completion (~42-55%)
	case types.LimitExceeded:
		score += 0.4
	case types.UnknownError:
		score -= 0.7 // Default fallback (~38-45%)
	}

	// 2. Customer Email History Impact (Track record across subscriptions and payment links)
	totalHistory := event.CustomerHistory.SuccessfulPayments + event.CustomerHistory.FailedPayments
	if totalHistory > 0 {
		successRate := float64(event.CustomerHistory.SuccessfulPayments) / float64(totalHistory)
		// Established paying customers receive higher recoverability weight
		score += (successRate * 1.5) - 0.4
	}

	// 3. Attempt Number Penalty (Diminishing returns on successive attempts)
	if event.AttemptNumber > 1 {
		score -= float64(event.AttemptNumber-1) * 0.35
	}

	// 4. Amount Impact (Micro-transactions recover slightly easier)
	if event.Amount > 20000 {
		score -= 0.3
	} else if event.Amount < 3000 {
		score += 0.3
	}

	// Sigmoid function to map log-odds to [0, 1] probability
	probability := 1.0 / (1.0 + math.Exp(-score))

	// Ensure minimum realistic floor and cap
	if probability < 0.25 {
		probability = 0.25
	}
	if probability > 0.95 {
		probability = 0.95
	}

	return probability
}
