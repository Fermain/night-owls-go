package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url" // For URL query parameters
	"testing"

	"night-owls-go/internal/api" // For claims if needed for other tests
	"night-owls-go/internal/service"

	// For AvailableShiftSlot struct
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportCreationAndValidation(t *testing.T) { // Renamed to fix redeclaration issue
	app := newTestApp(t)
	defer app.DB.Close()

	// --- User Registration and Login (to get a token for protected report endpoint) ---
	userPhone := "+442079460005" // Valid UK-style number
	userName := "Report User UK"
	registerPayload := api.RegisterRequest{Phone: userPhone, Name: userName}
	regPayloadBytes, _ := json.Marshal(registerPayload)
	rr := app.makeRequest(t, "POST", "/auth/register", bytes.NewBuffer(regPayloadBytes), "")
	require.Equal(t, http.StatusOK, rr.Code, "Register failed: %s", rr.Body.String())

	outboxItems, err := app.Querier.GetPendingOutboxItems(context.Background(), 10)
	require.NoError(t, err)
	var otpValue string
	foundOTP := false
	for _, item := range outboxItems {
		if item.Recipient == userPhone && item.MessageType == "OTP_VERIFICATION" {
			var otpPayload struct {
				OTP string `json:"otp"`
			}
			err = json.Unmarshal([]byte(item.Payload.String), &otpPayload)
			require.NoError(t, err)
			otpValue = otpPayload.OTP
			foundOTP = true
			break
		}
	}
	require.True(t, foundOTP, "OTP not found for report user %s in outbox", userPhone)
	require.NotEmpty(t, otpValue)

	verifyPayload := api.VerifyRequest{Phone: userPhone, Code: otpValue}
	verPayloadBytes, _ := json.Marshal(verifyPayload)
	rr = app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(verPayloadBytes), "")
	require.Equal(t, http.StatusOK, rr.Code, "Verify failed: %s", rr.Body.String())
	var verifyResp api.VerifyResponse
	err = json.Unmarshal(rr.Body.Bytes(), &verifyResp)
	require.NoError(t, err)
	userToken := verifyResp.Token
	require.NotEmpty(t, userToken)

	// --- Setup: Book a known valid shift for the report ---
	// Use one of the seeded schedules. Daily Evening Patrol (ID 1): '0 18 * * *'
	// Active Jan 1, 2025 - Dec 31, 2025.
	// Target shift: Monday, Jan 6, 2025, at 18:00 Johannesburg time (16:00 UTC)
	targetScheduleID := int64(1)                // Assuming Daily Evening Patrol is ID 1 from seed

	createBookingReq := api.CreateBookingRequest{
		ShiftID: targetScheduleID, // Using schedule ID as shift ID for migration
	}
	bookingPayloadBytes, _ := json.Marshal(createBookingReq)
	rrBooking := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingPayloadBytes), userToken)
	require.Equal(t, http.StatusCreated, rrBooking.Code, "Booking the chosen slot failed for report test: %s", rrBooking.Body.String())
	var createdBooking api.BookingResponse
	err = json.Unmarshal(rrBooking.Body.Bytes(), &createdBooking)
	require.NoError(t, err)

	// --- Test POST /bookings/{id}/report ---
	reportReq := api.CreateReportRequest{
		Severity: 1, // Moderate
		Message:  "Found a lost cat near the park entrance.",
	}
	reportPayloadBytes, _ := json.Marshal(reportReq)
	reportPath := fmt.Sprintf("/bookings/%d/report", createdBooking.BookingID)
	rrReport := app.makeRequest(t, "POST", reportPath, bytes.NewBuffer(reportPayloadBytes), userToken)

	assert.Equal(t, http.StatusCreated, rrReport.Code, "Create report failed: %s", rrReport.Body.String())
	var createdReport api.ReportResponse
	err = json.Unmarshal(rrReport.Body.Bytes(), &createdReport)
	require.NoError(t, err)
	assert.Equal(t, createdBooking.BookingID, createdReport.BookingID)
	assert.Equal(t, int64(reportReq.Severity), createdReport.Severity)
	assert.Equal(t, reportReq.Message, createdReport.Message)

	// --- Test POST /bookings/{id}/report (Invalid severity) ---
	invalidReportReq := api.CreateReportRequest{Severity: 5, Message: "Invalid severity report."}
	invalidReportPayloadBytes, _ := json.Marshal(invalidReportReq)
	rrInvalidReport := app.makeRequest(t, "POST", reportPath, bytes.NewBuffer(invalidReportPayloadBytes), userToken)
	assert.Equal(t, http.StatusBadRequest, rrInvalidReport.Code, "Expected 400 for invalid severity: %s", rrInvalidReport.Body.String())

	// --- Test POST /bookings/{id}/report (For a booking not owned by user) ---
	otherUserPhone := "+14155550999" // Using US format that passes validation
	err = app.UserService.RegisterOrLoginUser(context.Background(), otherUserPhone, sql.NullString{String: "Another Reporter", Valid: true})
	require.NoError(t, err, "Failed to register other user")

	// Get a fresh look at outbox items for the other user
	outboxItemsOther, err := app.Querier.GetPendingOutboxItems(context.Background(), 10)
	require.NoError(t, err)

	// Find the OTP for the other user
	var otherOtpValue string
	for _, item := range outboxItemsOther {
		if item.Recipient == otherUserPhone && item.MessageType == "OTP_VERIFICATION" {
			var otpPayload struct {
				OTP string `json:"otp"`
			}
			err = json.Unmarshal([]byte(item.Payload.String), &otpPayload)
			require.NoError(t, err)
			otherOtpValue = otpPayload.OTP
			break
		}
	}
	require.NotEmpty(t, otherOtpValue, "OTP not found for other reporter %s", otherUserPhone)

	// Verify the other user to get their token
	verifyOtherPayload := api.VerifyRequest{Phone: otherUserPhone, Code: otherOtpValue}
	verOtherPayloadBytes, _ := json.Marshal(verifyOtherPayload)
	rrOtherVerify := app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(verOtherPayloadBytes), "")
	require.Equal(t, http.StatusOK, rrOtherVerify.Code, "Verify for other user failed: %s", rrOtherVerify.Body.String())

	var verifyOtherResp api.VerifyResponse
	err = json.Unmarshal(rrOtherVerify.Body.Bytes(), &verifyOtherResp)
	require.NoError(t, err)
	otherUserToken := verifyOtherResp.Token
	require.NotEmpty(t, otherUserToken, "Token for other user should not be empty")

	rrForbiddenReport := app.makeRequest(t, "POST", reportPath, bytes.NewBuffer(reportPayloadBytes), otherUserToken)
	assert.Equal(t, http.StatusForbidden, rrForbiddenReport.Code, "Expected 403 when reporting on non-owned booking: %s", rrForbiddenReport.Body.String())
}

