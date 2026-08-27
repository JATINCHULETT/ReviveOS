package paymentprovider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RazorpayPaymentProvider implements PaymentProvider for the real Razorpay API.
type RazorpayPaymentProvider struct {
	KeyID      string
	KeySecret  string
	BaseURL    string
	HTTPClient *http.Client
}

// RazorpayPaymentResponse maps the standard Razorpay payment object.
type RazorpayPaymentResponse struct {
	ID               string `json:"id"`
	Entity           string `json:"entity"`
	Amount           int64  `json:"amount"` // in smallest currency sub-unit (paise for INR)
	Currency         string `json:"currency"`
	Status           string `json:"status"` // created, authorized, captured, refunded, failed
	Method           string `json:"method"`
	Captured         bool   `json:"captured"`
	ErrorCode        string `json:"error_code,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorSource      string `json:"error_source,omitempty"`
	ErrorStep        string `json:"error_step,omitempty"`
	ErrorReason      string `json:"error_reason,omitempty"`
	CreatedAt        int64  `json:"created_at"`
}

// RazorpayErrorResponse maps standard Razorpay error objects.
type RazorpayErrorResponse struct {
	Error struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Source      string `json:"source,omitempty"`
		Step        string `json:"step,omitempty"`
		Reason      string `json:"reason,omitempty"`
	} `json:"error"`
}

// NewRazorpayPaymentProvider initializes a new Razorpay payment provider instance.
func NewRazorpayPaymentProvider(keyID, keySecret, baseURL string) *RazorpayPaymentProvider {
	if keyID == "" {
		keyID = os.Getenv("RAZORPAY_KEY_ID")
	}
	if keySecret == "" {
		keySecret = os.Getenv("RAZORPAY_KEY_SECRET")
	}
	if baseURL == "" {
		baseURL = os.Getenv("RAZORPAY_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://api.razorpay.com/v1"
	}

	return &RazorpayPaymentProvider{
		KeyID:     keyID,
		KeySecret: keySecret,
		BaseURL:   strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ValidateCredentials checks if the configured Razorpay API credentials are non-empty.
func (r *RazorpayPaymentProvider) ValidateCredentials() error {
	if r.KeyID == "" {
		return errors.New("missing RAZORPAY_KEY_ID")
	}
	if r.KeySecret == "" {
		return errors.New("missing RAZORPAY_KEY_SECRET")
	}
	return nil
}

// GetPayment retrieves the authoritative status of a payment directly from Razorpay.
func (r *RazorpayPaymentProvider) GetPayment(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	if err := r.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("razorpay credentials error: %w", err)
	}

	url := fmt.Sprintf("%s/payments/%s", r.BaseURL, paymentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(r.KeyID, r.KeySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("razorpay request timeout for payment %s", paymentID)
		}
		return nil, fmt.Errorf("razorpay HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read razorpay response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp RazorpayErrorResponse
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Error.Code != "" {
			return nil, fmt.Errorf("razorpay API error (status %d): code=%s, desc=%s, reason=%s",
				resp.StatusCode, errResp.Error.Code, errResp.Error.Description, errResp.Error.Reason)
		}
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("razorpay authentication failed (401): invalid credentials")
		case http.StatusNotFound:
			return nil, fmt.Errorf("razorpay payment not found (404): %s", paymentID)
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("razorpay rate limit exceeded (429)")
		default:
			return nil, fmt.Errorf("razorpay returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	var pResp RazorpayPaymentResponse
	if err := json.Unmarshal(bodyBytes, &pResp); err != nil {
		return nil, fmt.Errorf("malformed razorpay response: %w", err)
	}

	statusUpper := strings.ToUpper(pResp.Status)
	isCaptured := statusUpper == "CAPTURED" || pResp.Captured

	amountFloat := float64(pResp.Amount) / 100.0 // Razorpay amounts are in paise

	return &PaymentStatus{
		PaymentID:         pResp.ID,
		ProviderPaymentID: pResp.ID,
		Status:            statusUpper,
		Amount:            amountFloat,
		Currency:          pResp.Currency,
		Method:            pResp.Method,
		FailureCode:       pResp.ErrorCode,
		FailureReason:     pResp.ErrorDescription,
		Captured:          isCaptured,
		UpdatedAt:         time.Now().UTC(),
	}, nil
}

// VerifyPayment checks and verifies the authoritative payment state from Razorpay.
func (r *RazorpayPaymentProvider) VerifyPayment(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	return r.GetPayment(ctx, paymentID)
}

// CreateRetryAttempt creates a payment link / retry attempt via Razorpay API.
func (r *RazorpayPaymentProvider) CreateRetryAttempt(ctx context.Context, paymentID string, amount float64) (*RetryResult, error) {
	if err := r.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("razorpay credentials error: %w", err)
	}

	// Amount in smallest unit (paise)
	amountPaise := int64(amount * 100)

	payload := map[string]interface{}{
		"amount":       amountPaise,
		"currency":     "INR",
		"description":  fmt.Sprintf("Payment recovery for %s", paymentID),
		"reference_id": fmt.Sprintf("rec_%s_%d", paymentID, time.Now().Unix()),
		"notify": map[string]bool{
			"sms":   true,
			"email": true,
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/payment_links", r.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create retry request: %w", err)
	}

	req.SetBasicAuth(r.KeyID, r.KeySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("razorpay retry request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read retry response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp RazorpayErrorResponse
		if jsonErr := json.Unmarshal(respBytes, &errResp); jsonErr == nil && errResp.Error.Code != "" {
			return &RetryResult{
				PaymentID:    paymentID,
				Status:       "FAILED",
				Amount:       amount,
				ErrorMessage: errResp.Error.Description,
				CreatedAt:    time.Now().UTC(),
			}, fmt.Errorf("razorpay retry API error (%d): %s", resp.StatusCode, errResp.Error.Description)
		}
		return &RetryResult{
			PaymentID:    paymentID,
			Status:       "FAILED",
			Amount:       amount,
			ErrorMessage: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBytes)),
			CreatedAt:    time.Now().UTC(),
		}, fmt.Errorf("razorpay returned status %d", resp.StatusCode)
	}

	var linkResp struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		ShortURL  string `json:"short_url"`
		CreatedAt int64  `json:"created_at"`
	}

	if err := json.Unmarshal(respBytes, &linkResp); err != nil {
		return nil, fmt.Errorf("malformed razorpay retry response: %w", err)
	}

	return &RetryResult{
		AttemptID:         linkResp.ID,
		PaymentID:         paymentID,
		ProviderPaymentID: linkResp.ID,
		Status:            "SUCCESS",
		Amount:            amount,
		PaymentLinkURL:    linkResp.ShortURL,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// VerifyWebhookSignature verifies the X-Razorpay-Signature using HMAC-SHA256.
func VerifyWebhookSignature(body []byte, signature, secret string) bool {
	if secret == "" || signature == "" || len(body) == 0 {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}
