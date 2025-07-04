/**
 * Admin Calendar Utilities
 *
 * Extracted business logic for admin calendar generation, ensuring proper
 * date ranges, month calculations, and shift organization for admin workflows.
 */

import type { AdminShiftSlot } from '$lib/types';
import type {
	AdminCalendarDay,
	AdminCalendarCell,
	AdminMonthGrid
} from '$lib/components/admin/calendar/admin-calendar-types';

// === CONSTANTS ===
export const MIN_MONTHS_TO_SHOW = 2; // Always show at least 2 months for admin view
export const MAX_MONTHS_TO_SHOW = 3; // Cap at 3 months for performance
export const MAX_DAY_RANGE = 365; // Maximum 1 year for performance
export const DEFAULT_DAY_RANGE = '60'; // 2 months default

// === DATE UTILITIES ===

/**
 * Format date as YYYY-MM-DD in local timezone for calendar keys
 * Production-safe with error handling
 */
export function formatLocalDate(date: Date): string {
	try {
		if (!date || isNaN(date.getTime())) {
			console.warn('Invalid date provided to formatLocalDate:', date);
			return '';
		}

		return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
	} catch (error) {
		console.error('Error formatting local date:', error);
		return '';
	}
}

/**
 * Calculate optimal number of months to display
 * Ensures at least 2 months for proper admin overview
 */
export function calculateMonthsToShow(dayRange: string): number {
	const days = parseInt(dayRange);

	// Validate input
	if (isNaN(days) || days <= 0) {
		console.warn('Invalid day range provided:', dayRange);
		return MIN_MONTHS_TO_SHOW;
	}

	// Performance cap
	if (days > MAX_DAY_RANGE) {
		console.warn('Day range too large, capping at maximum:', MAX_DAY_RANGE);
		return MAX_MONTHS_TO_SHOW;
	}

	// For short ranges, still show 2 months minimum for admin context
	if (days <= 7) return MIN_MONTHS_TO_SHOW;
	if (days <= 31) return MIN_MONTHS_TO_SHOW;
	if (days <= 62) return MAX_MONTHS_TO_SHOW;
	return MAX_MONTHS_TO_SHOW;
}

/**
 * Calculate date range based on day range parameter
 * Always ensures we cover at least 2 months worth of data
 * Production-safe with boundary validation
 */
export function calculateDateRange(dayRange: string): { startDate: Date; endDate: Date } {
	try {
		const days = parseInt(dayRange);

		// Input validation
		if (isNaN(days) || days <= 0) {
			console.warn('Invalid day range, using default:', DEFAULT_DAY_RANGE);
			return calculateDateRange(DEFAULT_DAY_RANGE);
		}

		// Performance protection
		const safeDays = Math.min(days, MAX_DAY_RANGE);

		const startDate = new Date();
		startDate.setHours(0, 0, 0, 0); // Start of day for consistency

		// Calculate end date based on day range
		let endDate = new Date(startDate.getTime() + safeDays * 24 * 60 * 60 * 1000);

		// Ensure we always cover at least 2 full months for admin context
		const twoMonthsFromNow = new Date(startDate);
		twoMonthsFromNow.setMonth(twoMonthsFromNow.getMonth() + 2);
		twoMonthsFromNow.setDate(0); // Last day of second month
		twoMonthsFromNow.setHours(23, 59, 59, 999); // End of day

		if (endDate < twoMonthsFromNow) {
			endDate = twoMonthsFromNow;
		}

		return { startDate, endDate };
	} catch (error) {
		console.error('Error calculating date range:', error);
		// Fallback to safe defaults
		const startDate = new Date();
		startDate.setHours(0, 0, 0, 0);
		const endDate = new Date(startDate);
		endDate.setMonth(endDate.getMonth() + 2);
		endDate.setHours(23, 59, 59, 999);

		return { startDate, endDate };
	}
}

/**
 * Group shifts by date for efficient calendar rendering
 * Performance optimized with error handling
 */
export function groupShiftsByDate(shifts: AdminShiftSlot[]): Map<string, AdminShiftSlot[]> {
	const shiftsByDate = new Map<string, AdminShiftSlot[]>();

	if (!Array.isArray(shifts)) {
		console.warn('Invalid shifts array provided to groupShiftsByDate');
		return shiftsByDate;
	}

	shifts.forEach((shift) => {
		try {
			if (!shift?.start_time) {
				console.warn('Shift missing start_time:', shift);
				return;
			}

			const shiftDate = new Date(shift.start_time);
			if (isNaN(shiftDate.getTime())) {
				console.warn('Invalid shift date:', shift.start_time);
				return;
			}

			const dateKey = formatLocalDate(shiftDate);
			if (!dateKey) return; // Skip invalid dates

			if (!shiftsByDate.has(dateKey)) {
				shiftsByDate.set(dateKey, []);
			}
			shiftsByDate.get(dateKey)!.push(shift);
		} catch (error) {
			console.warn('Error processing shift:', shift, error);
		}
	});

	return shiftsByDate;
}

