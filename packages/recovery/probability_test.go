package recovery

import (
	"testing"

	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
)

func TestProbabilityModelPredict(t *testing.T) {
	model := NewProbabilityModel()

	event := schemas.PaymentFailureEvent{
		Amount: 1500,
		CustomerHistory: schemas.CustomerHistory{
			SuccessfulPayments: 5,
			FailedPayments:     1,
		},
		AttemptNumber: 1,
	}

	prob := model.Predict(event, types.InsufficientFunds)
	if prob <= 0 || prob >= 1 {
		t.Fatalf("expected probability between 0 and 1, got %f", prob)
	}

	if prob < 0.5 {
		t.Fatalf("expected high probability for customer with good history and insufficient funds, got %f", prob)
	}
}
