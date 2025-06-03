<script lang="ts">
	import CalendarShiftButton from './CalendarShiftButton.svelte';
	import type { AvailableShiftSlot, UserBooking } from '$lib/services/api/user';

	let {
		day,
		isPast,
		isToday,
		isOnDuty,
		monthOffset,
		shifts,
		userShifts,
		onShiftSelect
	}: {
		day: number;
		isPast: boolean;
		isToday: boolean;
		isOnDuty: boolean;
		monthOffset: number;
		shifts: AvailableShiftSlot[];
		userShifts: UserBooking[];
		onShiftSelect: (shift: AvailableShiftSlot) => void;
	} = $props();

	// Create a combined list of all shifts for this day
	// Convert user bookings to AvailableShiftSlot format and combine with available shifts
	const allShiftsForDay = $derived.by(() => {
		const shiftMap = new Map<string, AvailableShiftSlot>();

		// Add available shifts
		shifts.forEach((shift) => {
			const key = `${shift.schedule_id}-${shift.start_time}`;
			shiftMap.set(key, shift);
		});

		// Add user bookings as pseudo-shifts (these will show with owl emoji)
		userShifts.forEach((booking) => {
			const key = `${booking.schedule_id}-${booking.shift_start}`;
			// Only add if not already in available shifts (to avoid duplicates)
			if (!shiftMap.has(key)) {
				const pseudoShift: AvailableShiftSlot = {
					schedule_id: booking.schedule_id,
					schedule_name: booking.schedule_name,
					start_time: booking.shift_start,
					end_time: booking.shift_end,
					is_booked: true // This will trigger the owl emoji
				};
				shiftMap.set(key, pseudoShift);
			}
		});

		// Convert back to array and sort by start time
		return Array.from(shiftMap.values()).sort(
			(a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
		);
	});

	// Get background class for month distinction
	function getMonthBackground(monthOffset: number): string {
		if (monthOffset === 0) return ''; // Current month - no background
		if (monthOffset === 1) return 'bg-muted/20'; // Next month - subtle background
		return 'bg-muted/40'; // Third month - slightly more background
	}
</script>

<div
	class="border-2 rounded relative p-1
		{isPast ? 'bg-muted/30 border-muted/50' : 'border-muted/30'}
		{isToday ? 'ring-2 ring-primary ring-offset-1' : ''}
		{isOnDuty ? 'bg-green-100 border-green-400' : ''}
		{getMonthBackground(monthOffset)}
	"
>
	<!-- Day number -->
	<div
		class="absolute top-1 left-1 text-xs font-medium leading-none
		{isOnDuty ? 'text-green-800' : 'text-muted-foreground'}
	"
	>
		{day}
	</div>

	<!-- Shift slots container -->
	<div class="pt-4 h-full flex flex-col gap-0.5 overflow-hidden">
		{#each allShiftsForDay as shift (shift.start_time)}
			<CalendarShiftButton {shift} {userShifts} {isPast} {onShiftSelect} />
		{/each}
	</div>
</div>
