package aiprovider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/reviveos/packages/recovery"
)

type OllamaProvider struct {
	URL    string
	Model  string
	Client *http.Client
}

func NewOllamaProvider(url, model string) *OllamaProvider {
	if url == "" {
		url = os.Getenv("OLLAMA_URL")
	}
	if url == "" {
		url = "http://localhost:11434"
	}
	if model == "" {
		model = os.Getenv("OLLAMA_MODEL")
	}
	if model == "" {
		model = "deepseek-r1:1.5b"
	}
	return &OllamaProvider{
		URL:   url,
		Model: model,
		Client: &http.Client{
			Timeout: 30 * time.Second, // Timeout for local LLM inference
		},
	}
}

var (
	thinkTagRegex   = regexp.MustCompile(`(?s)<think>.*?(?:</think>|$)`)
	numberRegex     = regexp.MustCompile(`\d+`)
	percentageRegex = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%`)
)

// CheckModelAvailable verifies whether Ollama is reachable and the configured model is installed.
func (o *OllamaProvider) CheckModelAvailable(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/tags", o.URL), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create tags request: %w", err)
	}

	resp, err := o.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("ollama unreachable at %s: %w", o.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("ollama returned status %d from /api/tags", resp.StatusCode)
	}

	var tagsResp struct {
		Models []struct {
			Name  string `json:"name"`
			Model string `json:"model"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return false, fmt.Errorf("failed to decode /api/tags response: %w", err)
	}

	for _, m := range tagsResp.Models {
		if strings.EqualFold(m.Name, o.Model) || strings.EqualFold(m.Model, o.Model) ||
			strings.HasPrefix(strings.ToLower(m.Name), strings.ToLower(o.Model)) {
			return true, nil
		}
	}

	return false, nil
}

// cleanJSON extracts and sanitizes raw JSON from LLM output.
func cleanJSON(raw string) string {
	cleaned := thinkTagRegex.ReplaceAllString(raw, "")
	cleaned = strings.TrimSpace(cleaned)

	// Remove markdown code blocks if wrapped
	if strings.Contains(cleaned, "```json") {
		parts := strings.Split(cleaned, "```json")
		if len(parts) > 1 {
			cleaned = strings.Split(parts[1], "```")[0]
		}
	} else if strings.Contains(cleaned, "```") {
		parts := strings.Split(cleaned, "```")
		if len(parts) > 1 {
			cleaned = parts[1]
		}
	}

	cleaned = strings.TrimSpace(cleaned)

	// Find outermost { and }
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start != -1 && end != -1 && end > start {
		cleaned = cleaned[start : end+1]
	}

	return cleaned
}

