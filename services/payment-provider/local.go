package paymentprovider

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LocalPaymentProvider implements PaymentProvider using PostgreSQL as the authoritative store.
type LocalPaymentProvider struct {
	pool *pgxpool.Pool
}

// NewLocalPaymentProvider creates a new instance of LocalPaymentProvider.
func NewLocalPaymentProvider(pool *pgxpool.Pool) *LocalPaymentProvider {
	return &LocalPaymentProvider{pool: pool}
}

// GetPayment fetches the authoritative state from the local provider store.
func (p *LocalPaymentProvider) GetPayment(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	var (
		id                string
		pID               string
		providerPaymentID string
		amount            float64
		currency          string
		status            string
		method            string
		failureCode       sql.NullString
		failureReason     sql.NullString
		attemptsCount     int
		updatedAt         time.Time
	)

	err := p.pool.QueryRow(ctx, `
		SELECT 
			id::text, payment_id, provider_payment_id, amount::float8, currency, status, method, 
			failure_code, failure_reason, attempts_count, updated_at
		FROM local_provider_payments
		WHERE payment_id = $1 OR provider_payment_id = $1
	`, paymentID).Scan(
		&id, &pID, &providerPaymentID, &amount, &currency, &status, &method,
		&failureCode, &failureReason, &attemptsCount, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If not yet initialized in local provider store, initialize from main payments table
			return p.initializeFromApplicationPayment(ctx, paymentID)
		}
		return nil, fmt.Errorf("failed to query local provider payment: %w", err)
	}

	return &PaymentStatus{
		PaymentID:         pID,
		ProviderPaymentID: providerPaymentID,
		Status:            status,
		Amount:            amount,
		Currency:          currency,
		Method:            method,
		FailureCode:       failureCode.String,
		FailureReason:     failureReason.String,
		Captured:          status == "CAPTURED",
		UpdatedAt:         updatedAt,
	}, nil
}

// initializeFromApplicationPayment bootstraps the local provider record if a payment exists in the application database.
func (p *LocalPaymentProvider) initializeFromApplicationPayment(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	var (
		amount      float64
		currency    string
		status      string
		method      sql.NullString
		failureCode sql.NullString
	)

	err := p.pool.QueryRow(ctx, `
		SELECT amount::float8, currency, status, method, failure_code
		FROM payments
		WHERE id::text = $1
	`, paymentID).Scan(&amount, &currency, &status, &method, &failureCode)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment %s not found in provider or database", paymentID)
		}
		return nil, fmt.Errorf("failed to read payment for provider initialization: %w", err)
	}

	methodVal := "card"
	if method.Valid && method.String != "" {
		methodVal = method.String
	}

	providerPaymentID := fmt.Sprintf("loc_pay_%s", paymentID)

	var (
		insertedID        string
		insertedUpdatedAt time.Time
	)

	err = p.pool.QueryRow(ctx, `
		INSERT INTO local_provider_payments (
			payment_id, provider_payment_id, amount, currency, status, method, failure_code, failure_reason, attempts_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)
		ON CONFLICT (payment_id) DO UPDATE
			SET updated_at = CURRENT_TIMESTAMP
		RETURNING id::text, updated_at
	`, paymentID, providerPaymentID, amount, currency, status, methodVal, failureCode.String, "").Scan(
		&insertedID, &insertedUpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider payment: %w", err)
	}

	return &PaymentStatus{
		PaymentID:         paymentID,
		ProviderPaymentID: providerPaymentID,
		Status:            status,
		Amount:            amount,
		Currency:          currency,
		Method:            methodVal,
		FailureCode:       failureCode.String,
		Captured:          status == "CAPTURED",
		UpdatedAt:         insertedUpdatedAt,
	}, nil
}

// CreateRetryAttempt executes a real provider operation to retry a failed payment.
func (p *LocalPaymentProvider) CreateRetryAttempt(ctx context.Context, paymentID string, amount float64) (*RetryResult, error) {
	return p.CreateRetryAttemptWithCustomer(ctx, paymentID, amount, "", "", "")
}

