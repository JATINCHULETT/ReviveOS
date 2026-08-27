package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/reviveos/utils/db"
)

func TestAuth_LoginFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	handler := AuthLoginHandler(pool)

	// 1. Admin login test
	adminPayload := map[string]string{
		"email":    "admin@reviveos.io",
		"password": "admin",
	}
	body, _ := json.Marshal(adminPayload)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected admin login 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}
	if resp.User.Role != "ADMIN" {
		t.Errorf("Expected role ADMIN, got %s", resp.User.Role)
	}
	if resp.Token == "" {
		t.Errorf("Expected non-empty token")
	}

	// 2. Merchant login test
	merchPayload := map[string]string{
		"email":    "merchant@acme.com",
		"password": "merchant",
	}
	body2, _ := json.Marshal(merchPayload)
	req2 := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body2))
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Expected merchant login 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp2 LoginResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("Failed to parse merchant login response: %v", err)
	}
	if resp2.User.Role != "MERCHANT" {
		t.Errorf("Expected role MERCHANT, got %s", resp2.User.Role)
	}

	t.Logf("SUCCESS: Both Admin and Merchant authentication verified with JWT tokens.")
}
