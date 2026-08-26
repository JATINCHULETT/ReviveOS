package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LoggingMiddleware logs incoming requests with duration and status
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// CORSMiddleware sets common CORS headers for frontend access
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Razorpay-Event-Id, X-Razorpay-Signature")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RegisterRoutes registers all API routes and returns the configured Handler
func RegisterRoutes(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	// 0. Root Endpoint
	mux.HandleFunc("/", RootHandler)

	// 1. Health & Readiness
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/ready", HealthHandler)

	// 2. Workflows
	mux.HandleFunc("/workflows", WorkflowsHandler(pool))
	mux.HandleFunc("/workflows/", WorkflowsHandler(pool))

	// 3. Analytics
	mux.HandleFunc("/analytics", AnalyticsHandler(pool))
	mux.HandleFunc("/analytics/", AnalyticsHandler(pool))

	// 4. System / Infrastructure Status
	mux.HandleFunc("/system/queues", SystemQueuesHandler)
	mux.HandleFunc("/system/health", SystemHealthHandler(pool))

	// 5. Payment Provider Webhooks
	mux.HandleFunc("/webhooks/razorpay", RazorpayWebhookHandler(pool))
	mux.HandleFunc("/api/v1/webhooks/razorpay", RazorpayWebhookHandler(pool))

	return LoggingMiddleware(CORSMiddleware(mux))
}
