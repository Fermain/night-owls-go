/**
 * Calendar Integration Utilities for Night Owls Shift Bookings
 * Supports .ics generation, calendar URLs, and popular calendar app integration
 *
 * TODO: Extract shared calendar generation logic to avoid duplication with backend
 * This frontend implementation should eventually use a shared library for consistency
 */

import type { UserBooking } from '$lib/services/api/user';
import { formatDateTime } from '$lib/utils/datetime';

// === TYPES ===

export interface CalendarEvent {
	title: string;
	description: string;
	startTime: Date;
	endTime: Date;
	location?: string;
	organizer?: {
		name: string;
		email: string;
	};
	attendees?: Array<{
		name: string;
		email?: string;
	}>;
	reminder?: {
		minutes: number;
	};
	uid: string;
}

export interface CalendarProvider {
	name: string;
	url: string;
	icon: string;
}

// === CONSTANTS ===

const CALENDAR_PROVIDERS = {
	google: {
		name: 'Google Calendar',
		icon: '📅',
		baseUrl: 'https://calendar.google.com/calendar/render'
	},
	outlook: {
		name: 'Outlook',
		icon: '📘',
		baseUrl: 'https://outlook.live.com/calendar/0/deeplink/compose'
	},
	apple: {
		name: 'Apple Calendar',
		icon: '🍎',
		baseUrl: 'data:text/calendar;charset=utf8'
	},
	yahoo: {
		name: 'Yahoo Calendar',
		icon: '🟣',
		baseUrl: 'https://calendar.yahoo.com'
	}
} as const;

// === CORE CALENDAR FUNCTIONS ===

/**
 * Convert UserBooking to CalendarEvent
 */
export function bookingToCalendarEvent(booking: UserBooking): CalendarEvent {
	const startTime = new Date(booking.shift_start);
	const endTime = new Date(booking.shift_end);

	// Create detailed description
	let description = `Night Owls Community Watch - ${booking.schedule_name}\n\n`;
	description += `⏰ Time: ${formatDateTime(booking.shift_start)} - ${formatDateTime(booking.shift_end)}\n`;

	if (booking.buddy_name) {
		description += `👥 Buddy: ${booking.buddy_name}\n`;
	}

	description += `\n📱 Check in on the Night Owls app when your shift starts`;
	description += `\n🚨 Report any incidents through the app`;
	description += `\n\n🔗 Night Owls App: ${typeof window !== 'undefined' ? window.location.origin : 'https://your-domain.com'}`;

	// Create attendees list
	const attendees: CalendarEvent['attendees'] = [];
	if (booking.buddy_name) {
		attendees.push({ name: booking.buddy_name });
	}

	return {
		title: `Night Owls Shift - ${booking.schedule_name}`,
		description,
		startTime,
		endTime,
		location: 'Mount Moreland Community Watch Area',
		organizer: {
			name: 'Night Owls Scheduler',
			email: 'noreply@mm.nightowls.app'
		},
		attendees,
		reminder: {
			minutes: 60 // 1 hour before shift
		},
		uid: `nightowls-shift-${booking.booking_id}@mm.nightowls.app`
	};
}

/**
 * Generate .ics file content for a single event
 */
export function generateICSEvent(event: CalendarEvent): string {
	const formatICSDate = (date: Date): string => {
		return date
			.toISOString()
			.replace(/[:-]/g, '')
			.replace(/\.\d{3}/, '');
	};

	const escapeICSText = (text: string): string => {
		return text
			.replace(/\\/g, '\\\\')
			.replace(/;/g, '\\;')
			.replace(/,/g, '\\,')
			.replace(/\n/g, '\\n');
	};

	const now = new Date();
	const icsContent = [
		'BEGIN:VCALENDAR',
		'VERSION:2.0',
		'PRODID:-//Night Owls//Shift Calendar//EN',
		'CALSCALE:GREGORIAN',
		'METHOD:PUBLISH',
		'BEGIN:VEVENT',
		`UID:${event.uid}`,
		`DTSTAMP:${formatICSDate(now)}`,
		`DTSTART:${formatICSDate(event.startTime)}`,
		`DTEND:${formatICSDate(event.endTime)}`,
		`SUMMARY:${escapeICSText(event.title)}`,
		`DESCRIPTION:${escapeICSText(event.description)}`,
		...(event.location ? [`LOCATION:${escapeICSText(event.location)}`] : []),
		...(event.organizer
			? [`ORGANIZER;CN=${escapeICSText(event.organizer.name)}:mailto:${event.organizer.email}`]
			: []),
		...(event.attendees?.map(
			(attendee) =>
				`ATTENDEE;CN=${escapeICSText(attendee.name)}${attendee.email ? `:mailto:${attendee.email}` : ''}`
		) || []),
		...(event.reminder
			? [
					'BEGIN:VALARM',
					'TRIGGER:-PT' + event.reminder.minutes + 'M',
					'ACTION:DISPLAY',
					`DESCRIPTION:${escapeICSText(event.title)} starts in ${event.reminder.minutes} minutes`,
					'END:VALARM'
				]
			: []),
		'END:VEVENT',
		'END:VCALENDAR'
	];

	return icsContent.join('\r\n');
}

