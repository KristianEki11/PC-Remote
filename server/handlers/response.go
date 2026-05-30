package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the standard error response format for all endpoints.
type ErrorBody struct {
	Error string `json:"error"`
}

// sendJSON writes a JSON response with the given status code and payload.
func sendJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
