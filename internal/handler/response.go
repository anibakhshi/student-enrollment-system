package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// APIResponse defines the standard structure of API responses.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// writeJSON sends a JSON response to the client.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error(
			"failed to encode JSON response",
			"error", err,
		)
	}
}
