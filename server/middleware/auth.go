package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"pcremote-server/config"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		configuredPIN := config.App.PIN

		if configuredPIN == "" {
			sendError(w, http.StatusForbidden, "no pin configured")
			return
		}

		providedPIN := r.Header.Get("X-PIN")
		
		// Constant time compare to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedPIN), []byte(configuredPIN)) != 1 {
			sendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

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
