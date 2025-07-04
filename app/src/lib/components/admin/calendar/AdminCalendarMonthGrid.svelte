<script lang="ts">
	import AdminCalendarCell from './AdminCalendarCell.svelte';
	import type { AdminShiftSlot } from '$lib/types';
	import type { AdminMonthGrid } from './admin-calendar-types.js';

	let {
		monthGrid,
		onFilledShiftClick,
		onUnfilledShiftClick
	}: {
		monthGrid: AdminMonthGrid;
		onFilledShiftClick?: (shift: AdminShiftSlot) => void;
		onUnfilledShiftClick?: (shift: AdminShiftSlot) => void;
	} = $props();
</script>

<div>
	<div class="space-y-1">
		{#each monthGrid.weeks as week, weekIndex (weekIndex)}
			<div class="grid grid-cols-7 gap-1">
				{#each week as cell, cellIndex (cellIndex)}
					<AdminCalendarCell {cell} {onFilledShiftClick} {onUnfilledShiftClick} />
				{/each}
			</div>
		{/each}
	</div>
</div>
