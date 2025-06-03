import type { AvailableShiftSlot, UserBooking } from '$lib/services/api/user';

// Type for calendar day data
export interface CalendarDay {
	day: number;
	date: Date;
	dateString: string;
	shifts: AvailableShiftSlot[];
	userShifts: UserBooking[];
	isToday: boolean;
	isPast: boolean;
	isOnDuty: boolean;
	monthOffset: number;
}

// Type for calendar cell (can be day or month title)
export interface CalendarCell {
	type: 'day' | 'month-title' | 'empty';
	dayData?: CalendarDay;
	monthName?: string;
	monthOffset?: number;
}

// Type for month grid
export interface MonthGrid {
	monthName: string;
	monthOffset: number;
	weeks: CalendarCell[][];
}
