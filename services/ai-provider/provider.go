package aiprovider

import (
	"context"

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
}

// Provider defines the interface for interacting with any AI model provider.
type Provider interface {
	RecommendStrategy(ctx context.Context, input AIContext) (*AIRecommendation, error)
	CheckModelAvailable(ctx context.Context) (bool, error)
}
