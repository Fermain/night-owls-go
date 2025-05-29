package outbox

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// MessageSender defines an interface for sending messages.
type MessageSender interface {
	Send(recipient, messageType, payload string) error
}

// LogFileMessageSender is an implementation of MessageSender that writes messages to a log file.
type LogFileMessageSender struct {
	logFilePath string
	logger      *slog.Logger
}

// NewLogFileMessageSender creates a new LogFileMessageSender.
// It ensures the directory for the log file exists.
func NewLogFileMessageSender(logFilePath string, logger *slog.Logger) (*LogFileMessageSender, error) {
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for OTP log file %s: %w", dir, err)
	}

	return &LogFileMessageSender{
		logFilePath: logFilePath,
		logger:      logger.With("component", "LogFileMessageSender"),
	}, nil
}

// Send writes the message details to the configured log file.
func (s *LogFileMessageSender) Send(recipient, messageType, payload string) error {
	logMessage := fmt.Sprintf("[%s] To: %s, Type: %s, Payload: %s\n",
		time.Now().Format(time.RFC3339),
		recipient,
		messageType,
		payload,
	)

	file, err := os.OpenFile(s.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		s.logger.Error("Failed to open OTP log file for writing", "path", s.logFilePath, "error", err)
		return fmt.Errorf("failed to open OTP log file %s: %w", s.logFilePath, err)
	}
	defer file.Close()

	if _, err := file.WriteString(logMessage); err != nil {
		s.logger.Error("Failed to write to OTP log file", "path", s.logFilePath, "error", err)
		return fmt.Errorf("failed to write to OTP log file %s: %w", s.logFilePath, err)
	}

	s.logger.Info("Message written to log file (mock send)", "recipient", recipient, "type", messageType, "path", s.logFilePath)
	return nil
}
