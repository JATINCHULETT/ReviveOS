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
	// Base log-odds (intercept)
	score := -0.5 

	// 1. Failure Code Impact
	switch category {
	case types.InsufficientFunds:
		score += 1.2
	case types.NetworkError, types.Timeout:
		score += 2.0 // highly recoverable
	case types.BankUnavailable:
		score += 1.5
	case types.ExpiredCard, types.MandateFailed:
		score -= 1.0 // requires customer action
	case types.AuthenticationFailed, types.CustomerActionRequired:
		score -= 1.5
	case types.LimitExceeded:
		score += 0.5
	case types.UnknownError:
		score -= 2.0
	}

	// 2. Customer History Impact
	totalHistory := event.CustomerHistory.SuccessfulPayments + event.CustomerHistory.FailedPayments
	if totalHistory > 0 {
		successRate := float64(event.CustomerHistory.SuccessfulPayments) / float64(totalHistory)
		// Strong historical success increases probability
		score += (successRate * 2.0) - 1.0 
	} else {
		// New customer penalty
		score -= 0.5
	}

	// 3. Attempt Number Penalty (Diminishing returns on retries)
	if event.AttemptNumber > 1 {
		score -= float64(event.AttemptNumber) * 0.5
	}

	// 4. Amount Impact (Higher amounts have slightly lower natural recovery probability)
	if event.Amount > 10000 { // e.g. > ₹10,000
		score -= 0.5
	} else if event.Amount < 1000 {
		score += 0.5
	}

	// Sigmoid function to map log-odds to [0, 1] probability
	probability := 1.0 / (1.0 + math.Exp(-score))

	return probability
}
