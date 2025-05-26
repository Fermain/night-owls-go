<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CheckIcon from '@lucide/svelte/icons/check';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import SearchIcon from '@lucide/svelte/icons/search';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import { notificationStore } from '$lib/services/notificationService';
	import type { UserNotification } from '$lib/services/notificationService';

	// Real notification state
	let notificationState = $state(notificationStore);
	let showUnreadOnly = $state(false);
	let searchQuery = $state('');
	let selectedAudience = $state('all');
	let markingAsRead = $state(new Set<number>());

	onMount(() => {
		// Load notifications when page loads
		notificationStore.fetchNotifications();

		// Set up periodic refresh every 10 seconds on the broadcasts page
		const interval = setInterval(() => {
			notificationStore.fetchNotifications();
		}, 10000);

		// Keyboard shortcuts
		const handleKeydown = (e: KeyboardEvent) => {
			// Ctrl/Cmd + K to focus search
			if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
				e.preventDefault();
				const searchInput = document.querySelector('input[placeholder="Search messages..."]') as HTMLInputElement;
				searchInput?.focus();
			}
			// Escape to clear search
			if (e.key === 'Escape' && searchQuery) {
				searchQuery = '';
			}
			// Ctrl/Cmd + A to mark all as read
			if ((e.ctrlKey || e.metaKey) && e.key === 'a' && e.shiftKey) {
				e.preventDefault();
				markAllAsRead();
			}
		};

		document.addEventListener('keydown', handleKeydown);

		return () => {
			clearInterval(interval);
			document.removeEventListener('keydown', handleKeydown);
		};
	});

	function refreshNotifications() {
		notificationStore.fetchNotifications(true); // Force refresh
	}

	// Computed - using real notifications with filtering and search
	const filteredNotifications = $derived.by(() => {
		let notifications = $notificationState.notifications;
		
		// Filter by read status
		if (showUnreadOnly) {
			notifications = notifications.filter((n) => !n.read);
		}
		
		// Filter by audience
		if (selectedAudience !== 'all') {
			notifications = notifications.filter((n) => n.data?.audience === selectedAudience);
		}
		
		// Filter by search query
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase().trim();
			notifications = notifications.filter((n) => 
				n.message.toLowerCase().includes(query) ||
				n.title.toLowerCase().includes(query)
			);
		}
		
		return notifications;
	});

	// Enhanced search using Dexie (for better performance with large datasets)
	async function performDexieSearch(query: string) {
		if (!query.trim()) return;
		
		try {
			const { messageStorage } = await import('$lib/services/messageStorageService');
			const results = await messageStorage.searchMessages(query);
			
			// Convert back to UserNotification format and update store
			const searchNotifications = results.map(stored => ({
				id: stored.id,
				type: 'broadcast' as const,
				title: stored.title,
				message: stored.message,
				timestamp: stored.timestamp,
				read: stored.read,
				data: {
					broadcastId: stored.id,
					audience: stored.audience
				}
			}));
			
			// You could update the notification store with search results here
			// For now, we'll stick with the in-memory filtering for simplicity
		} catch (error) {
			console.error('Dexie search failed:', error);
		}
	}

	const unreadCount = $derived($notificationState.unreadCount);
	const isLoading = $derived($notificationState.isLoading);

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
			case 'all':
				return 'Everyone';
			case 'owls':
				return 'Volunteers';
			case 'admins':
				return 'Admins';
			case 'active':
				return 'Active Members';
			default:
				return 'Unknown';
		}
	}

	async function markAsRead(notificationId: number) {
		markingAsRead.add(notificationId);
		markingAsRead = markingAsRead; // Trigger reactivity
		try {
			await notificationStore.markAsRead(notificationId);
		} finally {
			markingAsRead.delete(notificationId);
			markingAsRead = markingAsRead; // Trigger reactivity
		}
	}

	async function markAllAsRead() {
		await notificationStore.markAllAsRead();
	}

	function toggleShowUnread() {
		showUnreadOnly = !showUnreadOnly;
	}

	function clearFilters() {
		searchQuery = '';
		selectedAudience = 'all';
		showUnreadOnly = false;
	}

	// Get unique audiences from notifications
	const availableAudiences = $derived.by(() => {
		const audiences = new Set(
			$notificationState.notifications
				.map(n => n.data?.audience)
				.filter((audience): audience is string => Boolean(audience))
		);
		return Array.from(audiences).sort();
	});
</script>

<svelte:head>
	<title>Messages - Night Owls</title>
</svelte:head>

<div
	class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800"
