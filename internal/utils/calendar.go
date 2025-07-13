// Package utils provides utility functions for calendar generation
package utils

import (
	"fmt"
	"strings"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	UID         string
	Organizer   CalendarContact
	Attendees   []CalendarContact
	ReminderMin int // Minutes before event
}

// CalendarContact represents a contact in calendar
type CalendarContact struct {
	Name  string
	Email string
}

// ICSData contains calendar file information
type ICSData struct {
	Filename string
	Content  string
	MIME     string
}

// BookingToCalendarEvent converts a booking to a calendar event
func BookingToCalendarEvent(booking db.Booking, scheduleName string) CalendarEvent {
	title := fmt.Sprintf("Night Owls Shift - %s", scheduleName)
	
	// Build detailed description
	var descBuilder strings.Builder
	descBuilder.WriteString(fmt.Sprintf("Night Owls Community Watch - %s\\n\\n", scheduleName))
	descBuilder.WriteString(fmt.Sprintf("⏰ Time: %s - %s\\n", 
		booking.ShiftStart.Format("15:04"), 
		booking.ShiftEnd.Format("15:04")))
	
	if booking.BuddyName.Valid && booking.BuddyName.String != "" {
		descBuilder.WriteString(fmt.Sprintf("👥 Buddy: %s\\n", booking.BuddyName.String))
	}
	
	descBuilder.WriteString("\\n📱 Check in on the Night Owls app when your shift starts")
	descBuilder.WriteString("\\n🚨 Report any incidents through the app")
	descBuilder.WriteString("\\n\\n🔗 Night Owls App: https://mm.nightowls.app")
	
	// Create attendees list
	var attendees []CalendarContact
	if booking.BuddyName.Valid && booking.BuddyName.String != "" {
		attendees = append(attendees, CalendarContact{
			Name:  booking.BuddyName.String,
			Email: "",
		})
	}
	
	return CalendarEvent{
		Title:       title,
		Description: descBuilder.String(),
		StartTime:   booking.ShiftStart,
		EndTime:     booking.ShiftEnd,
		Location:    "Mount Moreland Community Watch Area",
		UID:         fmt.Sprintf("nightowls-shift-%d@mm.nightowls.app", booking.BookingID),
		Organizer: CalendarContact{
			Name:  "Night Owls Scheduler",
			Email: "noreply@mm.nightowls.app",
		},
		Attendees:   attendees,
		ReminderMin: 60, // 1 hour before shift
	}
}

// GenerateICS creates an .ics file content from a calendar event
func GenerateICS(event CalendarEvent) string {
	now := time.Now().UTC()
	
	var icsBuilder strings.Builder
	
	// ICS Header
	icsBuilder.WriteString("BEGIN:VCALENDAR\r\n")
	icsBuilder.WriteString("VERSION:2.0\r\n")
	icsBuilder.WriteString("PRODID:-//Night Owls//Shift Calendar//EN\r\n")
	icsBuilder.WriteString("CALSCALE:GREGORIAN\r\n")
	icsBuilder.WriteString("METHOD:PUBLISH\r\n")
	
	// Event
	icsBuilder.WriteString("BEGIN:VEVENT\r\n")
	icsBuilder.WriteString(fmt.Sprintf("UID:%s\r\n", event.UID))
	icsBuilder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", formatICSDate(now)))
	icsBuilder.WriteString(fmt.Sprintf("DTSTART:%s\r\n", formatICSDate(event.StartTime)))
	icsBuilder.WriteString(fmt.Sprintf("DTEND:%s\r\n", formatICSDate(event.EndTime)))
	icsBuilder.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICSText(event.Title)))
	icsBuilder.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICSText(event.Description)))
	
	if event.Location != "" {
		icsBuilder.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICSText(event.Location)))
	}
	
	// Organizer
	if event.Organizer.Email != "" {
		icsBuilder.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:mailto:%s\r\n", 
			escapeICSText(event.Organizer.Name), event.Organizer.Email))
	}
	
	// Attendees
	for _, attendee := range event.Attendees {
		if attendee.Email != "" {
			icsBuilder.WriteString(fmt.Sprintf("ATTENDEE;CN=%s:mailto:%s\r\n", 
				escapeICSText(attendee.Name), attendee.Email))
		} else {
			icsBuilder.WriteString(fmt.Sprintf("ATTENDEE;CN=%s\r\n", 
				escapeICSText(attendee.Name)))
		}
	}
	
	// Reminder
	if event.ReminderMin > 0 {
		icsBuilder.WriteString("BEGIN:VALARM\r\n")
		icsBuilder.WriteString(fmt.Sprintf("TRIGGER:-PT%dM\r\n", event.ReminderMin))
		icsBuilder.WriteString("ACTION:DISPLAY\r\n")
		icsBuilder.WriteString(fmt.Sprintf("DESCRIPTION:%s starts in %d minutes\r\n", 
			escapeICSText(event.Title), event.ReminderMin))
		icsBuilder.WriteString("END:VALARM\r\n")
	}
	
	icsBuilder.WriteString("END:VEVENT\r\n")
	icsBuilder.WriteString("END:VCALENDAR\r\n")
	
	return icsBuilder.String()
}

// GenerateBookingICS creates an ICS file for a single booking
func GenerateBookingICS(booking db.Booking, scheduleName string) ICSData {
	event := BookingToCalendarEvent(booking, scheduleName)
	content := GenerateICS(event)
	filename := fmt.Sprintf("night-owls-shift-%d.ics", booking.BookingID)
	
	return ICSData{
		Filename: filename,
		Content:  content,
		MIME:     "text/calendar; charset=utf-8",
	}
}

