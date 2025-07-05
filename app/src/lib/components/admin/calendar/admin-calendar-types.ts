import type { AdminShiftSlot } from '$lib/types';

// Type for admin calendar day data
export interface AdminCalendarDay {
	day: number;
	date: Date;
	dateString: string;
	shifts: AdminShiftSlot[];
	isToday: boolean;
	isPast: boolean;
	monthOffset: number;
	isWithinRange: boolean; // Whether this day is within the selected date range
}

// Type for admin calendar cell (can be day or month title)
export interface AdminCalendarCell {
	type: 'day' | 'month-title' | 'empty';
	dayData?: AdminCalendarDay;
	monthName?: string;
	monthOffset?: number;
}

// Type for admin month grid
export interface AdminMonthGrid {
	monthName: string;
	monthOffset: number;
	weeks: AdminCalendarCell[][];
}
