package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"night-owls-go/internal/api"
	db "night-owls-go/internal/db/sqlc_generated"

	// For AvailableShiftSlot struct
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportEndpoints(t *testing.T) { // Renamed to focus on reports
	app := newTestApp(t)
	defer app.DB.Close()

	// --- User Registration and Login (to get a token for protected report endpoint) ---
	userPhone := "+16554433220"
	userName := "Report User"
	registerPayload := api.RegisterRequest{Phone: userPhone, Name: userName}
	regPayloadBytes, _ := json.Marshal(registerPayload)
	rr := app.makeRequest(t, "POST", "/auth/register", bytes.NewBuffer(regPayloadBytes), "")
	require.Equal(t, http.StatusOK, rr.Code, "Register failed: %s", rr.Body.String())

	outboxItems, err := app.Querier.GetPendingOutboxItems(context.Background(), 10)
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
	require.NotEmpty(t, otpValue, "OTP not found for report user")

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
	// Use one of the seeded schedules. Summer Patrol (ID 1): '0 0,2 * 11-12,1-4 6,0,1'
	// Active Nov 1, 2024 - Apr 30, 2025.
	// Target shift: Monday, Nov 4, 2024, at 00:00:00Z.
	targetScheduleID := int64(1) // Assuming Summer Patrol is ID 1 from seed
	shiftStartTimeStr := "2024-11-04T00:00:00Z"
	shiftStartTime, _ := time.Parse(time.RFC3339, shiftStartTimeStr)

	createBookingReq := api.CreateBookingRequest{
		ScheduleID: targetScheduleID,
		StartTime:  shiftStartTime,
	}
	bookingPayloadBytes, _ := json.Marshal(createBookingReq)
	rrBooking := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingPayloadBytes), userToken)
	require.Equal(t, http.StatusCreated, rrBooking.Code, "Booking the chosen slot failed for report test: %s", rrBooking.Body.String())
	var createdBooking db.Booking
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
	var createdReport db.Report
	err = json.Unmarshal(rrReport.Body.Bytes(), &createdReport)
	require.NoError(t, err)
	assert.Equal(t, createdBooking.BookingID, createdReport.BookingID)
	assert.Equal(t, int64(reportReq.Severity), createdReport.Severity)
	assert.Equal(t, reportReq.Message, createdReport.Message.String)

	// --- Test POST /bookings/{id}/report (Invalid severity) ---
	invalidReportReq := api.CreateReportRequest{ Severity: 5, Message: "Invalid severity report." }
	invalidReportPayloadBytes, _ := json.Marshal(invalidReportReq)
	rrInvalidReport := app.makeRequest(t, "POST", reportPath, bytes.NewBuffer(invalidReportPayloadBytes), userToken)
	assert.Equal(t, http.StatusBadRequest, rrInvalidReport.Code, "Expected 400 for invalid severity: %s", rrInvalidReport.Body.String())

	// --- Test POST /bookings/{id}/report (For a booking not owned by user) ---
	otherUserPhone := "+16554433221"
	err = app.UserService.RegisterOrLoginUser(context.Background(), otherUserPhone, sql.NullString{String:"Another Reporter", Valid:true})
    require.NoError(t, err)

    outboxItemsOther, err := app.Querier.GetPendingOutboxItems(context.Background(), 10)
	require.NoError(t, err)
	var otherOtpValue string
	for _, item := range outboxItemsOther {
		if item.Recipient == otherUserPhone && item.MessageType == "OTP_VERIFICATION" {
			var otpP struct{ OTP string `json:"otp"` }; _ = json.Unmarshal([]byte(item.Payload.String), &otpP); otherOtpValue = otpP.OTP; break
		}
	}
    require.NotEmpty(t, otherOtpValue, "OTP not found for other reporter")
    
	verifyOtherPayload := api.VerifyRequest{Phone: otherUserPhone, Code: otherOtpValue}
	verOtherPayloadBytes, _ := json.Marshal(verifyOtherPayload)
	rrOtherVerify := app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(verOtherPayloadBytes), "")
	require.Equal(t, http.StatusOK, rrOtherVerify.Code)
	var verifyOtherResp api.VerifyResponse; _ = json.Unmarshal(rrOtherVerify.Body.Bytes(), &verifyOtherResp)
	otherUserToken := verifyOtherResp.Token

	rrForbiddenReport := app.makeRequest(t, "POST", reportPath, bytes.NewBuffer(reportPayloadBytes), otherUserToken)
	assert.Equal(t, http.StatusForbidden, rrForbiddenReport.Code, "Expected 403 when reporting on non-owned booking: %s", rrForbiddenReport.Body.String())
}

// TODO: Test GET /schedules (currently placeholder in handler)
// TODO: Write separate, focused tests for /shifts/available (filtering, limits, no active schedules, time window effects) 