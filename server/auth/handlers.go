package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"pcremote-server/config"
)

// AuthHandlers provides HTTP handlers for authentication endpoints.
type AuthHandlers struct {
	Sessions *SessionManager
	Pairing  *PairingManager
	Limiter  AuthFailureRecorder
}

// AuthFailureRecorder is an interface for recording authentication failures/successes
type AuthFailureRecorder interface {
	RecordAuthFailure(ip string)
	ResetAuthFailures(ip string)
	IsLocked(ip string) bool
}

// ── Request / Response types ─────────────────────────────────

type loginRequest struct {
	PIN string `json:"pin"`
}

type pairRequest struct {
	PairToken string `json:"pair_token"`
}

type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type authErrorResponse struct {
	Error string `json:"error"`
}

func authSendJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// ── POST /auth/login ─────────────────────────────────────────
// Validates PIN (bcrypt), returns a Bearer session token.
// This is the backward-compatible login endpoint for manual IP+PIN entry.
func (ah *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		authSendJSON(w, http.StatusMethodNotAllowed, authErrorResponse{Error: "method not allowed"})
		return
	}

	ip := r.RemoteAddr

	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		authSendJSON(w, http.StatusBadRequest, authErrorResponse{Error: "invalid request body"})
		return
	}

	if req.PIN == "" {
		authSendJSON(w, http.StatusBadRequest, authErrorResponse{Error: "pin is required"})
		return
	}

	// Validate PIN against stored bcrypt hash
	if !config.ValidatePIN(req.PIN) {
		slog.Warn("Failed login attempt", "remote_addr", r.RemoteAddr)
		if ah.Limiter != nil {
			ah.Limiter.RecordAuthFailure(ip)
		}
		authSendJSON(w, http.StatusUnauthorized, authErrorResponse{Error: "invalid pin"})
		return
	}

	if ah.Limiter != nil {
		ah.Limiter.ResetAuthFailures(ip)
	}

	deviceName := r.Header.Get("X-Device-Name")
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	session, err := ah.Sessions.CreateSession(deviceName, r.RemoteAddr)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		authSendJSON(w, http.StatusInternalServerError, authErrorResponse{Error: "internal server error"})
		return
	}

	slog.Info("Device logged in via PIN", "device", deviceName, "remote_addr", r.RemoteAddr)

	authSendJSON(w, http.StatusOK, tokenResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

// ── POST /auth/pair ──────────────────────────────────────────
// Validates one-time pairing token from QR code, returns a session token.
// This is the primary pairing flow: scan QR → call /auth/pair → get token.
func (ah *AuthHandlers) PairHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		authSendJSON(w, http.StatusMethodNotAllowed, authErrorResponse{Error: "method not allowed"})
		return
	}

	var req pairRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
		authSendJSON(w, http.StatusBadRequest, authErrorResponse{Error: "invalid request body"})
		return
	}

	if req.PairToken == "" {
		authSendJSON(w, http.StatusBadRequest, authErrorResponse{Error: "pair_token is required"})
		return
	}

	if !ah.Pairing.ValidateAndConsume(req.PairToken) {
		slog.Warn("Invalid pairing attempt", "remote_addr", r.RemoteAddr)
		authSendJSON(w, http.StatusUnauthorized, authErrorResponse{Error: "invalid or expired pairing token"})
		return
	}

	deviceName := r.Header.Get("X-Device-Name")
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	session, err := ah.Sessions.CreateSession(deviceName, r.RemoteAddr)
	if err != nil {
		slog.Error("Failed to create session after pairing", "error", err)
		authSendJSON(w, http.StatusInternalServerError, authErrorResponse{Error: "internal server error"})
		return
	}

	slog.Info("Device paired via QR code", "device", deviceName, "remote_addr", r.RemoteAddr)

	authSendJSON(w, http.StatusOK, tokenResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

// ── POST /auth/logout ────────────────────────────────────────
// Revokes the current session. Requires valid Bearer token (enforced by middleware).
func (ah *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		authSendJSON(w, http.StatusMethodNotAllowed, authErrorResponse{Error: "method not allowed"})
		return
	}

	ah.Sessions.RevokeAll()
	slog.Info("Device logged out", "remote_addr", r.RemoteAddr)
	authSendJSON(w, http.StatusOK, map[string]any{"success": true})
}

// ── GET /auth/sessions ───────────────────────────────────────
// Returns the active session status. Used by the system tray and internal QR page.
// This endpoint should be restricted to localhost via middleware.
func (ah *AuthHandlers) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		authSendJSON(w, http.StatusMethodNotAllowed, authErrorResponse{Error: "method not allowed"})
		return
	}

	session := ah.Sessions.GetActiveSession()
	if session == nil {
		authSendJSON(w, http.StatusOK, map[string]any{
			"connected": false,
			"session":   nil,
		})
		return
	}

	authSendJSON(w, http.StatusOK, map[string]any{
		"connected": session.IsOnline(),
		"session": map[string]any{
			"device_name":  session.DeviceName,
			"remote_addr":  session.RemoteAddr,
			"created_at":   session.CreatedAt.Format(time.RFC3339),
			"last_seen_at": session.LastSeenAt.Format(time.RFC3339),
			"online":       session.IsOnline(),
		},
	})
}
