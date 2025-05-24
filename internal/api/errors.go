package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

// APIError represents a structured API error response with detailed context
type APIError struct {
	// Core error information
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Code      string    `json:"code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	
	// Request context for debugging
	RequestID string `json:"request_id,omitempty"`
	Path      string `json:"path,omitempty"`
	Method    string `json:"method,omitempty"`
	
	// Detailed error information (development only)
	Details   *ErrorDetails `json:"details,omitempty"`
	
	// Validation errors (for input validation failures)
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

// ErrorDetails provides additional debugging information (development mode only)
type ErrorDetails struct {
	Type       string                 `json:"type,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	InternalMsg string                `json:"internal_message,omitempty"`
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ErrorCode constants for standardized error identification
const (
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeAuthentication   = "AUTHENTICATION_ERROR" 
	ErrCodeAuthorization    = "AUTHORIZATION_ERROR"
	ErrCodeNotFound         = "RESOURCE_NOT_FOUND"
	ErrCodeConflict         = "RESOURCE_CONFLICT"
	ErrCodeRateLimit        = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalServer   = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
)

// GetRequestID extracts or generates a request ID for tracking
func GetRequestID(r *http.Request) string {
	// Try to get request ID from header first
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	
	// Try correlation ID header
	if correlationID := r.Header.Get("X-Correlation-ID"); correlationID != "" {
		return correlationID
	}
	
	// Generate a new request ID if none exists
	return generateRequestID()
}

// generateRequestID creates a new unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return "req_" + hex.EncodeToString(bytes)
}

// Enhanced error response functions

// RespondWithAPIError sends a structured API error response with full context
func RespondWithAPIError(w http.ResponseWriter, r *http.Request, statusCode int, message, errorCode string, logger *slog.Logger, internalErr error, context map[string]interface{}) {
	apiError := APIError{
		Error:     http.StatusText(statusCode),
		Message:   message,
		Code:      errorCode,
		Timestamp: time.Now().UTC(),
		RequestID: GetRequestID(r),
		Path:      r.URL.Path,
		Method:    r.Method,
	}

	// Add detailed debugging information in development mode
	if shouldIncludeDetails(r) && internalErr != nil {
		apiError.Details = &ErrorDetails{
			Type:        fmt.Sprintf("%T", internalErr),
			InternalMsg: internalErr.Error(),
			Context:     context,
		}
		
		// Capture stack trace for internal errors
		if statusCode >= 500 {
			apiError.Details.StackTrace = captureStackTrace()
		}
	}

	// Log the error with full context
	logFields := []interface{}{
		"status_code", statusCode,
		"error_code", errorCode,
		"message", message,
		"request_id", apiError.RequestID,
		"path", r.URL.Path,
		"method", r.Method,
	}
	
	if internalErr != nil {
		logFields = append(logFields, "internal_error", internalErr.Error())
	}
	
	if context != nil {
		for k, v := range context {
			logFields = append(logFields, k, v)
		}
	}

	if statusCode >= 500 {
		logger.ErrorContext(r.Context(), "API Error", logFields...)
	} else {
		logger.WarnContext(r.Context(), "API Error", logFields...)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(apiError); err != nil {
		logger.ErrorContext(r.Context(), "Failed to encode error response", "error", err)
		// Fallback to basic error response
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"Internal Server Error","message":"Failed to encode error response"}`))
	}
}

// RespondWithValidationError sends a detailed validation error response
func RespondWithValidationError(w http.ResponseWriter, r *http.Request, validationErrors []ValidationError, logger *slog.Logger) {
	apiError := APIError{
		Error:            "Bad Request",
		Message:          "Request validation failed",
		Code:             ErrCodeValidation,
		Timestamp:        time.Now().UTC(),
		RequestID:        GetRequestID(r),
		Path:             r.URL.Path,
		Method:           r.Method,
		ValidationErrors: validationErrors,
	}

	// Log validation errors
	logger.WarnContext(r.Context(), "Validation Error", 
		"request_id", apiError.RequestID,
		"path", r.URL.Path,
		"validation_errors", validationErrors)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	
	if err := json.NewEncoder(w).Encode(apiError); err != nil {
		logger.ErrorContext(r.Context(), "Failed to encode validation error response", "error", err)
	}
}

// Helper functions

// shouldIncludeDetails determines if detailed error information should be included
func shouldIncludeDetails(r *http.Request) bool {
	// Check for dev mode header or query parameter
	return r.Header.Get("X-Debug-Mode") == "true" || 
		   r.URL.Query().Get("debug") == "true" ||
		   strings.Contains(r.Header.Get("User-Agent"), "development")
}

// captureStackTrace captures the current stack trace for debugging
func captureStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// getErrorCodeFromStatus maps HTTP status codes to error codes
func getErrorCodeFromStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrCodeBadRequest
	case http.StatusUnauthorized:
		return ErrCodeAuthentication
	case http.StatusForbidden:
		return ErrCodeAuthorization
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusMethodNotAllowed:
		return ErrCodeMethodNotAllowed
	case http.StatusConflict:
		return ErrCodeConflict
	case http.StatusTooManyRequests:
		return ErrCodeRateLimit
	case http.StatusInternalServerError:
		return ErrCodeInternalServer
	default:
		return fmt.Sprintf("HTTP_%d", statusCode)
	}
}

// buildContextFromDetails builds a context map from variadic details
func buildContextFromDetails(details []any) map[string]interface{} {
	if len(details) == 0 {
		return nil
	}
	
	context := make(map[string]interface{})
	for i := 0; i < len(details)-1; i += 2 {
		if key, ok := details[i].(string); ok && i+1 < len(details) {
			context[key] = details[i+1]
		}
	}
	return context
}

// extractRequestFromContext attempts to extract request from various sources
func extractRequestFromContext(w http.ResponseWriter, details []any) *http.Request {
	// Try to create a minimal request for error context
	r := &http.Request{
		Method: "UNKNOWN",
		URL:    &url.URL{Path: "unknown"},
		Header: make(http.Header),
	}
	
	// Look for request in details
	for _, detail := range details {
		if req, ok := detail.(*http.Request); ok {
			return req
		}
	}
	
	return r
}

// Error recovery and circuit breaker helpers

// RecoverFromPanic recovers from panics and returns a structured error response
func RecoverFromPanic(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	if err := recover(); err != nil {
		context := map[string]interface{}{
			"panic_value": fmt.Sprintf("%v", err),
		}
		
		RespondWithAPIError(w, r, http.StatusInternalServerError, 
			"Internal server error occurred", ErrCodeInternalServer, 
			logger, fmt.Errorf("panic: %v", err), context)
	}
}

// WithErrorRecovery wraps a handler with panic recovery
func WithErrorRecovery(handler http.HandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer RecoverFromPanic(w, r, logger)
		handler(w, r)
	}
} 