// GenerateUserCalendarFeed creates a multi-event ICS feed for WebCal subscription
func GenerateUserCalendarFeed(bookings []db.ListBookingsByUserIDWithScheduleRow, userID int64) ICSData {
	if len(bookings) == 0 {
		// Empty calendar
		content := generateEmptyCalendar(userID)
		return ICSData{
			Filename: fmt.Sprintf("night-owls-calendar-%d.ics", userID),
			Content:  content,
			MIME:     "text/calendar; charset=utf-8",
		}
	}
	
	// Generate header
	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Night Owls Community Watch//Calendar Feed//EN\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")
	sb.WriteString("X-WR-CALNAME:Night Owls Shifts\r\n")
	sb.WriteString("X-WR-CALDESC:Your upcoming Night Owls Community Watch shifts\r\n")
	sb.WriteString("X-WR-TIMEZONE:Africa/Johannesburg\r\n")
	
	// Add each booking as an event
	for _, booking := range bookings {
		// Only include future shifts in WebCal feed
		if booking.ShiftStart.After(time.Now()) {
			event := bookingRowToCalendarEvent(booking)
			eventContent := generateVEvent(event)
			sb.WriteString(eventContent)
		}
	}
	
	sb.WriteString("END:VCALENDAR\r\n")
	
	return ICSData{
		Filename: fmt.Sprintf("night-owls-calendar-%d.ics", userID),
		Content:  sb.String(),
		MIME:     "text/calendar; charset=utf-8",
	}
}

// bookingRowToCalendarEvent converts a booking row to a calendar event
func bookingRowToCalendarEvent(booking db.ListBookingsByUserIDWithScheduleRow) CalendarEvent {
	// Calculate shift end time (assuming 2-hour shifts if not specified)
	shiftEnd := booking.ShiftEnd
	if shiftEnd.IsZero() {
		shiftEnd = booking.ShiftStart.Add(2 * time.Hour)
	}
	
	// Build description
	var description strings.Builder
	description.WriteString(fmt.Sprintf("Night Owls Community Watch shift for %s\\n\\n", booking.ScheduleName))
	description.WriteString(fmt.Sprintf("📅 Date: %s\\n", booking.ShiftStart.Format("Monday, 2 January 2006")))
	description.WriteString(fmt.Sprintf("🕒 Time: %s - %s\\n", 
		booking.ShiftStart.Format("15:04"), 
		shiftEnd.Format("15:04")))
	
	if booking.BuddyName.Valid && booking.BuddyName.String != "" {
		description.WriteString(fmt.Sprintf("👥 Buddy: %s\\n", escapeICSText(booking.BuddyName.String)))
	}
	
	description.WriteString("\\n📱 Check in through the Night Owls app when your shift starts.")
	description.WriteString("\\n🔗 App: https://mm.nightowls.app")
	
	// Generate unique UID for this booking
	uid := fmt.Sprintf("night-owls-booking-%d@mm.nightowls.app", booking.BookingID)
	
	return CalendarEvent{
		Title:       fmt.Sprintf("Night Owls Shift - %s", booking.ScheduleName),
		Description: description.String(),
		StartTime:   booking.ShiftStart,
		EndTime:     shiftEnd,
		Location:    "Mount Moreland Community Watch Area",
		UID:         uid,
		Organizer: CalendarContact{
			Name:  "Night Owls Scheduler",
			Email: "noreply@mm.nightowls.app",
		},
		Attendees:   []CalendarContact{},
		ReminderMin: 60, // 1 hour before
	}
}

// generateEmptyCalendar creates an empty calendar for users with no shifts
func generateEmptyCalendar(userID int64) string {
	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Night Owls Community Watch//Calendar Feed//EN\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")
	sb.WriteString("X-WR-CALNAME:Night Owls Shifts\r\n")
	sb.WriteString("X-WR-CALDESC:Your upcoming Night Owls Community Watch shifts\r\n")
	sb.WriteString("X-WR-TIMEZONE:Africa/Johannesburg\r\n")
	sb.WriteString("END:VCALENDAR\r\n")
	
	return sb.String()
}

// generateVEvent creates a VEVENT block for a calendar event
func generateVEvent(event CalendarEvent) string {
	var sb strings.Builder
	
	sb.WriteString("BEGIN:VEVENT\r\n")
	sb.WriteString(fmt.Sprintf("UID:%s\r\n", event.UID))
	sb.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))
	sb.WriteString(fmt.Sprintf("DTSTART:%s\r\n", event.StartTime.UTC().Format("20060102T150405Z")))
	sb.WriteString(fmt.Sprintf("DTEND:%s\r\n", event.EndTime.UTC().Format("20060102T150405Z")))
	sb.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICSText(event.Title)))
	sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", event.Description))
	sb.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICSText(event.Location)))
	
	if event.Organizer.Email != "" {
		sb.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:MAILTO:%s\r\n", 
			escapeICSText(event.Organizer.Name), event.Organizer.Email))
	}
	
	// Add reminder alarm
	if event.ReminderMin > 0 {
		sb.WriteString("BEGIN:VALARM\r\n")
		sb.WriteString(fmt.Sprintf("TRIGGER:-PT%dM\r\n", event.ReminderMin))
		sb.WriteString("ACTION:DISPLAY\r\n")
		sb.WriteString(fmt.Sprintf("DESCRIPTION:Night Owls shift starts in %d minutes\r\n", event.ReminderMin))
		sb.WriteString("END:VALARM\r\n")
	}
	
	sb.WriteString("END:VEVENT\r\n")
	
	return sb.String()
}

// Helper functions

// formatICSDate formats a time for ICS format (YYYYMMDDTHHMMSSZ)
func formatICSDate(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// escapeICSText escapes special characters for ICS format
func escapeICSText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
} 