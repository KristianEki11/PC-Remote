package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"pcremote-server/auth"
	"pcremote-server/config"
)

// ErrorResponse is the standard JSON error format.
type ErrorResponse struct {
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// WithAuth creates authentication middleware that validates Bearer tokens.
// It also supports the legacy X-PIN header for backward compatibility,
// logging a deprecation warning when X-PIN is used.
//
// Authentication priority:
//  1. Authorization: Bearer <token>  →  validate via SessionManager
//  2. X-PIN: <pin>                   →  validate via bcrypt (deprecated, creates temp session)
//  3. Neither present                →  401 Unauthorized
func WithAuth(sessions *auth.SessionManager, authLimiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ExtractIP(r)

			// Check if IP is locked out from too many failures
			if authLimiter.IsLocked(ip) {
				sendError(w, http.StatusTooManyRequests, "too many failed attempts, try again later")
				return
			}

			// ── Strategy 1: Bearer token ─────────────────────
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if _, valid := sessions.ValidateToken(token); valid {
					authLimiter.ResetAuthFailures(ip)
					next.ServeHTTP(w, r)
					return
				}
				authLimiter.RecordAuthFailure(ip)
				sendError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// ── Strategy 2: Legacy X-PIN header (deprecated) ─
			providedPIN := r.Header.Get("X-PIN")
			if providedPIN != "" {
				slog.Warn("Deprecated X-PIN auth used — migrate to Bearer token",
					"remote_addr", r.RemoteAddr,
					"path", r.URL.Path,
				)

				if config.ValidatePIN(providedPIN) {
					authLimiter.ResetAuthFailures(ip)

					// Create a temporary session for X-PIN users
					// so they show up in the tray connection status
					deviceName := r.Header.Get("X-Device-Name")
					if deviceName == "" {
						deviceName = "Legacy Client"
					}
					// Only create session if none exists (avoid churning)
					if sessions.GetActiveSession() == nil {
						sessions.CreateSession(deviceName, r.RemoteAddr)
					} else {
						// Touch existing session's LastSeenAt
						if s := sessions.GetActiveSession(); s != nil {
							sessions.ValidateToken(s.Token)
						}
					}

					next.ServeHTTP(w, r)
					return
				}
				authLimiter.RecordAuthFailure(ip)
				sendError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			// ── No credentials provided ──────────────────────
			if config.App.PIN == "" {
				sendError(w, http.StatusForbidden, "no pin configured")
				return
			}

			sendError(w, http.StatusUnauthorized, "authorization required")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code for logging.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// WithLogging logs every HTTP request with method, path, status, and duration.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		level := slog.LevelInfo
		if rw.status >= 500 {
			level = slog.LevelError
		}

		slog.Log(r.Context(), level, "HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// WithSecurityHeaders adds standard security headers to all responses.
func WithSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}
