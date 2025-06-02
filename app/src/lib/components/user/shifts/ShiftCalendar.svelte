<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import { formatTime } from '$lib/utils/shiftFormatting';
	import type { AvailableShiftSlot, UserBooking } from '$lib/services/api/user';

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

	// Don't render anything if there are no shifts
	const hasAnyShifts = $derived(shifts.length > 0);

	// Type for calendar day data
	interface CalendarDay {
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
	interface CalendarCell {
		type: 'day' | 'month-title' | 'empty';
		dayData?: CalendarDay;
		monthName?: string;
		monthOffset?: number;
	}

	// Type for month grid
	interface MonthGrid {
		monthName: string;
		monthOffset: number;
		weeks: CalendarCell[][];
	}

	// Calculate how many months to show based on selected day range
	const monthsToShow = $derived.by(() => {
		const days = parseInt(selectedDayRange);
		if (days <= 7) return 1;
		if (days <= 31) return 2;
		if (days <= 62) return 3;
		return 3; // Max 3 months
	});

	// Get current date for calendar
	const today = new Date();

	// Generate separate month grids with month names in empty cells
	const calendarData = $derived.by(() => {
		const monthGrids: MonthGrid[] = [];
		let firstMonthName = '';
		let cumulativeDayCount = 0; // Track total days to maintain tessellation

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
				startingDayOfWeek = cumulativeDayCount % 7;
			}

			// Get days in this month
			const daysInMonth = new Date(year, month + 1, 0).getDate();

			// Update cumulative count for next month's tessellation
			cumulativeDayCount += startingDayOfWeek + daysInMonth;

			// Create month grid
			const monthCells: CalendarCell[] = [];

			// Add empty cells for days before month starts
			for (let i = 0; i < startingDayOfWeek; i++) {
				if (monthOffset > 0 && i === Math.floor(startingDayOfWeek / 2)) {
					// Use middle empty cell for month name (only for non-first months)
					monthCells.push({
						type: 'month-title',
						monthName: shortMonthName,
						monthOffset
					});
				} else {
					monthCells.push({ type: 'empty' });
				}
			}

			// Add days of the month
			for (let day = 1; day <= daysInMonth; day++) {
				const date = new Date(year, month, day);
				const dateString = date.toISOString().split('T')[0];

				// Find available shifts for this date
				const dayShifts = shifts.filter((shift) => {
					const shiftDate = new Date(shift.start_time).toISOString().split('T')[0];
					return shiftDate === dateString;
				});

				// Find user's bookings for this date
				const dayUserShifts = userBookings.filter((booking) => {
					const bookingDate = new Date(booking.shift_start).toISOString().split('T')[0];
					return bookingDate === dateString;
				});

				// Check if user is currently on duty (active shift)
				const now = new Date();
				const isOnDuty = dayUserShifts.some((booking) => {
					const shiftStart = new Date(booking.shift_start);
					const shiftEnd = new Date(booking.shift_end);
					return now >= shiftStart && now <= shiftEnd;
				});

				const dayData: CalendarDay = {
					day,
					date,
					dateString,
					shifts: dayShifts,
					userShifts: dayUserShifts,
					isToday: date.toDateString() === today.toDateString(),
					isPast: date < today && date.toDateString() !== today.toDateString(),
					isOnDuty,
					monthOffset
				};

				monthCells.push({
					type: 'day',
					dayData
				});
			}

			// Group cells into weeks
			const weeks: CalendarCell[][] = [];
			for (let i = 0; i < monthCells.length; i += 7) {
				const week = monthCells.slice(i, i + 7);
				// Pad the last week if necessary
				while (week.length < 7) {
					week.push({ type: 'empty' });
				}
				weeks.push(week);
			}

			monthGrids.push({
				monthName,
				monthOffset,
				weeks
			});
		}

		return { monthGrids, firstMonthName };
	});

	// Day names
	const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

	// Check if user is booked for a specific shift
	function isUserBookedForShift(shift: AvailableShiftSlot, userShifts: UserBooking[]): boolean {
		return userShifts.some((booking) => {
			const shiftStart = new Date(shift.start_time).getTime();
			const bookingStart = new Date(booking.shift_start).getTime();
			return Math.abs(shiftStart - bookingStart) < 60000; // Within 1 minute tolerance
		});
	}

	// Check if a specific shift is currently active
	function isShiftActive(shift: AvailableShiftSlot, userShifts: UserBooking[]): boolean {
		const now = new Date();
		return userShifts.some((booking) => {
			const shiftStart = new Date(shift.start_time).getTime();
			const bookingStart = new Date(booking.shift_start).getTime();
			const bookingEnd = new Date(booking.shift_end);
			return (
				Math.abs(shiftStart - bookingStart) < 60000 &&
				now >= new Date(booking.shift_start) &&
				now <= bookingEnd
			);
		});
	}

	// Get background class for month distinction
	function getMonthBackground(monthOffset: number): string {
		if (monthOffset === 0) return ''; // Current month - no background
		if (monthOffset === 1) return 'bg-muted/20'; // Next month - subtle background
		return 'bg-muted/40'; // Third month - slightly more background
	}
