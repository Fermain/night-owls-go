package outbox_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"

	"night-owls-go/internal/config"
	db "night-owls-go/internal/db/sqlc_generated"
	"night-owls-go/internal/outbox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOutboxQuerier is a focused mock for Outbox Dispatcher tests.
type MockOutboxQuerier struct {
	mock.Mock
}

func (m *MockOutboxQuerier) GetPendingOutboxItems(ctx context.Context, limit int64) ([]db.Outbox, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]db.Outbox), args.Error(1)
}

func (m *MockOutboxQuerier) UpdateOutboxItemStatus(ctx context.Context, arg db.UpdateOutboxItemStatusParams) (db.Outbox, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Outbox), args.Error(1)
}

// Stubs for other Querier methods (not used by OutboxDispatcherService)
func (m *MockOutboxQuerier) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {panic("not impl")} 
func (m *MockOutboxQuerier) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {panic("not impl")} 
func (m *MockOutboxQuerier) CreateOutboxItem(ctx context.Context, arg db.CreateOutboxItemParams) (db.Outbox, error) {panic("not impl")} 
func (m *MockOutboxQuerier) CreateBooking(ctx context.Context, arg db.CreateBookingParams) (db.Booking, error) {panic("not impl")} 
func (m *MockOutboxQuerier) GetBookingByID(ctx context.Context, bookingID int64) (db.Booking, error) {panic("not impl")} 
func (m *MockOutboxQuerier) GetReportByBookingID(ctx context.Context, bookingID int64) (db.Report, error) {panic("not impl")} 
func (m *MockOutboxQuerier) GetScheduleByID(ctx context.Context, scheduleID int64) (db.Schedule, error) {panic("not impl")} 
func (m *MockOutboxQuerier) GetUserByID(ctx context.Context, userID int64) (db.User, error) {panic("not impl")} 
func (m *MockOutboxQuerier) ListActiveSchedules(ctx context.Context, arg db.ListActiveSchedulesParams) ([]db.Schedule, error) {panic("not impl")} 
func (m *MockOutboxQuerier) ListBookingsByUserID(ctx context.Context, userID int64) ([]db.Booking, error) {panic("not impl")} 
func (m *MockOutboxQuerier) ListReportsByUserID(ctx context.Context, userID int64) ([]db.Report, error) {panic("not impl")} 
func (m *MockOutboxQuerier) UpdateBookingAttendance(ctx context.Context, arg db.UpdateBookingAttendanceParams) (db.Booking, error) {panic("not impl")} 
func (m *MockOutboxQuerier) CreateReport(ctx context.Context, arg db.CreateReportParams) (db.Report, error) {panic("not impl")} 
func (m *MockOutboxQuerier) CreateSchedule(ctx context.Context, arg db.CreateScheduleParams) (db.Schedule, error) {panic("not impl")} 
func (m *MockOutboxQuerier) GetBookingByScheduleAndStartTime(ctx context.Context, arg db.GetBookingByScheduleAndStartTimeParams) (db.Booking, error) {panic("not impl")}
func (m *MockOutboxQuerier) ListAllSchedules(ctx context.Context) ([]db.Schedule, error) {panic("not impl")}
func (m *MockOutboxQuerier) AdminBulkDeleteSchedules(ctx context.Context, scheduleIds []int64) error {panic("not impl")}
func (m *MockOutboxQuerier) DeleteSchedule(ctx context.Context, scheduleID int64) error {panic("not impl")}
func (m *MockOutboxQuerier) DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) error {panic("not impl")}
func (m *MockOutboxQuerier) DeleteUser(ctx context.Context, userID int64) error {panic("not impl")}
func (m *MockOutboxQuerier) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]db.GetSubscriptionsByUserRow, error) {panic("not impl")}
func (m *MockOutboxQuerier) ListUsers(ctx context.Context, searchTerm interface{}) ([]db.User, error) {panic("not impl")}
func (m *MockOutboxQuerier) UpdateSchedule(ctx context.Context, arg db.UpdateScheduleParams) (db.Schedule, error) {panic("not impl")}
func (m *MockOutboxQuerier) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {panic("not impl")}
func (m *MockOutboxQuerier) UpsertSubscription(ctx context.Context, arg db.UpsertSubscriptionParams) error {panic("not impl")}

// MockMessageSender is a mock implementation of the MessageSender interface.
type MockMessageSender struct {
	mock.Mock
}

