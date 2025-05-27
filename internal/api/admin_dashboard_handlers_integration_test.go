package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminDashboardHandlers_GetDashboard_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get dashboard metrics
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var dashboard map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &dashboard)
	require.NoError(t, err)

	// Verify dashboard structure
	assert.Contains(t, dashboard, "metrics")
	assert.Contains(t, dashboard, "member_contributions")
	assert.Contains(t, dashboard, "quality_metrics")
	assert.Contains(t, dashboard, "problematic_slots")
	assert.Contains(t, dashboard, "generated_at")

	// Verify metrics structure
	metrics, ok := dashboard["metrics"].(map[string]interface{})
	require.True(t, ok, "metrics should be an object")
	
	expectedMetrics := []string{
		"total_shifts", "booked_shifts", "unfilled_shifts", 
		"checked_in_shifts", "completed_shifts", "fill_rate",
		"check_in_rate", "completion_rate", "next_week_unfilled",
		"this_weekend_status",
	}
	
	for _, metric := range expectedMetrics {
		assert.Contains(t, metrics, metric, "Missing metric: %s", metric)
	}

	// Verify numeric metrics are reasonable
	assert.IsType(t, float64(0), metrics["total_shifts"])
	assert.IsType(t, float64(0), metrics["booked_shifts"])
	assert.IsType(t, float64(0), metrics["unfilled_shifts"])
	assert.IsType(t, float64(0), metrics["fill_rate"])
	assert.IsType(t, float64(0), metrics["check_in_rate"])
	assert.IsType(t, float64(0), metrics["completion_rate"])

	// Verify member contributions structure
	memberContributions, ok := dashboard["member_contributions"].([]interface{})
	require.True(t, ok, "member_contributions should be an array")
	
	if len(memberContributions) > 0 {
		contribution := memberContributions[0].(map[string]interface{})
		expectedFields := []string{
			"user_id", "name", "phone", "shifts_booked",
			"shifts_attended", "shifts_completed", "attendance_rate",
			"completion_rate", "last_shift_date", "contribution_category",
		}
		
		for _, field := range expectedFields {
			assert.Contains(t, contribution, field, "Missing contribution field: %s", field)
		}
	}

	// Verify quality metrics structure
	qualityMetrics, ok := dashboard["quality_metrics"].(map[string]interface{})
	require.True(t, ok, "quality_metrics should be an object")
	
	expectedQualityMetrics := []string{"no_show_rate", "incomplete_rate", "reliability_score"}
	for _, metric := range expectedQualityMetrics {
		assert.Contains(t, qualityMetrics, metric, "Missing quality metric: %s", metric)
		assert.IsType(t, float64(0), qualityMetrics[metric])
	}

	// Verify problematic slots is an array
	problematicSlots, ok := dashboard["problematic_slots"].([]interface{})
	require.True(t, ok, "problematic_slots should be an array")
	assert.NotNil(t, problematicSlots)

	// Verify generated_at is present
	assert.Contains(t, dashboard, "generated_at")
	assert.IsType(t, "", dashboard["generated_at"])
}

func TestAdminDashboardHandlers_GetDashboard_MetricsValidation(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get dashboard metrics
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code)

	var dashboard map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &dashboard)
	require.NoError(t, err)

	metrics := dashboard["metrics"].(map[string]interface{})

	// Validate metric ranges and relationships
	totalShifts := metrics["total_shifts"].(float64)
	bookedShifts := metrics["booked_shifts"].(float64)
	unfilledShifts := metrics["unfilled_shifts"].(float64)
	fillRate := metrics["fill_rate"].(float64)

	// Basic validation
	assert.GreaterOrEqual(t, totalShifts, float64(0), "Total shifts should be non-negative")
	assert.GreaterOrEqual(t, bookedShifts, float64(0), "Booked shifts should be non-negative")
	assert.GreaterOrEqual(t, unfilledShifts, float64(0), "Unfilled shifts should be non-negative")
	
	// Relationship validation
	if totalShifts > 0 {
		assert.LessOrEqual(t, bookedShifts, totalShifts, "Booked shifts should not exceed total shifts")
		assert.LessOrEqual(t, unfilledShifts, totalShifts, "Unfilled shifts should not exceed total shifts")
		
		// Fill rate should be between 0 and 100
		assert.GreaterOrEqual(t, fillRate, float64(0), "Fill rate should be at least 0%")
		assert.LessOrEqual(t, fillRate, float64(100), "Fill rate should not exceed 100%")
	}

	// Validate percentage metrics
	checkInRate := metrics["check_in_rate"].(float64)
	completionRate := metrics["completion_rate"].(float64)
	
	assert.GreaterOrEqual(t, checkInRate, float64(0), "Check-in rate should be at least 0%")
	assert.LessOrEqual(t, checkInRate, float64(100), "Check-in rate should not exceed 100%")
	assert.GreaterOrEqual(t, completionRate, float64(0), "Completion rate should be at least 0%")
	assert.LessOrEqual(t, completionRate, float64(100), "Completion rate should not exceed 100%")
}

