package paymentprovider

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPaymentProvider instantiates the configured payment provider based on the environment or parameter.
func NewPaymentProvider(providerType string, pool *pgxpool.Pool) (PaymentProvider, error) {
	if providerType == "" {
		providerType = os.Getenv("PAYMENT_PROVIDER")
	}
	if providerType == "" {
		if os.Getenv("RAZORPAY_KEY_ID") != "" && os.Getenv("RAZORPAY_KEY_SECRET") != "" {
			providerType = "razorpay"
		} else {
			providerType = "local"
		}
	}

	switch strings.ToLower(providerType) {
	case "local":
		return NewLocalPaymentProvider(pool), nil
	case "razorpay":
		return NewRazorpayPaymentProvider(
			os.Getenv("RAZORPAY_KEY_ID"),
			os.Getenv("RAZORPAY_KEY_SECRET"),
			os.Getenv("RAZORPAY_BASE_URL"),
		), nil
	default:
		return nil, fmt.Errorf("unknown payment provider type: %s", providerType)
	}
}
