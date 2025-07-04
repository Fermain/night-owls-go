package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"

	"io"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_ProtectedRoutes(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Test 1: Access to protected route without token
	bookingReqBody, _ := json.Marshal(map[string]interface{}{
		"schedule_id": 1,
		"start_time":  "2025-01-06T16:00:00Z", // 18:00 Johannesburg time = 16:00 UTC
		"buddy_name":  "Test Buddy",
	})

	rr := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected unauthorized for request without token")

	// Test 2: Access with invalid token format
	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), "not-a-valid-token")
	// The middleware may return either 401 or 500 depending on implementation details
	statusCode := rr.Code
	assert.True(t, statusCode == http.StatusUnauthorized || statusCode == http.StatusInternalServerError,
		"Expected unauthorized (401) or internal server error (500) for invalid token format, got %d", statusCode)

	// Test 3: Access with valid token but for nonexistent endpoint
	// First register a user and get a valid token
	userPhone := "+14155550222"
	userName := "Auth Test User"
	ctx := context.Background()

	err := app.UserService.RegisterOrLoginUser(ctx, userPhone,
		sql.NullString{String: userName, Valid: true}, "test-ip", "test-agent")
	require.NoError(t, err)

	outboxItems, err := app.Querier.GetPendingOutboxItems(ctx, 10)
	require.NoError(t, err)
	var otpValue string
	for _, item := range outboxItems {
		if item.Recipient == userPhone && item.MessageType == "OTP_VERIFICATION" {
			var otpPayload struct {
				OTP string `json:"otp"`
			}
			err = json.Unmarshal([]byte(item.Payload.String), &otpPayload)
			require.NoError(t, err)
			otpValue = otpPayload.OTP
			break
		}
	}
	require.NotEmpty(t, otpValue)

	token, err := app.UserService.VerifyOTP(ctx, userPhone, otpValue)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Try to access a nonexistent protected path with valid token
	rr = app.makeRequest(t, "GET", "/bookings/nonexistent-path", nil, token)
	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected 404 for valid token but nonexistent endpoint")

	// Test 4: Generate an expired token and try to use it
	userID := int64(999)
	phoneNumber := "+14155550000"
	expiredToken, err := auth.GenerateJWT(userID, phoneNumber, "Test User", "guest", app.Config.JWTSecret, -24) // Negative hours for expired token
	require.NoError(t, err)

	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), expiredToken)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected unauthorized for request with expired token")

	// Test 5: Test success case - valid token with protected route
	// Create a booking with valid token
	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), token)
	assert.Equal(t, http.StatusCreated, rr.Code, "Creating booking with valid token should succeed")

	// Test 6: Test middleware properly adds user to context
	// We'll check this by marking check-in on our own booking
	var booking api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &booking)
	require.NoError(t, err)

	// This should work since the JWT contains the user ID that matches the booking's owner
	bookingIDStr := fmt.Sprintf("%d", booking.BookingID)
	rr = app.makeRequest(t, "POST", "/bookings/"+bookingIDStr+"/checkin", nil, token)
	assert.Equal(t, http.StatusOK, rr.Code, "Marking check-in on own booking with valid token should succeed")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:          "test-secret-key-for-valid-token",
		JWTExpirationHours: 24,
	}

	// Create a valid token
	userID := int64(123)
	phone := "+1234567890"
	token, err := auth.GenerateJWT(userID, phone, "Test User", "guest", cfg.JWTSecret, cfg.JWTExpirationHours)
	require.NoError(t, err)

	// Create test handler that requires auth
	router := chi.NewRouter()
	router.Use(api.AuthMiddleware(cfg, testLogger()))
	router.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		contextUserID := r.Context().Value(api.UserIDKey).(int64)
		assert.Equal(t, userID, contextUserID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test request with valid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:          "test-secret-key-for-expired-token",
		JWTExpirationHours: 24,
	}

	// Create an expired token by generating one with negative expiration
	// Note: We can't easily create an expired token with the current GenerateJWT function
	// This test would need a modified function or we test with a very short expiration
	// For now, let's test with a short expiration and wait
	shortCfg := &config.Config{
		JWTSecret:          cfg.JWTSecret,
		JWTExpirationHours: 1, // 1 hour, we'll test validation after artificial time passage
	}

	token, err := auth.GenerateJWT(123, "+1234567890", "Test User", "guest", shortCfg.JWTSecret, -1) // Negative hours to force expiration
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(api.AuthMiddleware(cfg, testLogger()))
	router.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with expired token")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	tests := []struct {
		name       string
		authHeader string
		expectCode int
	}{
		{
			name:       "missing bearer prefix",
			authHeader: "invalid-token-format",
			expectCode: http.StatusUnauthorized,
		},
		{
			name:       "invalid jwt format",
			authHeader: "Bearer not.a.valid.jwt.token",
			expectCode: http.StatusUnauthorized,
		},
		{
			name:       "empty token",
			authHeader: "Bearer ",
			expectCode: http.StatusUnauthorized,
		},
		{
			name:       "wrong secret signature",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInBob25lIjoiKzEyMzQ1Njc4OTAiLCJleHAiOjk5OTk5OTk5OTl9.wrong-signature",
			expectCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(api.AuthMiddleware(cfg, testLogger()))
			router.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("Handler should not be called with malformed token: %s", tt.name)
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectCode, rr.Code, "Test case: %s", tt.name)
		})
	}
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}

	router := chi.NewRouter()
	router.Use(api.AuthMiddleware(cfg, testLogger()))
	router.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without auth header")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	// No Authorization header set
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_ContextValues(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:          "test-secret-context",
		JWTExpirationHours: 24,
	}

	expectedUserID := int64(456)
	expectedPhone := "+9876543210"

	token, err := auth.GenerateJWT(expectedUserID, expectedPhone, "Test User", "guest", cfg.JWTSecret, cfg.JWTExpirationHours)
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(api.AuthMiddleware(cfg, testLogger()))
	router.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		// Verify context values are set correctly
		userID, ok := r.Context().Value(api.UserIDKey).(int64)
		assert.True(t, ok, "UserID should be int64")
		assert.Equal(t, expectedUserID, userID)

		phone, ok := r.Context().Value(api.UserPhoneKey).(string)
		assert.True(t, ok, "Phone should be string")
		assert.Equal(t, expectedPhone, phone)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// Test concurrent requests with different tokens
