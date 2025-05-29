# Advanced Error Handling & Refinements

This document describes the sophisticated error handling system implemented in the Night Owls Go API.

## Overview

The error handling system provides:
- **Structured Error Responses** with standardized format
- **Request Tracing** with correlation IDs
- **Comprehensive Validation** with detailed field-level errors
- **Rate Limiting** with proper error responses
- **Security Headers** and CORS handling
- **Panic Recovery** with graceful error responses
- **Enhanced Logging** with full context preservation

## Core Components

### 1. Structured Error Responses

All API errors return a consistent format with rich context:

```json
{
  "error": "Bad Request",
  "message": "Request validation failed",
  "code": "VALIDATION_ERROR",
  "timestamp": "2024-11-27T10:30:00Z",
  "request_id": "req_abc123def456",
  "path": "/api/bookings",
  "method": "POST",
  "validation_errors": [
    {
      "field": "start_time",
      "value": "invalid-date",
      "message": "start_time format is invalid",
      "code": "INVALID_FORMAT"
    }
  ],
  "details": {
    "type": "*errors.errorString",
    "context": {
      "user_id": 123,
      "schedule_id": 456
    },
    "internal_message": "parsing time \"invalid-date\": invalid format"
  }
}
```

### 2. Error Codes

Standardized error codes for programmatic handling:

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `AUTHENTICATION_ERROR` | 401 | Authentication required or failed |
| `AUTHORIZATION_ERROR` | 403 | Insufficient permissions |
| `RESOURCE_NOT_FOUND` | 404 | Requested resource not found |
| `RESOURCE_CONFLICT` | 409 | Resource conflict (duplicate, etc.) |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Internal server error |

### 3. Request Tracing

Every request gets a unique correlation ID for tracking:

```go
// Automatically added to all requests
requestID := GetRequestID(r)

// Available in logs and error responses
{
  "request_id": "req_abc123def456",
  "message": "Request processed",
  "duration_ms": 125
}
```

## Usage Patterns

### 1. Basic Error Response

```go
func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
    // Use enhanced error response with context
    context := map[string]interface{}{
        "user_id": getUserID(r),
        "resource_type": "booking",
    }
    
    RespondWithAPIError(w, r, http.StatusBadRequest, 
        "Invalid resource data", ErrCodeValidation,
        h.logger, err, context)
}
```

### 2. Validation Errors

```go
func (h *Handler) ValidateAndCreate(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    
    validationErrors := ValidateRequest(r, &req, validationRules, h.logger)
    if len(validationErrors) > 0 {
        RespondWithValidationError(w, r, validationErrors, h.logger)
        return
    }
    
    // Process validated request...
}
```

### 3. Service Error Mapping

```go
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
    booking, err := h.service.CreateBooking(ctx, params)
    if err != nil {
        switch {
        case errors.Is(err, service.ErrScheduleNotFound):
            RespondWithAPIError(w, r, http.StatusNotFound, 
                "Schedule not found", ErrCodeNotFound,
                h.logger, err, map[string]interface{}{
                    "schedule_id": params.ScheduleID,
                })
        case errors.Is(err, service.ErrBookingConflict):
            RespondWithAPIError(w, r, http.StatusConflict, 
                "Shift slot already booked", ErrCodeConflict,
                h.logger, err, map[string]interface{}{
                    "schedule_id": params.ScheduleID,
                    "start_time": params.StartTime,
                })
        default:
            RespondWithAPIError(w, r, http.StatusInternalServerError, 
                "Failed to create booking", ErrCodeInternalServer,
                h.logger, err, nil)
        }
        return
    }
    
    RespondWithJSON(w, http.StatusCreated, booking, h.logger)
}
```

## Middleware Stack

### 1. Complete Middleware Chain

```go
func SetupAdvancedRoutes(mux *http.ServeMux, cfg *config.Config, logger *slog.Logger) {
    // Validation rules
    rules := []ValidationRule{
        {Field: "Name", Required: true, MinLen: 2, MaxLen: 100},
        {Field: "Email", Required: true, Pattern: EmailPattern},
    }
    
    // Complete middleware chain
    handler := SecurityHeadersMiddleware(cfg)(           // Security headers
        RequestTracingMiddleware(logger)(                // Request ID & timing
            RateLimitMiddleware(60, logger)(             // Rate limiting
                ValidationMiddleware(rules, Request{}, logger)( // Validation
                    WithErrorRecovery(actualHandler, logger))))) // Panic recovery
    
    mux.Handle("/api/resource", handler)
}
```

### 2. Individual Middleware Components

#### Request Tracing
- Generates/extracts request IDs
- Adds timing information
- Logs request start/completion
- Sets response headers for debugging

