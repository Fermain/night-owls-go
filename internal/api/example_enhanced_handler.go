package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"night-owls-go/internal/config"
)

// ExampleRequest demonstrates the validation framework
type ExampleRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Age   int    `json:"age"`
}

// ExampleResponse demonstrates structured API responses
type ExampleResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	RequestID string    `json:"request_id"`
}

// ExampleHandler demonstrates advanced error handling and validation patterns
type ExampleHandler struct {
	logger *slog.Logger
	config *config.Config
}

// NewExampleHandler creates a new example handler
func NewExampleHandler(logger *slog.Logger, cfg *config.Config) *ExampleHandler {
	return &ExampleHandler{
		logger: logger.With("handler", "ExampleHandler"),
		config: cfg,
	}
}

// ValidationRules for the example request
var exampleValidationRules = []ValidationRule{
	{
		Field:    "Name",
		Required: true,
		MinLen:   2,
		MaxLen:   100,
	},
	{
		Field:    "Email",
		Required: true,
		Pattern:  EmailPattern,
	},
	{
		Field:    "Phone",
		Required: true,
		Pattern:  PhonePattern,
	},
	{
		Field: "Age",
		Custom: func(value interface{}) error {
			if age, ok := value.(int); ok {
				if age < 0 || age > 150 {
					return fmt.Errorf("age must be between 0 and 150")
				}
			}
			return nil
		},
	},
}

// CreateExampleWithValidation demonstrates the validation middleware
// @Summary Create example with validation
// @Description Creates an example resource with comprehensive request validation
// @Tags examples
// @Accept json
// @Produce json
// @Param request body ExampleRequest true "Example data"
// @Success 201 {object} ExampleResponse "Example created successfully"
// @Failure 400 {object} APIError "Validation error with detailed field information"
// @Failure 500 {object} APIError "Internal server error with request tracking"
// @Router /examples [post]
func (h *ExampleHandler) CreateExampleWithValidation(w http.ResponseWriter, r *http.Request) {
	// The validation middleware has already validated the request
	var req ExampleRequest
	if !GetValidatedData(r, &req) {
		// This shouldn't happen if middleware is configured correctly
		RespondWithAPIError(w, r, http.StatusInternalServerError, 
			"Failed to get validated data", ErrCodeInternalServer, 
			h.logger, fmt.Errorf("validation middleware not configured"), nil)
		return
	}

	// Simulate processing
	response := ExampleResponse{
		ID:        generateRequestID(), // Reuse the ID generation
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Age:       req.Age,
		CreatedAt: time.Now().UTC(),
		RequestID: GetRequestID(r),
	}

	h.logger.InfoContext(r.Context(), "Example created successfully",
		"example_id", response.ID,
		"name", req.Name,
		"request_id", response.RequestID)

	RespondWithJSON(w, http.StatusCreated, response, h.logger.With("operation", "create_example"))
}

