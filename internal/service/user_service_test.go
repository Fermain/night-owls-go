package service_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"night-owls-go/internal/auth"
	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuerier is a mock implementation of the db.Querier interface for testing.
// We are using testify/mock to simplify mock creation.
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockQuerier) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {
	args := m.Called(ctx, phone)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockQuerier) CreateOutboxItem(ctx context.Context, arg db.CreateOutboxItemParams) (db.Outbox, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Outbox), args.Error(1)
}

func (m *MockQuerier) GetRecentOutboxItemsByRecipient(ctx context.Context, arg db.GetRecentOutboxItemsByRecipientParams) ([]db.Outbox, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Outbox), args.Error(1)
}

// Implement other Querier methods as needed for future tests, returning nil/error or using mock.Anything
// For now, these are the ones used by UserService. To satisfy the interface for general use, we'd add all.
// To keep this focused for UserService, we only implement what's directly used by the service under test.
// However, the db.Querier interface is large. For UserService, these 3 are enough.
// If we were testing a component that needs the *full* Querier, we'd need all methods.
// Let's add stubs for the rest to make the mock a complete Querier for broader reusability, even if UserService doesn't use them.

func (m *MockQuerier) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Booking), args.Error(1)
}
func (m *MockQuerier) GetBookingByID(ctx context.Context, bookingID int64) (db.Booking, error) {
	args := m.Called(ctx, bookingID)
	return args.Get(0).(db.Booking), args.Error(1)
}
func (m *MockQuerier) GetBookingByScheduleAndStartTime(ctx context.Context, arg db.GetBookingByScheduleAndStartTimeParams) (db.Booking, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Booking), args.Error(1)
}
func (m *MockQuerier) GetPendingOutboxItems(ctx context.Context, limit int64) ([]db.Outbox, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]db.Outbox), args.Error(1)
}
func (m *MockQuerier) GetReportByBookingID(ctx context.Context, bookingID int64) (db.Report, error) {
	args := m.Called(ctx, bookingID)
	return args.Get(0).(db.Report), args.Error(1)
}
func (m *MockQuerier) GetScheduleByID(ctx context.Context, scheduleID int64) (db.Schedule, error) {
	args := m.Called(ctx, scheduleID)
	return args.Get(0).(db.Schedule), args.Error(1)
}
func (m *MockQuerier) GetUserByID(ctx context.Context, userID int64) (db.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.User), args.Error(1)
}
func (m *MockQuerier) ListActiveSchedules(ctx context.Context, arg db.ListActiveSchedulesParams) ([]db.Schedule, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Schedule), args.Error(1)
}
func (m *MockQuerier) ListBookingsByUserID(ctx context.Context, userID int64) ([]db.Booking, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Booking), args.Error(1)
}
func (m *MockQuerier) ListReportsByUserID(ctx context.Context, userID int64) ([]db.Report, error) {
	args := m.Called(ctx, userID)
	get0 := args.Get(0)
	if get0 == nil {
		return nil, args.Error(1)
	}
	return get0.([]db.Report), args.Error(1)
}
func (m *MockQuerier) UpdateBookingCheckIn(ctx context.Context, arg db.UpdateBookingCheckInParams) (db.Booking, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Booking), args.Error(1)
}
func (m *MockQuerier) UpdateOutboxItemStatus(ctx context.Context, arg db.UpdateOutboxItemStatusParams) (db.Outbox, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Outbox), args.Error(1)
}
func (m *MockQuerier) CreateReport(ctx context.Context, arg db.CreateReportParams) (db.Report, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Report), args.Error(1)
}
func (m *MockQuerier) CreateSchedule(ctx context.Context, arg db.CreateScheduleParams) (db.Schedule, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Schedule), args.Error(1)
}
func (m *MockQuerier) ListAllSchedules(ctx context.Context) ([]db.Schedule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Schedule), args.Error(1)
}
func (m *MockQuerier) AdminBulkDeleteSchedules(ctx context.Context, scheduleIds []int64) error {
	args := m.Called(ctx, scheduleIds)
	return args.Error(0)
}
func (m *MockQuerier) DeleteSchedule(ctx context.Context, scheduleID int64) error {
	args := m.Called(ctx, scheduleID)
	return args.Error(0)
}
func (m *MockQuerier) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockQuerier) DeleteUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
func (m *MockQuerier) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]db.GetSubscriptionsByUserRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.GetSubscriptionsByUserRow), args.Error(1)
}
func (m *MockQuerier) ListUsers(ctx context.Context, searchTerm interface{}) ([]db.User, error) {
	args := m.Called(ctx, searchTerm)
	return args.Get(0).([]db.User), args.Error(1)
}
func (m *MockQuerier) UpdateSchedule(ctx context.Context, arg db.UpdateScheduleParams) (db.Schedule, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Schedule), args.Error(1)
}
func (m *MockQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}
func (m *MockQuerier) UpsertSubscription(ctx context.Context, arg db.UpsertSubscriptionParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
func (m *MockQuerier) AdminBulkDeleteUsers(ctx context.Context, userIds []int64) error {
	args := m.Called(ctx, userIds)
	return args.Error(0)
}
func (m *MockQuerier) CreateRecurringAssignment(ctx context.Context, arg db.CreateRecurringAssignmentParams) (db.RecurringAssignment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.RecurringAssignment), args.Error(1)
}
func (m *MockQuerier) DeleteRecurringAssignment(ctx context.Context, recurringAssignmentID int64) error {
	args := m.Called(ctx, recurringAssignmentID)
	return args.Error(0)
}
func (m *MockQuerier) GetRecurringAssignmentByID(ctx context.Context, recurringAssignmentID int64) (db.RecurringAssignment, error) {
	args := m.Called(ctx, recurringAssignmentID)
	return args.Get(0).(db.RecurringAssignment), args.Error(1)
}
func (m *MockQuerier) GetRecurringAssignmentsByPattern(ctx context.Context, arg db.GetRecurringAssignmentsByPatternParams) ([]db.GetRecurringAssignmentsByPatternRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.GetRecurringAssignmentsByPatternRow), args.Error(1)
}
func (m *MockQuerier) ListRecurringAssignments(ctx context.Context) ([]db.RecurringAssignment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.RecurringAssignment), args.Error(1)
}
func (m *MockQuerier) ListRecurringAssignmentsByUserID(ctx context.Context, userID int64) ([]db.RecurringAssignment, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.RecurringAssignment), args.Error(1)
}
func (m *MockQuerier) UpdateRecurringAssignment(ctx context.Context, arg db.UpdateRecurringAssignmentParams) (db.RecurringAssignment, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.RecurringAssignment), args.Error(1)
}