#### Rate Limiting
- IP-based rate limiting
- Configurable requests per minute
- Automatic cleanup of old entries
- Structured rate limit error responses

#### Validation
- Reflection-based field validation
- Custom validation functions
- Detailed field-level error messages
- Pattern matching (regex) support

#### Security Headers
- Content type protection
- Frame options
- XSS protection
- CORS configuration for development

#### Panic Recovery
- Catches and logs panics
- Returns structured 500 errors
- Includes stack trace in dev mode
- Prevents server crashes

## Advanced Features

### 1. Development Mode Debugging

When debugging is enabled (via headers or config), error responses include:

```json
{
  "details": {
    "type": "*errors.errorString",
    "stack_trace": "goroutine 1 [running]:\n...",
    "context": {
      "database_query": "SELECT * FROM bookings WHERE...",
      "parameters": {...}
    },
    "internal_message": "connection timeout after 30s"
  }
}
```

Enable debugging:
- Header: `X-Debug-Mode: true`
- Query param: `?debug=true`
- User-Agent containing "development"

### 2. Context Preservation

Errors maintain full context throughout the stack:

```go
// At service layer
return fmt.Errorf("booking conflict: %w", originalErr)

// At API layer - context is preserved and enhanced
context := map[string]interface{}{
    "schedule_id": params.ScheduleID,
    "requested_time": params.StartTime,
    "existing_booking_id": existingBooking.ID,
}
RespondWithAPIError(w, r, http.StatusConflict, 
    "Time slot unavailable", ErrCodeConflict,
    logger, err, context)
```

### 3. Backward Compatibility

The enhanced system maintains compatibility with existing code:

```go
// Old style - still works
RespondWithError(w, http.StatusBadRequest, "Invalid input", logger)

// Enhanced automatically when request is available
RespondWithError(w, http.StatusBadRequest, "Invalid input", logger, 
    "user_id", 123, r) // Request enables enhanced features
```

## Error Response Examples

### Validation Error

```json
{
  "error": "Bad Request",
  "message": "Request validation failed",
  "code": "VALIDATION_ERROR",
  "timestamp": "2024-11-27T10:30:00Z",
  "request_id": "req_abc123",
  "path": "/api/bookings",
  "method": "POST",
  "validation_errors": [
    {
      "field": "start_time",
      "value": "2024-13-45T25:00:00Z",
      "message": "start_time format is invalid",
      "code": "INVALID_FORMAT"
    },
    {
      "field": "buddy_name",
      "message": "buddy_name must be at least 2 characters",
      "code": "MIN_LENGTH"
    }
  ]
}
```

### Business Logic Error

```json
{
  "error": "Conflict", 
  "message": "Shift slot is already booked",
  "code": "RESOURCE_CONFLICT",
  "timestamp": "2024-11-27T10:30:00Z",
  "request_id": "req_def456",
  "path": "/api/bookings",
  "method": "POST",
  "details": {
    "type": "*service.BookingConflictError",
    "context": {
      "schedule_id": 42,
      "start_time": "2024-11-28T18:00:00Z",
      "existing_booking_id": 789,
      "existing_user": "Alice Smith"
    },
    "internal_message": "booking exists for schedule 42 at 2024-11-28T18:00:00Z"
  }
}
```

### Rate Limit Error

```json
{
  "error": "Too Many Requests",
  "message": "Rate limit exceeded", 
  "code": "RATE_LIMIT_EXCEEDED",
  "timestamp": "2024-11-27T10:30:00Z",
  "request_id": "req_ghi789",
  "path": "/api/bookings",
  "method": "POST",
  "details": {
    "context": {
      "client_ip": "192.168.1.100",
      "limit": 60,
      "reset_time": "2024-11-27T10:31:00Z"
    }
  }
}
```

## Integration Guidelines

### 1. Converting Existing Handlers

**Before:**
```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondWithError(w, http.StatusBadRequest, "Invalid JSON", h.logger)
        return
    }
    
    if req.Name == "" {
        RespondWithError(w, http.StatusBadRequest, "Name required", h.logger)
        return  
    }
    
    // Process...
}
```

