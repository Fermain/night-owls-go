package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	// For JWT tests if any, not directly by BookingService though
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBookingQuerier is a focused mock for BookingService tests.
type MockBookingQuerier struct {
	mock.Mock
}

func (m *MockBookingQuerier) GetScheduleByID(ctx context.Context, scheduleID int64) (db.Schedule, error) {
	args := m.Called(ctx, scheduleID)
	return args.Get(0).(db.Schedule), args.Error(1)
}

func (m *MockBookingQuerier) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {
	args := m.Called(ctx, phone)
	if args.Error(1) != nil && errors.Is(args.Error(1), sql.ErrNoRows) {
        return db.User{}, args.Error(1)
    }
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockBookingQuerier) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	args := m.Called(ctx, arg)
	// Simulate unique constraint error if err is not sql.ErrNoRows and not nil
    // This is a bit simplistic; real error checking is in the service.
    // Here, we just ensure the mock can return various errors.
    if args.Error(1) != nil && !errors.Is(args.Error(1), sql.ErrNoRows) { // any error other than NoRows is passed through
        return db.Booking{}, args.Error(1)
    }
	return args.Get(0).(db.Booking), args.Error(1)
}

func (m *MockBookingQuerier) CreateOutboxItem(ctx context.Context, arg db.CreateOutboxItemParams) (db.Outbox, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Outbox), args.Error(1)
}

func (m *MockBookingQuerier) GetBookingByID(ctx context.Context, bookingID int64) (db.Booking, error) {
	args := m.Called(ctx, bookingID)
	if args.Error(1) != nil && errors.Is(args.Error(1), sql.ErrNoRows) {
        return db.Booking{}, args.Error(1)
    }
	return args.Get(0).(db.Booking), args.Error(1)
}

func (m *MockBookingQuerier) UpdateBookingAttendance(ctx context.Context, arg db.UpdateBookingAttendanceParams) (db.Booking, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Booking), args.Error(1)
}

