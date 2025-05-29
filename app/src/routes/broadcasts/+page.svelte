<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import CheckIcon from '@lucide/svelte/icons/check';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import SearchIcon from '@lucide/svelte/icons/search';
	import { notificationStore } from '$lib/services/notificationService';
	import { userSession } from '$lib/stores/authStore';

	// Real notification state
	let notificationState = $state(notificationStore);
	let showUnreadOnly = $state(false);
	let searchQuery = $state('');
	let markingAsRead = $state(new Set<number>());

	onMount(() => {
		// Only load notifications if user is authenticated
		if ($userSession.isAuthenticated) {
			notificationStore.fetchNotifications();

			// Set up periodic refresh every 10 seconds on the broadcasts page
			const interval = setInterval(() => {
				if ($userSession.isAuthenticated) {
					notificationStore.fetchNotifications();
				}
			}, 10000);

			// Keyboard shortcuts
			const handleKeydown = (e: KeyboardEvent) => {
				// Ctrl/Cmd + K to focus search
				if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
					e.preventDefault();
					const searchInput = document.querySelector(
						'input[placeholder="Search messages..."]'
					) as HTMLInputElement;
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
		} else {
			// For unauthenticated users, still set up keyboard shortcuts
			const handleKeydown = (e: KeyboardEvent) => {
				// Ctrl/Cmd + K to focus search
				if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
					e.preventDefault();
					const searchInput = document.querySelector(
						'input[placeholder="Search messages..."]'
					) as HTMLInputElement;
					searchInput?.focus();
				}
				// Escape to clear search
				if (e.key === 'Escape' && searchQuery) {
					searchQuery = '';
				}
			};

			document.addEventListener('keydown', handleKeydown);

			return () => {
				document.removeEventListener('keydown', handleKeydown);
			};
		}
	});

	function refreshNotifications() {
		if ($userSession.isAuthenticated) {
			notificationStore.fetchNotifications(true); // Force refresh
		}
	}

	// Computed - using real notifications with filtering and search
	const filteredNotifications = $derived.by(() => {
		let notifications = $notificationState.notifications;

		// Filter by read status
		if (showUnreadOnly) {
			notifications = notifications.filter((n) => !n.read);
		}

		// Filter by search query
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase().trim();
			notifications = notifications.filter(
				(n) => n.message.toLowerCase().includes(query) || n.title.toLowerCase().includes(query)
			);
		}

		return notifications;
	});

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
		showUnreadOnly = false;
	}
</script>

<svelte:head>
	<title>Messages - Night Owls</title>
</svelte:head>