**After:**
```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req Request
    
    // Use validation framework
    validationErrors := ValidateRequest(r, &req, h.validationRules, h.logger)
    if len(validationErrors) > 0 {
        RespondWithValidationError(w, r, validationErrors, h.logger)
        return
    }
    
    // Process with enhanced error handling...
    result, err := h.service.Create(r.Context(), req)
    if err != nil {
        h.handleServiceError(w, r, err, map[string]interface{}{
            "request_data": req,
        })
        return
    }
    
    RespondWithJSON(w, http.StatusCreated, result, h.logger)
}

func (h *Handler) handleServiceError(w http.ResponseWriter, r *http.Request, err error, context map[string]interface{}) {
    switch {
    case errors.Is(err, service.ErrNotFound):
        RespondWithAPIError(w, r, http.StatusNotFound, 
            "Resource not found", ErrCodeNotFound, h.logger, err, context)
    case errors.Is(err, service.ErrConflict):
        RespondWithAPIError(w, r, http.StatusConflict, 
            "Resource conflict", ErrCodeConflict, h.logger, err, context)
    default:
        RespondWithAPIError(w, r, http.StatusInternalServerError, 
            "Internal server error", ErrCodeInternalServer, h.logger, err, context)
    }
}
```

### 2. Testing Error Responses

```go
func TestAdvancedErrorHandling(t *testing.T) {
    // Test validation errors
    t.Run("validation error", func(t *testing.T) {
        body := `{"name": "", "email": "invalid"}`
        resp := makeRequest("POST", "/api/examples", body)
        
        assert.Equal(t, 400, resp.StatusCode)
        
        var apiError APIError
        json.Unmarshal(resp.Body, &apiError)
        
        assert.Equal(t, "VALIDATION_ERROR", apiError.Code)
        assert.Len(t, apiError.ValidationErrors, 2)
        assert.NotEmpty(t, apiError.RequestID)
    })
    
    // Test business logic errors  
    t.Run("conflict error", func(t *testing.T) {
        body := `{"name": "error"}` // Triggers reserved name error
        resp := makeRequest("POST", "/api/examples", body)
        
        assert.Equal(t, 409, resp.StatusCode)
        
        var apiError APIError
        json.Unmarshal(resp.Body, &apiError)
        
        assert.Equal(t, "RESOURCE_CONFLICT", apiError.Code)
        assert.Contains(t, apiError.Message, "reserved")
        assert.Equal(t, "reserved_name", apiError.Details.Context["reason"])
    })
}
```

## Best Practices

### 1. Error Context

Always provide relevant context in error responses:

```go
// Good - provides context for debugging
context := map[string]interface{}{
    "user_id": userID,
    "schedule_id": scheduleID,
    "requested_time": startTime,
    "available_slots": availableSlots,
}
RespondWithAPIError(w, r, http.StatusConflict, "No available slots", 
    ErrCodeConflict, logger, err, context)

// Bad - no context
RespondWithAPIError(w, r, http.StatusConflict, "Conflict", 
    ErrCodeConflict, logger, err, nil)
```

### 2. Validation Rules

Define validation rules declaratively:

```go
var createBookingRules = []ValidationRule{
    {Field: "ScheduleID", Required: true},
    {Field: "StartTime", Required: true},
    {Field: "BuddyName", MaxLen: 100},
    {Field: "Notes", MaxLen: 500},
    {
        Field: "StartTime",
        Custom: func(value interface{}) error {
            if t, ok := value.(time.Time); ok {
                if t.Before(time.Now()) {
                    return fmt.Errorf("start time cannot be in the past")
                }
            }
            return nil
        },
    },
}
```

### 3. Service Error Design

Design service errors for easy API mapping:

```go
// Service layer
var (
    ErrBookingNotFound    = errors.New("booking not found")
    ErrBookingConflict    = errors.New("booking conflict")
    ErrScheduleNotFound   = errors.New("schedule not found")
    ErrUserNotAuthorized  = errors.New("user not authorized")
)

// API layer mapping
func (h *BookingHandler) mapServiceError(err error) (int, string, string) {
    switch {
    case errors.Is(err, ErrBookingNotFound):
        return http.StatusNotFound, "Booking not found", ErrCodeNotFound
    case errors.Is(err, ErrBookingConflict):
        return http.StatusConflict, "Booking conflict", ErrCodeConflict
    case errors.Is(err, ErrScheduleNotFound):
        return http.StatusNotFound, "Schedule not found", ErrCodeNotFound
    case errors.Is(err, ErrUserNotAuthorized):
        return http.StatusForbidden, "Access denied", ErrCodeAuthorization
    default:
        return http.StatusInternalServerError, "Internal error", ErrCodeInternalServer
    }
}
```

## Migration Strategy

1. **Phase 1**: Add middleware stack to new endpoints
2. **Phase 2**: Gradually migrate existing handlers to use enhanced error responses
3. **Phase 3**: Update client applications to handle new error format
4. **Phase 4**: Remove legacy error handling patterns

The system is designed for gradual adoption while maintaining backward compatibility throughout the migration process. 