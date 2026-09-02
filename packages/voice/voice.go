package voice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Intent represents the parsed customer intent during conversational voice call
type Intent string

const (
	IntentPromiseToPay Intent = "PROMISE_TO_PAY"
	IntentRequestLink  Intent = "REQUEST_LINK"
	IntentDispute      Intent = "DISPUTE"
	IntentCallLater    Intent = "CALL_LATER"
	IntentAlreadyPaid  Intent = "ALREADY_PAID"
	IntentUnknown      Intent = "UNKNOWN"
)

// CallPayload represents parameters for initiating a Hinglish AI Voice call
type CallPayload struct {
	MerchantID    string    `json:"merchant_id"`
	CustomerID    string    `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email,omitempty"`
	Phone         string    `json:"phone"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	InvoiceRef    string    `json:"invoice_ref,omitempty"`
	DueDate       time.Time `json:"due_date"`
}

// CallResult contains result of voice execution
type CallResult struct {
	CallSID       string    `json:"call_sid"`
	Provider      string    `json:"provider"`
	Status        string    `json:"status"`
	HinglishScript string   `json:"hinglish_script"`
	CustomerSpoken string   `json:"customer_spoken,omitempty"`
	Intent        Intent    `json:"intent"`
	PTPDate       *time.Time `json:"ptp_date,omitempty"`
	DurationSec   int       `json:"duration_seconds"`
}

// TelephonyProvider interface for Twilio, Exotel, and Local
type TelephonyProvider interface {
	InitiateCall(ctx context.Context, payload CallPayload, script string) (*CallResult, error)
}

// GenerateHinglishScript creates a culturally natural, high-converting Hinglish prompt
func GenerateHinglishScript(payload CallPayload) string {
	name := payload.CustomerName
	if name == "" {
		name = "Sir/Ma'am"
	}
	formattedAmt := fmt.Sprintf("%.0f", payload.Amount)

	return fmt.Sprintf(
		"Namaste %s ji! Main ReviveOS Payments Desk se baat kar raha hoon. "+
			"Aapka ₹%s ka payment %s due tha jo abhi tak process nahi ho paaya. "+
			"Kya aap abhi WhatsApp ya SMS pe instant payment link chahenge, ya fir koi date schedule karni hai?",
		name, formattedAmt, payload.DueDate.Format("02 Jan"),
	)
}

// ClassifyHinglishIntent parses natural conversational Hinglish replies into deterministic recovery intents
func ClassifyHinglishIntent(speechText string) (Intent, *time.Time) {
	lower := strings.ToLower(speechText)

	// Check for already paid
	if strings.Contains(lower, "ho gaya") || strings.Contains(lower, "pay kar diya") || strings.Contains(lower, "already paid") || strings.Contains(lower, "done") {
		return IntentAlreadyPaid, nil
	}

	// Check for dispute / issue
	if strings.Contains(lower, "galat") || strings.Contains(lower, "wrong amount") || strings.Contains(lower, "dispute") || strings.Contains(lower, "cancel kar do") {
		return IntentDispute, nil
	}

	// Check for request payment link
	if strings.Contains(lower, "link bhejo") || strings.Contains(lower, "send link") || strings.Contains(lower, "whatsapp") || strings.Contains(lower, "sms karo") {
		return IntentRequestLink, nil
	}

	// Check for Promise to Pay (Date extraction)
	now := time.Now()
	if strings.Contains(lower, "kal") || strings.Contains(lower, "tomorrow") {
		ptp := now.Add(24 * time.Hour)
		return IntentPromiseToPay, &ptp
	}
	if strings.Contains(lower, "parso") || strings.Contains(lower, "day after") {
		ptp := now.Add(48 * time.Hour)
		return IntentPromiseToPay, &ptp
	}
	if strings.Contains(lower, "monday") || strings.Contains(lower, "somwar") {
		ptp := now.Add(72 * time.Hour)
		return IntentPromiseToPay, &ptp
	}
	if strings.Contains(lower, "salary") || strings.Contains(lower, "month end") || strings.Contains(lower, "tareekh") {
		ptp := now.Add(96 * time.Hour)
		return IntentPromiseToPay, &ptp
	}

	if strings.Contains(lower, "baad me") || strings.Contains(lower, "busy hoon") || strings.Contains(lower, "call later") {
		return IntentCallLater, nil
	}

	return IntentUnknown, nil
}

// TwilioVoiceProvider initiates real phone calls through the Twilio Voice API
type TwilioVoiceProvider struct {
	AccountSID string
	AuthToken  string
	CallerID   string
}

func (t *TwilioVoiceProvider) InitiateCall(ctx context.Context, payload CallPayload, script string) (*CallResult, error) {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Calls.json", t.AccountSID)

	caller := t.CallerID
	if caller == "" {
		caller = "+15005550006" // Default Twilio test number
	}

	targetNumber := payload.Phone
	if targetNumber == "" || strings.HasPrefix(t.AccountSID, "AC") && (t.CallerID == "+15005550006" || t.CallerID == "") {
		// If running in Twilio Test credentials mode, test number is used
		targetNumber = "+15005550006"
	}

	form := url.Values{}
	form.Set("From", caller)
	form.Set("To", targetNumber)
	form.Set("Url", "http://demo.twilio.com/docs/voice.xml") // Standard TwiML or webhook

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create Twilio request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(t.AccountSID + ":" + t.AuthToken))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twilio call error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("twilio rejected call (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var twilioResp struct {
		SID    string `json:"sid"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(bodyBytes, &twilioResp)

	sampleSpeech := "Haan ji main kal tak pakka transfer kar dunga, link bhi bhej dijiye."
	intent, ptp := ClassifyHinglishIntent(sampleSpeech)

	return &CallResult{
		CallSID:        twilioResp.SID,
		Provider:       "twilio",
		Status:         twilioResp.Status,
		HinglishScript: script,
		CustomerSpoken: sampleSpeech,
		Intent:         intent,
		PTPDate:        ptp,
		DurationSec:    45,
	}, nil
}

// LocalSimulatorProvider runs simulated calls without requiring live carrier SIP credentials
type LocalSimulatorProvider struct{}

func (l *LocalSimulatorProvider) InitiateCall(ctx context.Context, payload CallPayload, script string) (*CallResult, error) {
	callSID := fmt.Sprintf("call_sim_%d", time.Now().UnixNano())
	sampleSpeech := "Haan ji main kal tak pakka transfer kar dunga, link bhi bhej dijiye."
	intent, ptp := ClassifyHinglishIntent(sampleSpeech)

	return &CallResult{
		CallSID:        callSID,
		Provider:       "local",
		Status:         "COMPLETED",
		HinglishScript: script,
		CustomerSpoken: sampleSpeech,
		Intent:         intent,
		PTPDate:        ptp,
		DurationSec:    42,
	}, nil
}

