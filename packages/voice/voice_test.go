package voice_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/reviveos/voice"
)

func TestGenerateHinglishScript(t *testing.T) {
	payload := voice.CallPayload{
		CustomerName: "Rahul Sharma",
		Phone:        "+919876543210",
		Amount:       4999,
		Currency:     "INR",
		DueDate:      time.Now(),
	}

	script := voice.GenerateHinglishScript(payload)
	if !strings.Contains(script, "Rahul Sharma") {
		t.Errorf("Expected script to include customer name, got %s", script)
	}
	if !strings.Contains(script, "₹4999") {
		t.Errorf("Expected script to include ₹4999, got %s", script)
	}
	if !strings.Contains(script, "Namaste") {
		t.Errorf("Expected Hinglish greeting, got %s", script)
	}
}

func TestClassifyHinglishIntent(t *testing.T) {
	// Promise to Pay
	intent, ptp := voice.ClassifyHinglishIntent("Haan sir kal pakka payment kar dunga")
	if intent != voice.IntentPromiseToPay || ptp == nil {
		t.Errorf("Expected PROMISE_TO_PAY with date, got %v", intent)
	}

	// Request Link
	intent, _ = voice.ClassifyHinglishIntent("Mujhe WhatsApp pe payment link bhej do")
	if intent != voice.IntentRequestLink {
		t.Errorf("Expected REQUEST_LINK, got %v", intent)
	}

	// Already paid
	intent, _ = voice.ClassifyHinglishIntent("Main to already pay kar diya subah")
	if intent != voice.IntentAlreadyPaid {
		t.Errorf("Expected ALREADY_PAID, got %v", intent)
	}
}

func TestLocalSimulatorProvider(t *testing.T) {
	provider := &voice.LocalSimulatorProvider{}
	res, err := provider.InitiateCall(context.Background(), voice.CallPayload{
		CustomerName: "Priya",
		Amount:       2500,
	}, "Namaste Priya ji")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Status != "COMPLETED" {
		t.Errorf("Expected COMPLETED call status, got %s", res.Status)
	}
	if res.Intent != voice.IntentPromiseToPay {
		t.Errorf("Expected simulated PROMISE_TO_PAY intent, got %s", res.Intent)
	}
}
