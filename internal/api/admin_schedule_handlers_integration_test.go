package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"night-owls-go/internal/api"
	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminScheduleHandlers_ListSchedules_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test schedules
	ctx := context.Background()
	schedule1, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Morning Watch",
		CronExpr:        "0 6 * * *",
		DurationMinutes: 120,
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)

	schedule2, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Evening Watch",
		CronExpr:        "0 18 * * *",
		DurationMinutes: 180,
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)

	// Test list schedules
	rr := app.makeRequest(t, "GET", "/api/admin/schedules", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var schedules []api.ScheduleResponse
	err = json.Unmarshal(rr.Body.Bytes(), &schedules)
	require.NoError(t, err)

	// Should have at least 2 schedules
	assert.GreaterOrEqual(t, len(schedules), 2)

	// Find our test schedules
	var foundSchedule1, foundSchedule2 bool
	for _, schedule := range schedules {
		switch schedule.ScheduleID {
		case schedule1.ScheduleID:
			foundSchedule1 = true
			assert.Equal(t, "Morning Watch", schedule.Name)
			assert.Equal(t, "0 6 * * *", schedule.CronExpr)
			assert.Equal(t, int64(120), schedule.DurationMinutes)
		case schedule2.ScheduleID:
			foundSchedule2 = true
			assert.Equal(t, "Evening Watch", schedule.Name)
			assert.Equal(t, "0 18 * * *", schedule.CronExpr)
			assert.Equal(t, int64(180), schedule.DurationMinutes)
		}
	}
	assert.True(t, foundSchedule1, "Schedule 1 not found in response")
	assert.True(t, foundSchedule2, "Schedule 2 not found in response")
}

func TestAdminScheduleHandlers_GetSchedule_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test schedule
	ctx := context.Background()
	testSchedule, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Test Schedule",
		CronExpr:        "0 12 * * *",
		DurationMinutes: 90,
		Timezone:        sql.NullString{String: "America/New_York", Valid: true},
		StartDate:       sql.NullTime{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		EndDate:         sql.NullTime{Time: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), Valid: true},
	})
	require.NoError(t, err)

	// Test get schedule by ID
	rr := app.makeRequest(t, "GET", fmt.Sprintf("/api/admin/schedules/%d", testSchedule.ScheduleID), nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var schedule api.ScheduleResponse
	err = json.Unmarshal(rr.Body.Bytes(), &schedule)
	require.NoError(t, err)

	assert.Equal(t, testSchedule.ScheduleID, schedule.ScheduleID)
	assert.Equal(t, "Test Schedule", schedule.Name)
	assert.Equal(t, "0 12 * * *", schedule.CronExpr)
	assert.Equal(t, int64(90), schedule.DurationMinutes)
	assert.Equal(t, "America/New_York", schedule.Timezone)
	assert.NotNil(t, schedule.StartDate)
	assert.NotNil(t, schedule.EndDate)
}

