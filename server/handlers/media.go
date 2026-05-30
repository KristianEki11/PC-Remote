package handlers

import (
	"log/slog"
	"net/http"

	winapi "pcremote-server/windows"
)

// ──────────────────────────────────────
// Handlers
// ──────────────────────────────────────

// MediaPlayHandler handles POST /media/play
// Simulates a media Play/Pause keypress.
func MediaPlayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.SendMediaKey("play_pause"); err != nil {
		slog.Error("SendMediaKey play_pause failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}

// MediaNextHandler handles POST /media/next
// Simulates a media Next Track keypress.
func MediaNextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.SendMediaKey("next"); err != nil {
		slog.Error("SendMediaKey next failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}

// MediaPrevHandler handles POST /media/prev
// Simulates a media Previous Track keypress.
func MediaPrevHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	if err := winapi.SendMediaKey("prev"); err != nil {
		slog.Error("SendMediaKey prev failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true})
}

// MediaStatusHandler handles GET /media/status
// Retrieves the real-time playback status and metadata of active media sessions.
func MediaStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	status, err := winapi.GetMediaStatus()
	if err != nil {
		slog.Error("GetMediaStatus failed", "error", err)
		sendJSON(w, http.StatusInternalServerError, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, status)
}

