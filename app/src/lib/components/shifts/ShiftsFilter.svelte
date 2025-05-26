<script lang="ts">
	import type { FilterOption } from '$lib/types/shifts';

	export let selectedFilter: FilterOption = 'all';
	export let onFilterChange: (filter: FilterOption) => void = () => {};

	const filterOptions = [
		{ value: 'all', label: 'All Shifts' },
		{ value: 'tonight', label: 'Tonight Only' },
		{ value: 'available', label: 'Available Only' },
		{ value: 'urgent', label: 'Urgent Need' }
	] as const;

	function handleFilterChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		onFilterChange(target.value as FilterOption);
	}
</script>

<div class="mt-3 pt-3 border-t border-border">
	<select
		bind:value={selectedFilter}
		on:change={handleFilterChange}
		class="w-full px-3 py-2 border border-input rounded bg-background text-foreground focus:ring-1 focus:ring-ring"
	>
		{#each filterOptions as option (option.value)}
			<option value={option.value}>
				{option.label}
			</option>
		{/each}
	</select>
</div> 