package aiprovider

import (
	"context"
	"os"

	"github.com/reviveos/packages/recovery"
	"github.com/reviveos/packages/types"
	"github.com/reviveos/schemas"
)

// AIRecommendation represents the structured output required from the LLM.
type AIRecommendation struct {
	Diagnosis             string              `json:"diagnosis"`
	Recoverability        float64             `json:"recoverability"`
	RecommendedAction     recovery.ActionType `json:"recommended_action"`
	RecommendedDelayHours int                 `json:"recommended_delay_hours"`
	Reasoning             string              `json:"reason"`
	Confidence            float64             `json:"confidence"`

	// Execution & Audit Metadata
	RawResponse         string `json:"raw_response,omitempty"`
	PromptHash          string `json:"prompt_hash,omitempty"`
	InferenceDurationMs int    `json:"inference_duration_ms,omitempty"`
	Provider            string `json:"provider,omitempty"`
	Model               string `json:"model,omitempty"`
	IsFallback          bool   `json:"is_fallback,omitempty"`
}

// AIContext is the structured input passed to the LLM.
type AIContext struct {
	FailureEvent       schemas.PaymentFailureEvent `json:"failure_event"`
	DeterministicClass types.FailureCategory       `json:"deterministic_class"`
	StatisticalProb    float64                     `json:"statistical_probability"`
	EmpiricalStats     map[string]float64          `json:"empirical_stats"`
	FraudRiskScore     float64                     `json:"fraud_risk_score,omitempty"`
	FraudRiskLevel     string                      `json:"fraud_risk_level,omitempty"`
	ReturnRiskScore    float64                     `json:"return_risk_score,omitempty"`
}

// Provider defines the interface for interacting with any AI model provider.
type Provider interface {
	RecommendStrategy(ctx context.Context, input AIContext) (*AIRecommendation, error)
	CheckModelAvailable(ctx context.Context) (bool, error)
}

// NewAIProvider creates the active AI provider based on environment configuration.
func NewAIProvider() Provider {
	// If NVIDIA API Key is provided, prioritize NVIDIA NIM
	if os.Getenv("NVIDIA_API_KEY") != "" {
		return NewNvidiaNIMProvider(
			os.Getenv("NVIDIA_BASE_URL"),
			os.Getenv("NVIDIA_API_KEY"),
			os.Getenv("NVIDIA_MODEL"),
		)
	}

	// Fallback to Ollama
	return NewOllamaProvider(
		os.Getenv("OLLAMA_URL"),
		os.Getenv("OLLAMA_MODEL"),
	)
}