func (m *MockMessageSender) Send(recipient, messageType, payload string) error {
	args := m.Called(recipient, messageType, payload)
	return args.Error(0)
}

// newOutboxTestDeps creates logger and a basic config for outbox tests.
func newOutboxTestDeps() (*slog.Logger, *config.Config) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{ // Provide defaults relevant to outbox
		OutboxBatchSize: 10, 
		OutboxMaxRetries: 3,
	}
	return logger, cfg
}

func TestDispatcherService_ProcessPendingOutboxMessages_NoItems(t *testing.T) {
	mockQuerier := new(MockOutboxQuerier)
	mockSender := new(MockMessageSender)
	testLogger, cfg := newOutboxTestDeps()

	dispatcher := outbox.NewDispatcherService(mockQuerier, mockSender, nil, testLogger, cfg)

	mockQuerier.On("GetPendingOutboxItems", mock.Anything, int64(cfg.OutboxBatchSize)).Return([]db.Outbox{}, nil).Once()

	processed, errors := dispatcher.ProcessPendingOutboxMessages(context.Background())

	assert.Equal(t, 0, processed)
	assert.Equal(t, 0, errors)
	mockQuerier.AssertExpectations(t)
	mockSender.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestDispatcherService_ProcessPendingOutboxMessages_SendSuccess(t *testing.T) {
	mockQuerier := new(MockOutboxQuerier)
	mockSender := new(MockMessageSender)
	testLogger, cfg := newOutboxTestDeps()
	dispatcher := outbox.NewDispatcherService(mockQuerier, mockSender, nil, testLogger, cfg)

	item1 := db.Outbox{OutboxID: 1, Recipient: "r1", MessageType: "sms", Payload: sql.NullString{String: "p1", Valid: true}, Status: "pending"}
	item2 := db.Outbox{OutboxID: 2, Recipient: "r2", MessageType: "sms", Payload: sql.NullString{String: "p2", Valid: true}, Status: "pending"}
	pendingItems := []db.Outbox{item1, item2}

	mockQuerier.On("GetPendingOutboxItems", mock.Anything, int64(cfg.OutboxBatchSize)).Return(pendingItems, nil).Once()

	// Expect Send to be called for each item
	mockSender.On("Send", item1.Recipient, item1.MessageType, item1.Payload.String).Return(nil).Once()
	mockSender.On("Send", item2.Recipient, item2.MessageType, item2.Payload.String).Return(nil).Once()

	// Expect UpdateOutboxItemStatus to be called for each item with status "sent"
	mockQuerier.On("UpdateOutboxItemStatus", mock.Anything, mock.MatchedBy(func(params db.UpdateOutboxItemStatusParams) bool {
		return params.OutboxID == item1.OutboxID && params.Status == "sent" && params.SentAt.Valid
	})).Return(db.Outbox{}, nil).Once()
	mockQuerier.On("UpdateOutboxItemStatus", mock.Anything, mock.MatchedBy(func(params db.UpdateOutboxItemStatusParams) bool {
		return params.OutboxID == item2.OutboxID && params.Status == "sent" && params.SentAt.Valid
	})).Return(db.Outbox{}, nil).Once()

	processed, errCount := dispatcher.ProcessPendingOutboxMessages(context.Background())

	assert.Equal(t, 2, processed)
	assert.Equal(t, 0, errCount)
	mockQuerier.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestDispatcherService_ProcessPendingOutboxMessages_SendFailure_Retry(t *testing.T) {
	mockQuerier := new(MockOutboxQuerier)
	mockSender := new(MockMessageSender)
	testLogger, cfg := newOutboxTestDeps()
	dispatcher := outbox.NewDispatcherService(mockQuerier, mockSender, nil, testLogger, cfg)

	item1 := db.Outbox{OutboxID: 1, Recipient: "r1", MessageType: "sms", Payload: sql.NullString{String: "p1", Valid: true}, Status: "pending", RetryCount: sql.NullInt64{Int64: 0, Valid: true}}
	pendingItems := []db.Outbox{item1}

	mockQuerier.On("GetPendingOutboxItems", mock.Anything, int64(cfg.OutboxBatchSize)).Return(pendingItems, nil).Once()

	// Send fails for item1
	mockSender.On("Send", item1.Recipient, item1.MessageType, item1.Payload.String).Return(errors.New("send failed")).Once()

	// Expect UpdateOutboxItemStatus to be called with status "failed" and incremented retry count
	mockQuerier.On("UpdateOutboxItemStatus", mock.Anything, mock.MatchedBy(func(params db.UpdateOutboxItemStatusParams) bool {
		return params.OutboxID == item1.OutboxID && params.Status == "failed" && params.RetryCount.Int64 == 1
	})).Return(db.Outbox{}, nil).Once()

	processed, errCount := dispatcher.ProcessPendingOutboxMessages(context.Background())

	assert.Equal(t, 0, processed)
	assert.Equal(t, 1, errCount)
	mockQuerier.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestDispatcherService_ProcessPendingOutboxMessages_SendFailure_MaxRetries(t *testing.T) {
	mockQuerier := new(MockOutboxQuerier)
	mockSender := new(MockMessageSender)
	testLogger, cfg := newOutboxTestDeps()
	// Use a specific config for this test to control max retries easily
	customCfg := *cfg // copy base
	customCfg.OutboxMaxRetries = 1 // Set max retries to 1 for this test
	dispatcher := outbox.NewDispatcherService(mockQuerier, mockSender, nil, testLogger, &customCfg)

	item1 := db.Outbox{OutboxID: 1, Recipient: "r1", MessageType: "sms", Payload: sql.NullString{String: "p1", Valid: true}, Status: "pending", RetryCount: sql.NullInt64{Int64: int64(customCfg.OutboxMaxRetries -1) , Valid: true}} 
	pendingItems := []db.Outbox{item1}

	mockQuerier.On("GetPendingOutboxItems", mock.Anything, int64(customCfg.OutboxBatchSize)).Return(pendingItems, nil).Once()
	mockSender.On("Send", item1.Recipient, item1.MessageType, item1.Payload.String).Return(errors.New("send failed again")).Once()
	mockQuerier.On("UpdateOutboxItemStatus", mock.Anything, mock.MatchedBy(func(params db.UpdateOutboxItemStatusParams) bool {
		return params.OutboxID == item1.OutboxID && params.Status == "permanently_failed" && params.RetryCount.Int64 == int64(customCfg.OutboxMaxRetries)
	})).Return(db.Outbox{}, nil).Once()

	processed, errCount := dispatcher.ProcessPendingOutboxMessages(context.Background())

	assert.Equal(t, 0, processed)
	assert.Equal(t, 1, errCount)
	mockQuerier.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestDispatcherService_ProcessPendingOutboxMessages_GetPendingError(t *testing.T) {
	mockQuerier := new(MockOutboxQuerier)
	mockSender := new(MockMessageSender)
	testLogger, cfg := newOutboxTestDeps()
	dispatcher := outbox.NewDispatcherService(mockQuerier, mockSender, nil, testLogger, cfg)

	mockQuerier.On("GetPendingOutboxItems", mock.Anything, int64(cfg.OutboxBatchSize)).Return([]db.Outbox{}, errors.New("db error")).Once()

	processed, errCount := dispatcher.ProcessPendingOutboxMessages(context.Background())

	assert.Equal(t, 0, processed)
	assert.Equal(t, 1, errCount) // Error from GetPendingOutboxItems itself
	mockQuerier.AssertExpectations(t)
	mockSender.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestDispatcherService_ProcessPendingOutboxMessages_UpdateStatusError(t *testing.T) {
	mockQuerier := new(MockOutboxQuerier)
	mockSender := new(MockMessageSender)
	testLogger, cfg := newOutboxTestDeps()
	dispatcher := outbox.NewDispatcherService(mockQuerier, mockSender, nil, testLogger, cfg)

	item1 := db.Outbox{OutboxID: 1, Recipient: "r1", MessageType: "sms", Payload: sql.NullString{String: "p1", Valid: true}, Status: "pending"}
	pendingItems := []db.Outbox{item1}

	mockQuerier.On("GetPendingOutboxItems", mock.Anything, int64(cfg.OutboxBatchSize)).Return(pendingItems, nil).Once()
	mockSender.On("Send", item1.Recipient, item1.MessageType, item1.Payload.String).Return(nil).Once()
	mockQuerier.On("UpdateOutboxItemStatus", mock.Anything, mock.AnythingOfType("db.UpdateOutboxItemStatusParams")).Return(db.Outbox{}, errors.New("update failed")).Once()

	processed, errCount := dispatcher.ProcessPendingOutboxMessages(context.Background())

	assert.Equal(t, 1, processed) // Processed because Send was successful
	assert.Equal(t, 1, errCount)   // Error from UpdateOutboxItemStatus
	mockQuerier.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

// TODO: Test context cancellation during Send operation if DispatcherService respects it. 