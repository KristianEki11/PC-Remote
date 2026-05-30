package handlers

import (
	"bufio"
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pcremote-server/config"
	winapi "pcremote-server/windows"
)

// ──────────────────────────────────────
// Request Model
// ──────────────────────────────────────

type ShutdownRequest struct {
	DelaySeconds int `json:"delay_seconds"`
	DelayMinutes int `json:"delay_minutes"`
}

// ──────────────────────────────────────
// Handlers
// ──────────────────────────────────────

// SystemLockHandler handles POST /system/lock
// Locks the Windows workstation (shows the lock screen).
func SystemLockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.LockWorkstation(); err != nil {
		slog.Error("LockWorkstation failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}

func HandleChangePIN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	// 1. Parse JSON body
	var req struct {
		CurrentPIN string `json:"current_pin"`
		NewPIN     string `json:"new_pin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "malformed JSON body"})
		return
	}

	// 4. Validate current PIN is not empty
	if config.App.PIN == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "No PIN currently set. Please reinstall or manually set PIN in .env"})
		return
	}

	// 2. Validate current_pin matches X-PIN header (constant-time compare)
	xPin := r.Header.Get("X-PIN")
	if subtle.ConstantTimeCompare([]byte(req.CurrentPIN), []byte(xPin)) != 1 {
		sendJSON(w, http.StatusUnauthorized, ErrorBody{Error: "current PIN incorrect"})
		return
	}

	// 3. Validate new_pin: 4-8 chars, digits only
	if req.NewPIN == req.CurrentPIN {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "new PIN cannot be same as current PIN"})
		return
	}

	if len(req.NewPIN) < 4 || len(req.NewPIN) > 8 {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "new PIN must be 4-8 digits"})
		return
	}

	for _, c := range req.NewPIN {
		if c < '0' || c > '9' {
			sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "new PIN must be 4-8 digits"})
			return
		}
	}

	// 5. Update config.App.PIN in-memory
	config.App.PIN = req.NewPIN

	// 6. Write updated .env file (preserve other keys like PORT)
	envPath := ".env"
	exePath, err := os.Executable()
	if err == nil {
		targetPath := filepath.Join(filepath.Dir(exePath), ".env")
		if _, err := os.Stat(targetPath); err == nil {
			envPath = targetPath
		}
	}

	file, err := os.Open(envPath)
	if err != nil {
		slog.Error("PIN change failed: Failed to open .env for reading", "error", err, "path", envPath)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: "Failed to save PIN to file"})
		return
	}

	var newContent bytes.Buffer
	scanner := bufio.NewScanner(file)
	keyFound := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		var pinKeyName string
		if strings.HasPrefix(trimmedLine, "PIN=") {
			pinKeyName = "PIN"
		} else if strings.HasPrefix(trimmedLine, "APP_PIN=") {
			pinKeyName = "APP_PIN"
		}

		if pinKeyName != "" {
			newContent.WriteString(pinKeyName + "=" + req.NewPIN + "\n")
			keyFound = true
		} else {
			newContent.WriteString(line + "\n")
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		slog.Error("PIN change failed: Failed to read .env", "error", err, "path", envPath)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: "Failed to save PIN to file"})
		return
	}

	if !keyFound {
		newContent.WriteString("APP_PIN=" + req.NewPIN + "\n")
	}

	// Write to temporary file in the same directory as .env
	tmpPath := envPath + ".tmp"
	if err := os.WriteFile(tmpPath, newContent.Bytes(), 0600); err != nil {
		slog.Error("PIN change failed: Failed to write temp file", "error", err, "path", tmpPath)
		os.Remove(tmpPath)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: "Failed to save PIN to file"})
		return
	}

	// Validate temp file is readable
	if _, err := os.ReadFile(tmpPath); err != nil {
		slog.Error("PIN change failed: Failed to read temp file", "error", err, "path", tmpPath)
		os.Remove(tmpPath)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: "Failed to save PIN to file"})
		return
	}

	// ATOMIC RENAME
	if err := os.Rename(tmpPath, envPath); err != nil {
		slog.Error("PIN change failed: Failed to rename temp file to env", "error", err, "from", tmpPath, "to", envPath)
		os.Remove(tmpPath)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: "Failed to save PIN to file"})
		return
	}

	// 7. Log success
	slog.Info("PIN changed successfully")

	// 8. Return 200 with success message
	sendJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "PIN changed successfully",
	})
}

// SystemShutdownHandler handles POST /system/shutdown
// Schedules a system shutdown. Expects JSON body: {"delay_seconds": N} or {"delay_minutes": M}
func SystemShutdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	var req ShutdownRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.DelaySeconds < 0 || req.DelayMinutes < 0 {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "delay must be >= 0"})
		return
	}

	seconds := req.DelaySeconds
	if seconds == 0 && req.DelayMinutes > 0 {
		seconds = req.DelayMinutes * 60
	}

	if err := winapi.ScheduleShutdown(seconds); err != nil {
		slog.Error("ScheduleShutdown failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{
		"success":       true,
		"delay_seconds": seconds,
	})
}

// SystemShutdownCancelHandler handles POST /system/shutdown/cancel
// Cancels a previously scheduled shutdown.
func SystemShutdownCancelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.CancelShutdown(); err != nil {
		slog.Error("CancelShutdown failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}

// SystemSleepHandler handles POST /system/sleep
// Puts the Windows PC to sleep.
func SystemSleepHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.Sleep(); err != nil {
		slog.Error("Sleep failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}

// SystemRestartHandler handles POST /system/restart
// Restarts the Windows PC.
func SystemRestartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.Restart(); err != nil {
		slog.Error("Restart failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}
