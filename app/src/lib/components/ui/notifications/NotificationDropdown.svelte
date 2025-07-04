<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import BellIcon from '@lucide/svelte/icons/bell';
	import BellRingIcon from '@lucide/svelte/icons/bell-ring';
	import MarkAsReadIcon from '@lucide/svelte/icons/check-check';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import { notificationStore, hasUnread, unreadCount } from '$lib/services/notificationService';
	import { formatDistanceToNow } from 'date-fns';
	import { userSession } from '$lib/stores/authStore';

	// Subscribe to notification state
	let notificationState = $state(notificationStore);
	let showUnread = $state(hasUnread);
	let badgeCount = $state(unreadCount);

	onMount(() => {
		// Only load notifications if user is authenticated
		if ($userSession.isAuthenticated) {
			notificationStore.fetchNotifications();

			// Set up periodic refresh every 15 seconds for more responsive updates
			const interval = setInterval(() => {
				if ($userSession.isAuthenticated) {
					notificationStore.fetchNotifications();
				}
			}, 15000);

			return () => clearInterval(interval);
		}
	});

	async function handleNotificationClick(notificationId: number) {
		await notificationStore.markAsRead(notificationId);
	}

	async function handleMarkAllRead() {
		await notificationStore.markAllAsRead();
	}

	function formatTimestamp(timestamp: string): string {
		try {
			return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
		} catch {
			return 'Recently';
		}
	}

	function getNotificationIcon(type: string) {
		switch (type) {
			case 'broadcast':
				return MessageSquareIcon;
			case 'shift_reminder':
				return ClockIcon;
			case 'shift_assignment':
				return BellRingIcon;
			default:
				return BellIcon;
		}
	}
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger>
		<Button variant="ghost" size="sm" class="relative h-9 w-9 p-0">
			{#if $showUnread}
				<BellRingIcon class="h-4 w-4" />
				{#if $badgeCount > 0}
					<Badge
						variant="destructive"
						class="absolute -top-2 -right-2 h-5 w-5 flex items-center justify-center p-0 text-xs"
					>
						{$badgeCount > 99 ? '99+' : $badgeCount}
					</Badge>
				{/if}
			{:else}
				<BellIcon class="h-4 w-4" />
			{/if}
			<span class="sr-only">Alerts</span>
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="w-72" align="end">
		<div class="flex items-center justify-between px-2 py-1">
			<h3 class="text-sm font-medium">Alerts</h3>
			{#if $badgeCount > 0}
				<Button variant="ghost" size="sm" onclick={handleMarkAllRead} class="h-5 px-1.5 text-xs">
					<MarkAsReadIcon class="h-3 w-3 mr-1" />
					Mark read
				</Button>
			{/if}
		</div>

		<Separator />

		<div class="max-h-64 overflow-y-auto">
			{#if $notificationState.isLoading && $notificationState.notifications.length === 0}
				<div class="flex items-center justify-center py-4">
					<div class="text-xs text-muted-foreground">Loading...</div>
				</div>
			{:else if $notificationState.notifications.length === 0}
				<div class="flex flex-col items-center justify-center py-4 text-center">
					<BellIcon class="h-6 w-6 text-muted-foreground mb-1" />
					<div class="text-xs text-muted-foreground">No alerts</div>
				</div>
			{:else}
				{#each $notificationState.notifications as notification (notification.id)}
					{@const IconComponent = getNotificationIcon(notification.type)}
					<DropdownMenu.Item
						class="flex items-start gap-2 p-2 cursor-pointer hover:bg-muted/50 {!notification.read
							? 'bg-muted/30'
							: ''}"
						onclick={() => handleNotificationClick(notification.id)}
					>
						<div class="flex-shrink-0 mt-0.5">
							<IconComponent
								class="h-3 w-3 {!notification.read ? 'text-primary' : 'text-muted-foreground'}"
							/>
						</div>

						<div class="flex-1 min-w-0">
							<div class="flex items-start justify-between gap-1">
								<p
									class="text-xs font-medium leading-tight {!notification.read
										? 'text-foreground'
										: 'text-muted-foreground'}"
								>
									{notification.title}
								</p>
								{#if !notification.read}
									<div class="h-1.5 w-1.5 bg-primary rounded-full flex-shrink-0 mt-0.5"></div>
								{/if}
							</div>

							<p class="text-xs text-muted-foreground mt-0.5 line-clamp-2 leading-tight">
								{notification.message}
							</p>

							<p class="text-xs text-muted-foreground/80 mt-0.5">
								{formatTimestamp(notification.timestamp)}
							</p>
						</div>
					</DropdownMenu.Item>
				{/each}
			{/if}
		</div>

		{#if $notificationState.notifications.length > 0}
			<Separator />
			<DropdownMenu.Item class="justify-center text-center py-1">
				<a href="/broadcasts" class="w-full">
					<Button variant="ghost" size="sm" class="text-xs w-full h-6">View all</Button>
				</a>
			</DropdownMenu.Item>
		{/if}
	</DropdownMenu.Content>
</DropdownMenu.Root>
