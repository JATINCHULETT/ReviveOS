package risk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type RiskScore struct {
	Probability  float64 `json:"probability"`
	RiskLevel    string  `json:"risk_level"`
	ModelVersion string  `json:"model_version"`
}

type RiskAnalysisRequest struct {
	EventType               string                 `json:"event_type"`
	PaymentID               string                 `json:"payment_id,omitempty"`
	MerchantID              string                 `json:"merchant_id,omitempty"`
	CustomerID              string                 `json:"customer_id,omitempty"`
	CustomerEmail           string                 `json:"customer_email,omitempty"`
	Amount                  float64                `json:"amount"`
	Currency                string                 `json:"currency"`
	FailureCode             string                 `json:"failure_code,omitempty"`
	AttemptNumber           int                    `json:"attempt_number"`
	CustomerFailedCount     int                    `json:"customer_failed_count"`
	CustomerSuccessCount    int                    `json:"customer_success_count"`
	CustomerPreviousReturns int                    `json:"customer_previous_returns"`
	Velocity1h              int                    `json:"velocity_1h"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
}

type RiskAnalysisResponse struct {
	EventType         string     `json:"event_type"`
	PaymentID         string     `json:"payment_id,omitempty"`
	Fraud             RiskScore  `json:"fraud"`
	ReturnRisk        *RiskScore `json:"return_risk,omitempty"`
	OverallRisk       string     `json:"overall_risk"`
	ExpectedLoss      float64    `json:"expected_loss"`
	RecommendedAction string     `json:"recommended_action"`
	Reason            string     `json:"reason"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("ML_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8001"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// AnalyzeRisk queries the ML Service or uses deterministic heuristics as fallback
func (c *Client) AnalyzeRisk(ctx context.Context, req RiskAnalysisRequest) (*RiskAnalysisResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return c.fallbackHeuristic(req), nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/risk/analyze", c.baseURL), bytes.NewReader(reqBody))
	if err != nil {
		return c.fallbackHeuristic(req), nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Graceful fallback to heuristic if ML service is not active
		return c.fallbackHeuristic(req), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallbackHeuristic(req), nil
	}

	var riskResp RiskAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&riskResp); err != nil {
		return c.fallbackHeuristic(req), nil
	}

	return &riskResp, nil
}

func (c *Client) fallbackHeuristic(req RiskAnalysisRequest) *RiskAnalysisResponse {
	// Deterministic fallback heuristic
	fraudScore := 0.08
	if req.Amount > 50000 {
		fraudScore += 0.40
	} else if req.Amount > 20000 {
		fraudScore += 0.20
	}
	if req.CustomerFailedCount > 3 {
		fraudScore += 0.30
	}
	if req.Velocity1h > 2 {
		fraudScore += 0.25
	}
	if req.AttemptNumber > 3 {
		fraudScore += 0.15
	}
	if fraudScore > 0.95 {
		fraudScore = 0.95
	}

	riskLevel := "LOW"
	if fraudScore >= 0.70 {
		riskLevel = "HIGH"
	} else if fraudScore >= 0.35 {
		riskLevel = "MEDIUM"
	}

	return &RiskAnalysisResponse{
		EventType: req.EventType,
		PaymentID: req.PaymentID,
		Fraud: RiskScore{
			Probability:  fraudScore,
			RiskLevel:    riskLevel,
			ModelVersion: "fallback-heuristic-v1.0",
		},
		OverallRisk:       riskLevel,
		ExpectedLoss:      req.Amount * fraudScore,
		RecommendedAction: "ALLOW_AUTONOMOUS_RECOVERY",
		Reason:            fmt.Sprintf("Heuristic fraud assessment score: %.1f%%", fraudScore*100),
	}
}