/**
 * Build calendar cell for a specific day
 * Production-safe with comprehensive error handling
 */
export function buildDayCell(
	day: number,
	year: number,
	month: number,
	monthOffset: number,
	dateRange: { startDate: Date; endDate: Date },
	shiftsByDate: Map<string, AdminShiftSlot[]>
): AdminCalendarCell {
	try {
		// Validate inputs
		if (day < 1 || day > 31 || year < 1900 || year > 2100 || month < 0 || month > 11) {
			console.warn('Invalid date parameters for buildDayCell:', { day, year, month });
			return { type: 'empty' };
		}

		const date = new Date(year, month, day);
		if (isNaN(date.getTime())) {
			console.warn('Invalid date created in buildDayCell:', { day, year, month });
			return { type: 'empty' };
		}

		const dateString = formatLocalDate(date);
		if (!dateString) {
			return { type: 'empty' };
		}

		const today = new Date();
		today.setHours(0, 0, 0, 0); // Normalize for comparison

		// Check if this day is within our target date range
		const isWithinRange = date >= dateRange.startDate && date <= dateRange.endDate;

		// Get shifts for this day (only if within range for performance)
		const dayShifts = isWithinRange ? shiftsByDate.get(dateString) || [] : [];

		const dayData: AdminCalendarDay = {
			day,
			date,
			dateString,
			shifts: dayShifts,
			isToday: date.toDateString() === today.toDateString(),
			isPast: date < today,
			monthOffset,
			isWithinRange
		};

		return {
			type: 'day',
			dayData
		};
	} catch (error) {
		console.error('Error building day cell:', error, { day, year, month });
		return { type: 'empty' };
	}
}

/**
 * Build calendar cells for a complete month
 * Production-optimized with error boundaries
 */
export function buildMonthCells(
	year: number,
	month: number,
	monthOffset: number,
	monthName: string,
	dateRange: { startDate: Date; endDate: Date },
	shiftsByDate: Map<string, AdminShiftSlot[]>
): AdminCalendarCell[] {
	const monthCells: AdminCalendarCell[] = [];

	try {
		// Add month title in first cell
		monthCells.push({
			type: 'month-title',
			monthName: monthName || 'Unknown Month'
		});

		// Add empty cells to align with calendar grid (Sunday = 0)
		const firstDayOfWeek = new Date(year, month, 1).getDay();
		for (let i = 1; i < firstDayOfWeek; i++) {
			monthCells.push({ type: 'empty' });
		}

		// Add day cells for the month
		const daysInMonth = new Date(year, month + 1, 0).getDate();
		for (let day = 1; day <= daysInMonth; day++) {
			monthCells.push(buildDayCell(day, year, month, monthOffset, dateRange, shiftsByDate));
		}

		return monthCells;
	} catch (error) {
		console.error('Error building month cells:', error, { year, month });
		// Return minimal safe structure
		return [
			{ type: 'month-title', monthName: 'Error' },
			...Array(42).fill({ type: 'empty' }) // Standard calendar grid
		];
	}
}

/**
 * Group cells into weeks for grid layout
 * Performance optimized
 */
export function groupCellsIntoWeeks(monthCells: AdminCalendarCell[]): AdminCalendarCell[][] {
	const weeks: AdminCalendarCell[][] = [];

	try {
		for (let i = 0; i < monthCells.length; i += 7) {
			const week = monthCells.slice(i, i + 7);

			// Pad the last week if necessary
			while (week.length < 7) {
				week.push({ type: 'empty' });
			}

			weeks.push(week);
		}

		return weeks;
	} catch (error) {
		console.error('Error grouping cells into weeks:', error);
		return [];
	}
}

/**
 * Generate complete calendar data for admin view
 *
 * This is the main function that orchestrates calendar generation
 * Production-ready with comprehensive error handling and performance optimization
 */
