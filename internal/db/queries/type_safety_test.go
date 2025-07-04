package queries

import (
	"testing"

	db "night-owls-go/internal/db/sqlc_generated"
)

// TestDaysFromNowTypeSafety verifies that DaysFromNow field is int64 (not interface{})
// This test ensures type safety at compile time after the Copilot issue fix
func TestDaysFromNowTypeSafety(t *testing.T) {
	// This test verifies compile-time type safety for DaysFromNow field
	var booking db.GetBookingsInDateRangeRow
	
	// This should compile without issues since DaysFromNow is int64
	var daysFromNow int64 = booking.DaysFromNow
	
	// This test passes if it compiles successfully
	_ = daysFromNow
	t.Log("DaysFromNow field is correctly typed as int64")
}

// TestDaysFromNowUsage demonstrates safe usage of the DaysFromNow field
func TestDaysFromNowUsage(t *testing.T) {
	var booking db.GetBookingsInDateRangeRow
	booking.DaysFromNow = 5
	
	// These operations should work with int64 type
	if booking.DaysFromNow > 0 {
		t.Logf("Booking is %d days from now", booking.DaysFromNow)
	}
	
	// Arithmetic operations should work
	weekFromNow := booking.DaysFromNow + 7
	_ = weekFromNow
	
	t.Log("DaysFromNow field supports proper int64 operations")
} 