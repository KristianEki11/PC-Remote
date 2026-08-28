package handlers

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

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
// Only http:// and https:// schemes are allowed to prevent SSRF attacks.
func BrowserOpenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, ErrorBody{Error: "method not allowed"})
		return
	}

	var req BrowserOpenRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "invalid request body"})
		return
	}

	if req.URL == "" {
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: "url is required"})
		return
	}

	// Validate URL scheme — only http and https are allowed
	if err := validateBrowserURL(req.URL); err != nil {
		slog.Warn("Browser URL rejected", "url", req.URL, "reason", err.Error())
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: err.Error()})
		return
	}

	if err := winapi.OpenBrowser(req.URL); err != nil {
		slog.Error("OpenBrowser failed", "error", err)
		sendJSON(w, http.StatusBadRequest, ErrorBody{Error: err.Error()})
		return
	}

	sendJSON(w, http.StatusOK, map[string]any{"success": true, "url": req.URL})
}

// validateBrowserURL checks that the URL uses a safe scheme and doesn't target internal IPs.
func validateBrowserURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return &urlError{"invalid URL format"}
	}

	// Only allow http and https schemes
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return &urlError{"only http:// and https:// URLs are allowed"}
	}

	// Block requests to loopback and link-local addresses
	hostname := parsed.Hostname()
	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return &urlError{"URLs targeting internal addresses are not allowed"}
		}
	}

	// Block common dangerous hostnames
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || lowerHost == "metadata.google.internal" {
		return &urlError{"URLs targeting internal addresses are not allowed"}
	}

	return nil
}

type urlError struct {
	msg string
}

func (e *urlError) Error() string { return e.msg }
