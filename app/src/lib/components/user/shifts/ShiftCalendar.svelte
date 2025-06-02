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

	// Generate calendar data for multiple months
	const calendarMonths = $derived.by(() => {
		const months = [];

		for (let monthOffset = 0; monthOffset < monthsToShow; monthOffset++) {
			const currentMonth = new Date(today.getFullYear(), today.getMonth() + monthOffset, 1);
			const year = currentMonth.getFullYear();
			const month = currentMonth.getMonth();

			// Get first day of month and days in month
			const firstDayOfMonth = new Date(year, month, 1);
			const lastDayOfMonth = new Date(year, month + 1, 0);
			const daysInMonth = lastDayOfMonth.getDate();
			const startingDayOfWeek = firstDayOfMonth.getDay(); // 0 = Sunday

			const days: (CalendarDay | null)[] = [];

			// Add empty cells for days before month starts
			for (let i = 0; i < startingDayOfWeek; i++) {
				days.push(null);
			}

			// Add days of the month
			for (let day = 1; day <= daysInMonth; day++) {
				const date = new Date(year, month, day);
				const dateString = date.toISOString().split('T')[0]; // YYYY-MM-DD format

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

				days.push({
					day,
					date,
					dateString,
					shifts: dayShifts,
					userShifts: dayUserShifts,
					isToday: date.toDateString() === today.toDateString(),
					isPast: date < today && date.toDateString() !== today.toDateString(),
					isOnDuty
				});
			}

			months.push({
				year,
				month,
				monthName: currentMonth.toLocaleDateString('en-GB', {
					month: 'long',
					year: monthsToShow > 1 ? 'numeric' : undefined
				}),
				days
			});
		}

		return months;
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
</script>

<!-- Only render if there are shifts -->
{#if hasAnyShifts}
	<div class="space-y-4">
		<!-- Header -->
		<div class="flex items-center gap-2 px-4">
			<CalendarIcon class="h-4 w-4" />
			<h3 class="text-base font-semibold">Shift Calendar</h3>
		</div>

		<!-- Calendar Grid(s) -->
		<div class="px-4 space-y-6">
			{#each calendarMonths as monthData, monthIndex (monthData.month)}
				<div class="space-y-2">
					<!-- Month header -->
					<h4 class="text-sm font-medium text-muted-foreground">
						{monthData.monthName}
					</h4>

					<!-- Day names header -->
					<div class="grid grid-cols-7 gap-1 text-center">
						{#each dayNames as dayName, index (index)}
							<div class="text-xs font-medium text-muted-foreground p-2">
								{dayName}
							</div>
						{/each}
					</div>

					<!-- Calendar days -->
					<div class="grid grid-cols-7 gap-1">
						{#each monthData.days as dayData, index (index)}
							{#if dayData === null}
								<!-- Empty cell for padding -->
								<div class="aspect-square"></div>
							{:else}
								<div
									class="aspect-square border-2 rounded relative p-1
										{dayData.isPast ? 'bg-muted/30 border-muted/50' : 'border-muted/30'}
										{dayData.isToday ? 'ring-2 ring-primary ring-offset-1' : ''}
										{dayData.isOnDuty ? 'bg-green-100 border-green-400' : ''}
									"
								>
									<!-- Day number - positioned absolutely so it doesn't interfere with content flow -->
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

										<!-- Show message if no shifts available -->
										{#if dayData.shifts.length === 0 && !dayData.isPast}
											<div class="text-xs text-muted-foreground text-center mt-2 leading-tight">
												No shifts
											</div>
										{/if}
									</div>
								</div>
							{/if}
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
			</div>
		</div>
	</div>
{/if}