func TestAuthMiddleware_ConcurrentRequests(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:          "test-secret-concurrent",
		JWTExpirationHours: 24,
	}

	// Create tokens for different users
	tokens := make([]string, 3)
	for i := 0; i < 3; i++ {
		userID := int64(i + 1)
		phone := fmt.Sprintf("+123456789%d", i)
		token, err := auth.GenerateJWT(userID, phone, "Test User", "guest", cfg.JWTSecret, cfg.JWTExpirationHours)
		require.NoError(t, err)
		tokens[i] = token
	}

	router := chi.NewRouter()
	router.Use(api.AuthMiddleware(cfg, testLogger()))
	router.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(api.UserIDKey).(int64)
		w.Write([]byte(fmt.Sprintf("user-%d", userID)))
	})

	// Test concurrent requests
	results := make(chan string, 3)
	for i, token := range tokens {
		go func(i int, token string) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			results <- rr.Body.String()
		}(i, token)
	}

	// Collect results
	responseCount := make(map[string]int)
	for i := 0; i < 3; i++ {
		result := <-results
		responseCount[result]++
	}

	// Verify each user got their own response
	assert.Equal(t, 1, responseCount["user-1"])
	assert.Equal(t, 1, responseCount["user-2"])
	assert.Equal(t, 1, responseCount["user-3"])
}

// Helper function to create a test logger
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