>
	<!-- Header -->
	<header
		class="bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border-b border-slate-200 dark:border-slate-700 sticky top-0 z-40"
	>
		<div class="px-4 py-3">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-lg font-semibold text-slate-900 dark:text-slate-100">Messages</h1>
					<p class="text-sm text-slate-600 dark:text-slate-400">
						{unreadCount > 0 ? `${unreadCount} unread` : 'All caught up'}
					</p>
				</div>
				<div class="flex items-center gap-2">
					<Button 
						variant="ghost" 
						size="sm" 
						onclick={refreshNotifications}
						disabled={isLoading}
					>
						<RefreshCwIcon class="h-4 w-4 mr-1 {isLoading ? 'animate-spin' : ''}" />
						Refresh
					</Button>
					{#if unreadCount > 0}
						<Button variant="ghost" size="sm" onclick={markAllAsRead}>
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
		<!-- Search and Filters -->
		<div class="mb-6 space-y-4">
			<!-- Search Bar -->
			<div class="relative">
				<SearchIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
				<input
					type="text"
					placeholder="Search messages... (⌘K)"
					bind:value={searchQuery}
					class="w-full pl-10 pr-4 py-2 border border-input rounded-md bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all duration-200"
				/>
				{#if searchQuery}
					<button
						onclick={() => searchQuery = ''}
						class="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground hover:text-foreground transition-colors"
						aria-label="Clear search"
					>
						✕
					</button>
				{/if}
			</div>

			<!-- Filters Row -->
			<div class="flex flex-wrap items-center gap-3">
				<!-- Audience Filter -->
				<div class="flex items-center gap-2">
					<FilterIcon class="h-4 w-4 text-muted-foreground" />
					<select
						bind:value={selectedAudience}
						class="px-3 py-1 text-sm border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					>
						<option value="all">All audiences</option>
						{#each availableAudiences as audience}
							<option value={audience}>{getAudienceLabel(audience)}</option>
						{/each}
					</select>
				</div>

				<!-- Clear Filters -->
				{#if searchQuery || selectedAudience !== 'all' || showUnreadOnly}
					<Button variant="outline" size="sm" onclick={clearFilters}>
						Clear filters
					</Button>
				{/if}

				<!-- Keyboard Shortcuts Hint -->
				<div class="hidden md:flex items-center gap-4 text-xs text-muted-foreground">
					<span>⌘K to search</span>
					<span>⇧⌘A to mark all read</span>
				</div>

				<!-- Results Count -->
				<div class="text-sm text-muted-foreground ml-auto">
					{filteredNotifications.length} of {$notificationState.notifications.length} messages
				</div>
			</div>
		</div>

		<!-- Quick Stats -->
		<div class="grid grid-cols-2 gap-3 mb-6">
			<Card.Root class="text-center">
				<Card.Content class="p-3">
					<div class="text-lg font-bold text-blue-600 dark:text-blue-400">
						{$notificationState.notifications.length}
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
			{#if isLoading && $notificationState.notifications.length === 0}
				<Card.Root class="text-center">
					<Card.Content class="p-8">
						<LoaderCircleIcon class="h-12 w-12 text-slate-400 mx-auto mb-3 animate-spin" />
						<h3 class="text-lg font-medium text-slate-900 dark:text-slate-100 mb-2">
							Loading messages...
						</h3>
						<p class="text-sm text-slate-600 dark:text-slate-400">
							Fetching your latest messages and updates.
						</p>
					</Card.Content>
				</Card.Root>
			{:else if filteredNotifications.length === 0}
				<Card.Root class="text-center">
					<Card.Content class="p-8">
						<MessageSquareIcon class="h-12 w-12 text-slate-400 mx-auto mb-3" />
						<h3 class="text-lg font-medium text-slate-900 dark:text-slate-100 mb-2">
							{#if $notificationState.notifications.length === 0}
								No messages yet
							{:else if searchQuery}
								No messages match your search
							{:else if showUnreadOnly}
								No unread messages
							{:else if selectedAudience !== 'all'}
								No messages for {getAudienceLabel(selectedAudience || 'all')}
							{:else}
								No messages found
							{/if}
						</h3>
						<p class="text-sm text-slate-600 dark:text-slate-400">
							{#if $notificationState.notifications.length === 0}
								Messages from coordinators will appear here.
							{:else if searchQuery}
								Try adjusting your search terms or clearing filters.
							{:else if showUnreadOnly}
								All caught up! Check back later for new updates.
							{:else}
								Try adjusting your filters to see more messages.
							{/if}
						</p>
						{#if searchQuery || selectedAudience !== 'all' || showUnreadOnly}
							<Button variant="outline" size="sm" onclick={clearFilters} class="mt-4">
								Clear all filters
							</Button>
						{/if}
					</Card.Content>
				</Card.Root>
			{:else}
				{#each filteredNotifications as notification (notification.id)}
					<Card.Root
						class="transition-all duration-200 hover:shadow-md hover:scale-[1.01] {!notification.read
							? 'ring-2 ring-blue-200 dark:ring-blue-800 bg-blue-50/50 dark:bg-blue-950/20'
							: ''}"
					>
						<Card.Content class="p-4">
							<div class="flex items-start gap-3">
								<!-- Status indicator -->
								<div class="mt-1">
									{#if !notification.read}
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
												{notification.title}
											</span>
											{#if notification.type === 'shift_reminder'}
												<Badge variant="destructive" class="text-xs">Urgent</Badge>
											{/if}
											<BellIcon class="h-3 w-3 text-slate-400" />
										</div>
										<div class="text-right">
											<div class="text-xs text-slate-500 dark:text-slate-400">
												{formatTimeAgo(notification.timestamp)}
											</div>
											<Badge variant="secondary" class="text-xs mt-1">
												{notification.data?.audience ? getAudienceLabel(notification.data.audience) : 'Message'}
											</Badge>
										</div>
									</div>

									<!-- Message content -->
									<div class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed mb-3">
										{notification.message}
									</div>

									<!-- Actions -->
									{#if !notification.read}
										<Button
											variant="ghost"
											size="sm"
											onclick={() => markAsRead(notification.id)}
											disabled={markingAsRead.has(notification.id)}
											class="text-xs h-auto p-1"
										>
											{#if markingAsRead.has(notification.id)}
												<LoaderCircleIcon class="h-3 w-3 mr-1 animate-spin" />
												Marking...
											{:else}
												<CheckIcon class="h-3 w-3 mr-1" />
												Mark as read
											{/if}
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
