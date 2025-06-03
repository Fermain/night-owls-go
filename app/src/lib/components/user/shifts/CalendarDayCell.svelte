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
		{#each shifts as shift (shift.start_time)}
			<CalendarShiftButton {shift} {userShifts} {isPast} {onShiftSelect} />
		{/each}
	</div>
</div>
