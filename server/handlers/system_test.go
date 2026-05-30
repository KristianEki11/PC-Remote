package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
		t.Errorf("expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "PIN changed successfully") {
		t.Errorf("expected body to contain success message, got %q", rr.Body.String())
	}
	if config.App.PIN != "5678" {
		t.Errorf("expected config.App.PIN to be 5678, got %q", config.App.PIN)
	}

	// Verify .env file updated
	envBytes, err := os.ReadFile(".env")
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}
	envStr := string(envBytes)
	if !strings.Contains(envStr, "PIN=5678") {
		t.Errorf("expected .env to contain PIN=5678, got: %q", envStr)
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

	handler := middleware.WithAuth(http.HandlerFunc(HandleChangePIN))

	body := `{"current_pin":"1234","new_pin":"5678"}`
	req := httptest.NewRequest("POST", "/system/pin", strings.NewReader(body))
	// Missing X-PIN header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
