package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"night-owls-go/internal/config"
)

// RequestContextKey represents a key for request context values
type RequestContextKey string

const (
	RequestIDKey        RequestContextKey = "request_id"
	RequestStartTimeKey RequestContextKey = "request_start_time"
)

// RequestTracingMiddleware adds request ID and timing to all requests
func RequestTracingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate or extract request ID
			requestID := GetRequestID(r)
			startTime := time.Now()

			// Add request ID to response headers for client debugging
			w.Header().Set("X-Request-ID", requestID)

			// Add to request context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			ctx = context.WithValue(ctx, RequestStartTimeKey, startTime)

			// Create wrapped response writer to capture status code
			wrapped := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Log request start
			logger.InfoContext(ctx, "Request started",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.Header.Get("User-Agent"))

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log request completion
			duration := time.Since(startTime)
			logger.InfoContext(ctx, "Request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status_code", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"duration", duration.String())
		})
	}
}

// responseWriterWrapper captures the status code for logging
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// RateLimitMiddleware implements simple rate limiting per IP
func RateLimitMiddleware(requestsPerMinute int, logger *slog.Logger) func(next http.Handler) http.Handler {
	type rateLimitEntry struct {
		count     int
		resetTime time.Time
	}

	var (
		mu      sync.RWMutex
		clients = make(map[string]*rateLimitEntry)
	)

	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, entry := range clients {
				if now.After(entry.resetTime) {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP
			clientIP := getClientIP(r)
			now := time.Now()

			mu.Lock()
			entry, exists := clients[clientIP]
			if !exists || now.After(entry.resetTime) {
				// Create new entry or reset expired entry
				clients[clientIP] = &rateLimitEntry{
					count:     1,
					resetTime: now.Add(time.Minute),
				}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if entry.count >= requestsPerMinute {
				mu.Unlock()
				// Rate limit exceeded
				context := map[string]interface{}{
					"client_ip":  clientIP,
					"limit":      requestsPerMinute,
					"reset_time": entry.resetTime.Format(time.RFC3339),
				}

				RespondWithAPIError(w, r, http.StatusTooManyRequests,
					"Rate limit exceeded", ErrCodeRateLimit,
					logger, nil, context)
				return
			}

			entry.count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from various headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if ip := strings.Split(r.RemoteAddr, ":"); len(ip) > 0 {
		return ip[0]
	}

	return r.RemoteAddr
}

// ValidationRule represents a validation rule for a field
type ValidationRule struct {
	Field    string
	Required bool
	MinLen   int
	MaxLen   int
	Pattern  *regexp.Regexp
	Custom   func(interface{}) error
}

// ValidateRequest validates a request body against validation rules
func ValidateRequest(r *http.Request, target interface{}, rules []ValidationRule, logger *slog.Logger) []ValidationError {
	var validationErrors []ValidationError

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "body",
			Message: "Invalid JSON format",
			Code:    "INVALID_JSON",
		})
		return validationErrors
	}

	// Use reflection to validate fields
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "body",
			Message: "Expected object",
			Code:    "INVALID_TYPE",
		})
		return validationErrors
	}

	// Validate each rule
	for _, rule := range rules {
		field := v.FieldByName(rule.Field)
		if !field.IsValid() {
			continue // Field doesn't exist in struct
		}

		var fieldValue string
		var isEmpty bool

		// Handle different field types
		switch field.Kind() {
		case reflect.String:
			fieldValue = field.String()
			isEmpty = fieldValue == ""
		case reflect.Ptr:
			if field.IsNil() {
				isEmpty = true
			} else {
				fieldValue = fmt.Sprintf("%v", field.Elem().Interface())
			}
		default:
			fieldValue = fmt.Sprintf("%v", field.Interface())
			isEmpty = fieldValue == "" || fieldValue == "0"
		}

		// Check required fields
		if rule.Required && isEmpty {
			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(rule.Field),
				Message: fmt.Sprintf("%s is required", rule.Field),
				Code:    "REQUIRED_FIELD",
			})
			continue
		}

		// Skip other validations if field is empty and not required
		if isEmpty && !rule.Required {
			continue
		}

		// Check minimum length
		if rule.MinLen > 0 && len(fieldValue) < rule.MinLen {
			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(rule.Field),
				Value:   fieldValue,
				Message: fmt.Sprintf("%s must be at least %d characters", rule.Field, rule.MinLen),
				Code:    "MIN_LENGTH",
			})
		}

		// Check maximum length
		if rule.MaxLen > 0 && len(fieldValue) > rule.MaxLen {
			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(rule.Field),
				Value:   fieldValue,
				Message: fmt.Sprintf("%s must be at most %d characters", rule.Field, rule.MaxLen),
				Code:    "MAX_LENGTH",
			})
		}

		// Check pattern match
		if rule.Pattern != nil && !rule.Pattern.MatchString(fieldValue) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(rule.Field),
				Value:   fieldValue,
				Message: fmt.Sprintf("%s format is invalid", rule.Field),
				Code:    "INVALID_FORMAT",
			})
		}

		// Check custom validation
		if rule.Custom != nil {
			if err := rule.Custom(field.Interface()); err != nil {
				validationErrors = append(validationErrors, ValidationError{
					Field:   strings.ToLower(rule.Field),
					Value:   fieldValue,
					Message: err.Error(),
					Code:    "CUSTOM_VALIDATION",
				})
			}
		}
	}

	return validationErrors
}

// Common validation patterns
var (
	PhonePattern = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
	EmailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	CronPattern  = regexp.MustCompile(`^(\*|[0-5]?\d)\s+(\*|1?\d|2[0-3])\s+(\*|[12]?\d|3[01])\s+(\*|[1-9]|1[0-2])\s+(\*|[0-6])$`)
)

// ValidationMiddleware creates a middleware for request validation
func ValidationMiddleware(rules []ValidationRule, target interface{}, logger *slog.Logger) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Create new instance of target type
			targetType := reflect.TypeOf(target)
			if targetType.Kind() == reflect.Ptr {
				targetType = targetType.Elem()
			}
			newTarget := reflect.New(targetType).Interface()

			// Validate request
			validationErrors := ValidateRequest(r, newTarget, rules, logger)
			if len(validationErrors) > 0 {
				RespondWithValidationError(w, r, validationErrors, logger)
				return
			}

			// Add validated data to context for handler use
			ctx := context.WithValue(r.Context(), "validated_data", newTarget)
			next(w, r.WithContext(ctx))
		}
	}
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// CORS headers for development
			if cfg.DevMode {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

				// Handle preflight requests
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to get validated data from context
func GetValidatedData(r *http.Request, target interface{}) bool {
	if data := r.Context().Value("validated_data"); data != nil {
		// Copy data to target
		targetValue := reflect.ValueOf(target)
		if targetValue.Kind() != reflect.Ptr {
			return false
		}

		dataValue := reflect.ValueOf(data)
		if dataValue.Kind() == reflect.Ptr {
			dataValue = dataValue.Elem()
		}

		targetValue.Elem().Set(dataValue)
		return true
	}
	return false
}
