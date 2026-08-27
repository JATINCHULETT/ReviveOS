package aiprovider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/reviveos/packages/recovery"
)

// NvidiaNIMProvider implements the Provider interface using NVIDIA NIM Cloud Endpoints.
type NvidiaNIMProvider struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewNvidiaNIMProvider creates a new instance of NvidiaNIMProvider.
func NewNvidiaNIMProvider(baseURL, apiKey, model string) *NvidiaNIMProvider {
	if baseURL == "" {
		baseURL = os.Getenv("NVIDIA_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://integrate.api.nvidia.com/v1"
	}
	if apiKey == "" {
		apiKey = os.Getenv("NVIDIA_API_KEY")
	}
	if model == "" {
		model = os.Getenv("NVIDIA_MODEL")
	}
	if model == "" {
		model = "deepseek-ai/deepseek-v4-flash-0731"
	}

	return &NvidiaNIMProvider{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CheckModelAvailable verifies connectivity and API key validity against NVIDIA NIM.
func (n *NvidiaNIMProvider) CheckModelAvailable(ctx context.Context) (bool, error) {
	if n.APIKey == "" {
		return false, fmt.Errorf("NVIDIA_API_KEY is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", n.BaseURL+"/models", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+n.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := n.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to reach NVIDIA NIM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, fmt.Errorf("invalid or unauthorized NVIDIA_API_KEY (HTTP %d)", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("NVIDIA NIM API returned HTTP status %d", resp.StatusCode)
	}

	return true, nil
}

// RecommendStrategy queries NVIDIA NIM DeepSeek model for an optimal recovery strategy.
func (n *NvidiaNIMProvider) RecommendStrategy(ctx context.Context, input AIContext) (*AIRecommendation, error) {
	promptData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI context: %w", err)
	}

	userPrompt := fmt.Sprintf(`Analyze the failed payment context and output an optimal recovery recommendation in JSON.
Context:
%s

Respond with ONLY valid JSON with keys:
{
  "diagnosis": "brief explanation",
  "recoverability": 0.8,
  "recommended_action": "DELAYED_RETRY",
  "recommended_delay_hours": 24,
  "reason": "brief reason",
  "confidence": 0.85
}
Allowed recommended_action values: IMMEDIATE_RETRY, DELAYED_RETRY, PAYMENT_METHOD_UPDATE, PAYMENT_LINK, CUSTOMER_NOTIFICATION, ESCALATION, NO_ACTION.`, string(promptData))

	promptHashBytes := sha256.Sum256([]byte(userPrompt))
	promptHash := hex.EncodeToString(promptHashBytes[:])

	requestBody := map[string]interface{}{
		"model": n.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are ReviveOS Payment Recovery Intelligence. You analyze payment failures and produce high-confidence recovery actions. Always return strictly valid JSON.",
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"temperature": 0.2,
		"top_p":       0.95,
		"max_tokens":  4096,
		"stream":      false,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "POST", n.BaseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.APIKey)

	resp, err := n.Client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return nil, fmt.Errorf("nvidia nim request failed (%dms): %w", latencyMs, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read nvidia response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nvidia nim returned status %d (%dms): %s", resp.StatusCode, latencyMs, string(respBody))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Role             string `json:"role"`
				Content          string `json:"content"`
				Reasoning        string `json:"reasoning,omitempty"`
				ReasoningContent string `json:"reasoning_content,omitempty"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode nvidia nim response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned by nvidia nim")
	}

	rawContent := chatResp.Choices[0].Message.Content
	reasoningContent := chatResp.Choices[0].Message.Reasoning
	if reasoningContent == "" {
		reasoningContent = chatResp.Choices[0].Message.ReasoningContent
	}

	cleanedJSON := CleanJSON(rawContent)

	var parsed struct {
		Diagnosis             string      `json:"diagnosis"`
		Recoverability        interface{} `json:"recoverability"`
		RecommendedAction     string      `json:"recommended_action"`
		RecommendedDelayHours interface{} `json:"recommended_delay_hours"`
		Reason                string      `json:"reason"`
		Confidence            interface{} `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(cleanedJSON), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse structured JSON: %w (Raw: %s)", err, rawContent)
	}

	// Normalizing numeric values
	recoverability := toFloat(parsed.Recoverability, 0.5)
	confidence := toFloat(parsed.Confidence, 0.8)
	delayHours := toInt(parsed.RecommendedDelayHours, 0)

	normAction := normalizeAction(parsed.RecommendedAction)
	reasonText := parsed.Reason
	if reasonText == "" && reasoningContent != "" {
		reasonText = reasoningContent
	}

	return &AIRecommendation{
		Diagnosis:             parsed.Diagnosis,
		Recoverability:        recoverability,
		RecommendedAction:     normAction,
		RecommendedDelayHours: delayHours,
		Reasoning:             reasonText,
		Confidence:            confidence,
		RawResponse:           rawContent,
		PromptHash:            promptHash,
		InferenceDurationMs:   latencyMs,
		Provider:              "nvidia-nim",
		Model:                 n.Model,
		IsFallback:            false,
	}, nil
}

func toFloat(val interface{}, def float64) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func toInt(val interface{}, def int) int {
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func normalizeAction(actionStr string) recovery.ActionType {
	upper := strings.ToUpper(strings.TrimSpace(actionStr))
	switch recovery.ActionType(upper) {
	case recovery.ActionImmediateRetry:
		return recovery.ActionImmediateRetry
	case recovery.ActionDelayedRetry:
		return recovery.ActionDelayedRetry
	case recovery.ActionPaymentUpdate:
		return recovery.ActionPaymentUpdate
	case recovery.ActionPaymentLink:
		return recovery.ActionPaymentLink
	case recovery.ActionCustomerNotify:
		return recovery.ActionCustomerNotify
	case recovery.ActionEscalation:
		return recovery.ActionEscalation
	case recovery.ActionNoAction:
		return recovery.ActionNoAction
	default:
		if strings.Contains(upper, "RETRY") {
			return recovery.ActionDelayedRetry
		}
		if strings.Contains(upper, "LINK") {
			return recovery.ActionPaymentLink
		}
		if strings.Contains(upper, "UPDATE") || strings.Contains(upper, "CARD") {
			return recovery.ActionPaymentUpdate
		}
		return recovery.ActionDelayedRetry
	}
}
