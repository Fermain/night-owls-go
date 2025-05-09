package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/gorhill/cronexpr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScheduleQuerier is a focused mock for ScheduleService tests.
// It only implements methods used by ScheduleService.
// If we need a full Querier mock, it should be in a shared location.
type MockScheduleQuerier struct {
	mock.Mock
}

func (m *MockScheduleQuerier) ListActiveSchedules(ctx context.Context, arg db.ListActiveSchedulesParams) ([]db.Schedule, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Schedule), args.Error(1)
}

func (m *MockScheduleQuerier) GetBookingByScheduleAndStartTime(ctx context.Context, arg db.GetBookingByScheduleAndStartTimeParams) (db.Booking, error) {
	args := m.Called(ctx, arg)
	// Handle sql.ErrNoRows specifically if the mock needs to simulate it
	if args.Error(1) != nil && errors.Is(args.Error(1), sql.ErrNoRows) {
		// Return zero value for db.Booking if error is sql.ErrNoRows
		return db.Booking{}, args.Error(1)
	}
	return args.Get(0).(db.Booking), args.Error(1)
}

// Stubs for other Querier methods to satisfy a more general (but unused here) Querier interface if we were to try and make this a full mock.
// For this specific test file, these are not strictly necessary as ScheduleService only uses the two above.
func (m *MockScheduleQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) GetUserByPhone(ctx context.Context, phone string) (db.User, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) CreateOutboxItem(ctx context.Context, arg db.CreateOutboxItemParams) (db.Outbox, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) GetBookingByID(ctx context.Context, bookingID int64) (db.Booking, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) GetPendingOutboxItems(ctx context.Context, limit int64) ([]db.Outbox, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) GetReportByBookingID(ctx context.Context, bookingID int64) (db.Report, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) GetScheduleByID(ctx context.Context, scheduleID int64) (db.Schedule, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) GetUserByID(ctx context.Context, userID int64) (db.User, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) ListBookingsByUserID(ctx context.Context, userID int64) ([]db.Booking, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) ListReportsByUserID(ctx context.Context, userID int64) ([]db.Report, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) UpdateBookingAttendance(ctx context.Context, arg db.UpdateBookingAttendanceParams) (db.Booking, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) UpdateOutboxItemStatus(ctx context.Context, arg db.UpdateOutboxItemStatusParams) (db.Outbox, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) CreateReport(ctx context.Context, arg db.CreateReportParams) (db.Report, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) CreateSchedule(ctx context.Context, arg db.CreateScheduleParams) (db.Schedule, error) { panic("not implemented by MockScheduleQuerier") }
func (m *MockScheduleQuerier) ListAllSchedules(ctx context.Context) ([]db.Schedule, error) { 
	args := m.Called(ctx)
	return args.Get(0).([]db.Schedule), args.Error(1)
}


