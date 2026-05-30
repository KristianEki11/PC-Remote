package handlers

import (
	"net/http"
)

// HealthHandler handles GET /health (no auth required).
// Returns server status, version, and platform info.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"version":  "2.1.1",
		"platform": "windows",
	})
}