/**
 * Generate .ics file content for multiple events
 */
export function generateICSCalendar(
	events: CalendarEvent[],
	calendarName = 'Night Owls Shifts'
): string {
	const formatICSDate = (date: Date): string => {
		return date
			.toISOString()
			.replace(/[:-]/g, '')
			.replace(/\.\d{3}/, '');
	};

	const escapeICSText = (text: string): string => {
		return text
			.replace(/\\/g, '\\\\')
			.replace(/;/g, '\\;')
			.replace(/,/g, '\\,')
			.replace(/\n/g, '\\n');
	};

	const now = new Date();

	const icsContent = [
		'BEGIN:VCALENDAR',
		'VERSION:2.0',
		'PRODID:-//Night Owls//Shift Calendar//EN',
		'CALSCALE:GREGORIAN',
		'METHOD:PUBLISH',
		`X-WR-CALNAME:${escapeICSText(calendarName)}`,
		'X-WR-CALDESC:Mount Moreland Night Owls Community Watch Shifts',
		'X-WR-TIMEZONE:Africa/Johannesburg'
	];

	// Add each event
	for (const event of events) {
		icsContent.push(
			'BEGIN:VEVENT',
			`UID:${event.uid}`,
			`DTSTAMP:${formatICSDate(now)}`,
			`DTSTART:${formatICSDate(event.startTime)}`,
			`DTEND:${formatICSDate(event.endTime)}`,
			`SUMMARY:${escapeICSText(event.title)}`,
			`DESCRIPTION:${escapeICSText(event.description)}`,
			...(event.location ? [`LOCATION:${escapeICSText(event.location)}`] : []),
			...(event.organizer
				? [`ORGANIZER;CN=${escapeICSText(event.organizer.name)}:mailto:${event.organizer.email}`]
				: []),
			...(event.attendees?.map(
				(attendee) =>
					`ATTENDEE;CN=${escapeICSText(attendee.name)}${attendee.email ? `:mailto:${attendee.email}` : ''}`
			) || []),
			...(event.reminder
				? [
						'BEGIN:VALARM',
						'TRIGGER:-PT' + event.reminder.minutes + 'M',
						'ACTION:DISPLAY',
						`DESCRIPTION:${escapeICSText(event.title)} starts in ${event.reminder.minutes} minutes`,
						'END:VALARM'
					]
				: []),
			'END:VEVENT'
		);
	}

	icsContent.push('END:VCALENDAR');
	return icsContent.join('\r\n');
}

// === CALENDAR PROVIDER URLS ===

/**
 * Generate Google Calendar URL
 */
export function generateGoogleCalendarURL(event: CalendarEvent): string {
	const formatGoogleDate = (date: Date): string => {
		return date
			.toISOString()
			.replace(/[:-]/g, '')
			.replace(/\.\d{3}/, '');
	};

	const params = new URLSearchParams({
		action: 'TEMPLATE',
		text: event.title,
		dates: `${formatGoogleDate(event.startTime)}/${formatGoogleDate(event.endTime)}`,
		details: event.description,
		...(event.location && { location: event.location })
	});

	return `${CALENDAR_PROVIDERS.google.baseUrl}?${params.toString()}`;
}

/**
 * Generate Outlook Calendar URL
 */
