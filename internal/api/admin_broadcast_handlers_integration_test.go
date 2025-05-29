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

func TestAdminBroadcastHandlers_CreateBroadcast_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create some test users for recipient count calculation
	ctx := context.Background()
	_, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Test Owl 1", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	_, err = app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001003",
		Name:  sql.NullString{String: "Test Owl 2", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	// Test create immediate broadcast
	createReq := map[string]interface{}{
		"message":      "Test broadcast message",
		"audience":     "all",
		"push_enabled": true,
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

	var broadcast api.BroadcastResponse
	err = json.Unmarshal(rr.Body.Bytes(), &broadcast)
	require.NoError(t, err)

	assert.Equal(t, "Test broadcast message", broadcast.Message)
	assert.Equal(t, "all", broadcast.Audience)
	assert.True(t, broadcast.PushEnabled)
	assert.Equal(t, "pending", broadcast.Status)
	assert.Equal(t, int64(3), broadcast.RecipientCount) // admin + 2 owls
	assert.Nil(t, broadcast.ScheduledAt)
	assert.Nil(t, broadcast.SentAt)
	assert.NotZero(t, broadcast.CreatedAt)
}

func TestAdminBroadcastHandlers_CreateBroadcast_Scheduled(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create scheduled broadcast
	scheduledTime := time.Now().UTC().Add(24 * time.Hour)
	createReq := map[string]interface{}{
		"message":      "Scheduled broadcast message",
		"audience":     "owls",
		"push_enabled": false,
		"scheduled_at": scheduledTime.Format(time.RFC3339),
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

	var broadcast api.BroadcastResponse
	err := json.Unmarshal(rr.Body.Bytes(), &broadcast)
	require.NoError(t, err)

	assert.Equal(t, "Scheduled broadcast message", broadcast.Message)
	assert.Equal(t, "owls", broadcast.Audience)
	assert.False(t, broadcast.PushEnabled)
	assert.Equal(t, "pending", broadcast.Status) // Broadcasts start as pending regardless of scheduling
	assert.NotNil(t, broadcast.ScheduledAt)
	assert.True(t, broadcast.ScheduledAt.Equal(scheduledTime.Truncate(time.Second)))
}

func TestAdminBroadcastHandlers_CreateBroadcast_DifferentAudiences(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test users with different roles
	ctx := context.Background()
	_, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Test Admin 2", Valid: true},
		Role:  sql.NullString{String: "admin", Valid: true},
	})
	require.NoError(t, err)

	_, err = app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001003",
		Name:  sql.NullString{String: "Test Owl", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	testCases := []struct {
		audience      string
		expectedCount int64
		description   string
	}{
		{"all", 3, "all users (2 admins + 1 owl)"},
		{"admins", 2, "admin users only"},
		{"owls", 1, "owl users only"},
		{"active", 3, "active users (currently all users)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			createReq := map[string]interface{}{
				"message":      fmt.Sprintf("Test message for %s", tc.audience),
				"audience":     tc.audience,
				"push_enabled": true,
			}
			reqBytes, _ := json.Marshal(createReq)

			rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer(reqBytes), adminToken)
			require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

			var broadcast api.BroadcastResponse
			err := json.Unmarshal(rr.Body.Bytes(), &broadcast)
			require.NoError(t, err)

			assert.Equal(t, tc.audience, broadcast.Audience)
			assert.Equal(t, tc.expectedCount, broadcast.RecipientCount, "Recipient count mismatch for audience: %s", tc.audience)
		})
	}
}

