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

	"github.com/jackc/pgx/v5/pgxpool"
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

// CreateRetryAttempt creates a payment link / retry attempt via Razorpay API without explicit customer metadata.
func (r *RazorpayPaymentProvider) CreateRetryAttempt(ctx context.Context, paymentID string, amount float64) (*RetryResult, error) {
	return r.CreateRetryAttemptWithCustomer(ctx, paymentID, amount, "", "", "")
}

// CreateRetryAttemptWithCustomer creates a payment link populated with customer email, phone, and name for Razorpay notifications.
func (r *RazorpayPaymentProvider) CreateRetryAttemptWithCustomer(ctx context.Context, paymentID string, amount float64, customerEmail, customerPhone, customerName string) (*RetryResult, error) {
	if err := r.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("razorpay credentials error: %w", err)
	}

	// Amount in smallest unit (paise)
	amountPaise := int64(amount * 100)

	notesMap := map[string]string{
		"payment_id":      paymentID,
		"recovery_origin": "reviveos",
	}
	if customerEmail != "" {
		notesMap["customer_email"] = customerEmail
	}
	if customerPhone != "" {
		notesMap["customer_phone"] = customerPhone
	}

	refID := fmt.Sprintf("rec_%d", time.Now().UnixNano())
	if len(refID) > 30 {
		refID = refID[:30]
	}

	payload := map[string]interface{}{
		"amount":       amountPaise,
		"currency":     "INR",
		"description":  fmt.Sprintf("Payment recovery for %s", paymentID),
		"reference_id": refID,
		"notify": map[string]bool{
			"sms":   customerPhone != "",
			"email": customerEmail != "",
		},
		"notes": notesMap,
	}

	if customerEmail != "" || customerPhone != "" || customerName != "" {
		custMap := map[string]string{}
		if customerEmail != "" {
			custMap["email"] = customerEmail
		}
		if customerPhone != "" {
			custMap["contact"] = customerPhone
		}
		if customerName != "" {
			custMap["name"] = customerName
		} else if customerEmail != "" {
			custMap["name"] = strings.Split(customerEmail, "@")[0]
		}
		payload["customer"] = custMap
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

// RazorpayPaymentLinkItem models a payment link object from Razorpay's API
type RazorpayPaymentLinkItem struct {
	ID          string                 `json:"id"`
	Amount      int64                  `json:"amount"` // in paise
	Currency    string                 `json:"currency"`
	Status      string                 `json:"status"` // created, paid, cancelled, expired
	Description string                 `json:"description"`
	ShortURL    string                 `json:"short_url"`
	Customer    map[string]interface{} `json:"customer"`
	Notes       map[string]interface{} `json:"notes"`
	CreatedAt   int64                  `json:"created_at"`
}

// FetchPaymentLinks fetches payment links directly from Razorpay REST API
func (r *RazorpayPaymentProvider) FetchPaymentLinks(ctx context.Context, count int) ([]RazorpayPaymentLinkItem, error) {
	if err := r.ValidateCredentials(); err != nil {
		return nil, err
	}
	if count <= 0 {
		count = 50
	}
	url := fmt.Sprintf("%s/payment_links?count=%d", r.BaseURL, count)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(r.KeyID, r.KeySecret)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("razorpay returned status %d", resp.StatusCode)
	}

	var result struct {
		Count        int                       `json:"count"`
		PaymentLinks []RazorpayPaymentLinkItem `json:"payment_links"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.PaymentLinks, nil
}

// SyncRazorpayPaymentLinks pulls recent payment links from Razorpay and ensures they are recorded in ReviveOS
func SyncRazorpayPaymentLinks(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}
	keyID := strings.TrimSpace(os.Getenv("RAZORPAY_KEY_ID"))
	keySecret := strings.TrimSpace(os.Getenv("RAZORPAY_KEY_SECRET"))
	baseURL := strings.TrimSpace(os.Getenv("RAZORPAY_BASE_URL"))
	if keyID == "" || keySecret == "" {
		return nil
	}

	rzp := NewRazorpayPaymentProvider(keyID, keySecret, baseURL)
	links, err := rzp.FetchPaymentLinks(ctx, 50)
	if err != nil {
		return err
	}

	var merchantID string
	_ = pool.QueryRow(ctx, "SELECT id::text FROM merchants ORDER BY created_at ASC LIMIT 1").Scan(&merchantID)
	if merchantID == "" {
		_ = pool.QueryRow(ctx, "INSERT INTO merchants (name) VALUES ('Default Merchant') RETURNING id::text").Scan(&merchantID)
	}

	for _, link := range links {
		if link.ID == "" {
			continue
		}
		amountFloat := float64(link.Amount) / 100.0
		currency := link.Currency
		if currency == "" {
			currency = "INR"
		}

		var custEmail, custPhone string
		if em, ok := link.Customer["email"].(string); ok {
			custEmail = strings.TrimSpace(em)
		}
		if cn, ok := link.Customer["contact"].(string); ok {
			custPhone = strings.TrimSpace(cn)
		}
		if custEmail == "" && link.Notes != nil {
			if em, ok := link.Notes["customer_email"].(string); ok {
				custEmail = strings.TrimSpace(em)
			}
		}
		if custEmail == "" {
			custEmail = fmt.Sprintf("cust_%s@revive-os.me", link.ID)
		}

		// 1. Customer
		var customerID string
		_ = pool.QueryRow(ctx, "SELECT id::text FROM customers WHERE merchant_id = $1 AND email = $2 LIMIT 1", merchantID, custEmail).Scan(&customerID)
		if customerID == "" {
			_ = pool.QueryRow(ctx, "INSERT INTO customers (merchant_id, email, phone) VALUES ($1, $2, $3) RETURNING id::text", merchantID, custEmail, custPhone).Scan(&customerID)
		}

		// 2. Payment status mapping
		payStatus := "PENDING"
		wfStatus := "SCHEDULED"
		action := "PAYMENT_LINK"
		prob := 0.80

		switch strings.ToLower(link.Status) {
		case "paid":
			payStatus = "CAPTURED"
			wfStatus = "RECOVERED"
		case "cancelled", "expired":
			payStatus = "CANCELLED"
			wfStatus = "HALTED"
		}

		// 3. Payments record
		var payID string
		_ = pool.QueryRow(ctx, "SELECT id::text FROM payments WHERE razorpay_payment_id = $1 LIMIT 1", link.ID).Scan(&payID)
		if payID == "" {
			_ = pool.QueryRow(ctx, `
				INSERT INTO payments (merchant_id, customer_id, amount, currency, status, method, razorpay_payment_id)
				VALUES ($1, $2, $3, $4, $5, 'payment_link', $6)
				RETURNING id::text
			`, merchantID, customerID, amountFloat, currency, payStatus, link.ID).Scan(&payID)
		} else {
			_, _ = pool.Exec(ctx, `
				UPDATE payments
				SET status = $1, customer_id = $2, updated_at = CURRENT_TIMESTAMP
				WHERE id::text = $3 AND status != 'CAPTURED'
			`, payStatus, customerID, payID)
		}

		// 4. Recovery Workflow record
		if payID != "" {
			var wfID string
			_ = pool.QueryRow(ctx, "SELECT id::text FROM recovery_workflows WHERE payment_id::text = $1 LIMIT 1", payID).Scan(&wfID)
			if wfID == "" {
				_, _ = pool.Exec(ctx, `
					INSERT INTO recovery_workflows (payment_id, merchant_id, status, selected_action, recovery_probability, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				`, payID, merchantID, wfStatus, action, prob)
			} else if wfStatus == "RECOVERED" || wfStatus == "HALTED" {
				_, _ = pool.Exec(ctx, `
					UPDATE recovery_workflows
					SET status = $1, updated_at = CURRENT_TIMESTAMP
					WHERE id::text = $2
				`, wfStatus, wfID)
			}
		}
	}
	return nil
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
