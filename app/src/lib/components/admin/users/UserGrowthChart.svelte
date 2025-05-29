<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';

	interface UserGrowthData {
		date: string;
		count: number;
		cumulative: number;
	}

	let {
		isLoading = false,
		data: _data = null
	}: { isLoading?: boolean; data?: UserGrowthData[] | null } = $props();
</script>

<Card.Root class="p-6">
	<div class="flex items-center gap-2 mb-4">
		<TrendingUpIcon class="h-5 w-5 text-muted-foreground" />
		<h2 class="text-lg font-semibold">User Growth</h2>
	</div>
	{#if isLoading}
		<Skeleton class="h-32 w-full" />
	{:else}
		<div class="text-center py-8 text-muted-foreground">
			<p class="text-sm">User growth chart will be displayed here</p>
		</div>
	{/if}
</Card.Root>
