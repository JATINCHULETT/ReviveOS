package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
	"github.com/reviveos/utils/queue"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Fatalf("expected status 'ok', got '%s'", resp.Status)
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestWorkflowsHandler_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not available (%v)", err)
	}
	defer pool.Close()

	handler := WorkflowsHandler(pool)

	// 1. Test listing workflows
	req := httptest.NewRequest(http.MethodGet, "/workflows?limit=5", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /workflows, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var listResp WorkflowsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode workflows list response: %v", err)
	}

	if listResp.Limit != 5 {
		t.Errorf("expected limit 5, got %d", listResp.Limit)
	}
}

func TestAnalyticsOverviewHandler_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Skipf("Skipping integration test: database not available (%v)", err)
	}
	defer pool.Close()

	handler := AnalyticsHandler(pool)

	req := httptest.NewRequest(http.MethodGet, "/analytics/overview", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /analytics/overview, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var overview AnalyticsOverviewResponse
	if err := json.NewDecoder(rec.Body).Decode(&overview); err != nil {
		t.Fatalf("failed to decode analytics overview response: %v", err)
	}

	if overview.TotalPayments < 0 {
		t.Errorf("expected total_payments >= 0, got %d", overview.TotalPayments)
	}
}

func TestSystemQueuesHandler_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := queue.PingRedis(ctx); err != nil {
		t.Skipf("Skipping integration test: redis not available (%v)", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/system/queues", nil)
	rec := httptest.NewRecorder()

	SystemQueuesHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from /system/queues, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var resp SystemQueuesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode system queues response: %v", err)
	}

	if resp.RedisStatus != "connected" {
		t.Errorf("expected redis_status='connected', got '%s'", resp.RedisStatus)
	}
}
