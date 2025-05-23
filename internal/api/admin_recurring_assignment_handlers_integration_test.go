package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/logging"
	"night-owls-go/internal/service"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test setup helpers
func setupIntegrationTest(t *testing.T) (*sql.DB, db.Querier, *AdminRecurringAssignmentHandlers, *service.ScheduleService) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { testDB.Close() })

	// Create tables with complete schema
	schema := `
	CREATE TABLE users (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone TEXT UNIQUE NOT NULL,
		name TEXT,
		role TEXT NOT NULL DEFAULT 'guest',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE schedules (
		schedule_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		cron_expr TEXT NOT NULL,
		start_date DATE,
		end_date DATE,
		duration_minutes INTEGER NOT NULL DEFAULT 120,
		timezone TEXT
	);

	CREATE TABLE bookings (
		booking_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(user_id),
		schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id),
		shift_start DATETIME NOT NULL,
		shift_end DATETIME NOT NULL,
		buddy_user_id INTEGER REFERENCES users(user_id),
		buddy_name TEXT,
		attended BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(schedule_id, shift_start)
	);

	CREATE TABLE recurring_assignments (
		recurring_assignment_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
		buddy_name TEXT,
		day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
		schedule_id INTEGER NOT NULL REFERENCES schedules(schedule_id) ON DELETE CASCADE,
		time_slot TEXT NOT NULL,
		description TEXT,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, day_of_week, schedule_id, time_slot)
	);
	`

	_, err = testDB.Exec(schema)
	require.NoError(t, err)

	querier := db.New(testDB)
	cfg := &config.Config{DevMode: true}
	logger := logging.NewLogger(cfg)

	recurringService := service.NewRecurringAssignmentService(querier, logger, cfg)
	scheduleService := service.NewScheduleService(querier, logger, cfg)
	handlers := NewAdminRecurringAssignmentHandlers(logger, recurringService, scheduleService)

	return testDB, querier, handlers, scheduleService
}

func createTestUserIntegration(t *testing.T, querier db.Querier, phone, name, role string) db.User {
	user, err := querier.CreateUser(context.Background(), db.CreateUserParams{
		Phone: phone,
		Name:  sql.NullString{String: name, Valid: name != ""},
		Role:  sql.NullString{String: role, Valid: true},
	})
	require.NoError(t, err)
	return user
}

func createTestScheduleIntegration(t *testing.T, querier db.Querier, name, cronExpr string) db.Schedule {
	schedule, err := querier.CreateSchedule(context.Background(), db.CreateScheduleParams{
		Name:     name,
		CronExpr: cronExpr,
	})
	require.NoError(t, err)
	return schedule
}