</script>

<!-- Only render if there are shifts -->
{#if hasAnyShifts}
	<div class="space-y-6">
		<!-- Header -->
		<div class="flex items-center gap-2 px-4">
			<CalendarIcon class="h-4 w-4" />
			<h3 class="text-base font-semibold">Shift Calendar</h3>
			<span class="text-sm text-muted-foreground">- {calendarData.firstMonthName}</span>
		</div>

		<!-- Day names header (only once) -->
		<div class="px-4">
			<div class="grid grid-cols-7 gap-1 text-center">
				{#each dayNames as dayName, index (index)}
					<div class="text-xs font-medium text-muted-foreground p-2">
						{dayName}
					</div>
				{/each}
			</div>
		</div>

		<!-- Seamless Month Grids -->
		<div class="space-y-8">
			{#each calendarData.monthGrids as monthGrid, monthIndex (monthGrid.monthOffset)}
				<div class="px-4">
					<!-- Month grid with embedded month names -->
					<div class="space-y-1">
						{#each monthGrid.weeks as week, weekIndex (weekIndex)}
							<div class="grid grid-cols-7 gap-1">
								{#each week as cell, cellIndex (cellIndex)}
									{#if cell.type === 'empty'}
										<!-- Empty cell for padding -->
										<div class="aspect-square"></div>
									{:else if cell.type === 'month-title'}
										<!-- Month name in empty cell -->
										<div
											class="aspect-square border-2 border-dashed border-muted/30 rounded relative flex items-center justify-center {getMonthBackground(
												cell.monthOffset || 0
											)}"
										>
											<span class="text-xs font-bold text-muted-foreground/70 transform -rotate-12">
												{cell.monthName}
											</span>
										</div>
									{:else if cell.type === 'day' && cell.dayData}
										<!-- Regular day cell -->
										<div
											class="aspect-square border-2 rounded relative p-1
												{cell.dayData.isPast ? 'bg-muted/30 border-muted/50' : 'border-muted/30'}
												{cell.dayData.isToday ? 'ring-2 ring-primary ring-offset-1' : ''}
												{cell.dayData.isOnDuty ? 'bg-green-100 border-green-400' : ''}
												{getMonthBackground(cell.dayData.monthOffset)}
											"
										>
											<!-- Day number -->
											<div
												class="absolute top-1 left-1 text-xs font-medium leading-none
												{cell.dayData.isOnDuty ? 'text-green-800' : 'text-muted-foreground'}
											"
											>
												{cell.dayData.day}
											</div>

											<!-- Shift slots container -->
											<div class="pt-4 h-full flex flex-col gap-0.5 overflow-hidden">
												{#each cell.dayData.shifts as shift, shiftIndex (shift.start_time)}
													{@const isBooked = isUserBookedForShift(shift, cell.dayData.userShifts)}
													{@const isActive = isShiftActive(shift, cell.dayData.userShifts)}

													<button
														class="text-xs px-1 py-0.5 rounded transition-colors flex items-center justify-between
															{isBooked
															? isActive
																? 'bg-green-600 text-white font-bold'
																: 'bg-blue-100 text-blue-800 border border-blue-300'
															: 'bg-primary/10 hover:bg-primary/20 text-primary border border-primary/30'}
														"
														onclick={() => !isBooked && onShiftSelect(shift)}
														disabled={isBooked || cell.dayData.isPast}
														title={isBooked
															? 'Already booked'
															: `${formatTime(shift.start_time)} - ${formatTime(shift.end_time)}`}
													>
														<span class="truncate flex-1 text-left leading-none">
															{formatTime(shift.start_time)}
														</span>
														{#if isBooked}
															<span class="ml-1 leading-none">🦉</span>
														{/if}
													</button>
												{/each}
											</div>
										</div>
									{/if}
								{/each}
							</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- Legend -->
		<div class="px-4">
			<div class="flex flex-wrap gap-4 text-xs text-muted-foreground">
				<div class="flex items-center gap-1">
					<div class="w-3 h-3 bg-primary/10 border border-primary/30 rounded"></div>
					<span>Available slot</span>
				</div>
				<div class="flex items-center gap-1">
					<span class="text-base">🦉</span>
					<span>My shift</span>
				</div>
				<div class="flex items-center gap-1">
					<div class="w-3 h-3 bg-green-600 rounded"></div>
					<span>Active now</span>
				</div>
				{#if monthsToShow > 1}
					<div class="flex items-center gap-1">
						<div class="w-3 h-3 bg-muted/20 border border-muted/30 rounded"></div>
						<span>Next month</span>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