// Add missing admin report methods
func (m *MockQuerier) AdminGetReportWithContext(ctx context.Context, reportID int64) (db.AdminGetReportWithContextRow, error) { panic("not implemented") }
func (m *MockQuerier) AdminListArchivedReportsWithContext(ctx context.Context) ([]db.AdminListArchivedReportsWithContextRow, error) { panic("not implemented") }
func (m *MockQuerier) AdminListReportsWithContext(ctx context.Context) ([]db.AdminListReportsWithContextRow, error) { panic("not implemented") }
func (m *MockQuerier) ArchiveReport(ctx context.Context, reportID int64) error { panic("not implemented") }
func (m *MockQuerier) BulkArchiveReports(ctx context.Context, reportIds []int64) error { panic("not implemented") }
func (m *MockQuerier) UnarchiveReport(ctx context.Context, reportID int64) error { panic("not implemented") }

// Add missing booking and broadcast methods
func (m *MockQuerier) GetBookingMetrics(ctx context.Context, arg db.GetBookingMetricsParams) (db.GetBookingMetricsRow, error) { panic("not implemented") }
func (m *MockQuerier) GetBookingPatternsByTimeSlot(ctx context.Context) ([]db.GetBookingPatternsByTimeSlotRow, error) { panic("not implemented") }
func (m *MockQuerier) GetBookingsInDateRange(ctx context.Context, arg db.GetBookingsInDateRangeParams) ([]db.GetBookingsInDateRangeRow, error) { panic("not implemented") }
func (m *MockQuerier) GetMemberContributions(ctx context.Context) ([]db.GetMemberContributionsRow, error) { panic("not implemented") }
func (m *MockQuerier) ListBookingsByUserIDWithSchedule(ctx context.Context, userID int64) ([]db.ListBookingsByUserIDWithScheduleRow, error) { panic("not implemented") }
func (m *MockQuerier) GetReportsForAutoArchiving(ctx context.Context) ([]db.GetReportsForAutoArchivingRow, error) { panic("not implemented") }

