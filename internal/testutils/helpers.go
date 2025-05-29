// Package testutils provides shared utility functions for tests and development tools
package testutils

import (
	"database/sql"
	db "night-owls-go/internal/db/sqlc_generated"
)

// NewNullInt64 creates a valid sql.NullInt64 from an int64 value
func NewNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

// NewNullString creates a valid sql.NullString from a string value
func NewNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// NewCreateReportParams creates a db.CreateReportParams with the common fields filled
func NewCreateReportParams(bookingID, userID int64, severity int64, message string) db.CreateReportParams {
	return db.CreateReportParams{
		BookingID: NewNullInt64(bookingID),
		UserID:    NewNullInt64(userID),
		Severity:  severity,
		Message:   NewNullString(message),
	}
}

// NewCreateUserParams creates a db.CreateUserParams with the common fields filled
func NewCreateUserParams(phone, name, role string) db.CreateUserParams {
	return db.CreateUserParams{
		Phone: phone,
		Name:  NewNullString(name),
		Role:  NewNullString(role),
	}
} 