<script lang="ts">
	import AdminShiftButton from './AdminShiftButton.svelte';
	import { getShiftBookingStatus } from '$lib/utils/shifts';
	import type { AdminShiftSlot } from '$lib/types';

	let {
		day,
		isPast,
		isToday,
		monthOffset,
		isWithinRange,
		shifts,
		onFilledShiftClick,
		onUnfilledShiftClick
	}: {
		day: number;
		isPast: boolean;
		isToday: boolean;
		monthOffset: number;
		isWithinRange: boolean;
		shifts: AdminShiftSlot[];
		onFilledShiftClick?: (shift: AdminShiftSlot) => void;
		onUnfilledShiftClick?: (shift: AdminShiftSlot) => void;
	} = $props();

	// Calculate shift statistics for this day
	const shiftStats = $derived.by(() => {
		const total = shifts.length;
		// Use the same working logic as other components
		const filled = shifts.filter(
			(shift) => getShiftBookingStatus(shift).status !== 'available'
		).length;
		const unfilled = total - filled;

		// Determine day status for background color
		let dayStatus: 'empty' | 'partial' | 'full' = 'empty';
		if (total > 0) {
			if (filled === total) {
				dayStatus = 'full';
			} else if (filled > 0) {
				dayStatus = 'partial';
			} else {
				dayStatus = 'empty';
			}
		}

		return { total, filled, unfilled, dayStatus };
	});

	// Get background class for day status and month distinction
	function getDayBackgroundClasses(
		dayStatus: 'empty' | 'partial' | 'full',
		monthOffset: number
	): string {
		let baseClasses = 'border-muted/30'; // Default neutral border

		// Add month offset styling (subtle overlay)
		if (monthOffset === 1) {
			baseClasses += ' opacity-80';
		} else if (monthOffset === 2) {
			baseClasses += ' opacity-60';
		}

		return baseClasses;
	}
</script>

<div
	class="border-2 rounded relative p-1 min-h-24
		{isPast
		? 'bg-muted/30 border-muted/50'
		: getDayBackgroundClasses(shiftStats.dayStatus, monthOffset)}
		{isToday ? 'ring-2 ring-primary ring-offset-1' : ''}
		{!isWithinRange ? 'opacity-30 bg-muted/50' : ''}
	"
>
	<!-- Day number and stats -->
	<div class="absolute top-1 left-1 text-xs font-medium leading-none">
		<div class="text-muted-foreground">{day}</div>
		{#if isWithinRange && shiftStats.total > 0}
			<div class="text-[10px] text-muted-foreground mt-0.5">
				{shiftStats.filled}/{shiftStats.total}
			</div>
		{/if}
	</div>

	<!-- Shift slots container - only show shifts for days within range -->
	{#if isWithinRange}
		<div class="pt-6 h-full flex flex-col gap-0.5 overflow-hidden">
			{#each shifts as shift (shift.start_time)}
				<AdminShiftButton {shift} {isPast} {onFilledShiftClick} {onUnfilledShiftClick} />
			{/each}
		</div>
	{:else}
		<!-- Empty space for out-of-range days to maintain consistent height -->
		<div class="pt-6 h-full"></div>
	{/if}
</div>