// Add missing broadcast methods
func (m *MockQuerier) CreateBroadcast(ctx context.Context, arg db.CreateBroadcastParams) (db.Broadcast, error) { panic("not implemented") }
func (m *MockQuerier) GetBroadcastByID(ctx context.Context, broadcastID int64) (db.Broadcast, error) { panic("not implemented") }
func (m *MockQuerier) ListBroadcasts(ctx context.Context) ([]db.Broadcast, error) { panic("not implemented") }
func (m *MockQuerier) ListBroadcastsWithSender(ctx context.Context) ([]db.ListBroadcastsWithSenderRow, error) { panic("not implemented") }
func (m *MockQuerier) ListPendingBroadcasts(ctx context.Context) ([]db.Broadcast, error) { panic("not implemented") }
func (m *MockQuerier) UpdateBroadcastStatus(ctx context.Context, arg db.UpdateBroadcastStatusParams) (db.Broadcast, error) { panic("not implemented") }

// Helper to create a test logger that discards output
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to create a default test config - ensure OTPValidityMinutes is present
func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret: "test-secret",
		OTPLogPath: "/dev/null", 
		DefaultShiftDuration: 2 * time.Hour,
		OTPValidityMinutes: 5, // Added for tests
        JWTExpirationHours: 24, // Added for tests that might use it
	}
}

// MockOTPStore allows for simpler testing of OTP expiry
type MockOTPStore struct {
	store   map[string]auth.OTPStoreEntry
	mu      sync.RWMutex
}

// Ensure MockOTPStore implements auth.OTPStore interface
var _ auth.OTPStore = (*MockOTPStore)(nil)

func NewMockOTPStore() *MockOTPStore {
	return &MockOTPStore{
		store: make(map[string]auth.OTPStoreEntry),
	}
}

func (s *MockOTPStore) StoreOTP(identifier string, otp string, validityDuration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[identifier] = auth.OTPStoreEntry{
		OTP:       otp,
		ExpiresAt: time.Now().Add(validityDuration),
	}
}

func (s *MockOTPStore) ValidateOTP(identifier string, otpToValidate string) bool {
	s.mu.RLock()
	entry, exists := s.store[identifier]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(entry.ExpiresAt) {
		s.mu.Lock()
		delete(s.store, identifier)
		s.mu.Unlock()
		return false
	}

	valid := entry.OTP == otpToValidate
	if valid {
		s.mu.Lock()
		delete(s.store, identifier) 
		s.mu.Unlock()
	}
	return valid
}

// ForceExpireOTP allows tests to explicitly expire an OTP
func (s *MockOTPStore) ForceExpireOTP(identifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, exists := s.store[identifier]; exists {
		entry.ExpiresAt = time.Now().Add(-1 * time.Minute) // Set to 1 minute in the past
		s.store[identifier] = entry
	}
}

// mockGenerateJWT always returns an error for testing
func mockGenerateJWT(userID int64, phone string, role string, secret string, expiryHours int) (string, error) {
	return "", errors.New("JWT generation failed")
}