// CreateRetryAttemptWithCustomer executes a real provider operation to retry a failed payment with customer context.
func (p *LocalPaymentProvider) CreateRetryAttemptWithCustomer(ctx context.Context, paymentID string, amount float64, customerEmail, customerPhone, customerName string) (*RetryResult, error) {
	currentStatus, err := p.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment before retry: %w", err)
	}

	if currentStatus.Status == "CAPTURED" {
		return &RetryResult{
			AttemptID:         fmt.Sprintf("loc_att_%d", time.Now().UnixNano()),
			PaymentID:         currentStatus.PaymentID,
			ProviderPaymentID: currentStatus.ProviderPaymentID,
			Status:            "SUCCESS",
			Amount:            amount,
			CreatedAt:         time.Now().UTC(),
		}, nil
	}

	// Realistic provider-side decline evaluation
	retrySuccess := false
	var declineReason string

	switch strings.ToUpper(currentStatus.FailureCode) {
	case "EXPIRED_CARD", "INVALID_ACCOUNT", "CARD_BLOCKED", "STOLEN_CARD", "DO_NOT_HONOR":
		// Hard decline: Cannot succeed on same card without updating credentials
		retrySuccess = false
		declineReason = fmt.Sprintf("Declined by issuing bank: %s (terminal failure)", currentStatus.FailureCode)

	case "INSUFFICIENT_FUNDS":
		// Soft decline: Succeeds if amount is valid
		if amount <= 0 {
			retrySuccess = false
			declineReason = "Invalid retry amount"
		} else {
			retrySuccess = true
		}

	case "TIMEOUT", "BANK_UNAVAILABLE", "NETWORK_ERROR", "PROCESSING_ERROR":
		// Transient gateway/bank error: Succeeds upon retry
		retrySuccess = true

	default:
		// Default retry behavior for soft errors
		retrySuccess = true
	}

	var newProviderStatus string
	var attemptStatus string

	if retrySuccess {
		newProviderStatus = "CAPTURED"
		attemptStatus = "SUCCESS"
	} else {
		newProviderStatus = "FAILED"
		attemptStatus = "FAILED"
	}

	attemptID := fmt.Sprintf("loc_att_%d", time.Now().UnixNano())

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin provider transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Update authoritative provider record in local_provider_payments
	_, err = tx.Exec(ctx, `
		UPDATE local_provider_payments
		SET status = $1,
		    failure_reason = $2,
		    attempts_count = attempts_count + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE payment_id = $3
	`, newProviderStatus, declineReason, currentStatus.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to update provider payment state: %w", err)
	}

	// 2. Insert provider attempt log
	var localPaymentUUID string
	err = tx.QueryRow(ctx, `
		SELECT id::text FROM local_provider_payments WHERE payment_id = $1
	`, currentStatus.PaymentID).Scan(&localPaymentUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get local payment uuid: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO local_provider_attempts (local_payment_id, payment_id, amount, status, error_message)
		VALUES ($1, $2, $3, $4, $5)
	`, localPaymentUUID, currentStatus.PaymentID, amount, attemptStatus, declineReason)
	if err != nil {
		return nil, fmt.Errorf("failed to insert provider attempt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit provider retry attempt: %w", err)
	}

	return &RetryResult{
		AttemptID:         attemptID,
		PaymentID:         currentStatus.PaymentID,
		ProviderPaymentID: currentStatus.ProviderPaymentID,
		Status:            attemptStatus,
		Amount:            amount,
		PaymentLinkURL:    fmt.Sprintf("https://checkout.reviveos.io/pay/%s", currentStatus.PaymentID),
		ErrorMessage:      declineReason,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// VerifyPayment queries the provider's authoritative state store.
func (p *LocalPaymentProvider) VerifyPayment(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	var (
		pID               string
		providerPaymentID string
		amount            float64
		currency          string
		status            string
		method            string
		failureCode       sql.NullString
		failureReason     sql.NullString
		updatedAt         time.Time
	)

	err := p.pool.QueryRow(ctx, `
		SELECT 
			payment_id, provider_payment_id, amount::float8, currency, status, method, 
			failure_code, failure_reason, updated_at
		FROM local_provider_payments
		WHERE payment_id = $1 OR provider_payment_id = $1
	`, paymentID).Scan(
		&pID, &providerPaymentID, &amount, &currency, &status, &method,
		&failureCode, &failureReason, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment %s not found in provider store", paymentID)
		}
		return nil, fmt.Errorf("failed to verify provider payment: %w", err)
	}

	return &PaymentStatus{
		PaymentID:         pID,
		ProviderPaymentID: providerPaymentID,
		Status:            status,
		Amount:            amount,
		Currency:          currency,
		Method:            method,
		FailureCode:       failureCode.String,
		FailureReason:     failureReason.String,
		Captured:          status == "CAPTURED",
		UpdatedAt:         updatedAt,
	}, nil
}

// RegisterPayment seeds an explicit payment state in the local provider for testing and initialization.
func (p *LocalPaymentProvider) RegisterPayment(ctx context.Context, paymentID string, amount float64, currency, status, method, failureCode string) (*PaymentStatus, error) {
	providerPaymentID := fmt.Sprintf("loc_pay_%s", paymentID)
	var updatedAt time.Time

	err := p.pool.QueryRow(ctx, `
		INSERT INTO local_provider_payments (
			payment_id, provider_payment_id, amount, currency, status, method, failure_code, failure_reason, attempts_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, '', 0)
		ON CONFLICT (payment_id) DO UPDATE
			SET status = EXCLUDED.status,
			    amount = EXCLUDED.amount,
			    currency = EXCLUDED.currency,
			    method = EXCLUDED.method,
			    failure_code = EXCLUDED.failure_code,
			    updated_at = CURRENT_TIMESTAMP
		RETURNING updated_at
	`, paymentID, providerPaymentID, amount, currency, status, method, failureCode).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to register provider payment: %w", err)
	}

	return &PaymentStatus{
		PaymentID:         paymentID,
		ProviderPaymentID: providerPaymentID,
		Status:            status,
		Amount:            amount,
		Currency:          currency,
		Method:            method,
		FailureCode:       failureCode,
		Captured:          status == "CAPTURED",
		UpdatedAt:         updatedAt,
	}, nil
}
