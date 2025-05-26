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
	import { notificationStore } from '$lib/services/notificationService';
	import type { UserNotification } from '$lib/services/notificationService';

	// Real notification state
	let notificationState = $state(notificationStore);
	let showUnreadOnly = $state(false);

	onMount(() => {
		// Load notifications when page loads
		notificationStore.fetchNotifications();

		// Set up periodic refresh every 10 seconds on the broadcasts page
		const interval = setInterval(() => {
			notificationStore.fetchNotifications();
		}, 10000);

		return () => clearInterval(interval);
	});

	function refreshNotifications() {
		notificationStore.fetchNotifications(true); // Force refresh
	}

	// Computed - using real notifications
	const filteredNotifications = $derived.by(() => {
		const notifications = $notificationState.notifications;
		if (showUnreadOnly) {
			return notifications.filter((n) => !n.read);
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

	function markAsRead(notificationId: number) {
		notificationStore.markAsRead(notificationId);
	}

	function markAllAsRead() {
		notificationStore.markAllAsRead();
	}

	function toggleShowUnread() {
		showUnreadOnly = !showUnreadOnly;
	}
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
							{showUnreadOnly ? 'No unread messages' : 'No messages yet'}
						</h3>
						<p class="text-sm text-slate-600 dark:text-slate-400">
							{showUnreadOnly
								? 'All caught up! Check back later for new updates.'
								: 'Messages from coordinators will appear here.'}
						</p>
					</Card.Content>
				</Card.Root>
			{:else}
				{#each filteredNotifications as notification (notification.id)}
					<Card.Root
						class="transition-all hover:shadow-md {!notification.read
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
