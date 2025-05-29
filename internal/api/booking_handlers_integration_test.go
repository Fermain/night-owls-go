package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"night-owls-go/internal/api"
	"night-owls-go/internal/auth"
	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookingEndpoints_CreateAndMarkAttendance(t *testing.T) {
	app := newTestApp(t)
	defer app.DB.Close()

	// --- User Registration and Login (to get a token) ---
	userPhone := "+14155550103"
	userName := "Booking User"
	registerPayload := api.RegisterRequest{Phone: userPhone, Name: userName}
	regPayloadBytes, _ := json.Marshal(registerPayload)
	rr := app.makeRequest(t, "POST", "/auth/register", bytes.NewBuffer(regPayloadBytes), "")
	require.Equal(t, http.StatusOK, rr.Code, "Register failed: %s", rr.Body.String())

	// Retrieve OTP from outbox (as in auth tests)
	outboxItems, err := app.Querier.GetPendingOutboxItems(context.Background(), 10)
	require.NoError(t, err)
	require.NotEmpty(t, outboxItems)
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
	require.NotEmpty(t, otpValue, "OTP not found in outbox for booking user")

	verifyPayload := api.VerifyRequest{Phone: userPhone, Code: otpValue}
	verPayloadBytes, _ := json.Marshal(verifyPayload)
	rr = app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(verPayloadBytes), "")
	require.Equal(t, http.StatusOK, rr.Code, "Verify failed: %s", rr.Body.String())
	var verifyResp api.VerifyResponse
	err = json.Unmarshal(rr.Body.Bytes(), &verifyResp)
	require.NoError(t, err)
	userToken := verifyResp.Token
	require.NotEmpty(t, userToken)

	// --- Get User ID from token for later assertions ---
	claims, err := auth.ValidateJWT(userToken, app.Config.JWTSecret)
	require.NoError(t, err)
	authUserID := claims.UserID

	// --- Ensure there's an available shift ---
	// Use one of the seeded schedules. Daily Evening Patrol: '0 18 * * *'
	// Let's find a Monday in January 2025. Jan 6, 2025 is a Monday.
	// Shift time: 2025-01-06 at 18:00.

	dailySchedule, err := app.Querier.GetScheduleByID(context.Background(), 1) // Assuming seeded daily schedule ID is 1
	if errors.Is(err, sql.ErrNoRows) {                                         // If GetScheduleByID is not implemented in mock or returns error
		// Fallback: find it by name if GetScheduleByID is an issue or for robustness
		schedules, errList := app.Querier.ListActiveSchedules(context.Background(), db.ListActiveSchedulesParams{
			Date:   sql.NullTime{Time: time.Now(), Valid: true},
			Date_2: sql.NullTime{Time: time.Now(), Valid: true},
		})
		require.NoError(t, errList)
		found := false
		for _, s := range schedules {
			if s.Name == "Daily Evening Patrol" {
				dailySchedule = s
				found = true
				break
			}
		}
		require.True(t, found, "Daily Evening Patrol schedule not found via ListActiveSchedules")
	} else {
		require.NoError(t, err, "Failed to get daily schedule by ID=1 for test setup")
	}

	// Use a time that matches the cron expression in Africa/Johannesburg timezone
	// Daily Evening Patrol runs at 18:00 in Johannesburg time, which is 16:00 UTC
	shiftStartTimeStr := "2025-01-06T16:00:00Z"
	shiftStartTime, _ := time.Parse(time.RFC3339, shiftStartTimeStr)

	// --- Test POST /bookings (Create Booking) ---
	buddyName := "Test Buddy"
	createBookingReq := api.CreateBookingRequest{
		ScheduleID: dailySchedule.ScheduleID,
		StartTime:  shiftStartTime,
		BuddyName:  &buddyName,
	}
	bookingPayloadBytes, _ := json.Marshal(createBookingReq)
	rr = app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingPayloadBytes), userToken)

	assert.Equal(t, http.StatusCreated, rr.Code, "Create booking failed: %s", rr.Body.String())
	var createdBooking api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &createdBooking)
	require.NoError(t, err)
	assert.Equal(t, dailySchedule.ScheduleID, createdBooking.ScheduleID)
	assert.Equal(t, shiftStartTime, createdBooking.ShiftStart.UTC())
	assert.Equal(t, authUserID, createdBooking.UserID)
	assert.Equal(t, "Test Buddy", createdBooking.BuddyName)
	assert.Nil(t, createdBooking.CheckedInAt)

	// --- Test POST /bookings (Attempt to book same slot - Conflict) ---
	rrConflict := app.makeRequest(t, "POST", "/bookings", bytes.NewBuffer(bookingPayloadBytes), userToken)
	assert.Equal(t, http.StatusConflict, rrConflict.Code, "Expected conflict when booking same slot: %s", rrConflict.Body.String())

	// --- Test POST /bookings/{id}/checkin ---
	checkinPath := fmt.Sprintf("/bookings/%d/checkin", createdBooking.BookingID)
	rr = app.makeRequest(t, "POST", checkinPath, nil, userToken)

	assert.Equal(t, http.StatusOK, rr.Code, "Mark check-in failed: %s", rr.Body.String())
	var updatedBooking api.BookingResponse
	err = json.Unmarshal(rr.Body.Bytes(), &updatedBooking)
	require.NoError(t, err)
	assert.NotNil(t, updatedBooking.CheckedInAt)
	assert.Equal(t, createdBooking.BookingID, updatedBooking.BookingID)

	// --- Test PATCH /bookings/{id}/attendance (by another user - Forbidden) ---
	// Register and login another user
	otherUserPhone := "+14155550104"
	err = app.UserService.RegisterOrLoginUser(context.Background(), otherUserPhone, sql.NullString{String: "Other User", Valid: true})
	require.NoError(t, err, "Failed to register other user")

	// Retrieve OTP for other user (directly from outbox table)
	outboxItemsOther, err := app.Querier.GetPendingOutboxItems(context.Background(), 10)
	require.NoError(t, err)
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
	require.NotEmpty(t, otherOtpValue, "OTP not found in outbox for other user")

	verifyOtherPayload := api.VerifyRequest{Phone: otherUserPhone, Code: otherOtpValue}
	verOtherPayloadBytes, _ := json.Marshal(verifyOtherPayload)
	rrOtherVerify := app.makeRequest(t, "POST", "/auth/verify", bytes.NewBuffer(verOtherPayloadBytes), "")
	require.Equal(t, http.StatusOK, rrOtherVerify.Code, "Verify for other user failed: %s", rrOtherVerify.Body.String())
	var verifyOtherResp api.VerifyResponse
	err = json.Unmarshal(rrOtherVerify.Body.Bytes(), &verifyOtherResp)
	require.NoError(t, err)
	otherUserToken := verifyOtherResp.Token

	rrForbidden := app.makeRequest(t, "POST", checkinPath, nil, otherUserToken)
	assert.Equal(t, http.StatusForbidden, rrForbidden.Code, "Expected forbidden when other user marks check-in: %s", rrForbidden.Body.String())
}

// TODO: Test other booking error cases (invalid schedule, invalid time format, etc.)
// TODO: Test booking with registered buddy (phone lookup success)
