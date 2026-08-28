package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"pcremote-server/auth"
	"pcremote-server/config"
	"pcremote-server/middleware"
)

func setupTestEnv(t *testing.T, content string) func() {
	err := os.WriteFile(".env", []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to setup test .env: %v", err)
	}
	return func() {
		os.Remove(".env")
		os.Remove(".env.tmp")
	}
}

func TestChangePIN_Success(t *testing.T) {
	defer setTestPIN("1234")()
	defer setupTestEnv(t, "PIN=1234\nPORT=8000\n")()

	body := `{"current_pin":"1234","new_pin":"5678"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	req.Header.Set("X-PIN", "1234")
	rr := httptest.NewRecorder()

	HandleChangePIN(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// In-memory update check
	if config.App.PIN == "5678" || config.App.PIN == "" {
		t.Errorf("expected config.App.PIN to be a bcrypt hash, got %q", config.App.PIN)
	}

	// Verify the hash is actually valid for the new PIN
	if !config.ValidatePIN("5678") {
		t.Errorf("expected new PIN '5678' to validate against the new stored hash")
	}

	// .env file check
	envBytes, err := os.ReadFile(".env")
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}
	envContent := string(envBytes)
	
	// Check that it contains PIN= and doesn't contain plaintext 5678
	if !strings.Contains(envContent, "$2a$") && !strings.Contains(envContent, "$2b$") {
		t.Errorf("expected .env to contain bcrypt hash, got: %q", envContent)
	}
	if !strings.Contains(envContent, "PORT=8000") {
		t.Errorf("expected .env to preserve PORT=8000, got: %q", envContent)
	}
}

func TestChangePIN_WrongCurrentPIN(t *testing.T) {
	defer setTestPIN("1234")()
	defer setupTestEnv(t, "PIN=1234\n")()

	body := `{"current_pin":"9999","new_pin":"5678"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	req.Header.Set("X-PIN", "1234")
	rr := httptest.NewRecorder()

	HandleChangePIN(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "current PIN incorrect") {
		t.Errorf("expected body to contain error, got %q", rr.Body.String())
	}
	if config.App.PIN != "1234" {
		t.Errorf("expected config.App.PIN to remain 1234, got %q", config.App.PIN)
	}
}

func TestChangePIN_InvalidNewPINFormat(t *testing.T) {
	defer setTestPIN("1234")()
	defer setupTestEnv(t, "PIN=1234\n")()

	body := `{"current_pin":"1234","new_pin":"abc"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	req.Header.Set("X-PIN", "1234")
	rr := httptest.NewRecorder()

	HandleChangePIN(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "must be 4-8 digits") {
		t.Errorf("expected body to contain error, got %q", rr.Body.String())
	}
}

func TestChangePIN_SameAsOldPIN(t *testing.T) {
	defer setTestPIN("1234")()
	defer setupTestEnv(t, "PIN=1234\n")()

	body := `{"current_pin":"1234","new_pin":"1234"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	req.Header.Set("X-PIN", "1234")
	rr := httptest.NewRecorder()

	HandleChangePIN(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "cannot be same as current PIN") {
		t.Errorf("expected body to contain error, got %q", rr.Body.String())
	}
}

func TestChangePIN_EmptyCurrentPIN(t *testing.T) {
	defer setTestPIN("")()
	defer setupTestEnv(t, "PIN=\n")()

	body := `{"current_pin":"","new_pin":"5678"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	req.Header.Set("X-PIN", "")
	rr := httptest.NewRecorder()

	HandleChangePIN(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "No PIN currently set") {
		t.Errorf("expected body to contain error, got %q", rr.Body.String())
	}
}

func TestChangePIN_MissingAuthHeader(t *testing.T) {
	defer setTestPIN("1234")()
	defer setupTestEnv(t, "PIN=1234\n")()

	// Import "pcremote-server/auth" will be added by goimports
	sm := auth.NewSessionManager()
	rl := middleware.NewAuthRateLimiter()
	handler := middleware.WithAuth(sm, rl)(http.HandlerFunc(HandleChangePIN))

	body := `{"current_pin":"1234","new_pin":"5678"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	// Missing X-PIN header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
