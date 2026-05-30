package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	winapi "pcremote-server/windows"
)

// ──────────────────────────────────────
// Request Model
// ──────────────────────────────────────

type BrowserOpenRequest struct {
	URL string `json:"url"`
}

// ──────────────────────────────────────
// Handler
// ──────────────────────────────────────

// BrowserOpenHandler handles POST /browser/open
// Opens a URL in the system default browser. Expects JSON body: {"url": "https://..."}
func BrowserOpenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	var req BrowserOpenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}

	if req.URL == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "url is required"})
		return
	}

	if err := winapi.OpenBrowser(req.URL); err != nil {
		slog.Error("OpenBrowser failed", "error", err)
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true, "url": req.URL})
}
