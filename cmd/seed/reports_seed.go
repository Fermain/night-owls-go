package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	db "night-owls-go/internal/db/sqlc_generated"
)

// Sample report messages for different severity levels
var reportMessages = map[int][]string{
	0: { // Normal level
		"All quiet on the patrol. No incidents to report.",
		"Routine patrol completed. Checked all designated areas.",
		"Minor maintenance issue noted - streetlight flickering on Oak Street.",
		"Friendly interaction with local residents. Community engagement positive.",
		"Weather conditions good. Visibility excellent throughout shift.",
		"Patrol vehicle fuel level low - recommend refueling before next shift.",
		"New volunteer completed first shift successfully. No issues.",
		"Regular check-in with local business owners. All secure.",
		"Community cat spotted in usual location - residents feeding regularly.",
		"Shift completed without incident. All areas secure.",
	},
	1: { // Suspicion level
		"Suspicious vehicle parked for extended period on Maple Avenue. License plate noted.",
		"Loud party reported by residents. Spoke with organizers - volume reduced.",
		"Broken fence noticed at community park. Potential security concern.",
		"Unattended bag found near bus stop. Monitored for 30 minutes before owner returned.",
		"Group of teenagers gathering late at night. Dispersed peacefully after conversation.",
		"Street light out on Pine Street creating dark spot. Reported to municipality.",
		"Possible drug paraphernalia found in park. Disposed of safely.",
		"Aggressive dog encountered during patrol. Owner contacted and situation resolved.",
		"Minor vehicle accident witnessed. Assisted until emergency services arrived.",
		"Vandalism discovered on community notice board. Photos taken for evidence.",
	},
	2: { // Incident level
		"Break-in attempt observed at local business. Police contacted immediately.",
		"Medical emergency - elderly resident found collapsed. Ambulance called.",
		"Fire spotted in abandoned building. Fire department notified urgently.",
		"Domestic disturbance with shouting and possible violence. Police dispatched.",
		"Armed robbery in progress at corner store. Emergency services contacted.",
		"Serious vehicle accident with injuries. First aid provided until paramedics arrived.",
		"Suspicious individual attempting to enter multiple homes. Police alerted.",
		"Gas leak smell detected near residential area. Utility company contacted immediately.",
		"Assault witnessed in park area. Victim assisted and police called.",
		"Major flooding due to burst water main. Multiple agencies notified.",
	},
}

// Helper functions to reduce boilerplate
func newNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

func newNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func newCreateReportParams(bookingID, userID int64, severity int64, message string) db.CreateReportParams {
	return db.CreateReportParams{
		BookingID: newNullInt64(bookingID),
		UserID:    newNullInt64(userID),
		Severity:  severity,
		Message:   newNullString(message),
	}
}

func seedReports(querier db.Querier) error {
	ctx := context.Background()

	// First, get all existing bookings to create reports for
	bookings, err := querier.GetBookingsInDateRange(ctx, db.GetBookingsInDateRangeParams{
		ShiftStart:   time.Now().AddDate(-1, 0, 0), // 1 year ago
		ShiftStart_2: time.Now().AddDate(0, 0, 30), // 30 days from now
	})
	if err != nil {
		return fmt.Errorf("failed to get bookings: %w", err)
	}

	if len(bookings) == 0 {
		log.Println("No bookings found. Please seed bookings first.")
		return nil
	}

	log.Printf("Found %d bookings to potentially create reports for", len(bookings))

	// Create reports for a subset of bookings (not all shifts have incidents)
	// Roughly 30% of shifts have some kind of report
	reportsCreated := 0

	for _, booking := range bookings {
		// Only create reports for past shifts (shifts that have already happened)
		if booking.ShiftStart.After(time.Now()) {
			continue
		}

		// 30% chance of having a report
		if rand.Float32() > 0.3 {
			continue
		}

		// Determine severity based on realistic distribution
		// 60% normal, 30% suspicion, 10% incident
		var severity int64
		randVal := rand.Float32()
		if randVal < 0.6 {
			severity = 0 // Normal
		} else if randVal < 0.9 {
			severity = 1 // Suspicion
		} else {
			severity = 2 // Incident
		}

		// Select a random message for this severity level
		messages := reportMessages[int(severity)]
		message := messages[rand.Intn(len(messages))]

		// Create the report
		reportParams := newCreateReportParams(booking.BookingID, booking.UserID, severity, message)

		report, err := querier.CreateReport(ctx, reportParams)
		if err != nil {
			log.Printf("Failed to create report for booking %d: %v", booking.BookingID, err)
			continue
		}

		reportsCreated++
		log.Printf("Created report %d (severity %d) for booking %d by %s",
			report.ReportID, severity, booking.BookingID, booking.UserName)

		// Add some time variation to make reports more realistic
		// Reports are typically submitted shortly after or during the shift
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("Successfully created %d reports", reportsCreated)
	return nil
}

// Additional function to create some recent high-priority reports for demo purposes
func seedRecentCriticalReports(querier db.Querier) error {
	ctx := context.Background()

	// Get recent bookings (last 6 months)
	recentBookings, err := querier.GetBookingsInDateRange(ctx, db.GetBookingsInDateRangeParams{
		ShiftStart:   time.Now().AddDate(0, -6, 0), // 6 months ago
		ShiftStart_2: time.Now(),                   // now
	})
	if err != nil {
		return fmt.Errorf("failed to get recent bookings: %w", err)
	}

	// Create a few critical reports for recent shifts to make the demo more interesting
	criticalMessages := []string{
		"Emergency: Multiple break-in attempts reported in residential area. Police response coordinated.",
		"Critical: Medical emergency during patrol. Elderly resident required immediate assistance.",
		"Urgent: Suspicious activity near school premises. Enhanced monitoring requested.",
		"Alert: Vehicle accident with potential injuries. Emergency services coordinated response.",
		"Critical: Fire hazard identified in commercial district. Fire department notified immediately.",
	}

	reportsCreated := 0
	for i, booking := range recentBookings {
		if i >= 3 { // Only create 3 critical reports
			break
		}

		// Only for past shifts
		if booking.ShiftStart.After(time.Now()) {
			continue
		}

		reportParams := newCreateReportParams(booking.BookingID, booking.UserID, 2, criticalMessages[i])

		report, err := querier.CreateReport(ctx, reportParams)
		if err != nil {
			log.Printf("Failed to create critical report for booking %d: %v", booking.BookingID, err)
			continue
		}

		reportsCreated++
		log.Printf("Created critical report %d for booking %d by %s",
			report.ReportID, booking.BookingID, booking.UserName)
	}

	log.Printf("Successfully created %d critical reports", reportsCreated)
	return nil
}
