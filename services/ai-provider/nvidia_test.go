package aiprovider

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
)

func TestNvidiaNIMProvider_Availability(t *testing.T) {
	apiKey := os.Getenv("NVIDIA_API_KEY")
	if apiKey == "" {
		apiKey = "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"
	}

	provider := NewNvidiaNIMProvider("https://integrate.api.nvidia.com/v1", apiKey, "deepseek-ai/deepseek-v4-flash-0731")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	available, err := provider.CheckModelAvailable(ctx)
	if err != nil {
		t.Fatalf("NVIDIA NIM availability check failed: %v", err)
	}

	if !available {
		t.Fatalf("Expected NVIDIA NIM model to be available")
	}

	t.Logf("NVIDIA NIM API is reachable and API key is valid")
}

func TestNvidiaNIMProvider_RealInference(t *testing.T) {
	apiKey := os.Getenv("NVIDIA_API_KEY")
	if apiKey == "" {
		apiKey = "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"
	}

	provider := NewNvidiaNIMProvider("https://integrate.api.nvidia.com/v1", apiKey, "deepseek-ai/deepseek-v4-flash-0731")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := AIContext{
		FailureEvent: schemas.PaymentFailureEvent{
			PaymentID:     "pay_test_nim_123",
			CustomerID:    "cust_test_123",
			Amount:        450000,
			Currency:      "INR",
			PaymentMethod: "card",
			FailureCode:   "insufficient_funds",
			Timestamp:     time.Now(),
		},
		DeterministicClass: types.InsufficientFunds,
		StatisticalProb:    0.72,
		EmpiricalStats: map[string]float64{
			"historical_success_rate": 0.80,
		},
	}

	rec, err := provider.RecommendStrategy(ctx, input)
	if err != nil {
		t.Fatalf("NVIDIA NIM inference failed: %v", err)
	}

	if rec == nil {
		t.Fatal("Expected non-nil recommendation")
	}

	if rec.RecommendedAction == "" {
		t.Fatal("Expected non-empty recommended_action")
	}

	if rec.Confidence <= 0 || rec.Confidence > 1.0 {
		t.Fatalf("Expected confidence in (0, 1.0], got %f", rec.Confidence)
	}

	if rec.Provider != "nvidia-nim" {
		t.Fatalf("Expected provider 'nvidia-nim', got '%s'", rec.Provider)
	}

	if rec.Model != "deepseek-ai/deepseek-v4-flash-0731" {
		t.Fatalf("Expected model 'deepseek-ai/deepseek-v4-flash-0731', got '%s'", rec.Model)
	}

	t.Logf("NVIDIA NIM DeepSeek Inference SUCCESS -> Action: %s, Conf: %.2f, Latency: %dms, Diagnosis: %s, Reasoning: %s",
		rec.RecommendedAction, rec.Confidence, rec.InferenceDurationMs, rec.Diagnosis, rec.Reasoning)
}
