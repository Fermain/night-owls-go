<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { Label } from '$lib/components/ui/label';
	import FilterIcon from '@lucide/svelte/icons/filter';

	let {
		showFilled = $bindable(),
		showUnfilled = $bindable()
	}: {
		showFilled: boolean;
		showUnfilled: boolean;
	} = $props();

	function clearFilters() {
		showFilled = true;
		showUnfilled = true;
	}

	const hasActiveFilters = $derived(!showFilled || !showUnfilled);
</script>

<div class="p-4 border-b">
	<div class="flex items-center gap-4">
		<div class="flex items-center gap-2">
			<FilterIcon class="h-4 w-4" />
			<Label class="text-sm font-medium">Filters:</Label>
		</div>

		<div class="flex items-center gap-4">
			<div class="flex items-center space-x-2">
				<Switch id="filled-filter" bind:checked={showFilled} />
				<Label for="filled-filter" class="text-sm cursor-pointer">Filled</Label>
			</div>
			<div class="flex items-center space-x-2">
				<Switch id="unfilled-filter" bind:checked={showUnfilled} />
				<Label for="unfilled-filter" class="text-sm cursor-pointer">Unfilled</Label>
			</div>
		</div>

		{#if hasActiveFilters}
			<Button variant="outline" size="sm" onclick={clearFilters}>
				<FilterIcon class="h-4 w-4 mr-2" />
				Reset
			</Button>
		{/if}
	</div>
</div>