<div class="min-h-screen bg-background">
	<!-- Header -->
	<header class="bg-background border-b border-border sticky top-0 z-40">
		<div class="px-4 py-3">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-lg font-semibold text-foreground">Messages</h1>
					<p class="text-sm text-muted-foreground">
						{$notificationState.notifications.length} total, {unreadCount > 0
							? `${unreadCount} unread`
							: 'all read'}
					</p>
				</div>
				<div class="flex items-center gap-1">
					<Button
						variant="ghost"
						size="sm"
						onclick={refreshNotifications}
						disabled={isLoading}
						class="h-8 px-2"
					>
						<RefreshCwIcon class="h-4 w-4 {isLoading ? 'animate-spin' : ''}" />
					</Button>
					{#if unreadCount > 0}
						<Button variant="ghost" size="sm" onclick={markAllAsRead} class="h-8 px-2">
							<CheckIcon class="h-4 w-4" />
						</Button>
					{/if}
					<Button
						variant={showUnreadOnly ? 'default' : 'outline'}
						size="sm"
						onclick={toggleShowUnread}
						class="h-8 px-2"
					>
						{showUnreadOnly ? 'All' : 'Unread'}
					</Button>
				</div>
			</div>
		</div>
	</header>

	<div class="px-4 py-3">
		<!-- Search -->
		<div class="mb-3">
			<div class="relative">
				<SearchIcon
					class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"
				/>
				<input
					type="text"
					placeholder="Search messages..."
					bind:value={searchQuery}
					class="w-full pl-10 pr-4 py-2 border border-input rounded bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
				/>
				{#if searchQuery}
					<button
						onclick={() => (searchQuery = '')}
						class="absolute right-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground hover:text-foreground"
						aria-label="Clear search"
					>
						✕
					</button>
				{/if}
			</div>

			<!-- Clear Filters -->
			{#if searchQuery || showUnreadOnly}
				<div class="flex items-center justify-between mt-2">
					<div class="text-sm text-muted-foreground">
						{filteredNotifications.length} of {$notificationState.notifications.length} messages
					</div>
					<Button variant="ghost" size="sm" onclick={clearFilters} class="h-6 px-2 text-xs">
						Clear filters
					</Button>
				</div>
			{/if}
		</div>

		<!-- Messages List -->
		<div class="border border-border rounded overflow-hidden">
			{#if isLoading && $notificationState.notifications.length === 0}
				<div class="p-6 text-center">
					<LoaderCircleIcon class="h-8 w-8 text-muted-foreground mx-auto mb-3 animate-spin" />
					<h3 class="text-base font-medium text-foreground mb-2">Loading messages...</h3>
					<p class="text-sm text-muted-foreground">Fetching your latest messages and updates.</p>
				</div>
			{:else if filteredNotifications.length === 0}
				<div class="p-6 text-center">
					<MessageSquareIcon class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
					<h3 class="text-base font-medium text-foreground mb-2">
						{#if $notificationState.notifications.length === 0}
							No messages yet
						{:else if searchQuery}
							No messages match your search
						{:else if showUnreadOnly}
							No unread messages
						{:else}
							No messages found
						{/if}
					</h3>
					<p class="text-sm text-muted-foreground">
						{#if $notificationState.notifications.length === 0}
							Messages from coordinators will appear here.
						{:else if searchQuery}
							Try adjusting your search terms.
						{:else if showUnreadOnly}
							All caught up! Check back later for new updates.
						{:else}
							Try adjusting your filters to see more messages.
						{/if}
					</p>
					{#if searchQuery || showUnreadOnly}
						<Button variant="outline" size="sm" onclick={clearFilters} class="mt-4 h-8 px-3">
							Clear filters
						</Button>
					{/if}
				</div>
			{:else}
				{#each filteredNotifications as notification, index (notification.id)}
					<div
						class="p-3 {!notification.read ? 'bg-muted/30' : 'bg-background'} {index > 0
							? 'border-t border-border'
							: ''}"
					>
						<div class="flex items-start gap-3">
							<!-- Status indicator -->
							<div class="mt-1">
								{#if !notification.read}
									<div class="w-2 h-2 bg-foreground rounded-full"></div>
								{:else}
									<div class="w-2 h-2 bg-muted-foreground rounded-full"></div>
								{/if}
							</div>

							<div class="flex-1 min-w-0">
								<!-- Header -->
								<div class="flex items-start justify-between mb-2">
									<div class="flex items-center gap-2 flex-wrap">
										<span class="text-sm font-medium text-foreground">
											{notification.title}
										</span>
										{#if notification.type === 'shift_reminder'}
											<Badge variant="secondary" class="text-xs h-4 px-1">Urgent</Badge>
										{/if}
									</div>
									<div class="text-xs text-muted-foreground">
										{formatTimeAgo(notification.timestamp)}
									</div>
								</div>

								<!-- Message content -->
								<div class="text-sm text-foreground leading-relaxed mb-2">
									{notification.message}
								</div>

								<!-- Actions -->
								{#if !notification.read}
									<Button
										variant="ghost"
										size="sm"
										onclick={() => markAsRead(notification.id)}
										disabled={markingAsRead.has(notification.id)}
										class="text-xs h-6 px-2"
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
					</div>
				{/each}
			{/if}
		</div>

		<!-- Bottom spacing for mobile navigation -->
		<div class="h-6"></div>
	</div>
</div>
