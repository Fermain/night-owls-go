package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RespondWithError sends a JSON error response with enhanced error handling
// This function maintains backward compatibility while leveraging the new error system
func RespondWithError(w http.ResponseWriter, code int, message string, logger *slog.Logger, details ...any) {
	// Try to extract request from details for enhanced error handling
	var r *http.Request

	// Parse details to find request and build context
	context := make(map[string]interface{})
	for i := 0; i < len(details); i++ {
		switch v := details[i].(type) {
		case *http.Request:
			r = v
		case string:
			// Treat as key-value pairs
			if i+1 < len(details) {
				context[v] = details[i+1]
				i++ // Skip next item as we used it as value
			}
		}
	}

	// If we have a request, use the advanced error handling
	if r != nil {
		errorCode := getErrorCodeFromStatus(code)
		RespondWithAPIError(w, r, code, message, errorCode, logger, nil, context)
		return
	}

	// Fallback to legacy behavior for backward compatibility
	RespondWithJSON(w, code, map[string]string{"error": message}, logger, details...)
}

// RespondWithJSON sends a JSON response with enhanced logging
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}, logger *slog.Logger, details ...any) {
	response, err := json.Marshal(payload)
	if err != nil {
		// Use the provided logger if available, otherwise default to slog's default logger.
		currentLogger := logger
		if currentLogger == nil {
			currentLogger = slog.Default()
		}

		// Enhanced error logging with more context
		logFields := []interface{}{
			"payload_type", getPayloadType(payload),
			"error", err,
		}

		// Add details to log fields
		for i := 0; i < len(details)-1; i += 2 {
			if key, ok := details[i].(string); ok && i+1 < len(details) {
				logFields = append(logFields, key, details[i+1])
			}
		}

		currentLogger.Error("Failed to marshal JSON response", logFields...)

		// Send structured error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Failed to marshal JSON response", "code": "JSON_MARSHAL_ERROR"}`))
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

		// Enhanced warning logging
		logFields := []interface{}{
			"status_code", code,
			"response_size", len(response),
			"error", err,
		}

		// Add details to log fields
		for i := 0; i < len(details)-1; i += 2 {
			if key, ok := details[i].(string); ok && i+1 < len(details) {
				logFields = append(logFields, key, details[i+1])
			}
		}

		currentLogger.Warn("Failed to write JSON response", logFields...)
	}
}

// getPayloadType returns a string representation of the payload type for logging
func getPayloadType(payload interface{}) string {
	if payload == nil {
		return "nil"
	}

	switch payload.(type) {
	case map[string]string:
		return "map[string]string"
	case map[string]interface{}:
		return "map[string]interface{}"
	case []interface{}:
		return "[]interface{}"
	case string:
		return "string"
	default:
		return "unknown"
	}
}
