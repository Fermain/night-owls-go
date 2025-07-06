package utils

import (
	"strings"
	"testing"
	"time"

	"database/sql"
	db "night-owls-go/internal/db/sqlc_generated"
)

func TestGenerateBookingICS(t *testing.T) {
	// Create a test booking
	shiftStart := time.Date(2024, 12, 25, 18, 0, 0, 0, time.UTC)
	shiftEnd := time.Date(2024, 12, 25, 20, 0, 0, 0, time.UTC)
	
	booking := db.Booking{
		BookingID:  12345,
		UserID:     1,
		ScheduleID: 1,
		ShiftStart: shiftStart,
		ShiftEnd:   shiftEnd,
		BuddyName:  sql.NullString{String: "John Doe", Valid: true},
		CreatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}
	
	scheduleName := "Evening Patrol"
	
	// Generate ICS data
	icsData := GenerateBookingICS(booking, scheduleName)
	
	// Verify filename
	expectedFilename := "night-owls-shift-12345.ics"
	if icsData.Filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, icsData.Filename)
	}
	
	// Verify MIME type
	if icsData.MIME != "text/calendar" {
		t.Errorf("Expected MIME type text/calendar, got %s", icsData.MIME)
	}
	
	// Verify ICS content structure
	content := icsData.Content
	
	// Check for required ICS headers
	if !strings.Contains(content, "BEGIN:VCALENDAR") {
		t.Error("ICS content missing BEGIN:VCALENDAR")
	}
	
	if !strings.Contains(content, "END:VCALENDAR") {
		t.Error("ICS content missing END:VCALENDAR")
	}
	
	if !strings.Contains(content, "BEGIN:VEVENT") {
		t.Error("ICS content missing BEGIN:VEVENT")
	}
	
	if !strings.Contains(content, "END:VEVENT") {
		t.Error("ICS content missing END:VEVENT")
	}
	
	// Check for required event fields
	if !strings.Contains(content, "SUMMARY:Night Owls Shift - Evening Patrol") {
		t.Error("ICS content missing expected summary")
	}
	
	if !strings.Contains(content, "LOCATION:Mount Moreland Community Watch Area") {
		t.Error("ICS content missing expected location")
	}
	
	if !strings.Contains(content, "UID:nightowls-shift-12345@nightowls.app") {
		t.Error("ICS content missing expected UID")
	}
	
	// Check for buddy information in description
	if !strings.Contains(content, "Buddy: John Doe") {
		t.Error("ICS content missing buddy information")
	}
	
	// Check for reminder
	if !strings.Contains(content, "BEGIN:VALARM") {
		t.Error("ICS content missing reminder alarm")
	}
	
	t.Logf("Generated ICS content:\n%s", content)
}

func TestBookingToCalendarEvent(t *testing.T) {
	// Create a test booking
	shiftStart := time.Date(2024, 12, 25, 18, 0, 0, 0, time.UTC)
	shiftEnd := time.Date(2024, 12, 25, 20, 0, 0, 0, time.UTC)
	
	booking := db.Booking{
		BookingID:  12345,
		UserID:     1,
		ScheduleID: 1,
		ShiftStart: shiftStart,
		ShiftEnd:   shiftEnd,
		BuddyName:  sql.NullString{String: "John Doe", Valid: true},
		CreatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}
	
	scheduleName := "Evening Patrol"
	
	// Convert to calendar event
	event := BookingToCalendarEvent(booking, scheduleName)
	
	// Verify event details
	expectedTitle := "Night Owls Shift - Evening Patrol"
	if event.Title != expectedTitle {
		t.Errorf("Expected title %s, got %s", expectedTitle, event.Title)
	}
	
	if event.StartTime != shiftStart {
		t.Errorf("Expected start time %v, got %v", shiftStart, event.StartTime)
	}
	
	if event.EndTime != shiftEnd {
		t.Errorf("Expected end time %v, got %v", shiftEnd, event.EndTime)
	}
	
	if event.Location != "Mount Moreland Community Watch Area" {
		t.Errorf("Expected location 'Mount Moreland Community Watch Area', got %s", event.Location)
	}
	
	expectedUID := "nightowls-shift-12345@nightowls.app"
	if event.UID != expectedUID {
		t.Errorf("Expected UID %s, got %s", expectedUID, event.UID)
	}
	
	// Check for buddy in attendees
	if len(event.Attendees) != 1 || event.Attendees[0].Name != "John Doe" {
		t.Error("Expected buddy in attendees list")
	}
	
	// Check reminder
	if event.ReminderMin != 60 {
		t.Errorf("Expected 60 minute reminder, got %d", event.ReminderMin)
	}
}

func TestEscapeICSText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple text", "Simple text"},
		{"Text with; semicolon", "Text with\\; semicolon"},
		{"Text with, comma", "Text with\\, comma"},
		{"Text with\nnewline", "Text with\\nnewline"},
		{"Text with \\ backslash", "Text with \\\\ backslash"},
	}
	
	for _, test := range tests {
		result := escapeICSText(test.input)
		if result != test.expected {
			t.Errorf("escapeICSText(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFormatICSDate(t *testing.T) {
	testTime := time.Date(2024, 12, 25, 18, 30, 45, 0, time.UTC)
	expected := "20241225T183045Z"
	
	result := formatICSDate(testTime)
	if result != expected {
		t.Errorf("formatICSDate(%v) = %s, expected %s", testTime, result, expected)
	}
} 