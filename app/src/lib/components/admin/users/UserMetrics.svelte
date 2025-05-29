<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import UsersIcon from '@lucide/svelte/icons/users';
	import type { UserMetrics } from '$lib/utils/userProcessing';

	let {
		isLoading = false,
		metrics: _metrics = null
	}: { isLoading?: boolean; metrics?: UserMetrics | null } = $props();
</script>

<Card.Root class="p-6">
	<div class="flex items-center gap-2 mb-4">
		<UsersIcon class="h-5 w-5 text-muted-foreground" />
		<h2 class="text-lg font-semibold">User Metrics</h2>
	</div>
	{#if isLoading}
		<Skeleton class="h-32 w-full" />
	{:else}
		<div class="text-center py-8 text-muted-foreground">
			<p class="text-sm">User metrics will be displayed here</p>
		</div>
	{/if}
</Card.Root>
