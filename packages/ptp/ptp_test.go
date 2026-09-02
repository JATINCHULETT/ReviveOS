package ptp_test

import (
	"testing"
	"time"

	"github.com/reviveos/ptp"
)

func TestEvaluateStatus(t *testing.T) {
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	rec := ptp.PromiseRecord{
		PromisedDate: now.Add(-48 * time.Hour), // 2 days past due
		Status:       ptp.StatusPending,
	}

	// Unpaid and overdue past 24h grace -> BROKEN
	status, err := ptp.EvaluateStatus(rec, now, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if status != ptp.StatusBroken {
		t.Errorf("Expected BROKEN, got %v", status)
	}

	// Paid -> HONORED
	status, err = ptp.EvaluateStatus(rec, now, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if status != ptp.StatusHonored {
		t.Errorf("Expected HONORED, got %v", status)
	}
}

func TestExtendPromise(t *testing.T) {
	rec := ptp.PromiseRecord{
		PromisedDate:   time.Now().Add(24 * time.Hour),
		Status:         ptp.StatusPending,
		ExtensionCount: 0,
	}

	futureDate := time.Now().Add(72 * time.Hour)
	err := ptp.ExtendPromise(&rec, futureDate, 2)
	if err != nil {
		t.Fatalf("Failed to extend promise: %v", err)
	}
	if rec.Status != ptp.StatusExtended {
		t.Errorf("Expected EXTENDED status, got %v", rec.Status)
	}
	if rec.ExtensionCount != 1 {
		t.Errorf("Expected ExtensionCount = 1, got %d", rec.ExtensionCount)
	}

	// Try exceeding max extensions
	_ = ptp.ExtendPromise(&rec, time.Now().Add(96*time.Hour), 2)
	err = ptp.ExtendPromise(&rec, time.Now().Add(120*time.Hour), 2)
	if err == nil {
		t.Errorf("Expected error on exceeding max extensions, got nil")
	}
}
