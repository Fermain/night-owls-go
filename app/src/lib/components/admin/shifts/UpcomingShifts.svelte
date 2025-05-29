<script lang="ts">
	import { createUpcomingShiftsQuery } from '$lib/queries/admin/shifts';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { goto } from '$app/navigation';
	import ShiftThumbnail from './ShiftThumbnail.svelte';
	import type { AdminShiftSlot } from '$lib/types';

	// Props for customization
	let {
		maxItems = 10,
		className = ''
	}: {
		maxItems?: number;
		className?: string;
	} = $props();

	// Query for upcoming shifts
	const upcomingShiftsQuery = $derived(createUpcomingShiftsQuery());

	const isLoading = $derived($upcomingShiftsQuery.isLoading);
	const isError = $derived($upcomingShiftsQuery.isError);
	const shifts = $derived($upcomingShiftsQuery.data?.slice(0, maxItems) || []);

	// Navigation function
	function navigateToShiftDetail(shift: AdminShiftSlot) {
		const shiftStartTime = encodeURIComponent(shift.start_time);
		goto(`/admin/shifts?shiftStartTime=${shiftStartTime}`);
	}
</script>

<div class="space-y-3 {className}">
	{#if isLoading}
		<!-- Loading skeletons -->
		<div class="space-y-2">
			{#each Array(3) as _, i (i)}
				<div class="space-y-2 p-3 border-b">
					<Skeleton class="h-4 w-3/4" />
					<Skeleton class="h-3 w-1/2" />
					<Skeleton class="h-3 w-2/3" />
				</div>
			{/each}
		</div>
	{:else if isError}
		<!-- Error state -->
		<div class="text-sm text-muted-foreground p-3 border-b border-destructive/20 bg-destructive/5">
			Failed to load upcoming shifts
		</div>
	{:else if shifts.length === 0}
		<!-- Empty state -->
		<div class="text-sm text-muted-foreground p-3 border-b text-center">
			No upcoming shifts in the next 2 weeks
		</div>
	{:else}
		<!-- Shifts list -->
		<div>
			{#each shifts as shift (shift.schedule_id + shift.start_time)}
				<ShiftThumbnail {shift} onSelect={navigateToShiftDetail} />
			{/each}
		</div>

		{#if shifts.length === maxItems && $upcomingShiftsQuery.data && $upcomingShiftsQuery.data.length > maxItems}
			<div class="text-xs text-muted-foreground text-center pt-2">
				Showing {maxItems} of {$upcomingShiftsQuery.data.length} upcoming shifts
			</div>
		{/if}
	{/if}
</div>
