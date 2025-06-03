<script lang="ts">
	import CalendarCell from './CalendarCell.svelte';
	import type { AvailableShiftSlot } from '$lib/services/api/user';
	import type { MonthGrid } from './calendar-types.js';

	let {
		monthGrid,
		onShiftSelect
	}: {
		monthGrid: MonthGrid;
		onShiftSelect: (shift: AvailableShiftSlot) => void;
	} = $props();
</script>

<div class="px-4">
	<div class="space-y-1">
		{#each monthGrid.weeks as week, weekIndex (weekIndex)}
			<div class="grid grid-cols-7 gap-1">
				{#each week as cell, cellIndex (cellIndex)}
					<CalendarCell {cell} {onShiftSelect} />
				{/each}
			</div>
		{/each}
	</div>
</div>
