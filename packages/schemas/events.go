package schemas

import "time"

// CustomerHistory contains the summary of customer's historical payment behavior.
type CustomerHistory struct {
	SuccessfulPayments int `json:"successful_payments"`
	FailedPayments     int `json:"failed_payments"`
}

// PaymentFailureEvent represents the normalized failure event ingested into the system.
type PaymentFailureEvent struct {
	PaymentID       string          `json:"payment_id"`
	CustomerID      string          `json:"customer_id"`
	SubscriptionID  string          `json:"subscription_id"`
	Amount          float64         `json:"amount"`
	Currency        string          `json:"currency"`
	PaymentMethod   string          `json:"payment_method"`
	FailureCode     string          `json:"failure_code"`
	AttemptNumber   int             `json:"attempt_number"`
	Timestamp       time.Time       `json:"timestamp"`
	CustomerHistory CustomerHistory `json:"customer_history"`

	// Ground Truth for Synthetic Generation/Evaluation (Not present in real webhook)
	SimulatedGroundTruth *SyntheticGroundTruth `json:"simulated_ground_truth,omitempty"`
}

// SyntheticGroundTruth provides the objective reality for the simulator to evaluate against.
type SyntheticGroundTruth struct {
	WouldRecoverNaturally      bool `json:"would_recover_naturally"`
	WouldRecoverAfterRetry     bool `json:"would_recover_after_retry"`
	WouldRecoverAfterCustomer  bool `json:"would_recover_after_customer"`
	WouldNeverRecover          bool `json:"would_never_recover"`
	OptimalRetryTimingOffsetHs int  `json:"optimal_retry_timing_offset_hs"` // e.g., +24h
}