export function generateAdminCalendarData(
	shifts: AdminShiftSlot[],
	dayRange: string
): { monthGrids: AdminMonthGrid[]; firstMonthName: string } {
	try {
		// Input validation
		if (!Array.isArray(shifts)) {
			console.warn('Invalid shifts provided to generateAdminCalendarData');
			return { monthGrids: [], firstMonthName: 'No Data' };
		}

		// Calculate optimal parameters with error handling
		const monthsToShow = calculateMonthsToShow(dayRange);
		const dateRange = calculateDateRange(dayRange);
		const shiftsByDate = groupShiftsByDate(shifts);

		// Performance logging for large datasets
		if (shifts.length > 1000) {
			console.info(`Processing large shift dataset: ${shifts.length} shifts`);
		}

		// Generate month grids
		const monthGrids: AdminMonthGrid[] = [];
		const firstMonthName = dateRange.startDate.toLocaleDateString('en-US', {
			month: 'long',
			year: 'numeric'
		});

		for (let monthOffset = 0; monthOffset < monthsToShow; monthOffset++) {
			try {
				const monthDate = new Date(
					dateRange.startDate.getFullYear(),
					dateRange.startDate.getMonth() + monthOffset,
					1
				);

				const year = monthDate.getFullYear();
				const month = monthDate.getMonth();
				const monthName = monthDate.toLocaleDateString('en-US', {
					month: 'long',
					year: 'numeric'
				});

				// Build cells for this month
				const monthCells = buildMonthCells(
					year,
					month,
					monthOffset,
					monthOffset === 0 ? firstMonthName : monthName,
					dateRange,
					shiftsByDate
				);

				// Group into weeks
				const weeks = groupCellsIntoWeeks(monthCells);

				// Add to grids only if valid
				if (weeks.length > 0) {
					monthGrids.push({
						monthName,
						monthOffset,
						weeks
					});
				}
			} catch (error) {
				console.error('Error processing month:', monthOffset, error);
				// Continue with other months
			}
		}

		return { monthGrids, firstMonthName };
	} catch (error) {
		console.error('Fatal error in generateAdminCalendarData:', error);
		return { monthGrids: [], firstMonthName: 'Error' };
	}
}

// === SHIFT ANALYSIS UTILITIES ===

/**
 * Analyze shift distribution across the calendar period
 * Production-ready with comprehensive error handling
 */
export interface ShiftAnalysis {
	totalShifts: number;
	filledShifts: number;
	unfilledShifts: number;
	fillRate: number;
	daysWithShifts: number;
	averageShiftsPerDay: number;
}

export function analyzeShifts(shifts: AdminShiftSlot[], dayRange: string): ShiftAnalysis {
	try {
		// Input validation
		if (!Array.isArray(shifts)) {
			return {
				totalShifts: 0,
				filledShifts: 0,
				unfilledShifts: 0,
				fillRate: 0,
				daysWithShifts: 0,
				averageShiftsPerDay: 0
			};
		}

		const dateRange = calculateDateRange(dayRange);
		const shiftsByDate = groupShiftsByDate(shifts);

		// Filter shifts within the target range
		const shiftsInRange = shifts.filter((shift) => {
			try {
				if (!shift?.start_time) return false;
				const shiftDate = new Date(shift.start_time);
				return (
					!isNaN(shiftDate.getTime()) &&
					shiftDate >= dateRange.startDate &&
					shiftDate <= dateRange.endDate
				);
			} catch {
				return false;
			}
		});

		const totalShifts = shiftsInRange.length;
		const filledShifts = shiftsInRange.filter(
			(shift) => shift?.is_booked && shift?.user_name
		).length;
		const unfilledShifts = totalShifts - filledShifts;
		const fillRate = totalShifts > 0 ? (filledShifts / totalShifts) * 100 : 0;
		const daysWithShifts = shiftsByDate.size;

		// Calculate range in days for average
		const rangeDays = Math.ceil(
			(dateRange.endDate.getTime() - dateRange.startDate.getTime()) / (1000 * 60 * 60 * 24)
		);
		const averageShiftsPerDay = rangeDays > 0 ? totalShifts / rangeDays : 0;

		return {
			totalShifts,
			filledShifts,
			unfilledShifts,
			fillRate: Math.round(fillRate * 100) / 100, // Round to 2 decimal places
			daysWithShifts,
			averageShiftsPerDay: Math.round(averageShiftsPerDay * 100) / 100
		};
	} catch (error) {
		console.error('Error analyzing shifts:', error);
		return {
			totalShifts: 0,
			filledShifts: 0,
			unfilledShifts: 0,
			fillRate: 0,
			daysWithShifts: 0,
			averageShiftsPerDay: 0
		};
	}
}

// === VALIDATION UTILITIES ===

/**
 * Validate day range input
 */
export function validateDayRange(dayRange: string): boolean {
	try {
		const days = parseInt(dayRange);
		return !isNaN(days) && days > 0 && days <= MAX_DAY_RANGE;
	} catch {
		return false;
	}
}

/**
 * Sanitize day range input with fallback
 */
export function sanitizeDayRange(dayRange: string, fallback = DEFAULT_DAY_RANGE): string {
	return validateDayRange(dayRange) ? dayRange : fallback;
}