// Mock auth middleware for testing
func mockAuthMiddleware(userID int64, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Test POST /api/admin/recurring-assignments
func TestAdminCreateRecurringAssignment(t *testing.T) {
	_, querier, handlers, _ := setupIntegrationTest(t)

	user := createTestUserIntegration(t, querier, "+1234567890", "John Doe", "owl")
	schedule := createTestScheduleIntegration(t, querier, "Night Watch", "0 18 * * 6")

	tests := []struct {
		name           string
		userID         int64
		requestBody    AdminCreateRecurringAssignmentRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful creation",
			userID: 1,
			requestBody: AdminCreateRecurringAssignmentRequest{
				UserID:      user.UserID,
				DayOfWeek:   6,
				ScheduleID:  schedule.ScheduleID,
				TimeSlot:    "18:00-20:00",
				BuddyName:   stringPtr("Jane Smith"),
				Description: stringPtr("Regular night patrol"),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "missing required fields",
			userID: 1,
			requestBody: AdminCreateRecurringAssignmentRequest{
				// Missing UserID
				DayOfWeek:  6,
				ScheduleID: schedule.ScheduleID,
				TimeSlot:   "18:00-20:00",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing or invalid required fields",
		},
		{
			name:   "invalid day of week",
			userID: 1,
			requestBody: AdminCreateRecurringAssignmentRequest{
				UserID:     user.UserID,
				DayOfWeek:  7, // Invalid
				ScheduleID: schedule.ScheduleID,
				TimeSlot:   "18:00-20:00",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing or invalid required fields",
		},
		{
			name:   "user not found",
			userID: 1,
			requestBody: AdminCreateRecurringAssignmentRequest{
				UserID:     99999, // Non-existent
				DayOfWeek:  6,
				ScheduleID: schedule.ScheduleID,
				TimeSlot:   "18:00-20:00",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "User or schedule not found",
		},
		{
			name:   "schedule not found",
			userID: 1,
			requestBody: AdminCreateRecurringAssignmentRequest{
				UserID:     user.UserID,
				DayOfWeek:  6,
				ScheduleID: 99999, // Non-existent
				TimeSlot:   "18:00-20:00",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "User or schedule not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/admin/recurring-assignments", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Add mock auth context
			ctx := context.WithValue(req.Context(), UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handlers.AdminCreateRecurringAssignment(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusCreated {
				var assignment db.RecurringAssignment
				err := json.Unmarshal(rr.Body.Bytes(), &assignment)
				require.NoError(t, err)
				assert.Equal(t, tt.requestBody.UserID, assignment.UserID)
				assert.Equal(t, tt.requestBody.DayOfWeek, assignment.DayOfWeek)
			}
		})
	}
}

// Test GET /api/admin/recurring-assignments
func TestAdminListRecurringAssignments(t *testing.T) {
	_, querier, handlers, _ := setupIntegrationTest(t)

	user := createTestUserIntegration(t, querier, "+1234567890", "John Doe", "owl")
	schedule := createTestScheduleIntegration(t, querier, "Night Watch", "0 18 * * 6")

	// Create some test assignments
	assignment1, err := querier.CreateRecurringAssignment(context.Background(), db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	})
	require.NoError(t, err)

	assignment2, err := querier.CreateRecurringAssignment(context.Background(), db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  0,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "19:00-21:00",
	})
	require.NoError(t, err)

	// Soft delete one assignment
	err = querier.DeleteRecurringAssignment(context.Background(), assignment2.RecurringAssignmentID)
	require.NoError(t, err)

	t.Run("list active assignments", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/recurring-assignments", nil)
		rr := httptest.NewRecorder()

		handlers.AdminListRecurringAssignments(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var assignments []db.RecurringAssignment
		err := json.Unmarshal(rr.Body.Bytes(), &assignments)
		require.NoError(t, err)
		
		// Should only return active assignment
		assert.Len(t, assignments, 1)
		assert.Equal(t, assignment1.RecurringAssignmentID, assignments[0].RecurringAssignmentID)
	})

	t.Run("materialize bookings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/recurring-assignments?materialize=true", nil)
		rr := httptest.NewRecorder()

		handlers.AdminListRecurringAssignments(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Bookings materialized successfully", response["message"])
	})

	t.Run("materialize with date range", func(t *testing.T) {
		from := time.Now().UTC().Format(time.RFC3339)
		to := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
		
		url := fmt.Sprintf("/api/admin/recurring-assignments?materialize=true&from=%s&to=%s", from, to)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		handlers.AdminListRecurringAssignments(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("invalid date format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/recurring-assignments?materialize=true&from=invalid-date", nil)
		rr := httptest.NewRecorder()

		handlers.AdminListRecurringAssignments(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// Test GET /api/admin/recurring-assignments/{id}
func TestAdminGetRecurringAssignment(t *testing.T) {
	_, querier, handlers, _ := setupIntegrationTest(t)

	user := createTestUserIntegration(t, querier, "+1234567890", "John Doe", "owl")
	schedule := createTestScheduleIntegration(t, querier, "Night Watch", "0 18 * * 6")

	assignment, err := querier.CreateRecurringAssignment(context.Background(), db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		assignmentID   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "found",
			assignmentID:   fmt.Sprintf("%d", assignment.RecurringAssignmentID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			assignmentID:   "99999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "Recurring assignment not found",
		},
		{
			name:           "invalid ID format",
			assignmentID:   "abc",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid assignment ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with URL parameters
			router := chi.NewRouter()
			router.Get("/api/admin/recurring-assignments/{id}", handlers.AdminGetRecurringAssignment)

			url := fmt.Sprintf("/api/admin/recurring-assignments/%s", tt.assignmentID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var returnedAssignment db.RecurringAssignment
				err := json.Unmarshal(rr.Body.Bytes(), &returnedAssignment)
				require.NoError(t, err)
				assert.Equal(t, assignment.RecurringAssignmentID, returnedAssignment.RecurringAssignmentID)
			}
		})
	}
}

// Test PUT /api/admin/recurring-assignments/{id}
func TestAdminUpdateRecurringAssignment(t *testing.T) {
	_, querier, handlers, _ := setupIntegrationTest(t)

	user1 := createTestUserIntegration(t, querier, "+1234567890", "John Doe", "owl")
	user2 := createTestUserIntegration(t, querier, "+0987654321", "Jane Smith", "owl")
	schedule := createTestScheduleIntegration(t, querier, "Night Watch", "0 18 * * 6")

	assignment, err := querier.CreateRecurringAssignment(context.Background(), db.CreateRecurringAssignmentParams{
		UserID:     user1.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	})
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updateReq := AdminUpdateRecurringAssignmentRequest{
			UserID:      user2.UserID,
			DayOfWeek:   0, // Change to Sunday
			ScheduleID:  schedule.ScheduleID,
			TimeSlot:    "19:00-21:00", // Change time
			BuddyName:   stringPtr("New Buddy"),
			Description: stringPtr("Updated description"),
		}

		body, _ := json.Marshal(updateReq)
		
		// Create router with URL parameters
		router := chi.NewRouter()
		router.Put("/api/admin/recurring-assignments/{id}", handlers.AdminUpdateRecurringAssignment)

		url := fmt.Sprintf("/api/admin/recurring-assignments/%d", assignment.RecurringAssignmentID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var updatedAssignment db.RecurringAssignment
		err := json.Unmarshal(rr.Body.Bytes(), &updatedAssignment)
		require.NoError(t, err)
		assert.Equal(t, user2.UserID, updatedAssignment.UserID)
		assert.Equal(t, int64(0), updatedAssignment.DayOfWeek)
		assert.Equal(t, "19:00-21:00", updatedAssignment.TimeSlot)
	})

	t.Run("assignment not found", func(t *testing.T) {
		updateReq := AdminUpdateRecurringAssignmentRequest{
			UserID:     user1.UserID,
			DayOfWeek:  6,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "18:00-20:00",
		}

		body, _ := json.Marshal(updateReq)
		
		router := chi.NewRouter()
		router.Put("/api/admin/recurring-assignments/{id}", handlers.AdminUpdateRecurringAssignment)

		req := httptest.NewRequest(http.MethodPut, "/api/admin/recurring-assignments/99999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// Test DELETE /api/admin/recurring-assignments/{id}
func TestAdminDeleteRecurringAssignment(t *testing.T) {
	_, querier, handlers, _ := setupIntegrationTest(t)

	user := createTestUserIntegration(t, querier, "+1234567890", "John Doe", "owl")
	schedule := createTestScheduleIntegration(t, querier, "Night Watch", "0 18 * * 6")

	assignment, err := querier.CreateRecurringAssignment(context.Background(), db.CreateRecurringAssignmentParams{
		UserID:     user.UserID,
		DayOfWeek:  6,
		ScheduleID: schedule.ScheduleID,
		TimeSlot:   "18:00-20:00",
	})
	require.NoError(t, err)

	t.Run("successful delete", func(t *testing.T) {
		router := chi.NewRouter()
		router.Delete("/api/admin/recurring-assignments/{id}", handlers.AdminDeleteRecurringAssignment)

		url := fmt.Sprintf("/api/admin/recurring-assignments/%d", assignment.RecurringAssignmentID)
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Recurring assignment deleted successfully", response["message"])

		// Verify soft delete - assignment should not appear in active list
		assignments, err := querier.ListRecurringAssignments(context.Background())
		require.NoError(t, err)
		assert.Len(t, assignments, 0)

		// But should still exist in DB
		deletedAssignment, err := querier.GetRecurringAssignmentByID(context.Background(), assignment.RecurringAssignmentID)
		require.NoError(t, err)
		assert.False(t, deletedAssignment.IsActive)
	})

	t.Run("delete preserves existing bookings", func(t *testing.T) {
		// Create another assignment for this test
		assignment2, err := querier.CreateRecurringAssignment(context.Background(), db.CreateRecurringAssignmentParams{
			UserID:     user.UserID,
			DayOfWeek:  0,
			ScheduleID: schedule.ScheduleID,
			TimeSlot:   "19:00-21:00",
		})
		require.NoError(t, err)

		// Create a booking that would have been materialized from this assignment
		booking, err := querier.CreateBooking(context.Background(), db.CreateBookingParams{
			UserID:     user.UserID,
			ScheduleID: schedule.ScheduleID,
			ShiftStart: time.Now().UTC().AddDate(0, 0, 1),
			ShiftEnd:   time.Now().UTC().AddDate(0, 0, 1).Add(2 * time.Hour),
		})
		require.NoError(t, err)

		// Delete the recurring assignment
		router := chi.NewRouter()
		router.Delete("/api/admin/recurring-assignments/{id}", handlers.AdminDeleteRecurringAssignment)

		url := fmt.Sprintf("/api/admin/recurring-assignments/%d", assignment2.RecurringAssignmentID)
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify booking still exists
		existingBooking, err := querier.GetBookingByID(context.Background(), booking.BookingID)
		require.NoError(t, err)
		assert.Equal(t, booking.BookingID, existingBooking.BookingID)
	})
}

// Test error handling for malformed requests
func TestMalformedRequests(t *testing.T) {
	_, _, handlers, _ := setupIntegrationTest(t)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recurring-assignments", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.AdminCreateRecurringAssignment(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("empty request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/recurring-assignments", bytes.NewBufferString(""))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handlers.AdminCreateRecurringAssignment(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
} 