func TestAdminBroadcastHandlers_CreateBroadcast_ValidationErrors(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	testCases := []struct {
		name        string
		request     map[string]interface{}
		expectedMsg string
	}{
		{
			name:        "missing message",
			request:     map[string]interface{}{"audience": "all", "push_enabled": true},
			expectedMsg: "Message is required",
		},
		{
			name:        "missing audience",
			request:     map[string]interface{}{"message": "Test", "push_enabled": true},
			expectedMsg: "Audience is required",
		},
		{
			name:        "invalid audience",
			request:     map[string]interface{}{"message": "Test", "audience": "invalid", "push_enabled": true},
			expectedMsg: "Invalid audience",
		},
		{
			name:        "empty message",
			request:     map[string]interface{}{"message": "", "audience": "all", "push_enabled": true},
			expectedMsg: "Message is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBytes, _ := json.Marshal(tc.request)
			rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer(reqBytes), adminToken)
			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}

func TestAdminBroadcastHandlers_CreateBroadcast_InvalidJSON(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test invalid JSON
	rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer([]byte("invalid json")), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminBroadcastHandlers_ListBroadcasts_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	adminUser, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test broadcasts
	ctx := context.Background()
	broadcast1, err := app.Querier.CreateBroadcast(ctx, db.CreateBroadcastParams{
		Message:        "First broadcast",
		Audience:       "all",
		SenderUserID:   adminUser.UserID,
		PushEnabled:    true,
		RecipientCount: sql.NullInt64{Int64: 5, Valid: true},
	})
	require.NoError(t, err)

	broadcast2, err := app.Querier.CreateBroadcast(ctx, db.CreateBroadcastParams{
		Message:        "Second broadcast",
		Audience:       "owls",
		SenderUserID:   adminUser.UserID,
		PushEnabled:    false,
		RecipientCount: sql.NullInt64{Int64: 3, Valid: true},
	})
	require.NoError(t, err)

	// Test list broadcasts
	rr := app.makeRequest(t, "GET", "/api/admin/broadcasts", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var broadcasts []api.BroadcastResponse
	err = json.Unmarshal(rr.Body.Bytes(), &broadcasts)
	require.NoError(t, err)

	// Should have at least 2 broadcasts
	assert.GreaterOrEqual(t, len(broadcasts), 2)

	// Find our test broadcasts
	var foundBroadcast1, foundBroadcast2 bool
	for _, broadcast := range broadcasts {
		switch broadcast.BroadcastID {
		case broadcast1.BroadcastID:
			foundBroadcast1 = true
			assert.Equal(t, "First broadcast", broadcast.Message)
			assert.Equal(t, "all", broadcast.Audience)
			assert.True(t, broadcast.PushEnabled)
			assert.Equal(t, int64(5), broadcast.RecipientCount)
			assert.Equal(t, adminUser.UserID, broadcast.SenderUserID)
			assert.Equal(t, "Test Admin", broadcast.SenderName)
		case broadcast2.BroadcastID:
			foundBroadcast2 = true
			assert.Equal(t, "Second broadcast", broadcast.Message)
			assert.Equal(t, "owls", broadcast.Audience)
			assert.False(t, broadcast.PushEnabled)
			assert.Equal(t, int64(3), broadcast.RecipientCount)
		}
	}
	assert.True(t, foundBroadcast1, "Broadcast 1 not found in response")
	assert.True(t, foundBroadcast2, "Broadcast 2 not found in response")
}

func TestAdminBroadcastHandlers_ListBroadcasts_Empty(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test list broadcasts when none exist
	rr := app.makeRequest(t, "GET", "/api/admin/broadcasts", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var broadcasts []api.BroadcastResponse
	err := json.Unmarshal(rr.Body.Bytes(), &broadcasts)
	require.NoError(t, err)

	// Should be empty or contain only pre-existing broadcasts
	// Note: API returns empty slice, not nil
	assert.Equal(t, 0, len(broadcasts))
}

func TestAdminBroadcastHandlers_GetBroadcast_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	adminUser, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test broadcast
	ctx := context.Background()
	scheduledTime := time.Now().UTC().Add(2 * time.Hour)
	testBroadcast, err := app.Querier.CreateBroadcast(ctx, db.CreateBroadcastParams{
		Message:        "Test broadcast for get",
		Audience:       "admins",
		SenderUserID:   adminUser.UserID,
		PushEnabled:    true,
		ScheduledAt:    sql.NullTime{Time: scheduledTime, Valid: true},
		RecipientCount: sql.NullInt64{Int64: 2, Valid: true},
	})
	require.NoError(t, err)

	// Test get broadcast by ID
	rr := app.makeRequest(t, "GET", fmt.Sprintf("/api/admin/broadcasts/%d", testBroadcast.BroadcastID), nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var broadcast api.BroadcastResponse
	err = json.Unmarshal(rr.Body.Bytes(), &broadcast)
	require.NoError(t, err)

	assert.Equal(t, testBroadcast.BroadcastID, broadcast.BroadcastID)
	assert.Equal(t, "Test broadcast for get", broadcast.Message)
	assert.Equal(t, "admins", broadcast.Audience)
	assert.Equal(t, adminUser.UserID, broadcast.SenderUserID)
	assert.True(t, broadcast.PushEnabled)
	assert.Equal(t, int64(2), broadcast.RecipientCount)
	assert.NotNil(t, broadcast.ScheduledAt)
	// Compare times with some tolerance for database precision
	assert.WithinDuration(t, scheduledTime, *broadcast.ScheduledAt, time.Second)
	assert.NotZero(t, broadcast.CreatedAt)
}

func TestAdminBroadcastHandlers_GetBroadcast_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get non-existent broadcast
	rr := app.makeRequest(t, "GET", "/api/admin/broadcasts/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminBroadcastHandlers_GetBroadcast_InvalidID(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get broadcast with invalid ID
	rr := app.makeRequest(t, "GET", "/api/admin/broadcasts/invalid", nil, adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminBroadcastHandlers_BroadcastWorkflow_Complete(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create some test users
	ctx := context.Background()
	_, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Test Owl", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	// Step 1: Create a broadcast
	createReq := map[string]interface{}{
		"message":      "Complete workflow test broadcast",
		"audience":     "all",
		"push_enabled": true,
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusCreated, rr.Code)

	var createdBroadcast api.BroadcastResponse
	err = json.Unmarshal(rr.Body.Bytes(), &createdBroadcast)
	require.NoError(t, err)

	// Step 2: Verify it appears in the list
	rr = app.makeRequest(t, "GET", "/api/admin/broadcasts", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code)

	var broadcasts []api.BroadcastResponse
	err = json.Unmarshal(rr.Body.Bytes(), &broadcasts)
	require.NoError(t, err)

	found := false
	for _, broadcast := range broadcasts {
		if broadcast.BroadcastID == createdBroadcast.BroadcastID {
			found = true
			assert.Equal(t, "Complete workflow test broadcast", broadcast.Message)
			break
		}
	}
	assert.True(t, found, "Created broadcast not found in list")

	// Step 3: Get the specific broadcast
	rr = app.makeRequest(t, "GET", fmt.Sprintf("/api/admin/broadcasts/%d", createdBroadcast.BroadcastID), nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code)

	var retrievedBroadcast api.BroadcastResponse
	err = json.Unmarshal(rr.Body.Bytes(), &retrievedBroadcast)
	require.NoError(t, err)

	assert.Equal(t, createdBroadcast.BroadcastID, retrievedBroadcast.BroadcastID)
	assert.Equal(t, "Complete workflow test broadcast", retrievedBroadcast.Message)
	assert.Equal(t, "all", retrievedBroadcast.Audience)
	assert.Equal(t, int64(2), retrievedBroadcast.RecipientCount) // admin + owl
}

func TestAdminBroadcastHandlers_RecipientCountCalculation(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test users with specific roles
	ctx := context.Background()

	// Create another admin
	_, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Admin 2", Valid: true},
		Role:  sql.NullString{String: "admin", Valid: true},
	})
	require.NoError(t, err)

	// Create owls
	for i := 0; i < 3; i++ {
		_, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
			Phone: fmt.Sprintf("+1555000100%d", i+3),
			Name:  sql.NullString{String: fmt.Sprintf("Owl %d", i+1), Valid: true},
			Role:  sql.NullString{String: "owl", Valid: true},
		})
		require.NoError(t, err)
	}

	// Test recipient count for different audiences
	testCases := []struct {
		audience      string
		expectedCount int64
	}{
		{"all", 5},    // 2 admins + 3 owls
		{"admins", 2}, // 2 admins
		{"owls", 3},   // 3 owls
		{"active", 5}, // all users (for now)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("audience_%s", tc.audience), func(t *testing.T) {
			createReq := map[string]interface{}{
				"message":      fmt.Sprintf("Test for %s", tc.audience),
				"audience":     tc.audience,
				"push_enabled": false,
			}
			reqBytes, _ := json.Marshal(createReq)

			rr := app.makeRequest(t, "POST", "/api/admin/broadcasts", bytes.NewBuffer(reqBytes), adminToken)
			require.Equal(t, http.StatusCreated, rr.Code)

			var broadcast api.BroadcastResponse
			err := json.Unmarshal(rr.Body.Bytes(), &broadcast)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCount, broadcast.RecipientCount,
				"Recipient count mismatch for audience %s", tc.audience)
		})
	}
}

func TestAdminBroadcastHandlers_Unauthorized_NonAdmin(t *testing.T) {
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
		{"GET", "/api/admin/broadcasts", nil},
		{"GET", "/api/admin/broadcasts/1", nil},
		{"POST", "/api/admin/broadcasts", []byte(`{"message":"test","audience":"all","push_enabled":true}`)},
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

func TestAdminBroadcastHandlers_Unauthorized_NoToken(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Test that requests without token are rejected
	rr := app.makeRequest(t, "GET", "/api/admin/broadcasts", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
