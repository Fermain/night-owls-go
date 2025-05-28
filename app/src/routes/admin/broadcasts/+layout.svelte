<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import MessageCircleIcon from '@lucide/svelte/icons/message-circle';
	import { getBroadcasts } from '$lib/queries/admin/broadcasts';
	import type { BroadcastData } from '$lib/schemas/broadcast';
	import BroadcastThumbnail from '$lib/components/admin/broadcasts/BroadcastThumbnail.svelte';
	import { AUDIENCE_OPTIONS } from '$lib/utils/broadcasts';

	let searchTerm = $state('');
	let { children } = $props();

	// Recent broadcasts query
	const recentBroadcastsQuery = $derived(
		createQuery<BroadcastData[], Error>({
			queryKey: ['broadcasts'],
			queryFn: getBroadcasts
		})
	);

	// Filter broadcasts based on search term
	const filteredBroadcasts = $derived.by(() => {
		const broadcasts = $recentBroadcastsQuery.data ?? [];
		if (!searchTerm) return broadcasts;

		return broadcasts.filter(
			(broadcast) =>
				broadcast.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
				(AUDIENCE_OPTIONS.find((opt) => opt.value === broadcast.audience)?.label || '')
					.toLowerCase()
					.includes(searchTerm.toLowerCase())
		);
	});

	// Navigation function for broadcast selection
	function handleBroadcastSelect(broadcast: BroadcastData) {
		// Future: Navigate to broadcast detail view
		console.log('Selected broadcast:', broadcast.broadcast_id);
	}
</script>

{#snippet broadcastsListContent()}
	<div class="flex flex-col h-full">
		<!-- Broadcasts List -->
		<div class="flex-grow overflow-y-auto">
			{#if $recentBroadcastsQuery.isLoading}
				<div class="p-3 space-y-3">
					{#each Array(3) as _, i (i)}
						<div class="border rounded-lg p-3">
							<Skeleton class="h-4 w-3/4 mb-2" />
							<Skeleton class="h-3 w-1/2 mb-1" />
							<Skeleton class="h-3 w-1/3" />
						</div>
					{/each}
				</div>
			{:else if $recentBroadcastsQuery.isError}
				<div class="p-3 text-center">
					<MessageCircleIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-xs text-muted-foreground">Failed to load broadcasts</p>
				</div>
			{:else if filteredBroadcasts.length === 0}
				<div class="p-3 text-center">
					<MessageCircleIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-xs text-muted-foreground">
						{searchTerm ? `No broadcasts match "${searchTerm}"` : 'No broadcasts sent yet'}
					</p>
				</div>
			{:else}
				<div>
					{#each filteredBroadcasts as broadcast (broadcast.broadcast_id)}
						<BroadcastThumbnail {broadcast} onSelect={handleBroadcastSelect} />
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={broadcastsListContent} title="Broadcasts" bind:searchTerm>
	{@render children()}
</SidebarPage>
