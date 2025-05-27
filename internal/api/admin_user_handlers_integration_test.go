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

	"night-owls-go/internal/api"
	db "night-owls-go/internal/db/sqlc_generated"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminUserHandlers_ListUsers_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	adminUser, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")
	require.Equal(t, "admin", adminUser.Role)

	// Create some test users
	ctx := context.Background()
	testUser1, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Test User 1", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	testUser2, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001003",
		Name:  sql.NullString{String: "Test User 2", Valid: true},
		Role:  sql.NullString{String: "guest", Valid: true},
	})
	require.NoError(t, err)

	// Test list all users
	rr := app.makeRequest(t, "GET", "/api/admin/users", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var users []api.UserAPIResponse
	err = json.Unmarshal(rr.Body.Bytes(), &users)
	require.NoError(t, err)

	// Should have at least 3 users (admin + 2 test users)
	assert.GreaterOrEqual(t, len(users), 3)

	// Find our test users in the response
	var foundUser1, foundUser2, foundAdmin bool
	for _, user := range users {
		switch user.ID {
		case testUser1.UserID:
			foundUser1 = true
			assert.Equal(t, testUser1.Phone, user.Phone)
			assert.Equal(t, "Test User 1", *user.Name)
			assert.Equal(t, "owl", user.Role)
		case testUser2.UserID:
			foundUser2 = true
			assert.Equal(t, testUser2.Phone, user.Phone)
			assert.Equal(t, "Test User 2", *user.Name)
			assert.Equal(t, "guest", user.Role)
		case adminUser.UserID:
			foundAdmin = true
			assert.Equal(t, adminUser.Phone, user.Phone)
			assert.Equal(t, "Test Admin", *user.Name)
			assert.Equal(t, "admin", user.Role)
		}
	}
	assert.True(t, foundUser1, "Test User 1 not found in response")
	assert.True(t, foundUser2, "Test User 2 not found in response")
	assert.True(t, foundAdmin, "Admin user not found in response")
}

func TestAdminUserHandlers_ListUsers_WithSearch(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test users with specific names for searching
	ctx := context.Background()
	searchUser, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "John Searchable", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	_, err = app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001003",
		Name:  sql.NullString{String: "Jane Different", Valid: true},
		Role:  sql.NullString{String: "guest", Valid: true},
	})
	require.NoError(t, err)

	// Test search by name
	rr := app.makeRequest(t, "GET", "/api/admin/users?search=Searchable", nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var users []api.UserAPIResponse
	err = json.Unmarshal(rr.Body.Bytes(), &users)
	require.NoError(t, err)

	// Should find the searchable user
	found := false
	for _, user := range users {
		if user.ID == searchUser.UserID {
			found = true
			assert.Equal(t, "John Searchable", *user.Name)
			break
		}
	}
	assert.True(t, found, "Searchable user not found in search results")
}

func TestAdminUserHandlers_GetUser_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test user
	ctx := context.Background()
	testUser, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Test User", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	// Test get user by ID
	rr := app.makeRequest(t, "GET", fmt.Sprintf("/api/admin/users/%d", testUser.UserID), nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var user api.UserAPIResponse
	err = json.Unmarshal(rr.Body.Bytes(), &user)
	require.NoError(t, err)

	assert.Equal(t, testUser.UserID, user.ID)
	assert.Equal(t, testUser.Phone, user.Phone)
	assert.Equal(t, "Test User", *user.Name)
	assert.Equal(t, "owl", user.Role)
	assert.NotEmpty(t, user.CreatedAt)
}

