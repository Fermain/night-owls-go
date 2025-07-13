package queries

import (
	"fmt"
	"testing"

	db "night-owls-go/internal/db/sqlc_generated"
)

// TestDaysFromNowTypeSafety verifies that DaysFromNow field can be safely converted to int64
// This test ensures type safety when working with the DaysFromNow field
// NOTE: DaysFromNow is interface{} due to complex SQL expression with COALESCE/CAST
// This is expected behavior - sqlc cannot infer concrete types from complex expressions
func TestDaysFromNowTypeSafety(t *testing.T) {
	// DaysFromNow is interface{} due to complex SQL COALESCE/CAST expression
	var booking db.GetBookingsInDateRangeRow

	// Simulate actual data that would come from the database
	booking.DaysFromNow = int64(5)

	// Safe type assertion with check
	if daysFromNow, ok := booking.DaysFromNow.(int64); ok {
		_ = daysFromNow
		t.Log("DaysFromNow field safely converted to int64")
	} else {
		t.Errorf("DaysFromNow field is not int64, got %T", booking.DaysFromNow)
	}
}

// TestDaysFromNowUsage demonstrates safe usage of the DaysFromNow field
func TestDaysFromNowUsage(t *testing.T) {
	var booking db.GetBookingsInDateRangeRow
	booking.DaysFromNow = int64(5)

	// Safe type assertion before operations
	if daysFromNow, ok := booking.DaysFromNow.(int64); ok {
		// These operations work safely with type assertion
		if daysFromNow > 0 {
			t.Logf("Booking is %d days from now", daysFromNow)
		}

		// Arithmetic operations work after type assertion
		weekFromNow := daysFromNow + 7
		_ = weekFromNow

		t.Log("DaysFromNow field supports proper int64 operations after type assertion")
	} else {
		t.Errorf("Failed to assert DaysFromNow as int64, got %T", booking.DaysFromNow)
	}
}

// TestDaysFromNowHelper demonstrates the recommended helper function pattern
func TestDaysFromNowHelper(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected int64
		hasError bool
	}{
		{"valid int64", int64(5), 5, false},
		{"valid int", int(3), 3, false},
		{"valid float64", float64(7), 7, false},
		{"nil value", nil, 0, false},
		{"invalid string", "invalid", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var booking db.GetBookingsInDateRangeRow
			booking.DaysFromNow = tc.value

			result, err := getDaysFromNowSafe(booking)

			if tc.hasError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %d, got %d", tc.expected, result)
			}
		})
	}
}

// getDaysFromNowSafe is a helper function that safely extracts DaysFromNow as int64
// This demonstrates the recommended pattern for production code
func getDaysFromNowSafe(booking db.GetBookingsInDateRangeRow) (int64, error) {
	if booking.DaysFromNow == nil {
		return 0, nil
	}

	switch v := booking.DaysFromNow.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("DaysFromNow has unexpected type: %T", v)
	}
}
