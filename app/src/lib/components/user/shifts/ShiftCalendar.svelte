<script lang="ts">
	import CalendarHeader from './CalendarHeader.svelte';
	import CalendarDayNames from './CalendarDayNames.svelte';
	import CalendarMonthGrid from './CalendarMonthGrid.svelte';
	import CalendarLegend from './CalendarLegend.svelte';
	import type { AvailableShiftSlot, UserBooking } from '$lib/services/api/user';
	import type { CalendarDay, CalendarCell, MonthGrid } from './calendar-types.js';

	let {
		shifts = [],
		userBookings = [],
		selectedDayRange = '7',
		onShiftSelect
	}: {
		shifts: AvailableShiftSlot[];
		userBookings: UserBooking[];
		selectedDayRange: string;
		onShiftSelect: (shift: AvailableShiftSlot) => void;
	} = $props();

	// Helper function to format date as YYYY-MM-DD in local timezone
	function formatLocalDate(date: Date): string {
		return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
	}

	// Don't render anything if there are no shifts
	const hasAnyShifts = $derived(shifts.length > 0);

	// Show calendar if there are shifts OR user bookings
	const shouldShowCalendar = $derived(shifts.length > 0 || userBookings.length > 0);

	// Calculate how many months to show based on selected day range
	const monthsToShow = $derived.by(() => {
		const days = parseInt(selectedDayRange);
		if (days <= 7) return 1;
		if (days <= 31) return 2;
		if (days <= 62) return 3;
		return 3; // Max 3 months
	});

	// Calculate the date range we should show based on selectedDayRange
	const dateRange = $derived.by(() => {
		const days = parseInt(selectedDayRange);
		const startDate = new Date();
		const endDate = new Date(Date.now() + days * 24 * 60 * 60 * 1000);
		return { startDate, endDate };
	});

	// Get current date for calendar
	const today = new Date();

	// Generate separate month grids with month names in empty cells
	const calendarData = $derived.by(() => {
		const monthGrids: MonthGrid[] = [];
		let firstMonthName = '';
		const { startDate, endDate } = dateRange;

		for (let monthOffset = 0; monthOffset < monthsToShow; monthOffset++) {
			const currentMonth = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
			const year = currentMonth.getFullYear();
			const month = currentMonth.getMonth();

			const monthName = currentMonth.toLocaleDateString('en-GB', {
				month: 'long',
				year: monthsToShow > 1 ? 'numeric' : undefined
			});

			// Short month name for empty cells
			const shortMonthName = currentMonth
				.toLocaleDateString('en-GB', {
					month: 'short'
				})
				.toUpperCase();

			if (monthOffset === 0) {
				firstMonthName = monthName;
			}

			// Calculate starting day of week for this month
			let startingDayOfWeek;
			if (monthOffset === 0) {
				// First month starts on its natural day
				startingDayOfWeek = new Date(year, month, 1).getDay();
			} else {
				// Subsequent months start where previous month would have ended
				// Need to account for the fact that each month ends on a different day
				const prevMonth = new Date(year, month, 0); // Last day of previous month
				startingDayOfWeek = (prevMonth.getDay() + 1) % 7;
			}

			// Get days in this month
			const daysInMonth = new Date(year, month + 1, 0).getDate();

			// Create month grid
			const monthCells: CalendarCell[] = [];

			// Add empty cells for days before month starts
			for (let i = 0; i < startingDayOfWeek; i++) {
				if (monthOffset > 0 && i === 0 && startingDayOfWeek > 0) {
					// Use first empty cell for full month name with overflow (only for non-first months)
					monthCells.push({
						type: 'month-title',
						monthName: shortMonthName,
						monthOffset
					});
				} else {
					monthCells.push({ type: 'empty' });
				}
			}

			// Add days of the month - include ALL days to maintain calendar structure
			for (let day = 1; day <= daysInMonth; day++) {
				const date = new Date(year, month, day);
				const dateString = formatLocalDate(date);

				// Always create calendar cells to maintain grid structure
				// but only include shift data for days within our selected date range
				const isWithinRange = date >= startDate && date <= endDate;

				let dayShifts: AvailableShiftSlot[] = [];
				let dayUserShifts: UserBooking[] = [];
				let isOnDuty = false;

				if (isWithinRange) {
					// Find available shifts for this date using local date comparison
					dayShifts = shifts.filter((shift) => {
						const shiftDate = new Date(shift.start_time);
						return formatLocalDate(shiftDate) === dateString;
					});

					// Find user's bookings for this date using local date comparison
					dayUserShifts = userBookings.filter((booking) => {
						const bookingDate = new Date(booking.shift_start);
						return formatLocalDate(bookingDate) === dateString;
					});

					// Check if user is currently on duty (active shift)
					const now = new Date();
					isOnDuty = dayUserShifts.some((booking) => {
						const shiftStart = new Date(booking.shift_start);
						const shiftEnd = new Date(booking.shift_end);
						return now >= shiftStart && now <= shiftEnd;
					});
				}

				const dayData: CalendarDay = {
					day,
					date,
					dateString,
					shifts: dayShifts,
					userShifts: dayUserShifts,
					isToday: date.toDateString() === today.toDateString(),
					isPast: date < today && date.toDateString() !== today.toDateString(),
					isOnDuty,
					monthOffset,
					isWithinRange // Add flag to indicate if day is within selected range
				};

				monthCells.push({
					type: 'day',
					dayData
				});
			}

			// Group cells into weeks - show complete months to maintain grid alignment
			const weeks: CalendarCell[][] = [];
			for (let i = 0; i < monthCells.length; i += 7) {
				const week = monthCells.slice(i, i + 7);
				// Pad the last week if necessary
				while (week.length < 7) {
					week.push({ type: 'empty' });
				}
				weeks.push(week);
			}

			// Always add the month grid to maintain proper calendar structure
			monthGrids.push({
				monthName,
				monthOffset,
				weeks
			});
		}

		return { monthGrids, firstMonthName };
	});
</script>

<!-- Only render if there are shifts -->
{#if shouldShowCalendar}
	<div class="space-y-6">
		<!-- Header -->
		<CalendarHeader firstMonthName={calendarData.firstMonthName} />

		<!-- Day names header -->
		<CalendarDayNames />

		<!-- Seamless Month Grids -->
		<div class="space-y-8">
			{#each calendarData.monthGrids as monthGrid (monthGrid.monthOffset)}
				<CalendarMonthGrid {monthGrid} {onShiftSelect} />
			{/each}
		</div>

		<!-- Legend -->
		<CalendarLegend />
	</div>
{/if}
