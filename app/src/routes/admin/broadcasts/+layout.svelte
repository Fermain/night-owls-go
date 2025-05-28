<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import MessageCircleIcon from '@lucide/svelte/icons/message-circle';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import { getBroadcasts } from '$lib/queries/admin/broadcasts';
	import type { BroadcastData } from '$lib/schemas/broadcast';

	let searchTerm = $state('');
	let { children } = $props();

	// Audience options for display
	const audienceOptions = [
		{ value: 'all', label: 'All Users' },
		{ value: 'admins', label: 'Admins Only' },
		{ value: 'owls', label: 'Owls Only' },
		{ value: 'active', label: 'Active Users (last 30 days)' }
	];

	// Recent broadcasts query
	const recentBroadcastsQuery = $derived(
		createQuery<BroadcastData[], Error>({
			queryKey: ['broadcasts'],
			queryFn: getBroadcasts
		})
	);

	function formatRelativeTime(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));

		if (diffInHours < 1) return 'Just now';
		if (diffInHours < 24) return `${diffInHours}h ago`;
		if (diffInHours < 48) return 'Yesterday';
		return `${Math.floor(diffInHours / 24)}d ago`;
	}

	// Filter broadcasts based on search term
	const filteredBroadcasts = $derived.by(() => {
		const broadcasts = $recentBroadcastsQuery.data ?? [];
		if (!searchTerm) return broadcasts;

		return broadcasts.filter(
			(broadcast) =>
				broadcast.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
				(audienceOptions.find((opt) => opt.value === broadcast.audience)?.label || '')
					.toLowerCase()
					.includes(searchTerm.toLowerCase())
		);
	});
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
						<div class="p-3 border-b hover:bg-muted/50 transition-all duration-200 cursor-pointer group">
							<div class="space-y-2">
								<p class="text-sm font-medium line-clamp-2">{broadcast.message}</p>
								<div class="flex items-center justify-between text-xs text-muted-foreground">
									<div class="flex items-center gap-2">
										<span class="flex items-center gap-1">
											<UsersIcon class="h-3 w-3" />
											{broadcast.recipient_count}
										</span>
										<span class="flex items-center gap-1">
											<ClockIcon class="h-3 w-3" />
											{formatRelativeTime(broadcast.created_at)}
										</span>
									</div>
									{#if broadcast.push_enabled}
										<span class="text-blue-600 font-medium">SMS</span>
									{/if}
								</div>
								<div class="text-xs text-muted-foreground">
									{audienceOptions.find((opt) => opt.value === broadcast.audience)?.label ??
										'Unknown'}
								</div>
								<div class="text-xs">
									<span
										class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium
										{broadcast.status === 'sent'
											? 'bg-green-100 text-green-800'
											: broadcast.status === 'pending'
												? 'bg-yellow-100 text-yellow-800'
												: broadcast.status === 'sending'
													? 'bg-blue-100 text-blue-800'
													: 'bg-red-100 text-red-800'}"
									>
										{broadcast.status}
									</span>
									{#if broadcast.status === 'sent' && broadcast.sent_count !== broadcast.recipient_count}
										<span class="ml-2 text-orange-600 text-xs">
											{broadcast.sent_count}/{broadcast.recipient_count} delivered
										</span>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={broadcastsListContent} title="Broadcasts" bind:searchTerm>
	{@render children()}
</SidebarPage>