// Stubs for other Querier methods
func (m *MockBookingQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) { panic("not implemented") }
func (m *MockBookingQuerier) GetPendingOutboxItems(ctx context.Context, limit int64) ([]db.Outbox, error) { panic("not implemented") }
func (m *MockBookingQuerier) GetReportByBookingID(ctx context.Context, bookingID int64) (db.Report, error) { panic("not implemented") }
func (m *MockBookingQuerier) GetUserByID(ctx context.Context, userID int64) (db.User, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListActiveSchedules(ctx context.Context, arg db.ListActiveSchedulesParams) ([]db.Schedule, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListBookingsByUserID(ctx context.Context, userID int64) ([]db.Booking, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListReportsByUserID(ctx context.Context, userID int64) ([]db.Report, error) { panic("not implemented") }
func (m *MockBookingQuerier) UpdateOutboxItemStatus(ctx context.Context, arg db.UpdateOutboxItemStatusParams) (db.Outbox, error) { panic("not implemented") }
func (m *MockBookingQuerier) CreateReport(ctx context.Context, arg db.CreateReportParams) (db.Report, error) { panic("not implemented") }
func (m *MockBookingQuerier) CreateSchedule(ctx context.Context, arg db.CreateScheduleParams) (db.Schedule, error) { panic("not implemented") }
func (m *MockBookingQuerier) GetBookingByScheduleAndStartTime(ctx context.Context, arg db.GetBookingByScheduleAndStartTimeParams) (db.Booking, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListAllSchedules(ctx context.Context) ([]db.Schedule, error) { panic("not implemented") }
func (m *MockBookingQuerier) AdminBulkDeleteSchedules(ctx context.Context, scheduleIds []int64) error { panic("not implemented") }
func (m *MockBookingQuerier) DeleteSchedule(ctx context.Context, scheduleID int64) error { panic("not implemented") }
func (m *MockBookingQuerier) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) error { panic("not implemented") }
func (m *MockBookingQuerier) DeleteUser(ctx context.Context, userID int64) error { panic("not implemented") }
func (m *MockBookingQuerier) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]db.GetSubscriptionsByUserRow, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListUsers(ctx context.Context, searchTerm interface{}) ([]db.User, error) { panic("not implemented") }
func (m *MockBookingQuerier) UpdateSchedule(ctx context.Context, arg db.UpdateScheduleParams) (db.Schedule, error) { panic("not implemented") }
func (m *MockBookingQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) { panic("not implemented") }
func (m *MockBookingQuerier) UpsertSubscription(ctx context.Context, arg db.UpsertSubscriptionParams) error { panic("not implemented") }
func (m *MockBookingQuerier) AdminBulkDeleteUsers(ctx context.Context, userIds []int64) error { panic("not implemented") }
func (m *MockBookingQuerier) CreateRecurringAssignment(ctx context.Context, arg db.CreateRecurringAssignmentParams) (db.RecurringAssignment, error) { panic("not implemented") }
func (m *MockBookingQuerier) DeleteRecurringAssignment(ctx context.Context, recurringAssignmentID int64) error { panic("not implemented") }
func (m *MockBookingQuerier) GetRecurringAssignmentByID(ctx context.Context, recurringAssignmentID int64) (db.RecurringAssignment, error) { panic("not implemented") }
func (m *MockBookingQuerier) GetRecurringAssignmentsByPattern(ctx context.Context, arg db.GetRecurringAssignmentsByPatternParams) ([]db.GetRecurringAssignmentsByPatternRow, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListRecurringAssignments(ctx context.Context) ([]db.RecurringAssignment, error) { panic("not implemented") }
func (m *MockBookingQuerier) ListRecurringAssignmentsByUserID(ctx context.Context, userID int64) ([]db.RecurringAssignment, error) { panic("not implemented") }
func (m *MockBookingQuerier) UpdateRecurringAssignment(ctx context.Context, arg db.UpdateRecurringAssignmentParams) (db.RecurringAssignment, error) { panic("not implemented") }
func (m *MockBookingQuerier) GetRecentOutboxItemsByRecipient(ctx context.Context, arg db.GetRecentOutboxItemsByRecipientParams) ([]db.Outbox, error) { panic("not implemented") }


// Re-define newTestLogger and newTestConfig as they are not in a shared test utility package.
func newBookingTestLogger() *slog.Logger { // Renamed to avoid conflict
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newBookingTestConfig() *config.Config { // Renamed to avoid conflict
	return &config.Config{
		JWTSecret: "test-secret-booking",
		OTPLogPath: "/dev/null",
		DefaultShiftDuration: 2 * time.Hour,
		// Add other config fields if BookingService uses them directly
	}
}

var testTimeParse = func(value string) time.Time {
    t, _ := time.Parse(time.RFC3339, value)
    return t
}

func TestBookingService_CreateBooking_Success(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	userID := int64(1)
	scheduleID := int64(10)
	startTime := testTimeParse("2024-07-15T10:00:00Z")

	schedule := db.Schedule{
		ScheduleID:      scheduleID,
		Name:            "Test Schedule",
		CronExpr:        "0 10 * * 1", // Mondays at 10:00
		DurationMinutes: 120,
		StartDate:       sql.NullTime{Time: testTimeParse("2024-01-01T00:00:00Z"), Valid: true},
		EndDate:         sql.NullTime{Time: testTimeParse("2024-12-31T00:00:00Z"), Valid: true},
	}
	mockQuerier.On("GetScheduleByID", mock.Anything, scheduleID).Return(schedule, nil).Once()

	// Buddy is not a registered user
	buddyPhone := sql.NullString{String: "+15551234", Valid: true}
	buddyName := sql.NullString{String: "Buddy Test", Valid: true}
	mockQuerier.On("GetUserByPhone", mock.Anything, buddyPhone.String).Return(db.User{}, sql.ErrNoRows).Once()

	expectedBooking := db.Booking{BookingID: 1, UserID: userID, ScheduleID: scheduleID, ShiftStart: startTime}
	mockQuerier.On("CreateBooking", mock.Anything, mock.AnythingOfType("db.CreateBookingParams")).
		Run(func(args mock.Arguments) {
			params := args.Get(1).(db.CreateBookingParams)
			assert.Equal(t, userID, params.UserID)
			assert.Equal(t, scheduleID, params.ScheduleID)
			assert.Equal(t, startTime, params.ShiftStart)
			assert.False(t, params.BuddyUserID.Valid) // Buddy not registered
			assert.Equal(t, buddyName.String, params.BuddyName.String)
		}).Return(expectedBooking, nil).Once()

	mockQuerier.On("CreateOutboxItem", mock.Anything, mock.AnythingOfType("db.CreateOutboxItemParams")).Return(db.Outbox{}, nil).Once()

	createdBooking, err := bookingService.CreateBooking(context.Background(), userID, scheduleID, startTime, buddyPhone, buddyName)

	assert.NoError(t, err)
	assert.Equal(t, expectedBooking.BookingID, createdBooking.BookingID)
	mockQuerier.AssertExpectations(t)
}

func TestBookingService_CreateBooking_WithRegisteredBuddy(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	userID := int64(1)
	scheduleID := int64(10)
	startTime := testTimeParse("2024-07-15T10:00:00Z")

	schedule := db.Schedule{ScheduleID: scheduleID, CronExpr: "0 10 * * 1", DurationMinutes: 120, StartDate: sql.NullTime{Time: testTimeParse("2024-01-01T00:00:00Z"), Valid: true}, EndDate: sql.NullTime{Time: testTimeParse("2024-12-31T00:00:00Z"), Valid: true}}
	mockQuerier.On("GetScheduleByID", mock.Anything, scheduleID).Return(schedule, nil).Once()

	buddyPhoneStr := "+1555REGISTERED"
	registeredBuddy := db.User{UserID: 99, Phone: buddyPhoneStr, Name: sql.NullString{String: "Registered Buddy", Valid: true}}
	mockQuerier.On("GetUserByPhone", mock.Anything, buddyPhoneStr).Return(registeredBuddy, nil).Once()

	expectedBooking := db.Booking{BookingID: 2}
	mockQuerier.On("CreateBooking", mock.Anything, mock.MatchedBy(func(params db.CreateBookingParams) bool {
        return params.BuddyUserID.Valid && params.BuddyUserID.Int64 == registeredBuddy.UserID &&
               params.BuddyName.String == registeredBuddy.Name.String
    })).Return(expectedBooking, nil).Once()
	mockQuerier.On("CreateOutboxItem", mock.Anything, mock.AnythingOfType("db.CreateOutboxItemParams")).Return(db.Outbox{}, nil).Once()

	_, err := bookingService.CreateBooking(context.Background(), userID, scheduleID, startTime, 
		sql.NullString{String: buddyPhoneStr, Valid: true}, 
		sql.NullString{String: "Provided Name Should Be Overridden", Valid: true}) // Name from request

	assert.NoError(t, err)
	mockQuerier.AssertExpectations(t)
}

func TestBookingService_CreateBooking_ScheduleNotFound(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	scheduleID := int64(999)
	mockQuerier.On("GetScheduleByID", mock.Anything, scheduleID).Return(db.Schedule{}, sql.ErrNoRows).Once()

	_, err := bookingService.CreateBooking(context.Background(), 1, scheduleID, time.Now(), sql.NullString{}, sql.NullString{})

	assert.Error(t, err)
	assert.Equal(t, service.ErrScheduleNotFound, err)
	mockQuerier.AssertExpectations(t)
}

func TestBookingService_CreateBooking_ShiftTimeInvalid_OutsideScheduleWindow(t *testing.T) {
    mockQuerier := new(MockBookingQuerier)
    cfg := newBookingTestConfig()
    testLogger := newBookingTestLogger()
    bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

    scheduleID := int64(1)
    // Schedule from 2024-01-01 to 2024-01-31
    schedule := db.Schedule{
        ScheduleID:      scheduleID,
        CronExpr:        "0 0 * * *", // Daily at midnight
        DurationMinutes: 60,
        StartDate:       sql.NullTime{Time: testTimeParse("2024-01-01T00:00:00Z"), Valid: true},
        EndDate:         sql.NullTime{Time: testTimeParse("2024-01-31T23:59:59Z"), Valid: true},
    }
    mockQuerier.On("GetScheduleByID", mock.Anything, scheduleID).Return(schedule, nil).Once()

    // Attempt to book for Feb 1st, 2024 (outside window)
    startTimeOutsideWindow := testTimeParse("2024-02-01T00:00:00Z")
    _, err := bookingService.CreateBooking(context.Background(), 1, scheduleID, startTimeOutsideWindow, sql.NullString{}, sql.NullString{})

    assert.Error(t, err)
    assert.True(t, errors.Is(err, service.ErrShiftTimeInvalid), "Error should be ErrShiftTimeInvalid")
    mockQuerier.AssertExpectations(t)
}

func TestBookingService_CreateBooking_ShiftTimeInvalid_NotMatchingCron(t *testing.T) {
    mockQuerier := new(MockBookingQuerier)
    cfg := newBookingTestConfig()
    testLogger := newBookingTestLogger()
    bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

    scheduleID := int64(1)
    // Schedule for Mondays at 10:00
    schedule := db.Schedule{
        ScheduleID:      scheduleID,
        CronExpr:        "0 10 * * 1", 
        DurationMinutes: 60,
        StartDate:       sql.NullTime{Time: testTimeParse("2024-01-01T00:00:00Z"), Valid: true},
        EndDate:         sql.NullTime{Time: testTimeParse("2024-12-31T23:59:59Z"), Valid: true},
    }
    mockQuerier.On("GetScheduleByID", mock.Anything, scheduleID).Return(schedule, nil).Once()

    // Attempt to book for a Tuesday at 10:00 (valid date, wrong day for cron)
    // Tuesday, July 16, 2024 10:00:00 AM GMT
    startTimeWrongDay := testTimeParse("2024-07-16T10:00:00Z") 
    _, err := bookingService.CreateBooking(context.Background(), 1, scheduleID, startTimeWrongDay, sql.NullString{}, sql.NullString{})

    assert.Error(t, err)
    assert.True(t, errors.Is(err, service.ErrShiftTimeInvalid), "Error should be ErrShiftTimeInvalid for non-cron time")
    mockQuerier.AssertExpectations(t)
}


func TestBookingService_CreateBooking_Conflict(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	scheduleID := int64(1)
	startTime := testTimeParse("2024-07-15T10:00:00Z")
	schedule := db.Schedule{ScheduleID: scheduleID, CronExpr: "0 10 * * 1", DurationMinutes: 120, StartDate: sql.NullTime{Time: testTimeParse("2024-01-01T00:00:00Z"), Valid: true}, EndDate: sql.NullTime{Time: testTimeParse("2024-12-31T00:00:00Z"), Valid: true}}
	mockQuerier.On("GetScheduleByID", mock.Anything, scheduleID).Return(schedule, nil).Once()

    // Simulate DB unique constraint error
	mockQuerier.On("CreateBooking", mock.Anything, mock.AnythingOfType("db.CreateBookingParams")).Return(db.Booking{}, errors.New("UNIQUE constraint failed: bookings.schedule_id, bookings.shift_start")).Once()

	// Passing sql.NullString{} for buddyPhone, so GetUserByPhone should not be called.
	_, err := bookingService.CreateBooking(context.Background(), 1, scheduleID, startTime, sql.NullString{}, sql.NullString{})

	assert.Error(t, err)
	assert.Equal(t, service.ErrBookingConflict, err)
	mockQuerier.AssertExpectations(t)
}

func TestBookingService_MarkAttendance_Success(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	bookingID := int64(100)
	authUserID := int64(50)
	attendedStatus := true

	existingBooking := db.Booking{BookingID: bookingID, UserID: authUserID, Attended: false}
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(existingBooking, nil).Once()

	updatedBookingFromDB := db.Booking{BookingID: bookingID, UserID: authUserID, Attended: attendedStatus}
	mockQuerier.On("UpdateBookingAttendance", mock.Anything, db.UpdateBookingAttendanceParams{BookingID: bookingID, Attended: attendedStatus}).Return(updatedBookingFromDB, nil).Once()

	resultBooking, err := bookingService.MarkAttendance(context.Background(), bookingID, authUserID, attendedStatus)

	assert.NoError(t, err)
	assert.Equal(t, updatedBookingFromDB, resultBooking)
	assert.True(t, resultBooking.Attended)
	mockQuerier.AssertExpectations(t)
}

func TestBookingService_MarkAttendance_BookingNotFound(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	bookingID := int64(101)
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(db.Booking{}, sql.ErrNoRows).Once()

	_, err := bookingService.MarkAttendance(context.Background(), bookingID, 50, true)

	assert.Error(t, err)
	assert.Equal(t, service.ErrBookingNotFound, err)
	mockQuerier.AssertExpectations(t)
}

func TestBookingService_MarkAttendance_Forbidden(t *testing.T) {
	mockQuerier := new(MockBookingQuerier)
	cfg := newBookingTestConfig()
	testLogger := newBookingTestLogger()
	bookingService := service.NewBookingService(mockQuerier, cfg, testLogger)

	bookingID := int64(102)
	actualOwnerUserID := int64(51)
	authUserIDTryingToUpdate := int64(52) // Different user

	existingBooking := db.Booking{BookingID: bookingID, UserID: actualOwnerUserID}
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(existingBooking, nil).Once()

	_, err := bookingService.MarkAttendance(context.Background(), bookingID, authUserIDTryingToUpdate, true)

	assert.Error(t, err)
	assert.Equal(t, service.ErrForbiddenUpdate, err)
	mockQuerier.AssertExpectations(t)
	mockQuerier.AssertNotCalled(t, "UpdateBookingAttendance", mock.Anything, mock.Anything)
}

// TODO: Add tests for CreateBooking error paths (e.g., GetScheduleByID fails, GetUserByPhone (for buddy) fails not with NoRows, CreateOutboxItem fails).
// TODO: Add tests for MarkAttendance error paths (e.g., UpdateBookingAttendance DB call fails).