package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"night-owls-go/internal/service"

	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleEndpoints_GetAvailableShifts(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Test 1: Basic request with no parameters (should return default window of upcoming shifts)
	rr := app.makeRequest(t, "GET", "/shifts/available", nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	
	var shifts []service.AvailableShiftSlot
	err := json.Unmarshal(rr.Body.Bytes(), &shifts)
	require.NoError(t, err)
	
	// Since we're using seeded schedules (Summer and Winter), there should be some shifts
	assert.NotEmpty(t, shifts, "Expected some available shifts with default parameters")
	
	// Validate shift structure
	if len(shifts) > 0 {
		firstShift := shifts[0]
		assert.Greater(t, firstShift.ScheduleID, int64(0), "Schedule ID should be positive")
		assert.NotEmpty(t, firstShift.ScheduleName, "Schedule name should not be empty")
		assert.False(t, firstShift.StartTime.IsZero(), "Start time should be set")
		assert.False(t, firstShift.EndTime.IsZero(), "End time should be set")
		assert.False(t, firstShift.IsBooked, "All returned shifts should be available (not booked)")
		
		// End time should be after start time
		assert.True(t, firstShift.EndTime.After(firstShift.StartTime), 
			"End time (%v) should be after start time (%v)", 
			firstShift.EndTime, firstShift.StartTime)
	}

	// Test 2: Request with specific date range - Summer schedule (Nov-Apr)
	// Our seeded summer schedule runs on weekends and Mondays at 00:00 and 02:00
	// Let's choose a range with known slots
	summerQueryParams := url.Values{}
	summerQueryParams.Add("from", "2024-11-01T00:00:00Z")
	summerQueryParams.Add("to", "2024-11-04T23:59:59Z") // Monday Nov 4, 2024
	
	rr = app.makeRequest(t, "GET", "/shifts/available?"+summerQueryParams.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	
	var summerShifts []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &summerShifts)
	require.NoError(t, err)
	
	// Should have shifts for Fri Nov 1, Sat Nov 2, Sun Nov 3, and Mon Nov 4 (each with slots at 00:00 and 02:00)
	// The actual number may vary based on the cron expression implementation details
	assert.GreaterOrEqual(t, len(summerShifts), 6, 
		"Expected at least 6 shifts for the first 4 days of November 2024")
	
	// Verify all returned shifts are within the requested date range
	for _, shift := range summerShifts {
		assert.True(t, shift.StartTime.After(time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)) || 
					shift.StartTime.Equal(time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)),
			"Shift start time %v should be on or after Nov 1, 2024", shift.StartTime)
		
		assert.True(t, shift.StartTime.Before(time.Date(2024, 11, 5, 0, 0, 0, 0, time.UTC)),
			"Shift start time %v should be before Nov 5, 2024", shift.StartTime)
	}

	// Test 3: Request with limit parameter
	limitParams := url.Values{}
	limitParams.Add("limit", "3")
	
	rr = app.makeRequest(t, "GET", "/shifts/available?"+limitParams.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	
	var limitedShifts []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &limitedShifts)
	require.NoError(t, err)
	
	// Should respect the limit parameter
	assert.LessOrEqual(t, len(limitedShifts), 3, "Number of shifts should respect the limit parameter")

	// Test 4: Request for a date range where no schedule is active
	farFutureParams := url.Values{}
	farFutureParams.Add("from", "2030-01-01T00:00:00Z")
	farFutureParams.Add("to", "2030-01-31T23:59:59Z")
	
	rr = app.makeRequest(t, "GET", "/shifts/available?"+farFutureParams.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	
	var farFutureShifts []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &farFutureShifts)
	require.NoError(t, err)
	
	// Should be empty for a date range with no active schedules
	assert.Empty(t, farFutureShifts, "Expected no shifts for a far future date range")

	// Test 5: Invalid date parameter format
	invalidParams := url.Values{}
	invalidParams.Add("from", "not-a-date")
	
	rr = app.makeRequest(t, "GET", "/shifts/available?"+invalidParams.Encode(), nil, "")
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for invalid date format")
}

func TestScheduleEndpoints_GetAvailableShifts_WithBooking(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Create a user and get token
	ctx := context.Background()
	userPhone := "+14155550111"
	userName := "Shift Test User"
	
	err := app.UserService.RegisterOrLoginUser(ctx, userPhone, 
		sql.NullString{String: userName, Valid: true})
	require.NoError(t, err)
	
	// Get OTP from outbox
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
	
	// Verify OTP to get token
	token, err := app.UserService.VerifyOTP(ctx, userPhone, otpValue)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Check available shifts first
	fromDate := "2024-12-01T00:00:00Z"
	toDate := "2024-12-02T23:59:59Z"
	qParams := url.Values{}
	qParams.Add("from", fromDate)
	qParams.Add("to", toDate)
	
	rr := app.makeRequest(t, "GET", "/shifts/available?"+qParams.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var initialShifts []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &initialShifts)
	require.NoError(t, err)
	require.NotEmpty(t, initialShifts, "Expected some shifts in December 2024")
	
	// Book one of the shifts
	shiftToBook := initialShifts[0]
	bookingReqBody, err := json.Marshal(map[string]interface{}{
		"schedule_id": shiftToBook.ScheduleID,
		"start_time": shiftToBook.StartTime,
		"buddy_name": "Test Buddy",
	})
	require.NoError(t, err)
	
	// Create booking
	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingReqBody), token)
	assert.Equal(t, http.StatusCreated, rr.Code, "Failed to create booking: %s", rr.Body.String())
	
	// Check available shifts again - the booked shift should be excluded
	rr = app.makeRequest(t, "GET", "/shifts/available?"+qParams.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var updatedShifts []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &updatedShifts)
	require.NoError(t, err)
	
	// Verify the booked shift no longer appears in available shifts
	shiftStillAvailable := false
	for _, shift := range updatedShifts {
		if shift.ScheduleID == shiftToBook.ScheduleID && 
		   shift.StartTime.Equal(shiftToBook.StartTime) {
			shiftStillAvailable = true
			break
		}
	}
	assert.False(t, shiftStillAvailable, "Booked shift should no longer appear in available shifts")
	assert.Equal(t, len(initialShifts)-1, len(updatedShifts), 
		"Available shifts should be reduced by exactly 1 after booking")
} 