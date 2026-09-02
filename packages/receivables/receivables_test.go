package receivables_test

import (
	"testing"
	"time"

	"github.com/reviveos/receivables"
)

func TestCalculateAging(t *testing.T) {
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

	// Due in future -> CURRENT
	future := now.Add(5 * 24 * time.Hour)
	days, bucket := receivables.CalculateAging(future, now)
	if bucket != receivables.BucketCurrent || days != 0 {
		t.Errorf("Expected CURRENT with 0 days, got %v with %d days", bucket, days)
	}

	// 15 days past due -> 1_30
	past15 := now.Add(-15 * 24 * time.Hour)
	days, bucket = receivables.CalculateAging(past15, now)
	if bucket != receivables.Bucket1To30 || days != 15 {
		t.Errorf("Expected 1_30 with 15 days, got %v with %d days", bucket, days)
	}

	// 45 days past due -> 31_60
	past45 := now.Add(-45 * 24 * time.Hour)
	days, bucket = receivables.CalculateAging(past45, now)
	if bucket != receivables.Bucket31To60 || days != 45 {
		t.Errorf("Expected 31_60 with 45 days, got %v with %d days", bucket, days)
	}
}

func TestNextDunningAction(t *testing.T) {
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	inv := receivables.Invoice{
		InvoiceNumber: "INV-101",
		BuyerCompany:  "Acme Corp",
		BuyerName:     "Amit",
		Amount:        50000,
		Currency:      "INR",
		DueDate:       now.Add(-20 * 24 * time.Hour),
	}

	action := receivables.NextDunningAction(inv, now)
	if action.Stage != 1 {
		t.Errorf("Expected Stage 1 dunning, got %d", action.Stage)
	}
	if action.Tone != "FIRM" {
		t.Errorf("Expected FIRM tone, got %s", action.Tone)
	}
	if !action.GeneratePaymentLink {
		t.Errorf("Expected GeneratePaymentLink to be true")
	}
}
