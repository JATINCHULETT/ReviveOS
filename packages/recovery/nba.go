package recovery

import (
	"time"
	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
)

// ActionType defines the possible recovery actions
type ActionType string

const (
	ActionImmediateRetry ActionType = "IMMEDIATE_RETRY"
	ActionDelayedRetry   ActionType = "DELAYED_RETRY"
	ActionPaymentUpdate  ActionType = "PAYMENT_METHOD_UPDATE"
	ActionPaymentLink    ActionType = "PAYMENT_LINK"
	ActionCustomerNotify ActionType = "CUSTOMER_NOTIFICATION"
	ActionEscalation     ActionType = "ESCALATION"
	ActionNoAction       ActionType = "NO_ACTION"
)

// CandidateIntervention represents a possible action with its score
type CandidateIntervention struct {
	Action      ActionType
	Score       float64
	DelayHours  int
	Reasoning   string
}

// NBAEngine (Next-Best-Action Engine) generates and ranks interventions.
type NBAEngine struct{}

func NewNBAEngine() *NBAEngine {
	return &NBAEngine{}
}

// GenerateCandidates yields potential actions based on the failure category and probability.
func (n *NBAEngine) GenerateCandidates(event schemas.PaymentFailureEvent, category types.FailureCategory, probability float64) []CandidateIntervention {
	var candidates []CandidateIntervention

	// Base logic based on the failure category
	switch category {
	case types.NetworkError, types.Timeout:
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionImmediateRetry,
			Score:      probability * 1.5,
			DelayHours: 0,
			Reasoning:  "Transient failure; immediate retry is highly effective.",
		})
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionDelayedRetry,
			Score:      probability * 1.0,
			DelayHours: 1,
			Reasoning:  "Delayed retry if immediate fails.",
		})

	case types.InsufficientFunds:
		// Dynamic timing based on typical salary/funding cycles (mocked logic)
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionDelayedRetry,
			Score:      probability * 1.2,
			DelayHours: 24, // T+1 day
			Reasoning:  "Insufficient funds; retrying next day.",
		})
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionCustomerNotify,
			Score:      probability * 0.8,
			DelayHours: 0,
			Reasoning:  "Notify customer immediately to top up balance.",
		})

	case types.ExpiredCard, types.MandateFailed, types.AuthenticationFailed:
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionPaymentUpdate,
			Score:      probability * 1.5,
			DelayHours: 0,
			Reasoning:  "Requires customer to update payment method.",
		})
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionPaymentLink,
			Score:      probability * 1.3,
			DelayHours: 0,
			Reasoning:  "Send alternative payment link.",
		})

	case types.LimitExceeded:
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionDelayedRetry,
			Score:      probability * 1.1,
			DelayHours: 24,
			Reasoning:  "Daily limit exceeded; retry tomorrow.",
		})

	default:
		// Ambiguous / Unknown
		candidates = append(candidates, CandidateIntervention{
			Action:     ActionEscalation,
			Score:      1.0,
			DelayHours: 0,
			Reasoning:  "Ambiguous failure requires human review.",
		})
	}

	// Always add a "No Action" baseline candidate
	candidates = append(candidates, CandidateIntervention{
		Action:     ActionNoAction,
		Score:      (1.0 - probability) * 0.5, // If probability is very low, No Action scores higher
		DelayHours: 0,
		Reasoning:  "Probability too low or cost too high to intervene.",
	})

	return candidates
}

// SelectBestAction ranks the candidates and picks the highest scoring one.
func (n *NBAEngine) SelectBestAction(candidates []CandidateIntervention) CandidateIntervention {
	if len(candidates) == 0 {
		return CandidateIntervention{Action: ActionNoAction, Score: 0, Reasoning: "No candidates generated"}
	}

	best := candidates[0]
	for _, c := range candidates {
		if c.Score > best.Score {
			best = c
		}
	}
	return best
}

// CalculateOptimalTiming provides dynamic timing logic
func (n *NBAEngine) CalculateOptimalTiming(action CandidateIntervention, event schemas.PaymentFailureEvent) time.Time {
	if action.DelayHours == 0 {
		return time.Now()
	}
	// For MVP, simply add the DelayHours
	// Future enhancement: Incorporate `time_of_day`, `day_of_month` from historical data
	return time.Now().Add(time.Duration(action.DelayHours) * time.Hour)
}
