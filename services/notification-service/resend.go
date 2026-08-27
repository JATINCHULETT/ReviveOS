package notificationservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ResendProvider implements NotificationProvider using the Resend REST API.
type ResendProvider struct {
	APIKey     string
	FromEmail  string
	BaseURL    string
	HTTPClient *http.Client
}

// NewResendProvider creates a new ResendProvider instance.
func NewResendProvider(apiKey, fromEmail string) *ResendProvider {
	if apiKey == "" {
		apiKey = os.Getenv("RESEND_API_KEY")
	}
	if fromEmail == "" {
		fromEmail = os.Getenv("RESEND_FROM_EMAIL")
	}
	if fromEmail == "" {
		fromEmail = "ReviveOS Recovery <recoveries@reviveos.io>"
	}

	return &ResendProvider{
		APIKey:    apiKey,
		FromEmail: fromEmail,
		BaseURL:   "https://api.resend.com",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type resendEmailResponse struct {
	ID    string `json:"id"`
	Error *struct {
		Message string `json:"message"`
		Name    string `json:"name"`
	} `json:"error,omitempty"`
}

// SendRecoveryNotification sends a recovery email with a payment link using Resend.
func (r *ResendProvider) SendRecoveryNotification(ctx context.Context, req NotificationRequest) (*NotificationResult, error) {
	if r.APIKey == "" {
		return &NotificationResult{
			Provider:  "resend",
			Status:    "FAILED",
			Recipient: req.CustomerEmail,
			SentAt:    time.Now().UTC(),
			Error:     "missing RESEND_API_KEY",
		}, errors.New("RESEND_API_KEY is not configured")
	}

	if req.CustomerEmail == "" {
		return &NotificationResult{
			Provider:  "resend",
			Status:    "FAILED",
			Recipient: "",
			SentAt:    time.Now().UTC(),
			Error:     "customer email is required",
		}, errors.New("cannot send recovery notification: customer email is empty")
	}

	merchantName := req.MerchantName
	if merchantName == "" {
		merchantName = "Your Merchant"
	}

	currency := req.Currency
	if currency == "" {
		currency = "INR"
	}

	subject := fmt.Sprintf("Action Required: Complete your %s payment for %s", currency, merchantName)
	htmlContent := renderRecoveryEmailHTML(req)

	payload := resendEmailRequest{
		From:    r.FromEmail,
		To:      []string{req.CustomerEmail},
		Subject: subject,
		HTML:    htmlContent,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email payload: %w", err)
	}

	url := fmt.Sprintf("%s/emails", strings.TrimSuffix(r.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.HTTPClient.Do(httpReq)
	if err != nil {
		return &NotificationResult{
			Provider:  "resend",
			Status:    "FAILED",
			Recipient: req.CustomerEmail,
			SentAt:    time.Now().UTC(),
			Error:     err.Error(),
		}, fmt.Errorf("resend HTTP call failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read resend response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var resendErr resendEmailResponse
		_ = json.Unmarshal(respBytes, &resendErr)
		errMsg := string(respBytes)
		if resendErr.Error != nil && resendErr.Error.Message != "" {
			errMsg = resendErr.Error.Message
		}
		return &NotificationResult{
			Provider:  "resend",
			Status:    "FAILED",
			Recipient: req.CustomerEmail,
			SentAt:    time.Now().UTC(),
			Error:     errMsg,
		}, fmt.Errorf("resend API returned status %d: %s", resp.StatusCode, errMsg)
	}

	var successResp resendEmailResponse
	if err := json.Unmarshal(respBytes, &successResp); err != nil {
		return nil, fmt.Errorf("malformed resend response: %w", err)
	}

	return &NotificationResult{
		MessageID: successResp.ID,
		Provider:  "resend",
		Status:    "SENT",
		Recipient: req.CustomerEmail,
		SentAt:    time.Now().UTC(),
	}, nil
}

func renderRecoveryEmailHTML(req NotificationRequest) string {
	merchantName := req.MerchantName
	if merchantName == "" {
		merchantName = "ReviveOS Merchant"
	}
	customerName := req.CustomerName
	if customerName == "" {
		customerName = "Valued Customer"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #0B0F17; color: #F1F5F9; padding: 20px; }
  .card { max-width: 540px; margin: 0 auto; background-color: #111827; border: 1px solid #1F2937; border-radius: 12px; padding: 32px; }
  .badge { display: inline-block; background-color: #1E293B; color: #38BDF8; font-size: 12px; font-weight: 600; padding: 4px 10px; border-radius: 9999px; margin-bottom: 16px; }
  h1 { font-size: 20px; color: #FFFFFF; margin: 0 0 12px; }
  p { font-size: 14px; color: #94A3B8; line-height: 1.6; margin: 0 0 16px; }
  .amount-box { background: #0F172A; border-left: 4px solid #38BDF8; padding: 12px 16px; border-radius: 6px; margin-bottom: 24px; }
  .amount-box span { font-size: 13px; color: #64748B; }
  .amount-box strong { display: block; font-size: 18px; color: #38BDF8; margin-top: 4px; }
  .btn { display: inline-block; background: linear-gradient(135deg, #2563EB, #1D4ED8); color: #FFFFFF; text-decoration: none; padding: 12px 24px; font-size: 14px; font-weight: 600; border-radius: 8px; box-shadow: 0 4px 14px rgba(37,99,235,0.4); text-align: center; }
  .footer { font-size: 12px; color: #64748B; margin-top: 32px; border-top: 1px solid #1E293B; padding-top: 16px; text-align: center; }
</style>
</head>
<body>
  <div class="card">
    <div class="badge">SECURE PAYMENT RECOVERY</div>
    <h1>Complete Your Payment for %s</h1>
    <p>Hi %s,</p>
    <p>We noticed an issue completing your recent transaction. To keep your subscription or order active without interruption, please click the secure link below to retry with your preferred payment method.</p>
    
    <div class="amount-box">
      <span>Amount Pending</span>
      <strong>%s %.2f</strong>
    </div>

    <div style="text-align: center; margin: 24px 0;">
      <a href="%s" class="btn" target="_blank">Complete Payment Securely &rarr;</a>
    </div>

    <p style="font-size: 12px; color: #64748B;">If the button above does not work, copy and paste this link into your browser:<br><a href="%s" style="color: #38BDF8; word-break: break-all;">%s</a></p>

    <div class="footer">
      Powered by ReviveOS &bull; Protected by end-to-end encryption.
    </div>
  </div>
</body>
</html>`,
		merchantName, customerName, req.Currency, req.Amount, req.PaymentLink, req.PaymentLink, req.PaymentLink,
	)
}