func TestShiftsAvailable_FilteringAndLimits(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// Seeded schedules:
	// ID 1: Daily Evening Patrol (Jan 1, 2025 - Dec 31, 2025), Cron: '0 18 * * *' (Daily 18:00)
	// ID 2: Weekend Morning Watch (Jan 1, 2025 - Dec 31, 2025), Cron: '0 6,10 * * 6,0' (Sat/Sun 06:00, 10:00)
	// ID 3: Weekday Lunch Security (Jan 1, 2025 - Dec 31, 2025), Cron: '0 12 * * 1-5' (Mon-Fri 12:00)

	// Test Case 1: Query specific range for Daily Evening Patrol
	fromJan2025 := "2025-01-06T00:00:00Z"
	toJan2025 := "2025-01-10T23:59:59Z" // Mon Jan 6 to Fri Jan 10
	qParams := url.Values{}
	qParams.Add("from", fromJan2025)
	qParams.Add("to", toJan2025)

	rr := app.makeRequest(t, "GET", "/shifts/available?"+qParams.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	var slotsJan []service.AvailableShiftSlot
	err := json.Unmarshal(rr.Body.Bytes(), &slotsJan)
	require.NoError(t, err)
	// Expect multiple slots: Daily Evening Patrol (5 days @ 18:00), Weekday Lunch Security (5 days @ 12:00), Weekend Morning Watch (Sat/Sun @ 6:00, 10:00)
	assert.GreaterOrEqual(t, len(slotsJan), 10, "Expected at least 10 slots for Jan 6-10 range")
	if len(slotsJan) >= 1 {
		// Should have Daily Evening Patrol slots
		foundDailySlot := false
		for _, slot := range slotsJan {
			if slot.ScheduleName == "Daily Evening Patrol" {
				foundDailySlot = true
				break
			}
		}
		assert.True(t, foundDailySlot, "Expected to find Daily Evening Patrol slots")
	}

	// Test Case 2: Query with limit
	qParamsLimit := url.Values{}
	qParamsLimit.Add("from", "2025-01-01T00:00:00Z") // Full January 2025
	qParamsLimit.Add("to", "2025-01-31T23:59:59Z")
	qParamsLimit.Add("limit", "3")
	rr = app.makeRequest(t, "GET", "/shifts/available?"+qParamsLimit.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	var slotsLimit []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &slotsLimit)
	require.NoError(t, err)
	assert.Len(t, slotsLimit, 3, "Expected 3 slots due to limit for Jan 2025")

	// Test Case 3: Query range where schedules are not active due to their own start/end dates
	qParamsFarFuture := url.Values{}
	qParamsFarFuture.Add("from", "2030-01-01T00:00:00Z")
	qParamsFarFuture.Add("to", "2030-01-07T23:59:59Z")
	rr = app.makeRequest(t, "GET", "/shifts/available?"+qParamsFarFuture.Encode(), nil, "")
	assert.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())
	var slotsFarFuture []service.AvailableShiftSlot
	err = json.Unmarshal(rr.Body.Bytes(), &slotsFarFuture)
	require.NoError(t, err)
	assert.Empty(t, slotsFarFuture, "Expected no slots for a far future date range where seeded schedules are not active")
}

func TestSchedulesEndpoint(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	rr := app.makeRequest(t, "GET", "/schedules", nil, "")
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp []api.ScheduleResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp, "Expected non-empty schedules response")
}

// TODO for this file was:
// TODO: Test GET /schedules (currently placeholder in handler) - Partially done by TestSchedulesEndpoint
// TODO: More detailed tests for /shifts/available (filtering, limits, no active schedules, time window effects)