func TestUserService_RegisterOrLoginUser_NewUser(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()

	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1234567890"
	name := "Test User"
	sqlName := sql.NullString{String: name, Valid: true}

	// Expected behavior for new user
	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(db.User{}, sql.ErrNoRows).Once()
	mockQuerier.On("CreateUser", mock.Anything, mock.AnythingOfType("db.CreateUserParams")).Run(func(args mock.Arguments) {
		params := args.Get(1).(db.CreateUserParams)
		assert.Equal(t, phone, params.Phone)
		assert.Equal(t, name, params.Name.String)
	}).Return(db.User{UserID: 1, Phone: phone, Name: sqlName}, nil).Once()
	mockQuerier.On("CreateOutboxItem", mock.Anything, mock.AnythingOfType("db.CreateOutboxItemParams")).Return(db.Outbox{}, nil).Once()

	err := userService.RegisterOrLoginUser(context.Background(), phone, sqlName)

	assert.NoError(t, err)
	mockQuerier.AssertExpectations(t)

	// Check if OTP was stored (OTP itself is random, so we check for existence and expiry logic implicitly)
	// This is a bit of an indirect check for s.otpStore.StoreOTP.
	// We can't get the OTP directly from the store without exporting fields or methods for testing.
	// For this test, we trust StoreOTP works if no error from RegisterOrLoginUser and CreateOutboxItem was called.
}

func TestUserService_RegisterOrLoginUser_ExistingUser(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()

	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1987654321"
	existingUser := db.User{UserID: 2, Phone: phone, Name: sql.NullString{String: "Existing User", Valid: true}}

	// Expected behavior for existing user
	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(existingUser, nil).Once()
	mockQuerier.On("CreateOutboxItem", mock.Anything, mock.AnythingOfType("db.CreateOutboxItemParams")).Return(db.Outbox{}, nil).Once()

	err := userService.RegisterOrLoginUser(context.Background(), phone, sql.NullString{})

	assert.NoError(t, err)
	mockQuerier.AssertExpectations(t)
	mockQuerier.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}

func TestUserService_VerifyOTP_Success(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore() 
	cfg := newTestConfig()
	testLogger := newTestLogger()
	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1122334455"
	otp, _ := auth.GenerateOTP()
	otpValidityDuration := time.Duration(cfg.OTPValidityMinutes) * time.Minute
	otpStore.StoreOTP(phone, otp, otpValidityDuration) // Pass duration

	expectedUser := db.User{UserID: 3, Phone: phone, Name: sql.NullString{String: "Verified User", Valid: true}}
	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(expectedUser, nil).Once()

	token, err := userService.VerifyOTP(context.Background(), phone, otp)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockQuerier.AssertExpectations(t)
	claims, err := auth.ValidateJWT(token, cfg.JWTSecret)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser.UserID, claims.UserID)
	assert.Equal(t, expectedUser.Phone, claims.Phone)
}

func TestUserService_VerifyOTP_InvalidOTP(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()
	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1555667788"
	otpValidityDuration := time.Duration(cfg.OTPValidityMinutes) * time.Minute
	otpStore.StoreOTP(phone, "123456", otpValidityDuration) // Pass duration

	token, err := userService.VerifyOTP(context.Background(), phone, "654321")

	assert.Error(t, err)
	assert.Equal(t, service.ErrOTPValidationFailed, err)
	assert.Empty(t, token)
	mockQuerier.AssertNotCalled(t, "GetUserByPhone", mock.Anything, mock.Anything)
}

