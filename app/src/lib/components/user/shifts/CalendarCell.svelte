<script lang="ts">
	import CalendarDayCell from './CalendarDayCell.svelte';
	import type { AvailableShiftSlot } from '$lib/services/api/user';
	import type { CalendarCell } from './calendar-types.js';

	let {
		cell,
		onShiftSelect
	}: {
		cell: CalendarCell;
		onShiftSelect: (shift: AvailableShiftSlot) => void;
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
	<CalendarDayCell
		day={cell.dayData.day}
		isPast={cell.dayData.isPast}
		isToday={cell.dayData.isToday}
		isOnDuty={cell.dayData.isOnDuty}
		monthOffset={cell.dayData.monthOffset}
		isWithinRange={cell.dayData.isWithinRange}
		shifts={cell.dayData.shifts}
		userShifts={cell.dayData.userShifts}
		{onShiftSelect}
	/>
{/if}