export function generateOutlookURL(event: CalendarEvent): string {
	const params = new URLSearchParams({
		subject: event.title,
		startdt: event.startTime.toISOString(),
		enddt: event.endTime.toISOString(),
		body: event.description,
		...(event.location && { location: event.location })
	});

	return `${CALENDAR_PROVIDERS.outlook.baseUrl}?${params.toString()}`;
}

/**
 * Generate Yahoo Calendar URL
 */
export function generateYahooURL(event: CalendarEvent): string {
	const formatYahooDate = (date: Date): string => {
		return date
			.toISOString()
			.replace(/[:-]/g, '')
			.replace(/\.\d{3}/, '');
	};

	const params = new URLSearchParams({
		v: '60',
		title: event.title,
		st: formatYahooDate(event.startTime),
		et: formatYahooDate(event.endTime),
		desc: event.description,
		...(event.location && { in_loc: event.location })
	});

	return `${CALENDAR_PROVIDERS.yahoo.baseUrl}?${params.toString()}`;
}

// === FILE DOWNLOAD UTILITIES ===

/**
 * Download .ics file for a single booking
 */
export function downloadShiftICS(booking: UserBooking): void {
	const event = bookingToCalendarEvent(booking);
	const icsContent = generateICSEvent(event);
	const fileName = `night-owls-shift-${booking.booking_id}.ics`;

	downloadICSFile(icsContent, fileName);
}

/**
 * Download .ics file for multiple bookings
 */
export function downloadAllShiftsICS(bookings: UserBooking[]): void {
	const events = bookings.map(bookingToCalendarEvent);
	const icsContent = generateICSCalendar(events, 'Night Owls Shifts');
	const fileName = `night-owls-all-shifts-${new Date().toISOString().split('T')[0]}.ics`;

	downloadICSFile(icsContent, fileName);
}

/**
 * Utility to trigger file download
 */
function downloadICSFile(content: string, fileName: string): void {
	const blob = new Blob([content], { type: 'text/calendar;charset=utf-8' });
	const url = URL.createObjectURL(blob);

	const link = document.createElement('a');
	link.href = url;
	link.download = fileName;
	link.style.display = 'none';

	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);

	URL.revokeObjectURL(url);
}

// === CALENDAR PROVIDER INTEGRATION ===

/**
 * Get all supported calendar providers with their URLs for an event
 */
export function getCalendarProviders(event: CalendarEvent): CalendarProvider[] {
	return [
		{
			name: CALENDAR_PROVIDERS.google.name,
			url: generateGoogleCalendarURL(event),
			icon: CALENDAR_PROVIDERS.google.icon
		},
		{
			name: CALENDAR_PROVIDERS.outlook.name,
			url: generateOutlookURL(event),
			icon: CALENDAR_PROVIDERS.outlook.icon
		},
		{
			name: CALENDAR_PROVIDERS.yahoo.name,
			url: generateYahooURL(event),
			icon: CALENDAR_PROVIDERS.yahoo.icon
		}
	];
}

/**
 * Open calendar provider in new window/tab
 */
export function openCalendarProvider(url: string): void {
	window.open(url, '_blank', 'noopener,noreferrer');
}

// === UTILITY FUNCTIONS ===

/**
 * Check if a booking is in the future (can be added to calendar)
 */
export function canAddToCalendar(booking: UserBooking): boolean {
	const shiftStart = new Date(booking.shift_start);
	const now = new Date();
	return shiftStart > now;
}

/**
 * Get calendar integration summary for bookings
 */
export function getCalendarSummary(bookings: UserBooking[]): {
	total: number;
	upcoming: number;
	thisWeek: number;
	thisMonth: number;
} {
	const now = new Date();
	const oneWeek = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
	const oneMonth = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);

	const upcoming = bookings.filter((booking) => {
		const shiftStart = new Date(booking.shift_start);
		return shiftStart > now;
	});

	return {
		total: bookings.length,
		upcoming: upcoming.length,
		thisWeek: upcoming.filter((booking) => {
			const shiftStart = new Date(booking.shift_start);
			return shiftStart <= oneWeek;
		}).length,
		thisMonth: upcoming.filter((booking) => {
			const shiftStart = new Date(booking.shift_start);
			return shiftStart <= oneMonth;
		}).length
	};
}
