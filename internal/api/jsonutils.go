package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RespondWithError sends a JSON error response.
func RespondWithError(w http.ResponseWriter, code int, message string, logger *slog.Logger, details ...any) {
	RespondWithJSON(w, code, map[string]string{"error": message}, logger, details...)
}

// RespondWithJSON sends a JSON response.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}, logger *slog.Logger, details ...any) {
	response, err := json.Marshal(payload)
	if err != nil {
		// Use the provided logger if available, otherwise default to slog's default logger.
		currentLogger := logger
		if currentLogger == nil {
			currentLogger = slog.Default()
		}
		currentLogger.Error("Failed to marshal JSON response", "payload", payload, "error", err, "details", details)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "failed to marshal JSON response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		currentLogger := logger
		if currentLogger == nil {
			currentLogger = slog.Default()
		}
		currentLogger.Warn("Failed to write JSON response", "error", err, "details", details)
	}
} 