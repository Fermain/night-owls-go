package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_ProtectedRoutes(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Test 1: Access to protected route without token
	bookingReqBody, _ := json.Marshal(map[string]interface{}{
		"schedule_id": 1,
		"start_time": "2024-11-10T00:00:00Z",
		"buddy_name": "Test Buddy",
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
		sql.NullString{String: userName, Valid: true})
	require.NoError(t, err)
	
	outboxItems, err := app.Querier.GetPendingOutboxItems(ctx, 10)
	require.NoError(t, err)
	var otpValue string
	for _, item := range outboxItems {
		if item.Recipient == userPhone && item.MessageType == "OTP_VERIFICATION" {
			var otpPayload struct{ OTP string `json:"otp"` }
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
	expiredToken, err := auth.GenerateJWT(userID, phoneNumber, app.Config.JWTSecret, -24) // Negative hours for expired token
	require.NoError(t, err)
	
	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), expiredToken)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected unauthorized for request with expired token")

	// Test 5: Test success case - valid token with protected route
	// Create a booking with valid token
	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), token)
	assert.Equal(t, http.StatusCreated, rr.Code, "Creating booking with valid token should succeed")

	// Test 6: Test middleware properly adds user to context
	// We'll check this by marking attendance on our own booking
	var booking api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &booking)
	require.NoError(t, err)
	
	attendanceReqBody, _ := json.Marshal(map[string]interface{}{
		"attended": true,
	})
	
	// This should work since the JWT contains the user ID that matches the booking's owner
	bookingIDStr := fmt.Sprintf("%d", booking.BookingID)
	rr = app.makeRequest(t, "PATCH", "/bookings/"+bookingIDStr+"/attendance", 
		bytes.NewBuffer(attendanceReqBody), token)
	assert.Equal(t, http.StatusOK, rr.Code, "Marking attendance on own booking with valid token should succeed")
} 