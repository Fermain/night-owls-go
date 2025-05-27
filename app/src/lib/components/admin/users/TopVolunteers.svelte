<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AwardIcon from '@lucide/svelte/icons/award';

	let { isLoading = false, volunteers = null }: { isLoading?: boolean; volunteers?: any } = $props();
</script>

<Card.Root class="p-6">
	<div class="flex items-center gap-2 mb-4">
		<AwardIcon class="h-5 w-5 text-muted-foreground" />
		<h2 class="text-lg font-semibold">Top Volunteers</h2>
	</div>
	{#if isLoading}
		<Skeleton class="h-32 w-full" />
	{:else}
		<div class="text-center py-8 text-muted-foreground">
			<p class="text-sm">Top volunteers will be displayed here</p>
		</div>
	{/if}
</Card.Root> 