func TestAdminDashboardHandlers_GetDashboard_QualityMetricsValidation(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get dashboard metrics
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code)

	var dashboard map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &dashboard)
	require.NoError(t, err)

	qualityMetrics := dashboard["quality_metrics"].(map[string]interface{})

	// Validate quality metric ranges
	noShowRate := qualityMetrics["no_show_rate"].(float64)
	incompleteRate := qualityMetrics["incomplete_rate"].(float64)
	reliabilityScore := qualityMetrics["reliability_score"].(float64)

	// All rates should be percentages (0-100)
	assert.GreaterOrEqual(t, noShowRate, float64(0), "No-show rate should be at least 0%")
	assert.LessOrEqual(t, noShowRate, float64(100), "No-show rate should not exceed 100%")
	
	assert.GreaterOrEqual(t, incompleteRate, float64(0), "Incomplete rate should be at least 0%")
	assert.LessOrEqual(t, incompleteRate, float64(100), "Incomplete rate should not exceed 100%")
	
	assert.GreaterOrEqual(t, reliabilityScore, float64(0), "Reliability score should be at least 0%")
	assert.LessOrEqual(t, reliabilityScore, float64(100), "Reliability score should not exceed 100%")
}

func TestAdminDashboardHandlers_GetDashboard_MemberContributionsStructure(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get dashboard metrics
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code)

	var dashboard map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &dashboard)
	require.NoError(t, err)

	memberContributions := dashboard["member_contributions"].([]interface{})
	
	// If there are member contributions, validate their structure
	if len(memberContributions) > 0 {
		contribution := memberContributions[0].(map[string]interface{})
		
		// Validate data types
		assert.IsType(t, float64(0), contribution["user_id"])
		assert.IsType(t, "", contribution["name"])
		assert.IsType(t, "", contribution["phone"])
		assert.IsType(t, float64(0), contribution["shifts_booked"])
		assert.IsType(t, float64(0), contribution["shifts_attended"])
		assert.IsType(t, float64(0), contribution["shifts_completed"])
		assert.IsType(t, float64(0), contribution["attendance_rate"])
		assert.IsType(t, float64(0), contribution["completion_rate"])
		assert.IsType(t, "", contribution["last_shift_date"])
		assert.IsType(t, "", contribution["contribution_category"])
		
		// Validate ranges
		attendanceRate := contribution["attendance_rate"].(float64)
		completionRate := contribution["completion_rate"].(float64)
		
		assert.GreaterOrEqual(t, attendanceRate, float64(0), "Attendance rate should be at least 0%")
		assert.LessOrEqual(t, attendanceRate, float64(100), "Attendance rate should not exceed 100%")
		assert.GreaterOrEqual(t, completionRate, float64(0), "Completion rate should be at least 0%")
		assert.LessOrEqual(t, completionRate, float64(100), "Completion rate should not exceed 100%")
		
		// Validate contribution category
		category := contribution["contribution_category"].(string)
		validCategories := []string{"excellent_contributor", "good_contributor", "fair_contributor", "poor_contributor", "inactive"}
		assert.Contains(t, validCategories, category, "Invalid contribution category")
	}
}

func TestAdminDashboardHandlers_GetDashboard_ResponseTime(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test response time (should be fast for dashboard)
	start := time.Now()
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, adminToken)
	duration := time.Since(start)

	require.Equal(t, http.StatusOK, rr.Code)
	
	// Dashboard should respond quickly (under 1 second for simple implementation)
	assert.Less(t, duration.Milliseconds(), int64(1000), "Dashboard response should be fast")
}

func TestAdminDashboardHandlers_Unauthorized_NonAdmin(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create non-admin user
	_, userToken := app.createTestUserAndLogin(t, "+15550001001", "Regular User", "owl")

	// Test that non-admin cannot access admin dashboard
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, userToken)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAdminDashboardHandlers_Unauthorized_NoToken(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Test that requests without token are rejected
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAdminDashboardHandlers_GetDashboard_JSONStructure(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get dashboard metrics
	rr := app.makeRequest(t, "GET", "/api/admin/dashboard", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify it's valid JSON
	var dashboard interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &dashboard)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
} 