package service_test

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"testing"

	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReportQuerier is a focused mock for ReportService tests.
type MockReportQuerier struct {
	mock.Mock
}

func (m *MockReportQuerier) GetBookingByID(ctx context.Context, bookingID int64) (db.Booking, error) {
	args := m.Called(ctx, bookingID)
	return args.Get(0).(db.Booking), args.Error(1)
}

func (m *MockReportQuerier) CreateReport(ctx context.Context, arg db.CreateReportParams) (db.Report, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Report), args.Error(1)
}

func (m *MockReportQuerier) GetReportByBookingID(ctx context.Context, bookingID int64) (db.Report, error) {
	args := m.Called(ctx, bookingID)
	return args.Get(0).(db.Report), args.Error(1)
}

func (m *MockReportQuerier) ListReportsByUserID(ctx context.Context, userID int64) ([]db.Report, error) {
	args := m.Called(ctx, userID)
	get0 := args.Get(0)
	if get0 == nil {
		return nil, args.Error(1)
	}
	return get0.([]db.Report), args.Error(1)
}

// Stub methods for other Querier interface methods not used by ReportService
func (m *MockReportQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) { panic("not implemented") }
func (m *MockReportQuerier) GetUserByPhone(ctx context.Context, phone string) (db.User, error) { panic("not implemented") }
func (m *MockReportQuerier) CreateOutboxItem(ctx context.Context, arg db.CreateOutboxItemParams) (db.Outbox, error) { panic("not implemented") }
func (m *MockReportQuerier) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) { panic("not implemented") }
func (m *MockReportQuerier) GetPendingOutboxItems(ctx context.Context, limit int64) ([]db.Outbox, error) { panic("not implemented") }
func (m *MockReportQuerier) GetScheduleByID(ctx context.Context, scheduleID int64) (db.Schedule, error) { panic("not implemented") }
func (m *MockReportQuerier) GetUserByID(ctx context.Context, userID int64) (db.User, error) { panic("not implemented") }
func (m *MockReportQuerier) ListActiveSchedules(ctx context.Context, arg db.ListActiveSchedulesParams) ([]db.Schedule, error) { panic("not implemented") }
func (m *MockReportQuerier) ListBookingsByUserID(ctx context.Context, userID int64) ([]db.Booking, error) { panic("not implemented") }
func (m *MockReportQuerier) UpdateBookingAttendance(ctx context.Context, arg db.UpdateBookingAttendanceParams) (db.Booking, error) { panic("not implemented") }
func (m *MockReportQuerier) UpdateOutboxItemStatus(ctx context.Context, arg db.UpdateOutboxItemStatusParams) (db.Outbox, error) { panic("not implemented") }
func (m *MockReportQuerier) CreateSchedule(ctx context.Context, arg db.CreateScheduleParams) (db.Schedule, error) { panic("not implemented") }
func (m *MockReportQuerier) GetBookingByScheduleAndStartTime(ctx context.Context, arg db.GetBookingByScheduleAndStartTimeParams) (db.Booking, error) { panic("not implemented") }
func (m *MockReportQuerier) ListAllSchedules(ctx context.Context) ([]db.Schedule, error) { panic("not implemented") }
func (m *MockReportQuerier) AdminBulkDeleteSchedules(ctx context.Context, scheduleIds []int64) error { panic("not implemented") }
func (m *MockReportQuerier) DeleteSchedule(ctx context.Context, scheduleID int64) error { panic("not implemented") }
func (m *MockReportQuerier) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) error { panic("not implemented") }
func (m *MockReportQuerier) DeleteUser(ctx context.Context, userID int64) error { panic("not implemented") }
func (m *MockReportQuerier) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]db.GetSubscriptionsByUserRow, error) { panic("not implemented") }
func (m *MockReportQuerier) ListUsers(ctx context.Context, searchTerm interface{}) ([]db.User, error) { panic("not implemented") }
func (m *MockReportQuerier) UpdateSchedule(ctx context.Context, arg db.UpdateScheduleParams) (db.Schedule, error) { panic("not implemented") }
func (m *MockReportQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) { panic("not implemented") }
func (m *MockReportQuerier) UpsertSubscription(ctx context.Context, arg db.UpsertSubscriptionParams) error { panic("not implemented") }

