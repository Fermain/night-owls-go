<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import MessageCircleIcon from '@lucide/svelte/icons/message-circle';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ClockIcon from '@lucide/svelte/icons/clock';

	let searchTerm = $state('');
	let { children } = $props();

	// Audience options for display
	const audienceOptions = [
		{ value: 'all', label: 'All Users' },
		{ value: 'admins', label: 'Admins Only' },
		{ value: 'owls', label: 'Owls Only' },
		{ value: 'active', label: 'Active Users (last 30 days)' }
	];

	// Recent broadcasts query (simulated for now)
	const recentBroadcastsQuery = $derived(
		createQuery({
			queryKey: ['recentBroadcasts'],
			queryFn: async () => {
				// Simulate API call
				await new Promise((resolve) => setTimeout(resolve, 500));
				return [
					{
						id: 1,
						message: 'Welcome to Night Owls! Thank you for volunteering.',
						audience: 'all',
						sentAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
						sentCount: 156,
						pushEnabled: true
					},
					{
						id: 2,
						message: 'Reminder: Please confirm your shifts for next week.',
						audience: 'owls',
						sentAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
						sentCount: 98,
						pushEnabled: false
					},
					{
						id: 3,
						message: 'System maintenance scheduled for this weekend. No action required.',
						audience: 'all',
						sentAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
						sentCount: 156,
						pushEnabled: false
					}
				];
			}
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
		<!-- Header -->
		<div class="p-3 border-b bg-muted/50">
			<div class="flex items-center gap-2">
				<MessageCircleIcon class="h-4 w-4" />
				<span class="text-sm font-medium">Recent Broadcasts</span>
			</div>
			<p class="text-xs text-muted-foreground">Previously sent messages</p>
		</div>

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
				<div class="p-2">
					{#each filteredBroadcasts as broadcast (broadcast.id)}
						<div class="border rounded-lg p-3 mb-2 hover:bg-accent transition-colors">
							<div class="space-y-2">
								<p class="text-sm font-medium line-clamp-2">{broadcast.message}</p>
								<div class="flex items-center justify-between text-xs text-muted-foreground">
									<div class="flex items-center gap-2">
										<span class="flex items-center gap-1">
											<UsersIcon class="h-3 w-3" />
											{broadcast.sentCount}
										</span>
										<span class="flex items-center gap-1">
											<ClockIcon class="h-3 w-3" />
											{formatRelativeTime(broadcast.sentAt)}
										</span>
									</div>
									{#if broadcast.pushEnabled}
										<span class="text-blue-600 font-medium">SMS</span>
									{/if}
								</div>
								<div class="text-xs text-muted-foreground">
									{audienceOptions.find((opt) => opt.value === broadcast.audience)?.label ??
										'Unknown'}
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