// RecommendStrategy queries the local Ollama instance for a recovery recommendation.
func (o *OllamaProvider) RecommendStrategy(ctx context.Context, input AIContext) (*AIRecommendation, error) {
	promptData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI context: %w", err)
	}

	userPrompt := fmt.Sprintf(`Analyze the failed payment context and output a JSON recovery recommendation.
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
		"model": o.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	bodyBytes, _ := json.Marshal(requestBody)

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/chat", o.URL), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return nil, fmt.Errorf("ollama request failed (%dms): %w", latencyMs, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d (%dms)", resp.StatusCode, latencyMs)
	}

	var chatResponse struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	rawContent := chatResponse.Message.Content
	cleanedOutput := cleanJSON(rawContent)

	// Flexible parsing map to handle string/numeric variations from LLM
	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedOutput), &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse structured JSON: %w (Raw: %s)", err, rawContent)
	}

	rec := &AIRecommendation{
		RawResponse:         rawContent,
		PromptHash:          promptHash,
		InferenceDurationMs: latencyMs,
		Provider:            "ollama",
		Model:               o.Model,
		IsFallback:          false,
	}

	// 1. Diagnosis
	if d, ok := rawMap["diagnosis"]; ok {
		rec.Diagnosis = fmt.Sprintf("%v", d)
	} else if errDiag, ok := rawMap["error_diagnosis"]; ok {
		rec.Diagnosis = fmt.Sprintf("%v", errDiag)
	} else {
		rec.Diagnosis = fmt.Sprintf("Payment failure classified as %s", input.DeterministicClass)
	}

	// 2. Reason
	if r, ok := rawMap["reason"]; ok {
		rec.Reasoning = fmt.Sprintf("%v", r)
	} else if r, ok := rawMap["reasoning"]; ok {
		rec.Reasoning = fmt.Sprintf("%v", r)
	} else if r, ok := rawMap["explanation"]; ok {
		rec.Reasoning = fmt.Sprintf("%v", r)
	}

	// 3. Confidence
	rec.Confidence = parseNumericField(rawMap, "confidence", 0.80)

	// 4. Recoverability
	rec.Recoverability = parseNumericField(rawMap, "recoverability", 0.70)

	// 5. Delay Hours
	rec.RecommendedDelayHours = parseDelayHours(rawMap)

	// 6. Recommended Action
	rec.RecommendedAction = parseAction(rawMap)

	return rec, nil
}

// SafeFallback returns a deterministic rule-based recommendation when LLM inference is unavailable.
func SafeFallback(input AIContext, err error) *AIRecommendation {
	nba := recovery.NewNBAEngine()
	candidates := nba.GenerateCandidates(input.FailureEvent, input.DeterministicClass, input.StatisticalProb)
	best := nba.SelectBestAction(candidates)

	reason := fmt.Sprintf("Safe rule-based fallback applied: %s", best.Reasoning)
	if err != nil {
		reason = fmt.Sprintf("AI_FALLBACK (%v): %s", err, best.Reasoning)
	}

	return &AIRecommendation{
		Diagnosis:             fmt.Sprintf("Deterministic classification: %s", input.DeterministicClass),
		Recoverability:        input.StatisticalProb,
		RecommendedAction:     best.Action,
		RecommendedDelayHours: best.DelayHours,
		Reasoning:             reason,
		Confidence:            input.StatisticalProb,
		RawResponse:           "",
		PromptHash:            "",
		InferenceDurationMs:   0,
		Provider:              "deterministic-fallback",
		Model:                 "rule-engine",
		IsFallback:            true,
	}
}

func parseNumericField(rawMap map[string]interface{}, key string, defaultVal float64) float64 {
	val, ok := rawMap[key]
	if !ok {
		return defaultVal
	}

	switch v := val.(type) {
	case float64:
		if v > 1.0 && v <= 100.0 {
			return v / 100.0
		}
		return v
	case int:
		if v > 1 && v <= 100 {
			return float64(v) / 100.0
		}
		return float64(v)
	case string:
		str := strings.TrimSpace(v)
		if matches := percentageRegex.FindStringSubmatch(str); len(matches) > 1 {
			if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
				return f / 100.0
			}
		}
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			if f > 1.0 && f <= 100.0 {
				return f / 100.0
			}
			return f
		}
		lower := strings.ToLower(str)
		if strings.Contains(lower, "high") || strings.Contains(lower, "likely") || strings.Contains(lower, "possible") {
			return 0.85
		} else if strings.Contains(lower, "medium") || strings.Contains(lower, "moderate") {
			return 0.70
		} else if strings.Contains(lower, "low") || strings.Contains(lower, "unlikely") {
			return 0.40
		}
	}

	return defaultVal
}

func parseDelayHours(rawMap map[string]interface{}) int {
	for _, key := range []string{"recommended_delay_hours", "delay_hours", "delay", "recommendedDelayHours"} {
		if val, ok := rawMap[key]; ok {
			switch v := val.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case string:
				if i, err := strconv.Atoi(v); err == nil {
					return i
				}
				if nums := numberRegex.FindString(v); nums != "" {
					if i, err := strconv.Atoi(nums); err == nil {
						return i
					}
				}
			}
		}
	}
	return 0
}

func parseAction(rawMap map[string]interface{}) recovery.ActionType {
	var actionStr string
	for _, key := range []string{"recommended_action", "recommendeded_action", "action", "recommendedAction"} {
		if val, ok := rawMap[key]; ok && val != nil {
			actionStr = strings.ToUpper(fmt.Sprintf("%v", val))
			break
		}
	}

	switch {
	case strings.Contains(actionStr, "IMMEDIATE"):
		return recovery.ActionImmediateRetry
	case strings.Contains(actionStr, "DELAY"):
		return recovery.ActionDelayedRetry
	case strings.Contains(actionStr, "METHOD") || strings.Contains(actionStr, "UPDATE"):
		return recovery.ActionPaymentUpdate
	case strings.Contains(actionStr, "LINK"):
		return recovery.ActionPaymentLink
	case strings.Contains(actionStr, "NOTIF"):
		return recovery.ActionCustomerNotify
	case strings.Contains(actionStr, "ESCALAT"):
		return recovery.ActionEscalation
	case strings.Contains(actionStr, "NO_ACTION"):
		return recovery.ActionNoAction
	default:
		return recovery.ActionDelayedRetry
	}
}