func newReportTestLogger() *slog.Logger { 
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestReportService_CreateReport_Success(t *testing.T) {
	mockQuerier := new(MockReportQuerier)
	testLogger := newReportTestLogger()
	reportService := service.NewReportService(mockQuerier, testLogger)

	authUserID := int64(1)
	bookingID := int64(10)
	severity := int32(1)
	message := "Test report message"

	existingBooking := db.Booking{BookingID: bookingID, UserID: authUserID}
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(existingBooking, nil).Once()

	expectedReport := db.Report{ReportID: 1, BookingID: bookingID, Severity: int64(severity), Message: sql.NullString{String: message, Valid: true}}
	mockQuerier.On("CreateReport", mock.Anything, mock.MatchedBy(func(params db.CreateReportParams) bool {
		return params.BookingID == bookingID && params.Severity == int64(severity) && params.Message.String == message
	})).Return(expectedReport, nil).Once()

	createdReport, err := reportService.CreateReport(context.Background(), authUserID, bookingID, severity, message)

	assert.NoError(t, err)
	assert.Equal(t, expectedReport.ReportID, createdReport.ReportID)
	assert.Equal(t, int64(severity), createdReport.Severity) // Ensure severity is correctly stored/returned
	mockQuerier.AssertExpectations(t)
}

func TestReportService_CreateReport_BookingNotFound(t *testing.T) {
	mockQuerier := new(MockReportQuerier)
	testLogger := newReportTestLogger()
	reportService := service.NewReportService(mockQuerier, testLogger)

	authUserID := int64(1)
	bookingID := int64(11)
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(db.Booking{}, sql.ErrNoRows).Once()

	_, err := reportService.CreateReport(context.Background(), authUserID, bookingID, 1, "Test")

	assert.Error(t, err)
	assert.Equal(t, service.ErrReportBookingAuth, err) // Or ErrBookingNotFound depending on desired specificity
	mockQuerier.AssertExpectations(t)
	mockQuerier.AssertNotCalled(t, "CreateReport", mock.Anything, mock.Anything)
}

func TestReportService_CreateReport_Forbidden(t *testing.T) {
	mockQuerier := new(MockReportQuerier)
	testLogger := newReportTestLogger()
	reportService := service.NewReportService(mockQuerier, testLogger)

	authUserID := int64(1)    // User trying to report
	otherUserID := int64(2) // Actual owner of the booking
	bookingID := int64(12)

	existingBooking := db.Booking{BookingID: bookingID, UserID: otherUserID}
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(existingBooking, nil).Once()

	_, err := reportService.CreateReport(context.Background(), authUserID, bookingID, 1, "Test")

	assert.Error(t, err)
	assert.Equal(t, service.ErrReportBookingAuth, err)
	mockQuerier.AssertExpectations(t)
	mockQuerier.AssertNotCalled(t, "CreateReport", mock.Anything, mock.Anything)
}

func TestReportService_CreateReport_SeverityOutOfRange(t *testing.T) {
	mockQuerier := new(MockReportQuerier)
	testLogger := newReportTestLogger()
	reportService := service.NewReportService(mockQuerier, testLogger)

	authUserID := int64(1)
	bookingID := int64(13)

	existingBooking := db.Booking{BookingID: bookingID, UserID: authUserID}
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(existingBooking, nil).Twice()

	// Test severity too low
	_, errLow := reportService.CreateReport(context.Background(), authUserID, bookingID, -1, "Severity too low")
	assert.Error(t, errLow)
	assert.Equal(t, service.ErrSeverityOutOfRange, errLow)

	// Test severity too high
	_, errHigh := reportService.CreateReport(context.Background(), authUserID, bookingID, 3, "Severity too high")
	assert.Error(t, errHigh)
	assert.Equal(t, service.ErrSeverityOutOfRange, errHigh)

	mockQuerier.AssertExpectations(t) // GetBookingByID should have been called twice
	mockQuerier.AssertNotCalled(t, "CreateReport", mock.Anything, mock.Anything)
}

func TestReportService_CreateReport_EmptyMessage(t *testing.T) {
	mockQuerier := new(MockReportQuerier)
	testLogger := newReportTestLogger()
	reportService := service.NewReportService(mockQuerier, testLogger)

	authUserID := int64(1)
	bookingID := int64(10)
	severity := int32(0)
	message := "" // Empty message

	existingBooking := db.Booking{BookingID: bookingID, UserID: authUserID}
	mockQuerier.On("GetBookingByID", mock.Anything, bookingID).Return(existingBooking, nil).Once()

	expectedReport := db.Report{ReportID: 2, BookingID: bookingID, Severity: int64(severity), Message: sql.NullString{String: message, Valid: false}}
	mockQuerier.On("CreateReport", mock.Anything, mock.MatchedBy(func(params db.CreateReportParams) bool {
		return params.BookingID == bookingID && params.Severity == int64(severity) && !params.Message.Valid && params.Message.String == ""
	})).Return(expectedReport, nil).Once()

	createdReport, err := reportService.CreateReport(context.Background(), authUserID, bookingID, severity, message)

	assert.NoError(t, err)
	assert.Equal(t, expectedReport.ReportID, createdReport.ReportID)
	assert.False(t, createdReport.Message.Valid) // Check that empty message is stored as NULL or Valid=false
	mockQuerier.AssertExpectations(t)
}

// TODO: Test DB error from GetBookingByID (not ErrNoRows)
// TODO: Test DB error from CreateReport
