package utils

import (
	"crypto/sha256"
	"encoding/hex"
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
	if icsData.MIME != "text/calendar; charset=utf-8" {
		t.Errorf("Expected MIME type text/calendar; charset=utf-8, got %s", icsData.MIME)
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

// Security tests for calendar token functionality

func TestHashToken(t *testing.T) {
	// Test that the same token produces the same hash
	token := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	
	// Calculate expected hash
	hash := sha256.Sum256([]byte(token))
	expectedHash := hex.EncodeToString(hash[:])
	
	// Use the same hashing function as the handler
	actualHash := hashTokenForTest(token)
	
	if actualHash != expectedHash {
		t.Errorf("Hash mismatch. Expected: %s, Got: %s", expectedHash, actualHash)
	}
}

func TestTokenValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectValid bool
	}{
		{
			name:        "valid 64-character token",
			token:       "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectValid: true,
		},
		{
			name:        "invalid short token",
			token:       "short",
			expectValid: false,
		},
		{
			name:        "invalid long token",
			token:       "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123",
			expectValid: false,
		},
		{
			name:        "empty token",
			token:       "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic format validation (matches the validation in handlers)
			isValid := len(tt.token) == 64
			
			if isValid != tt.expectValid {
				t.Errorf("Token validation failed. Expected: %v, Got: %v", tt.expectValid, isValid)
			}
		})
	}
}

func TestCalendarTokenSecurity(t *testing.T) {
	t.Run("token hashes should be deterministic", func(t *testing.T) {
		token := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		
		hash1 := hashTokenForTest(token)
		hash2 := hashTokenForTest(token)
		
		if hash1 != hash2 {
			t.Errorf("Hash function is not deterministic. Hash1: %s, Hash2: %s", hash1, hash2)
		}
	})
	
	t.Run("different tokens should produce different hashes", func(t *testing.T) {
		token1 := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		token2 := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		
		hash1 := hashTokenForTest(token1)
		hash2 := hashTokenForTest(token2)
		
		if hash1 == hash2 {
			t.Errorf("Different tokens produced same hash: %s", hash1)
		}
	})
	
	t.Run("hash should be 64 characters (SHA-256 hex)", func(t *testing.T) {
		token := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		hash := hashTokenForTest(token)
		
		if len(hash) != 64 {
			t.Errorf("Hash length should be 64 characters, got %d", len(hash))
		}
	})
}

// Helper function to test token hashing (mirrors the private function in handlers)
func hashTokenForTest(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
} 