// CreateExampleManualValidation demonstrates manual validation with detailed errors
func (h *ExampleHandler) CreateExampleManualValidation(w http.ResponseWriter, r *http.Request) {
	var req ExampleRequest
	
	// Manual validation using the validation framework
	validationErrors := ValidateRequest(r, &req, exampleValidationRules, h.logger)
	if len(validationErrors) > 0 {
		RespondWithValidationError(w, r, validationErrors, h.logger)
		return
	}

	// Simulate processing with potential business logic error
	if req.Name == "error" {
		// Demonstrate business logic error with context
		context := map[string]interface{}{
			"reason": "reserved_name",
			"name":   req.Name,
		}
		RespondWithAPIError(w, r, http.StatusConflict, 
			"Name 'error' is reserved and cannot be used", ErrCodeConflict,
			h.logger, fmt.Errorf("reserved name: %s", req.Name), context)
		return
	}

	// Simulate internal error
	if req.Name == "panic" {
		panic("Simulated panic for testing error recovery")
	}

	response := ExampleResponse{
		ID:        generateRequestID(),
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Age:       req.Age,
		CreatedAt: time.Now().UTC(),
		RequestID: GetRequestID(r),
	}

	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

// GetExampleWithErrorDemo demonstrates different error scenarios
func (h *ExampleHandler) GetExampleWithErrorDemo(w http.ResponseWriter, r *http.Request) {
	exampleID := r.URL.Query().Get("id")
	errorType := r.URL.Query().Get("error_type")

	switch errorType {
	case "not_found":
		context := map[string]interface{}{
			"example_id": exampleID,
			"searched_at": time.Now().UTC(),
		}
		RespondWithAPIError(w, r, http.StatusNotFound, 
			"Example not found", ErrCodeNotFound,
			h.logger, fmt.Errorf("example %s not found", exampleID), context)
		return
		
	case "forbidden":
		context := map[string]interface{}{
			"example_id": exampleID,
			"user_id": r.Context().Value(UserIDKey),
		}
		RespondWithAPIError(w, r, http.StatusForbidden, 
			"Access denied to this example", ErrCodeAuthorization,
			h.logger, fmt.Errorf("user not authorized for example %s", exampleID), context)
		return
		
	case "internal":
		context := map[string]interface{}{
			"example_id": exampleID,
			"operation": "get_example",
		}
		RespondWithAPIError(w, r, http.StatusInternalServerError, 
			"Internal server error occurred", ErrCodeInternalServer,
			h.logger, fmt.Errorf("database connection failed"), context)
		return
	}

	// Success case
	response := ExampleResponse{
		ID:        exampleID,
		Name:      "Example Item",
		Email:     "example@test.com",
		Phone:     "+1234567890",
		Age:       25,
		CreatedAt: time.Now().Add(-24 * time.Hour).UTC(),
		RequestID: GetRequestID(r),
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// SetupExampleRoutes demonstrates how to set up routes with the new middleware
func (h *ExampleHandler) SetupExampleRoutes(mux *http.ServeMux, cfg *config.Config, logger *slog.Logger) {
	// Route with validation middleware
	validationMiddleware := ValidationMiddleware(exampleValidationRules, ExampleRequest{}, logger)
	
	// Chain multiple middlewares for comprehensive request handling
	createWithValidation := SecurityHeadersMiddleware(cfg)(
		RequestTracingMiddleware(logger)(
			RateLimitMiddleware(60, logger)( // 60 requests per minute
				http.HandlerFunc(validationMiddleware(
					WithErrorRecovery(h.CreateExampleWithValidation, logger))))))

	// Route with manual validation and error recovery
	createManual := SecurityHeadersMiddleware(cfg)(
		RequestTracingMiddleware(logger)(
			RateLimitMiddleware(60, logger)(
				WithErrorRecovery(h.CreateExampleManualValidation, logger))))

	// Route for error demonstration
	getWithErrors := SecurityHeadersMiddleware(cfg)(
		RequestTracingMiddleware(logger)(
			WithErrorRecovery(h.GetExampleWithErrorDemo, logger)))

	// Register routes
	mux.Handle("/api/examples/validated", createWithValidation)
	mux.Handle("/api/examples/manual", createManual)
	mux.Handle("/api/examples/errors", getWithErrors)
}

// GetExampleErrorPatterns provides documentation for error patterns
func GetExampleErrorPatterns() map[string]interface{} {
	return map[string]interface{}{
		"validation_error": map[string]interface{}{
			"status_code": 400,
			"error_code":  ErrCodeValidation,
			"example": APIError{
				Error:     "Bad Request",
				Message:   "Request validation failed",
				Code:      ErrCodeValidation,
				Timestamp: time.Now().UTC(),
				RequestID: "req_example123",
				Path:      "/api/examples",
				Method:    "POST",
				ValidationErrors: []ValidationError{
					{
						Field:   "email",
						Value:   "invalid-email",
						Message: "email format is invalid",
						Code:    "INVALID_FORMAT",
					},
					{
						Field:   "phone",
						Message: "phone is required",
						Code:    "REQUIRED_FIELD",
					},
				},
			},
		},
		"business_logic_error": map[string]interface{}{
			"status_code": 409,
			"error_code":  ErrCodeConflict,
			"example": APIError{
				Error:     "Conflict",
				Message:   "Name 'error' is reserved and cannot be used",
				Code:      ErrCodeConflict,
				Timestamp: time.Now().UTC(),
				RequestID: "req_example456",
				Path:      "/api/examples",
				Method:    "POST",
				Details: &ErrorDetails{
					Type:    "*errors.errorString",
					Context: map[string]interface{}{
						"reason": "reserved_name",
						"name":   "error",
					},
					InternalMsg: "reserved name: error",
				},
			},
		},
		"not_found_error": map[string]interface{}{
			"status_code": 404,
			"error_code":  ErrCodeNotFound,
			"example": APIError{
				Error:     "Not Found",
				Message:   "Example not found",
				Code:      ErrCodeNotFound,
				Timestamp: time.Now().UTC(),
				RequestID: "req_example789",
				Path:      "/api/examples/123",
				Method:    "GET",
				Details: &ErrorDetails{
					Context: map[string]interface{}{
						"example_id":  "123",
						"searched_at": time.Now().UTC(),
					},
				},
			},
		},
		"rate_limit_error": map[string]interface{}{
			"status_code": 429,
			"error_code":  ErrCodeRateLimit,
			"example": APIError{
				Error:     "Too Many Requests",
				Message:   "Rate limit exceeded",
				Code:      ErrCodeRateLimit,
				Timestamp: time.Now().UTC(),
				RequestID: "req_example999",
				Path:      "/api/examples",
				Method:    "POST",
				Details: &ErrorDetails{
					Context: map[string]interface{}{
						"client_ip":  "192.168.1.100",
						"limit":      60,
						"reset_time": time.Now().Add(time.Minute).Format(time.RFC3339),
					},
				},
			},
		},
	}
} 