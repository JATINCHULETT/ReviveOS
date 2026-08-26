package aiprovider

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
)

func TestOllamaProvider_Availability(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	provider := NewOllamaProvider("http://localhost:11434", "deepseek-r1:1.5b")
	available, err := provider.CheckModelAvailable(ctx)
	if err != nil {
		t.Skipf("Ollama not reachable: %v", err)
	}

	if !available {
		t.Fatalf("Expected model deepseek-r1:1.5b to be available in local Ollama instance")
	}

	t.Logf("Ollama is reachable and deepseek-r1:1.5b is AVAILABLE")
}

func TestOllamaProvider_RealInference(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	provider := NewOllamaProvider("http://localhost:11434", "deepseek-r1:1.5b")
	available, err := provider.CheckModelAvailable(ctx)
	if err != nil || !available {
		t.Skipf("Skipping real inference test: Ollama / model not available (%v)", err)
	}

	input := AIContext{
		FailureEvent: schemas.PaymentFailureEvent{
			Amount:        2499.00,
			Currency:      "INR",
			PaymentMethod: "card",
			FailureCode:   "INSUFFICIENT_FUNDS",
			AttemptNumber: 1,
			CustomerHistory: schemas.CustomerHistory{
				SuccessfulPayments: 5,
				FailedPayments:     0,
			},
		},
		DeterministicClass: types.InsufficientFunds,
		StatisticalProb:    0.8455,
		EmpiricalStats: map[string]float64{
			"category_success_rate": 0.72,
			"customer_success_rate": 1.0,
		},
	}

	rec, err := provider.RecommendStrategy(ctx, input)
	if err != nil {
		t.Fatalf("RecommendStrategy failed: %v", err)
	}

	if rec.RawResponse == "" {
		t.Errorf("Expected non-empty RawResponse")
	}
	if rec.PromptHash == "" {
		t.Errorf("Expected non-empty PromptHash")
	}
	if rec.InferenceDurationMs <= 0 {
		t.Errorf("Expected positive InferenceDurationMs, got %d", rec.InferenceDurationMs)
	}
	if rec.Provider != "ollama" {
		t.Errorf("Expected provider 'ollama', got '%s'", rec.Provider)
	}
	if rec.Model != "deepseek-r1:1.5b" {
		t.Errorf("Expected model 'deepseek-r1:1.5b', got '%s'", rec.Model)
	}
	if rec.Confidence <= 0 || rec.Confidence > 1.0 {
		t.Errorf("Expected confidence in (0, 1], got %f", rec.Confidence)
	}
	if rec.IsFallback {
		t.Errorf("Expected real inference, but IsFallback is true")
	}

	t.Logf("Real Ollama Inference SUCCESS -> Action: %s, Conf: %.2f, Latency: %dms, Diagnosis: %s",
		rec.RecommendedAction, rec.Confidence, rec.InferenceDurationMs, rec.Diagnosis)
}

func TestOllamaProvider_FallbackOnUnavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Provider pointing to non-existent port
	badProvider := NewOllamaProvider("http://localhost:59999", "deepseek-r1:1.5b")

	input := AIContext{
		FailureEvent: schemas.PaymentFailureEvent{
			Amount:        1500.00,
			Currency:      "INR",
			PaymentMethod: "card",
			FailureCode:   "INSUFFICIENT_FUNDS",
			AttemptNumber: 1,
			CustomerHistory: schemas.CustomerHistory{
				SuccessfulPayments: 3,
				FailedPayments:     1,
			},
		},
		DeterministicClass: types.InsufficientFunds,
		StatisticalProb:    0.75,
	}

	_, err := badProvider.RecommendStrategy(ctx, input)
	if err == nil {
		t.Fatalf("Expected error when calling unreachable Ollama, got nil")
	}

	// Test SafeFallback
	fallbackRec := SafeFallback(input, err)
	if !fallbackRec.IsFallback {
		t.Errorf("Expected fallback recommendation to have IsFallback=true")
	}
	if fallbackRec.Provider != "deterministic-fallback" {
		t.Errorf("Expected provider 'deterministic-fallback', got '%s'", fallbackRec.Provider)
	}
	if fallbackRec.RecommendedAction == "" {
		t.Errorf("Expected fallback recommendation to have valid RecommendedAction")
	}

	t.Logf("Safe fallback verified: Action=%s, Provider=%s, Reasoning=%s",
		fallbackRec.RecommendedAction, fallbackRec.Provider, fallbackRec.Reasoning)
}

func TestCleanJSON(t *testing.T) {
	inputWithThink := "<think>\nThinking process...\nRecommended action is DELAYED_RETRY.\n</think>\n```json\n{\n  \"diagnosis\": \"Soft card decline due to timeout\",\n  \"recoverability\": 0.85,\n  \"recommended_action\": \"DELAYED_RETRY\",\n  \"recommended_delay_hours\": 12,\n  \"reason\": \"Temporary network timeout\",\n  \"confidence\": 0.90\n}\n```"

	cleaned := cleanJSON(inputWithThink)
	expectedStart := "{"
	expectedEnd := "}"

	if len(cleaned) == 0 || !strings.HasPrefix(cleaned, expectedStart) || !strings.HasSuffix(cleaned, expectedEnd) {
		t.Fatalf("cleanJSON failed: result was '%s'", cleaned)
	}
}
