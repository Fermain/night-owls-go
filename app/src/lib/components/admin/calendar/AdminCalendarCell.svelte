<script lang="ts">
	import AdminCalendarDayCell from './AdminCalendarDayCell.svelte';
	import type { AdminShiftSlot } from '$lib/types';
	import type { AdminCalendarCell } from './admin-calendar-types.js';

	let {
		cell,
		onFilledShiftClick,
		onUnfilledShiftClick
	}: {
		cell: AdminCalendarCell;
		onFilledShiftClick?: (shift: AdminShiftSlot) => void;
		onUnfilledShiftClick?: (shift: AdminShiftSlot) => void;
	} = $props();
</script>

{#if cell.type === 'empty'}
	<!-- Empty cell for padding -->
	<div></div>
{:else if cell.type === 'month-title'}
	<!-- Month name with overflow in first empty cell -->
	<div
		class="border-2 border-dashed border-muted/30 rounded relative flex items-center justify-start overflow-visible z-10"
	>
		<span class="text-sm font-bold text-muted-foreground whitespace-nowrap pl-1">
			{cell.monthName}
		</span>
	</div>
{:else if cell.type === 'day' && cell.dayData}
	<!-- Regular day cell -->
	<AdminCalendarDayCell
		day={cell.dayData.day}
		isPast={cell.dayData.isPast}
		isToday={cell.dayData.isToday}
		monthOffset={cell.dayData.monthOffset}
		isWithinRange={cell.dayData.isWithinRange}
		shifts={cell.dayData.shifts}
		{onFilledShiftClick}
		{onUnfilledShiftClick}
	/>
{/if}
