<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import type { AvailableShiftSlot, UserBooking } from '$lib/services/api/user';

	let {
		shifts = [],
		userBookings = [],
		onShiftSelect
	}: {
		shifts: AvailableShiftSlot[];
		userBookings: UserBooking[];
		onShiftSelect: (shift: AvailableShiftSlot) => void;
	} = $props();

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

	// Get current date for calendar
	const today = new Date();
	const currentMonth = today.getMonth();
	const currentYear = today.getFullYear();

	// Get first day of current month and days in month
	const firstDayOfMonth = new Date(currentYear, currentMonth, 1);
	const lastDayOfMonth = new Date(currentYear, currentMonth + 1, 0);
	const daysInMonth = lastDayOfMonth.getDate();
	const startingDayOfWeek = firstDayOfMonth.getDay(); // 0 = Sunday

	// Create calendar grid
	const calendarDays = $derived.by(() => {
		const days: (CalendarDay | null)[] = [];

		// Add empty cells for days before month starts
		for (let i = 0; i < startingDayOfWeek; i++) {
			days.push(null);
		}

		// Add days of the month
		for (let day = 1; day <= daysInMonth; day++) {
			const date = new Date(currentYear, currentMonth, day);
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

		return days;
	});

	// Month name
	const monthName = $derived(
		new Date(currentYear, currentMonth).toLocaleDateString('en-GB', {
			month: 'long',
			year: 'numeric'
		})
	);

	// Day names
	const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

	function handleDayClick(dayData: CalendarDay) {
		if (dayData?.shifts?.length > 0) {
			// If there's only one shift, select it directly
			if (dayData.shifts.length === 1) {
				onShiftSelect(dayData.shifts[0]);
			} else {
				// If multiple shifts, select the first one for now
				// In a more complex implementation, you might show a picker
				onShiftSelect(dayData.shifts[0]);
			}
		}
	}
</script>

<Card.Root>
	<Card.Header class="pb-3">
		<div class="flex items-center gap-2">
			<CalendarIcon class="h-4 w-4" />
			<Card.Title class="text-base">Shift Calendar</Card.Title>
		</div>
		<p class="text-sm text-muted-foreground">
			{monthName} • Click highlighted dates to commit to shifts
		</p>
	</Card.Header>
	<Card.Content>
		<!-- Calendar Grid -->
		<div class="space-y-2">
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
				{#each calendarDays as dayData, index (index)}
					{#if dayData === null}
						<!-- Empty cell for padding -->
						<div class="aspect-square"></div>
					{:else}
						<button
							class="aspect-square p-1 rounded border-2 text-sm transition-colors relative
								{dayData.shifts.length > 0
								? 'bg-primary/10 hover:bg-primary/20 text-primary font-medium border-primary/30'
								: dayData.isPast
									? 'text-muted-foreground bg-muted/30 cursor-not-allowed border-muted/50'
									: 'hover:bg-muted/50 text-muted-foreground border-muted/30'}
								{dayData.isToday ? 'ring-2 ring-primary ring-offset-1' : ''}
								{dayData.isOnDuty ? 'bg-green-100 border-green-400 text-green-800 font-bold' : ''}
							"
							onclick={() => handleDayClick(dayData)}
							disabled={dayData.shifts.length === 0}
						>
							<div class="flex flex-col items-center justify-center h-full">
								<span class="text-xs leading-none">
									{dayData.day}
								</span>
								{#if dayData.userShifts.length > 0}
									<span class="text-lg leading-none mt-0.5">🦉</span>
								{/if}
								{#if dayData.shifts.length > 1 && dayData.userShifts.length === 0}
									<div
										class="absolute -top-1 -right-1 bg-primary text-primary-foreground text-xs rounded-full h-4 w-4 flex items-center justify-center"
									>
										{dayData.shifts.length}
									</div>
								{/if}
							</div>
						</button>
					{/if}
				{/each}
			</div>
		</div>

		<!-- Legend -->
		<div class="mt-4 flex flex-wrap gap-4 text-xs text-muted-foreground">
			<div class="flex items-center gap-1">
				<div class="w-3 h-3 bg-primary/10 border-2 border-primary/30 rounded"></div>
				<span>Available shifts</span>
			</div>
			<div class="flex items-center gap-1">
				<span class="text-lg">🦉</span>
				<span>My shifts</span>
			</div>
			<div class="flex items-center gap-1">
				<div class="w-3 h-3 bg-green-100 border-2 border-green-400 rounded"></div>
				<span>On duty now</span>
			</div>
			<div class="flex items-center gap-1">
				<div class="w-3 h-3 bg-muted/30 border-2 border-muted/50 rounded"></div>
				<span>No shifts</span>
			</div>
		</div>
	</Card.Content>
</Card.Root>
