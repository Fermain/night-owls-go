<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import PieChartIcon from '@lucide/svelte/icons/pie-chart';

	let { isLoading = false, distribution = null }: { isLoading?: boolean; distribution?: any } = $props();
</script>

<Card.Root class="p-6">
	<div class="flex items-center gap-2 mb-4">
		<PieChartIcon class="h-5 w-5 text-muted-foreground" />
		<h2 class="text-lg font-semibold">Shift Distribution</h2>
	</div>
	{#if isLoading}
		<Skeleton class="h-32 w-full" />
	{:else}
		<div class="text-center py-8 text-muted-foreground">
			<p class="text-sm">Shift distribution chart will be displayed here</p>
		</div>
	{/if}
</Card.Root> 