func TestAdminUserHandlers_GetUser_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get non-existent user
	rr := app.makeRequest(t, "GET", "/api/admin/users/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminUserHandlers_GetUser_InvalidID(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test get user with invalid ID
	rr := app.makeRequest(t, "GET", "/api/admin/users/invalid", nil, adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_CreateUser_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create user
	createReq := map[string]interface{}{
		"phone": "+15550001002",
		"name":  "New Test User",
		"role":  "owl",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusCreated, rr.Code, "Response: %s", rr.Body.String())

	var user api.UserAPIResponse
	err := json.Unmarshal(rr.Body.Bytes(), &user)
	require.NoError(t, err)

	assert.Equal(t, "+15550001002", user.Phone)
	assert.Equal(t, "New Test User", *user.Name)
	assert.Equal(t, "owl", user.Role)
	assert.NotEmpty(t, user.CreatedAt)

	// Verify user was created in database
	ctx := context.Background()
	dbUser, err := app.Querier.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "+15550001002", dbUser.Phone)
	assert.Equal(t, "New Test User", dbUser.Name.String)
	assert.Equal(t, "owl", dbUser.Role)
}

func TestAdminUserHandlers_CreateUser_DuplicatePhone(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create first user
	ctx := context.Background()
	_, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Existing User", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	// Try to create user with same phone
	createReq := map[string]interface{}{
		"phone": "+15550001002",
		"name":  "Duplicate User",
		"role":  "guest",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_CreateUser_InvalidRole(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create user with invalid role
	createReq := map[string]interface{}{
		"phone": "+15550001002",
		"name":  "Test User",
		"role":  "invalid_role",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_CreateUser_MissingPhone(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test create user without phone
	createReq := map[string]interface{}{
		"name": "Test User",
		"role": "owl",
	}
	reqBytes, _ := json.Marshal(createReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_UpdateUser_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test user
	ctx := context.Background()
	testUser, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Original Name", Valid: true},
		Role:  sql.NullString{String: "guest", Valid: true},
	})
	require.NoError(t, err)

	// Test update user
	updateReq := map[string]interface{}{
		"phone": "+15550001003",
		"name":  "Updated Name",
		"role":  "owl",
	}
	reqBytes, _ := json.Marshal(updateReq)

	rr := app.makeRequest(t, "PUT", fmt.Sprintf("/api/admin/users/%d", testUser.UserID), bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var user api.UserAPIResponse
	err = json.Unmarshal(rr.Body.Bytes(), &user)
	require.NoError(t, err)

	assert.Equal(t, testUser.UserID, user.ID)
	assert.Equal(t, "+15550001003", user.Phone)
	assert.Equal(t, "Updated Name", *user.Name)
	assert.Equal(t, "owl", user.Role)

	// Verify user was updated in database
	dbUser, err := app.Querier.GetUserByID(ctx, testUser.UserID)
	require.NoError(t, err)
	assert.Equal(t, "+15550001003", dbUser.Phone)
	assert.Equal(t, "Updated Name", dbUser.Name.String)
	assert.Equal(t, "owl", dbUser.Role)
}

func TestAdminUserHandlers_UpdateUser_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test update non-existent user
	updateReq := map[string]interface{}{
		"phone": "+15550001002",
		"name":  "Updated Name",
		"role":  "owl",
	}
	reqBytes, _ := json.Marshal(updateReq)

	rr := app.makeRequest(t, "PUT", "/api/admin/users/99999", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminUserHandlers_UpdateUser_PhoneConflict(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create two test users
	ctx := context.Background()
	user1, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "User 1", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	user2, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001003",
		Name:  sql.NullString{String: "User 2", Valid: true},
		Role:  sql.NullString{String: "guest", Valid: true},
	})
	require.NoError(t, err)

	// Try to update user2 with user1's phone
	updateReq := map[string]interface{}{
		"phone": user1.Phone,
		"name":  "Updated User 2",
		"role":  "owl",
	}
	reqBytes, _ := json.Marshal(updateReq)

	rr := app.makeRequest(t, "PUT", fmt.Sprintf("/api/admin/users/%d", user2.UserID), bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_DeleteUser_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test user
	ctx := context.Background()
	testUser, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "Test User", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	// Test delete user
	rr := app.makeRequest(t, "DELETE", fmt.Sprintf("/api/admin/users/%d", testUser.UserID), nil, adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "User deleted successfully", response["message"])

	// Verify user was deleted from database
	_, err = app.Querier.GetUserByID(ctx, testUser.UserID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestAdminUserHandlers_DeleteUser_NotFound(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test delete non-existent user
	rr := app.makeRequest(t, "DELETE", "/api/admin/users/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestAdminUserHandlers_BulkDeleteUsers_Success(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Create test users
	ctx := context.Background()
	user1, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001002",
		Name:  sql.NullString{String: "User 1", Valid: true},
		Role:  sql.NullString{String: "owl", Valid: true},
	})
	require.NoError(t, err)

	user2, err := app.Querier.CreateUser(ctx, db.CreateUserParams{
		Phone: "+15550001003",
		Name:  sql.NullString{String: "User 2", Valid: true},
		Role:  sql.NullString{String: "guest", Valid: true},
	})
	require.NoError(t, err)

	// Test bulk delete
	deleteReq := map[string]interface{}{
		"user_ids": []int64{user1.UserID, user2.UserID},
	}
	reqBytes, _ := json.Marshal(deleteReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users/bulk-delete", bytes.NewBuffer(reqBytes), adminToken)
	require.Equal(t, http.StatusOK, rr.Code, "Response: %s", rr.Body.String())

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Successfully deleted 2 users", response["message"])

	// Verify users were deleted from database
	_, err = app.Querier.GetUserByID(ctx, user1.UserID)
	assert.Equal(t, sql.ErrNoRows, err)
	_, err = app.Querier.GetUserByID(ctx, user2.UserID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestAdminUserHandlers_BulkDeleteUsers_EmptyList(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test bulk delete with empty list
	deleteReq := map[string]interface{}{
		"user_ids": []int64{},
	}
	reqBytes, _ := json.Marshal(deleteReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users/bulk-delete", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_BulkDeleteUsers_TooMany(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Create admin user
	_, adminToken := app.createTestUserAndLogin(t, "+15550001001", "Test Admin", "admin")

	// Test bulk delete with too many users (over 100 limit)
	userIDs := make([]int64, 101)
	for i := range userIDs {
		userIDs[i] = int64(i + 1)
	}

	deleteReq := map[string]interface{}{
		"user_ids": userIDs,
	}
	reqBytes, _ := json.Marshal(deleteReq)

	rr := app.makeRequest(t, "POST", "/api/admin/users/bulk-delete", bytes.NewBuffer(reqBytes), adminToken)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdminUserHandlers_Unauthorized_NonAdmin(t *testing.T) {
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
		{"GET", "/api/admin/users", nil},
		{"GET", "/api/admin/users/1", nil},
		{"POST", "/api/admin/users", []byte(`{"phone":"+15550001002","name":"Test","role":"owl"}`)},
		{"PUT", "/api/admin/users/1", []byte(`{"phone":"+15550001002","name":"Test","role":"owl"}`)},
		{"DELETE", "/api/admin/users/1", nil},
		{"POST", "/api/admin/users/bulk-delete", []byte(`{"user_ids":[1,2]}`)},
	}

	for _, endpoint := range endpoints {
		var body io.Reader
		if endpoint.body != nil {
			body = bytes.NewBuffer(endpoint.body)
		}
		rr := app.makeRequest(t, endpoint.method, endpoint.path, body, userToken)
		// Note: The actual authorization check depends on middleware implementation
		// This test assumes admin-only routes return 403 for non-admin users
		assert.True(t, rr.Code == http.StatusForbidden || rr.Code == http.StatusUnauthorized,
			"Expected 403 or 401 for %s %s, got %d", endpoint.method, endpoint.path, rr.Code)
	}
}

func TestAdminUserHandlers_Unauthorized_NoToken(t *testing.T) {
	app := newAdminTestApp(t)
	defer app.DB.Close()

	// Test that requests without token are rejected
	rr := app.makeRequest(t, "GET", "/api/admin/users", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
} 