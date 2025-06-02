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
		isFirstDayOfMonth: boolean;
	}

	// Type for calendar week
	interface CalendarWeek {
		days: (CalendarDay | null)[];
		hasMonthBoundary: boolean;
		monthName?: string;
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

	// Generate calendar weeks with proper tessellation
	const calendarData = $derived.by(() => {
		const allDays: (CalendarDay | null)[] = [];
		let firstMonthName = '';

		// Start with the first day of the current month to get proper week alignment
		const startDate = new Date(today.getFullYear(), today.getMonth(), 1);
		const startDayOfWeek = startDate.getDay(); // 0 = Sunday

		// Add empty cells for days before the first month starts
		for (let i = 0; i < startDayOfWeek; i++) {
			allDays.push(null);
		}

		// Add all days from all months
		for (let monthOffset = 0; monthOffset < monthsToShow; monthOffset++) {
			const currentMonth = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
			const year = currentMonth.getFullYear();
			const month = currentMonth.getMonth();

			// Store first month name for title
			if (monthOffset === 0) {
				firstMonthName = currentMonth.toLocaleDateString('en-GB', {
					month: 'long',
					year: monthsToShow > 1 ? 'numeric' : undefined
				});
			}

			// Get days in this month
			const lastDayOfMonth = new Date(year, month + 1, 0);
			const daysInMonth = lastDayOfMonth.getDate();

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

				allDays.push({
					day,
					date,
					dateString,
					shifts: dayShifts,
					userShifts: dayUserShifts,
					isToday: date.toDateString() === today.toDateString(),
					isPast: date < today && date.toDateString() !== today.toDateString(),
					isOnDuty,
					monthOffset,
					isFirstDayOfMonth: day === 1
				});
			}
		}

		// Group days into weeks and identify month boundaries
		const weeks: CalendarWeek[] = [];
		for (let i = 0; i < allDays.length; i += 7) {
			const weekDays = allDays.slice(i, i + 7);

			// Pad the last week if necessary
			while (weekDays.length < 7) {
				weekDays.push(null);
			}

			// Check if this week contains the start of a new month
			const hasMonthBoundary = weekDays.some(
				(day) => day?.isFirstDayOfMonth && day.monthOffset > 0
			);
			let monthName = '';

			if (hasMonthBoundary) {
				const firstDayOfNewMonth = weekDays.find(
					(day) => day?.isFirstDayOfMonth && day.monthOffset > 0
				);
				if (firstDayOfNewMonth) {
					const monthDate = new Date(
						firstDayOfNewMonth.date.getFullYear(),
						firstDayOfNewMonth.date.getMonth(),
						1
					);
					monthName = monthDate.toLocaleDateString('en-GB', {
						month: 'long',
						year: 'numeric'
					});
				}
			}

			weeks.push({
				days: weekDays,
				hasMonthBoundary,
				monthName
			});
		}

		return { weeks, firstMonthName };
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
	<div class="space-y-4">
		<!-- Header -->
		<div class="flex items-center gap-2 px-4">
			<CalendarIcon class="h-4 w-4" />
			<h3 class="text-base font-semibold">Shift Calendar</h3>
			<span class="text-sm text-muted-foreground">- {calendarData.firstMonthName}</span>
		</div>

		<!-- Tessellated Calendar Grid with Month Separations -->
		<div class="px-4 space-y-2">
			<!-- Day names header -->
			<div class="grid grid-cols-7 gap-1 text-center">
				{#each dayNames as dayName, index (index)}
					<div class="text-xs font-medium text-muted-foreground p-2">
						{dayName}
					</div>
				{/each}
			</div>

			<!-- Calendar weeks with month boundaries -->
			{#each calendarData.weeks as week, weekIndex (weekIndex)}
				{#if week.hasMonthBoundary && weekIndex > 0}
					<!-- Month separator with title -->
					<div class="pt-8 pb-2">
						<div class="text-center">
							<div
								class="text-sm font-medium text-muted-foreground bg-background px-3 py-1 rounded-full border inline-block"
							>
								{week.monthName}
							</div>
						</div>
					</div>
				{/if}

				<!-- Week row -->
				<div class="grid grid-cols-7 gap-1">
					{#each week.days as dayData, dayIndex (dayIndex)}
						{#if dayData === null}
							<!-- Empty cell for padding -->
							<div class="aspect-square"></div>
						{:else}
							<div
								class="aspect-square border-2 rounded relative p-1
									{dayData.isPast ? 'bg-muted/30 border-muted/50' : 'border-muted/30'}
									{dayData.isToday ? 'ring-2 ring-primary ring-offset-1' : ''}
									{dayData.isOnDuty ? 'bg-green-100 border-green-400' : ''}
									{getMonthBackground(dayData.monthOffset)}
								"
							>
								<!-- Day number -->
								<div
									class="absolute top-1 left-1 text-xs font-medium leading-none
									{dayData.isOnDuty ? 'text-green-800' : 'text-muted-foreground'}
								"
								>
									{dayData.day}
								</div>

								<!-- Shift slots container -->
								<div class="pt-4 h-full flex flex-col gap-0.5 overflow-hidden">
									{#each dayData.shifts as shift, shiftIndex (shift.start_time)}
										{@const isBooked = isUserBookedForShift(shift, dayData.userShifts)}
										{@const isActive = isShiftActive(shift, dayData.userShifts)}

										<button
											class="text-xs px-1 py-0.5 rounded transition-colors flex items-center justify-between
												{isBooked
												? isActive
													? 'bg-green-600 text-white font-bold'
													: 'bg-blue-100 text-blue-800 border border-blue-300'
												: 'bg-primary/10 hover:bg-primary/20 text-primary border border-primary/30'}
											"
											onclick={() => !isBooked && onShiftSelect(shift)}
											disabled={isBooked || dayData.isPast}
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