// Re-define newTestLogger and newTestConfig if they are not in a shared test utility package.
// For now, assuming they are not, so re-defining for this test file.
func newScheduleTestLogger() *slog.Logger { // Renamed to avoid conflict if in same package via _test
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestScheduleService_GetUpcomingAvailableSlots_NoSchedules(t *testing.T) {
	mockQuerier := new(MockScheduleQuerier)
	testLogger := newScheduleTestLogger()
	scheduleService := service.NewScheduleService(mockQuerier, testLogger)

	mockQuerier.On("ListAllSchedules", mock.Anything).Return([]db.Schedule{}, nil).Once()

	slots, err := scheduleService.GetUpcomingAvailableSlots(context.Background(), nil, nil, nil)

	assert.NoError(t, err)
	assert.Empty(t, slots)
	mockQuerier.AssertExpectations(t)
}

func TestScheduleService_GetUpcomingAvailableSlots_SingleScheduleNoBookings(t *testing.T) {
	mockQuerier := new(MockScheduleQuerier)
	testLogger := newScheduleTestLogger()
	scheduleService := service.NewScheduleService(mockQuerier, testLogger)

	now := time.Now()
	// Schedule active today, runs every hour at :00
	schedule1 := db.Schedule{
		ScheduleID:      1,
		Name:            "Hourly Test Schedule",
		CronExpr:        "0 * * * *", // Every hour at minute 0
		StartDate:       sql.NullTime{Time: now.AddDate(0, 0, -1), Valid: true}, // Started yesterday
		EndDate:         sql.NullTime{Time: now.AddDate(0, 0, 1), Valid: true},   // Ends tomorrow
		DurationMinutes: 60,
	}
	mockQuerier.On("ListAllSchedules", mock.Anything).Return([]db.Schedule{schedule1}, nil).Once()

	// For each potential slot generated by cron, GetBookingByScheduleAndStartTime will be called.
	// We need to mock it to return sql.ErrNoRows (slot not booked).
	mockQuerier.On("GetBookingByScheduleAndStartTime", mock.Anything, mock.AnythingOfType("db.GetBookingByScheduleAndStartTimeParams")).Return(db.Booking{}, sql.ErrNoRows) // Note: No .Once() as it can be called multiple times

	// Query for slots in the next 3 hours
	queryFrom := now
	queryTo := now.Add(3 * time.Hour)
	limit := 5

	slots, err := scheduleService.GetUpcomingAvailableSlots(context.Background(), &queryFrom, &queryTo, &limit)

	assert.NoError(t, err)
	assert.NotEmpty(t, slots)

	// Expect roughly 3 slots (one per hour, could be 2, 3 or 4 depending on exact current time to minute 0)
	// For cron "0 * * * *" starting now, next occurrence is top of next hour.
	// If now is 10:30, query window to 13:30. Expected: 11:00, 12:00, 13:00.
	// This assertion is a bit loose due to cron and time.Now() dynamics.
	// A more precise test would fix time.Now() or calculate expected slots very carefully.
	assert.GreaterOrEqual(t, len(slots), 2, "Should find at least 2 slots in the next 3 hours for an hourly schedule")
	assert.LessOrEqual(t, len(slots), 4, "Should find at most 4 slots in the next 3 hours for an hourly schedule")

	for _, slot := range slots {
		assert.Equal(t, schedule1.ScheduleID, slot.ScheduleID)
		assert.False(t, slot.IsBooked)
		assert.Equal(t, time.Duration(schedule1.DurationMinutes)*time.Minute, slot.EndTime.Sub(slot.StartTime))
	}
	mockQuerier.AssertExpectations(t) // Will fail if GetBooking was not called for potential slots
}

func TestScheduleService_GetUpcomingAvailableSlots_WithBookedSlot(t *testing.T) {
	mockQuerier := new(MockScheduleQuerier)
	testLogger := newScheduleTestLogger()
	scheduleService := service.NewScheduleService(mockQuerier, testLogger)

	now := time.Now()
	schedule1 := db.Schedule{
		ScheduleID:      1, Name: "Hourly Test", CronExpr: "0 * * * *",
		StartDate:       sql.NullTime{Time: now.AddDate(0, 0, -1), Valid: true},
		EndDate:         sql.NullTime{Time: now.AddDate(0, 0, 1), Valid: true},
		DurationMinutes: 60,
	}
	mockQuerier.On("ListAllSchedules", mock.Anything).Return([]db.Schedule{schedule1}, nil).Once()

	// Determine a specific upcoming slot to mark as "booked"
	// Let's find the first upcoming hourly slot from now.
	cronInst, _ := cronexpr.Parse(schedule1.CronExpr)
	firstUpcomingSlotTime := cronInst.Next(now) 

	// Mock GetBookingByScheduleAndStartTime
	mockQuerier.On("GetBookingByScheduleAndStartTime", mock.Anything, mock.MatchedBy(func(params db.GetBookingByScheduleAndStartTimeParams) bool {
		return params.ScheduleID == schedule1.ScheduleID && params.ShiftStart.Equal(firstUpcomingSlotTime)
	})).Return(db.Booking{BookingID: 100, ScheduleID: schedule1.ScheduleID, ShiftStart: firstUpcomingSlotTime}, nil).Once() // This one is booked

	// For any other slot, it's not booked
	mockQuerier.On("GetBookingByScheduleAndStartTime", mock.Anything, mock.AnythingOfType("db.GetBookingByScheduleAndStartTimeParams")).Return(db.Booking{}, sql.ErrNoRows) // Default for others

	queryFrom := now
	queryTo := now.Add(3 * time.Hour)
	limit := 5

	slots, err := scheduleService.GetUpcomingAvailableSlots(context.Background(), &queryFrom, &queryTo, &limit)

	assert.NoError(t, err)
	assert.NotEmpty(t, slots)

	foundBookedSlotTimeInResults := false
	for _, slot := range slots {
		assert.Equal(t, schedule1.ScheduleID, slot.ScheduleID)
		assert.False(t, slot.IsBooked) // is_booked in AvailableShiftSlot struct should be false (as it's for available slots)
		if slot.StartTime.Equal(firstUpcomingSlotTime) {
			foundBookedSlotTimeInResults = true
		}
	}
	assert.False(t, foundBookedSlotTimeInResults, "The specifically booked slot should not appear in available slots")
	mockQuerier.AssertExpectations(t)
}

// TODO: Add more tests:
// - Schedule with no occurrences in the query window
// - Query window entirely outside schedule's active start/end dates
// - Limit parameter working correctly
// - Error from ListActiveSchedules
// - Error from GetBookingByScheduleAndStartTime (other than sql.ErrNoRows)
// - Cron expression parsing error (though service might treat as internal error for existing schedule) 