func TestAdminScheduleHandlers_GetSchedule_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get non-existent schedule
	rr := app.makeRequest(t, "GET", "/api/admin/schedules/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminScheduleHandlers_GetSchedule_InvalidID(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get schedule with invalid ID
	rr := app.makeRequest(t, "GET", "/api/admin/schedules/invalid", nil, adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_CreateSchedule_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create schedule
	createReq := map[string]interface{}{
		"name":      "New Schedule",
		"cron_expr": "0 14 * * *",
		"timezone":  "Europe/London",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

	var schedule api.ScheduleResponse
	err := json.Unmarshal(rr.Body.Bytes(), &schedule)
	require.NoError(t, err)

	assert.Equal(t, "New Schedule", schedule.Name)
	assert.Equal(t, "0 14 * * *", schedule.CronExpr)
	assert.Equal(t, "Europe/London", schedule.Timezone)

	// Verify schedule was created in database
	ctx := context.Background()
	dbSchedule, err := app.Querier.GetScheduleByID(ctx, schedule.ScheduleID)
	require.NoError(t, err)
	assert.Equal(t, "New Schedule", dbSchedule.Name)
	assert.Equal(t, "0 14 * * *", dbSchedule.CronExpr)
	assert.Equal(t, "Europe/London", dbSchedule.Timezone.String)
}

func TestAdminScheduleHandlers_CreateSchedule_WithDates(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create schedule with start and end dates
	createReq := map[string]interface{}{
		"name":       "Seasonal Schedule",
		"cron_expr":  "0 20 * * *",
		"timezone":   "UTC",
		"start_date": "2025-06-01",
		"end_date":   "2025-08-31",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

	var schedule api.ScheduleResponse
	err := json.Unmarshal(rr.Body.Bytes(), &schedule)
	require.NoError(t, err)

	assert.Equal(t, "Seasonal Schedule", schedule.Name)
	assert.NotNil(t, schedule.StartDate)
	assert.NotNil(t, schedule.EndDate)

	// Verify dates are parsed correctly (API returns YYYY-MM-DD format)
	startDate, err := time.Parse("2006-01-02", *schedule.StartDate)
	require.NoError(t, err)
	assert.Equal(t, 2025, startDate.Year())
	assert.Equal(t, time.June, startDate.Month())
	assert.Equal(t, 1, startDate.Day())

	endDate, err := time.Parse("2006-01-02", *schedule.EndDate)
	require.NoError(t, err)
	assert.Equal(t, 2025, endDate.Year())
	assert.Equal(t, time.August, endDate.Month())
	assert.Equal(t, 31, endDate.Day())
}

func TestAdminScheduleHandlers_CreateSchedule_InvalidCron(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create schedule with invalid cron expression
	createReq := map[string]interface{}{
		"name":      "Invalid Schedule",
		"cron_expr": "invalid cron",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_CreateSchedule_InvalidTimezone(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create schedule with invalid timezone
	createReq := map[string]interface{}{
		"name":      "Invalid Timezone Schedule",
		"cron_expr": "0 12 * * *",
		"timezone":  "Invalid/Timezone",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_CreateSchedule_InvalidDate(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create schedule with invalid date format
	createReq := map[string]interface{}{
		"name":       "Invalid Date Schedule",
		"cron_expr":  "0 12 * * *",
		"start_date": "invalid-date",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_CreateSchedule_MissingFields(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create schedule without required fields
	createReq := map[string]interface{}{
		"timezone": "UTC",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_UpdateSchedule_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test schedule
	ctx := context.Background()
	testSchedule, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Original Schedule",
		CronExpr:        "0 10 * * *",
		DurationMinutes: 60,
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)

	// Test update schedule
	updateReq := map[string]interface{}{
		"name":      "Updated Schedule",
		"cron_expr": "0 16 * * *",
		"timezone":  "America/Los_Angeles",
	}
	reqBytes, _ := json.Marshal(updateReq)

	rr := app.makeRequest(t, "PUT", fmt.Sprintf("/api/admin/schedules/%d", testSchedule.ScheduleID), bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var schedule api.ScheduleResponse
	err = json.Unmarshal(rr.Body.Bytes(), &schedule)
	require.NoError(t, err)

	assert.Equal(t, testSchedule.ScheduleID, schedule.ScheduleID)
	assert.Equal(t, "Updated Schedule", schedule.Name)
	assert.Equal(t, "0 16 * * *", schedule.CronExpr)
	assert.Equal(t, "America/Los_Angeles", schedule.Timezone)

	// Verify schedule was updated in database
	dbSchedule, err := app.Querier.GetScheduleByID(ctx, testSchedule.ScheduleID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Schedule", dbSchedule.Name)
	assert.Equal(t, "0 16 * * *", dbSchedule.CronExpr)
	assert.Equal(t, "America/Los_Angeles", dbSchedule.Timezone.String)
}

func TestAdminScheduleHandlers_UpdateSchedule_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test update non-existent schedule
	updateReq := map[string]interface{}{
		"name":      "Updated Schedule",
		"cron_expr": "0 16 * * *",
	}
	reqBytes, _ := json.Marshal(updateReq)

	rr := app.makeRequest(t, "PUT", "/api/admin/schedules/99999", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminScheduleHandlers_DeleteSchedule_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test schedule
	ctx := context.Background()
	testSchedule, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Schedule to Delete",
		CronExpr:        "0 8 * * *",
		DurationMinutes: 120,
	})
	require.NoError(t, err)

	// Test delete schedule
	rr := app.makeRequest(t, "DELETE", fmt.Sprintf("/api/admin/schedules/%d", testSchedule.ScheduleID), nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Schedule deleted successfully", response["message"])

	// Verify schedule was deleted from database
	_, err = app.Querier.GetScheduleByID(ctx, testSchedule.ScheduleID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestAdminScheduleHandlers_DeleteSchedule_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test delete non-existent schedule
	rr := app.makeRequest(t, "DELETE", "/api/admin/schedules/99999", nil, adminToken)
	// Note: Depending on implementation, this might return 404 or 200 (idempotent delete)
	assert.True(t, rr.Code == http.StatusNotFound || rr.Code == http.StatusOK)
}

func TestAdminScheduleHandlers_BulkDeleteSchedules_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test schedules
	ctx := context.Background()
	schedule1, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Schedule 1",
		CronExpr:        "0 9 * * *",
		DurationMinutes: 60,
	})
	require.NoError(t, err)

	schedule2, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Schedule 2",
		CronExpr:        "0 15 * * *",
		DurationMinutes: 90,
	})
	require.NoError(t, err)

	// Test bulk delete
	deleteReq := map[string]interface{}{
		"schedule_ids": []int64{schedule1.ScheduleID, schedule2.ScheduleID},
	}
	reqBytes, _ := json.Marshal(deleteReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules/bulk-delete", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Schedules deleted successfully", response["message"])

	// Verify schedules were deleted from database
	_, err = app.Querier.GetScheduleByID(ctx, schedule1.ScheduleID)
	assert.Equal(t, sql.ErrNoRows, err)
	_, err = app.Querier.GetScheduleByID(ctx, schedule2.ScheduleID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestAdminScheduleHandlers_BulkDeleteSchedules_EmptyList(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test bulk delete with empty list
	deleteReq := map[string]interface{}{
		"schedule_ids": []int64{},
	}
	reqBytes, _ := json.Marshal(deleteReq)

	rr := app.makeRequest(t, "POST", "/api/admin/schedules/bulk-delete", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_ListAllShiftSlots_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test schedule
	ctx := context.Background()
	_, err := app.Querier.CreateSchedule(ctx, db.CreateScheduleParams{
		Name:            "Test Schedule",
		CronExpr:        "0 12 * * *", // Daily at noon
		DurationMinutes: 120,
		Timezone:        sql.NullString{String: "UTC", Valid: true},
	})
	require.NoError(t, err)

	// Test list all shift slots
	rr := app.makeRequest(t, "GET", "/api/admin/schedules/all-slots", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	// The response format depends on the service implementation
	// For now, just verify we get a valid JSON response
	var slots interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &slots)
	require.NoError(t, err)
}

func TestAdminScheduleHandlers_ListAllShiftSlots_WithParams(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test with query parameters
	fromTime := time.Now().UTC().Format(time.RFC3339)
	toTime := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)

	url := fmt.Sprintf("/api/admin/schedules/all-slots?from=%s&to=%s&limit=10", fromTime, toTime)
	rr := app.makeRequest(t, "GET", url, nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var slots interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &slots)
	require.NoError(t, err)
}

func TestAdminScheduleHandlers_ListAllShiftSlots_InvalidParams(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test with invalid date format
	rr := app.makeRequest(t, "GET", "/api/admin/schedules/all-slots?from=invalid-date", nil, adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Test with invalid limit
	rr = app.makeRequest(t, "GET", "/api/admin/schedules/all-slots?limit=invalid", nil, adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminScheduleHandlers_Unauthorized_NonAdmin(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create non-admin user
	_, userToken := app.createTestUserAndLogin(t, "+15550001001", "Regular User", "owl")

	// Test that non-admin cannot access admin endpoints
	endpoints := []struct {
		method string
		path   string
		body   []byte
	}{
		{"GET", "/api/admin/schedules", nil},
		{"GET", "/api/admin/schedules/1", nil},
		{"POST", "/api/admin/schedules", []byte(`{"name":"Test","cron_expr":"0 12 * * *"}`)},
		{"PUT", "/api/admin/schedules/1", []byte(`{"name":"Test","cron_expr":"0 12 * * *"}`)},
		{"DELETE", "/api/admin/schedules/1", nil},
		{"POST", "/api/admin/schedules/bulk-delete", []byte(`{"schedule_ids":[1,2]}`)},
		{"GET", "/api/admin/schedules/all-slots", nil},
	}

	for _, endpoint := range endpoints {
		var body io.Reader
		if endpoint.body != nil {
			body = bytes.NewBuffer(endpoint.body)
		}
		rr := app.makeRequest(t, endpoint.method, endpoint.path, body, userToken)
		assert.Equal(t, http.StatusForbidden, rr.Code,
			"Expected 403 for %s %s, got %d", endpoint.method, endpoint.path, rr.Code)
	}
}

func TestAdminScheduleHandlers_Unauthorized_NoToken(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Test that requests without token are rejected
	rr := app.makeRequest(t, "GET", "/api/admin/schedules", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
