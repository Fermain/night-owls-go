<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CheckIcon from '@lucide/svelte/icons/check';

	// Mock broadcasts data
	const mockBroadcasts = [
		{
			id: 1,
			message: 'Weekly safety briefing tomorrow at 6 PM. All volunteers welcome! We\'ll cover new patrol routes and safety protocols.',
			sender_name: 'Sarah Admin',
			audience: 'all',
			push_enabled: true,
			priority: 'normal',
			created_at: '2025-05-24T18:00:00Z',
			read: false
		},
		{
			id: 2,
			message: 'URGENT: Increased security concerns in the Main Street area tonight. Extra volunteers needed for patrol shifts.',
			sender_name: 'Mike Coordinator',
			audience: 'owls',
			push_enabled: true,
			priority: 'high',
			created_at: '2025-05-24T15:30:00Z',
			read: false
		},
		{
			id: 3,
			message: 'Thanks to everyone who participated in last week\'s community safety training. Your dedication makes our neighborhood safer!',
			sender_name: 'Sarah Admin',
			audience: 'all',
			push_enabled: false,
			priority: 'normal',
			created_at: '2025-05-23T09:15:00Z',
			read: true
		},
		{
			id: 4,
			message: 'Reminder: Please check in at the start of your shift and check out when finished. This helps us track coverage.',
			sender_name: 'System',
			audience: 'owls',
			push_enabled: false,
			priority: 'normal',
			created_at: '2025-05-22T20:00:00Z',
			read: true
		},
		{
			id: 5,
			message: 'New patrol equipment available for pickup at the community center. Contact admin for scheduling.',
			sender_name: 'Sarah Admin',
			audience: 'all',
			push_enabled: true,
			priority: 'normal',
			created_at: '2025-05-21T14:30:00Z',
			read: true
		}
	];

	// State
	let broadcasts = $state([...mockBroadcasts]);
	let showUnreadOnly = $state(false);

	// Computed
	const filteredBroadcasts = $derived.by(() => {
		if (showUnreadOnly) {
			return broadcasts.filter(b => !b.read);
		}
		return broadcasts;
	});

	const unreadCount = $derived(broadcasts.filter(b => !b.read).length);

	// Helper functions
	function formatTimeAgo(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffDays = Math.floor(diffHours / 24);
		
		if (diffHours < 1) return 'Just now';
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays === 1) return 'Yesterday';
		if (diffDays < 7) return `${diffDays}d ago`;
		
		return date.toLocaleDateString('en-GB', { 
			day: 'numeric', 
			month: 'short',
			year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
		});
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'high': 
				return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
			case 'normal': 
			default:
				return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
		}
	}

	function getAudienceLabel(audience: string) {
		switch (audience) {
			case 'all': return 'Everyone';
			case 'owls': return 'Volunteers';
			case 'admins': return 'Admins';
			case 'active': return 'Active Members';
			default: return 'Unknown';
		}
	}

	function markAsRead(broadcastId: number) {
		broadcasts = broadcasts.map(b => 
			b.id === broadcastId ? { ...b, read: true } : b
		);
	}

	function markAllAsRead() {
		broadcasts = broadcasts.map(b => ({ ...b, read: true }));
	}

	function toggleShowUnread() {
		showUnreadOnly = !showUnreadOnly;
	}
</script>

<svelte:head>
	<title>Messages - Night Owls</title>
</svelte:head>

