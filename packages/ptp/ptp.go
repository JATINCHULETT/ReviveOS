package ptp

import (
	"errors"
	"time"
)

// PTPStatus represents status of a Promise-to-Pay record
type PTPStatus string

const (
	StatusPending   PTPStatus = "PENDING"
	StatusHonored   PTPStatus = "HONORED"
	StatusBroken    PTPStatus = "BROKEN"
	StatusExtended  PTPStatus = "EXTENDED"
	StatusCancelled PTPStatus = "CANCELLED"
)

// PromiseRecord represents a customer's formal promise to pay
type PromiseRecord struct {
	ID                 string     `json:"id"`
	MerchantID         string     `json:"merchant_id"`
	CustomerID         string     `json:"customer_id"`
	CustomerName       string     `json:"customer_name"`
	CustomerContact    string     `json:"customer_contact"`
	InvoiceID          *string    `json:"invoice_id,omitempty"`
	PaymentEventID     *string    `json:"payment_event_id,omitempty"`
	PromisedAmount     float64    `json:"promised_amount"`
	PromisedDate       time.Time  `json:"promised_date"`
	Status             PTPStatus  `json:"status"`
	RecordedChannel    string     `json:"recorded_channel"`
	ExtensionCount     int        `json:"extension_count"`
	LastReminderSentAt *time.Time `json:"last_reminder_sent_at,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// EvaluateStatus evaluates if a pending promise has lapsed into BROKEN based on current timestamp
func EvaluateStatus(record PromiseRecord, asOf time.Time, isPaid bool) (PTPStatus, error) {
	if isPaid {
		return StatusHonored, nil
	}
	if record.Status == StatusHonored {
		return StatusHonored, nil
	}

	// 24-hour grace period before marking broken
	graceDeadline := record.PromisedDate.Add(24 * time.Hour)
	if asOf.After(graceDeadline) {
		return StatusBroken, nil
	}

	return record.Status, nil
}

// ExtendPromise extends a customer promise date up to max extensions
func ExtendPromise(record *PromiseRecord, newDate time.Time, maxExtensions int) error {
	if record.ExtensionCount >= maxExtensions {
		return errors.New("maximum promise extensions reached")
	}
	if newDate.Before(time.Now()) {
		return errors.New("extended date must be in the future")
	}
	record.PromisedDate = newDate
	record.ExtensionCount++
	record.Status = StatusExtended
	record.UpdatedAt = time.Now()
	return nil
}
