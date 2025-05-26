<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import { Button } from '$lib/components/ui/button';
	import ShiftCard from './ShiftCard.svelte';
	import type { ProcessedShift } from '$lib/types/shifts';

	export let shifts: ProcessedShift[] = [];
	export let isLoading: boolean = false;
	export let error: string | null = null;
	export let selectedFilter: string = 'all';
	export let isAuthenticated: boolean = false;
	export let onBookShift: (shift: ProcessedShift) => void = () => {};
	export let onSignIn: () => void = () => {};
	export let onRetry: () => void = () => {};
	export let onShowAllShifts: () => void = () => {};

	function LoadingSkeleton() {
		return Array(3).fill(null).map((_, index) => ({
			id: index,
			content: `
				<div class="p-4 ${index > 0 ? 'border-t border-border' : ''}">
					<div class="animate-pulse space-y-3">
						<div class="h-4 bg-muted rounded w-1/4"></div>
						<div class="h-4 bg-muted rounded w-1/2"></div>
						<div class="h-4 bg-muted rounded w-1/3"></div>
					</div>
				</div>
			`
		}));
	}
</script>

<div class="border border-border rounded overflow-hidden">
	{#if isLoading}
		{#each Array(3) as _, index (index)}
			<div class="p-4 {index > 0 ? 'border-t border-border' : ''}">
				<div class="animate-pulse space-y-3">
					<div class="h-4 bg-muted rounded w-1/4"></div>
					<div class="h-4 bg-muted rounded w-1/2"></div>
					<div class="h-4 bg-muted rounded w-1/3"></div>
				</div>
			</div>
		{/each}
	{:else if error}
		<div class="p-6 text-center">
			<CalendarIcon class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
			<p class="text-foreground font-medium">
				Error loading shifts: {error}
			</p>
			<Button variant="outline" size="sm" onclick={onRetry} class="mt-3">
				Try again
			</Button>
		</div>
	{:else if shifts.length === 0}
		<div class="p-6 text-center">
			<CalendarIcon class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
			<h3 class="text-base font-medium text-foreground mb-2">No shifts found</h3>
			<p class="text-sm text-muted-foreground">
				{#if shifts.length === 0}
					No patrol shifts are currently scheduled.
				{:else}
					Try adjusting your filters to see more shifts.
				{/if}
			</p>
			{#if selectedFilter !== 'all'}
				<Button variant="outline" size="sm" onclick={onShowAllShifts} class="mt-3">
					Show all shifts
				</Button>
			{/if}
		</div>
	{:else}
		{#each shifts as shift, index (shift.schedule_id + '-' + shift.start_time)}
			<div class="{index > 0 ? 'border-t border-border' : ''}">
				<ShiftCard 
					{shift} 
					{isAuthenticated} 
					{onBookShift} 
					{onSignIn}
				/>
			</div>
		{/each}
	{/if}
</div> 