func TestUserService_VerifyOTP_OTPExpired(t *testing.T) {
    mockQuerier := new(MockQuerier)
    mockOTPStore := NewMockOTPStore() // Use our custom mock that allows force-expiry
    cfg := newTestConfig()
    testLogger := newTestLogger()
    userService := service.NewUserService(mockQuerier, mockOTPStore, cfg, testLogger)

    phone := "+1555667788"
    otp, _ := auth.GenerateOTP()
    otpValidityDuration := time.Duration(cfg.OTPValidityMinutes) * time.Minute
    mockOTPStore.StoreOTP(phone, otp, otpValidityDuration)
    
    // Force the OTP to expire
    mockOTPStore.ForceExpireOTP(phone)
    
    // Now try to verify the expired OTP
    token, err := userService.VerifyOTP(context.Background(), phone, otp)

    assert.Error(t, err)
    assert.Equal(t, service.ErrOTPValidationFailed, err)
    assert.Empty(t, token)
    mockQuerier.AssertNotCalled(t, "GetUserByPhone", mock.Anything, mock.Anything)
}

func TestUserService_VerifyOTP_UserNotFoundAfterValidOTP(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()
	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1777888999"
	otp, _ := auth.GenerateOTP()
	otpValidityDuration := time.Duration(cfg.OTPValidityMinutes) * time.Minute
	otpStore.StoreOTP(phone, otp, otpValidityDuration) // Pass duration

	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(db.User{}, sql.ErrNoRows).Once()

	token, err := userService.VerifyOTP(context.Background(), phone, otp)

	assert.Error(t, err)
	assert.Equal(t, service.ErrUserNotFound, err)
	assert.Empty(t, token)
	mockQuerier.AssertExpectations(t)
}

func TestUserService_RegisterOrLoginUser_GetUserError(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()

	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1234500000"
	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(db.User{}, errors.New("some db error")).Once()

	err := userService.RegisterOrLoginUser(context.Background(), phone, sql.NullString{})

	assert.Error(t, err)
	assert.Equal(t, service.ErrInternalServer, err)
	mockQuerier.AssertExpectations(t)
}

func TestUserService_RegisterOrLoginUser_CreateUserError(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()

	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)

	phone := "+1234511111"
	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(db.User{}, sql.ErrNoRows).Once()
	mockQuerier.On("CreateUser", mock.Anything, mock.AnythingOfType("db.CreateUserParams")).Return(db.User{}, errors.New("create failed")).Once()

	err := userService.RegisterOrLoginUser(context.Background(), phone, sql.NullString{})

	assert.Error(t, err)
	assert.Equal(t, service.ErrInternalServer, err)
	mockQuerier.AssertExpectations(t)
}

func TestUserService_VerifyOTP_JWTGenerationError(t *testing.T) {
	mockQuerier := new(MockQuerier)
	otpStore := auth.NewInMemoryOTPStore()
	cfg := newTestConfig()
	testLogger := newTestLogger()
	userService := service.NewUserService(mockQuerier, otpStore, cfg, testLogger)
	
	// Inject a custom JWT generator that always fails
	userService.SetJWTGenerator(func(userID int64, phone string, role string, secret string, expiryHours int) (string, error) {
		return "", errors.New("forced JWT generation error")
	})

	phone := "+1122334455"
	otp, _ := auth.GenerateOTP()
	otpValidityDuration := time.Duration(cfg.OTPValidityMinutes) * time.Minute
	otpStore.StoreOTP(phone, otp, otpValidityDuration)

	// Mock the database to return a valid user
	expectedUser := db.User{UserID: 3, Phone: phone}
	mockQuerier.On("GetUserByPhone", mock.Anything, phone).Return(expectedUser, nil).Once()
	
	// Try to verify the OTP - our mocked JWT generator will fail
	token, err := userService.VerifyOTP(context.Background(), phone, otp)

	// Check that we got the expected error and empty token
	assert.Empty(t, token, "Token should be empty when JWT generation fails")
	assert.Error(t, err, "An error should be returned when JWT generation fails")
	assert.Equal(t, service.ErrInternalServer, err, "Error should be ErrInternalServer when JWT generation fails")
	mockQuerier.AssertExpectations(t)
} 