<div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800">
	<!-- Header -->
	<header class="bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border-b border-slate-200 dark:border-slate-700 sticky top-0 z-40">
		<div class="px-4 py-3">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-lg font-semibold text-slate-900 dark:text-slate-100">Messages</h1>
					<p class="text-sm text-slate-600 dark:text-slate-400">
						{unreadCount > 0 ? `${unreadCount} unread` : 'All caught up'}
					</p>
				</div>
				<div class="flex items-center gap-2">
					{#if unreadCount > 0}
						<Button 
							variant="ghost" 
							size="sm"
							onclick={markAllAsRead}
						>
							<CheckIcon class="h-4 w-4 mr-1" />
							Mark all read
						</Button>
					{/if}
					<Button 
						variant="outline" 
						size="sm"
						onclick={toggleShowUnread}
						class={showUnreadOnly ? 'bg-blue-50 dark:bg-blue-950' : ''}
					>
						{showUnreadOnly ? 'Show All' : 'Unread Only'}
					</Button>
				</div>
			</div>
		</div>
	</header>

	<div class="px-4 py-6">
		<!-- Quick Stats -->
		<div class="grid grid-cols-2 gap-3 mb-6">
			<Card.Root class="text-center">
				<Card.Content class="p-3">
					<div class="text-lg font-bold text-blue-600 dark:text-blue-400">
						{broadcasts.length}
					</div>
					<div class="text-xs text-slate-600 dark:text-slate-400">Total Messages</div>
				</Card.Content>
			</Card.Root>
			
			<Card.Root class="text-center">
				<Card.Content class="p-3">
					<div class="text-lg font-bold text-orange-600 dark:text-orange-400">
						{unreadCount}
					</div>
					<div class="text-xs text-slate-600 dark:text-slate-400">Unread</div>
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Messages List -->
		<div class="space-y-3">
			{#if filteredBroadcasts.length === 0}
				<Card.Root class="text-center">
					<Card.Content class="p-8">
						<MessageSquareIcon class="h-12 w-12 text-slate-400 mx-auto mb-3" />
						<h3 class="text-lg font-medium text-slate-900 dark:text-slate-100 mb-2">
							{showUnreadOnly ? 'No unread messages' : 'No messages yet'}
						</h3>
						<p class="text-sm text-slate-600 dark:text-slate-400">
							{showUnreadOnly 
								? 'All caught up! Check back later for new updates.' 
								: 'Messages from coordinators will appear here.'
							}
						</p>
					</Card.Content>
				</Card.Root>
			{:else}
				{#each filteredBroadcasts as broadcast (broadcast.id)}
					<Card.Root 
						class="transition-all hover:shadow-md {!broadcast.read ? 'ring-2 ring-blue-200 dark:ring-blue-800 bg-blue-50/50 dark:bg-blue-950/20' : ''}"
					>
						<Card.Content class="p-4">
							<div class="flex items-start gap-3">
								<!-- Status indicator -->
								<div class="mt-1">
									{#if !broadcast.read}
										<div class="w-3 h-3 bg-blue-500 rounded-full"></div>
									{:else}
										<div class="w-3 h-3 bg-slate-300 dark:bg-slate-600 rounded-full"></div>
									{/if}
								</div>
								
								<div class="flex-1 min-w-0">
									<!-- Header -->
									<div class="flex items-start justify-between mb-2">
										<div class="flex items-center gap-2 flex-wrap">
											<span class="text-sm font-medium text-slate-900 dark:text-slate-100">
												{broadcast.sender_name}
											</span>
											{#if broadcast.priority === 'high'}
												<Badge variant="destructive" class="text-xs">Urgent</Badge>
											{/if}
											{#if broadcast.push_enabled}
												<BellIcon class="h-3 w-3 text-slate-400" />
											{/if}
										</div>
										<div class="text-right">
											<div class="text-xs text-slate-500 dark:text-slate-400">
												{formatTimeAgo(broadcast.created_at)}
											</div>
											<Badge variant="secondary" class="text-xs mt-1">
												{getAudienceLabel(broadcast.audience)}
											</Badge>
										</div>
									</div>
									
									<!-- Message content -->
									<div class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed mb-3">
										{broadcast.message}
									</div>
									
									<!-- Actions -->
									{#if !broadcast.read}
										<Button 
											variant="ghost" 
											size="sm" 
											onclick={() => markAsRead(broadcast.id)}
											class="text-xs h-auto p-1"
										>
											<CheckIcon class="h-3 w-3 mr-1" />
											Mark as read
										</Button>
									{/if}
								</div>
							</div>
						</Card.Content>
					</Card.Root>
				{/each}
			{/if}
		</div>

		<!-- Bottom spacing for mobile navigation -->
		<div class="h-6"></div>
	</div>
</div> 