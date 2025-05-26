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

	// Subscribe to notification state
	let notificationState = $state(notificationStore);
	let showUnread = $state(hasUnread);
	let badgeCount = $state(unreadCount);

	onMount(() => {
		// Load notifications on mount
		notificationStore.fetchNotifications();

		// Set up periodic refresh every 30 seconds
		const interval = setInterval(() => {
			notificationStore.fetchNotifications();
		}, 30000);

		return () => clearInterval(interval);
	});

	function handleNotificationClick(notificationId: number) {
		notificationStore.markAsRead(notificationId);
	}

	function handleMarkAllRead() {
		notificationStore.markAllAsRead();
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
		<Button 
			variant="ghost" 
			size="sm"
			class="relative h-9 w-9 p-0"
		>
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
			<span class="sr-only">Notifications</span>
		</Button>
	</DropdownMenu.Trigger>

	<DropdownMenu.Content class="w-80" align="end">
		<div class="flex items-center justify-between px-2 py-1.5">
			<h3 class="font-semibold">Notifications</h3>
			{#if $badgeCount > 0}
				<Button 
					variant="ghost" 
					size="sm" 
					onclick={handleMarkAllRead}
					class="h-6 px-2 text-xs"
				>
					<MarkAsReadIcon class="h-3 w-3 mr-1" />
					Mark all read
				</Button>
			{/if}
		</div>

		<Separator />

		<div class="max-h-80 overflow-y-auto">
			{#if $notificationState.isLoading && $notificationState.notifications.length === 0}
				<div class="flex items-center justify-center py-8">
					<div class="text-sm text-muted-foreground">Loading notifications...</div>
				</div>
			{:else if $notificationState.notifications.length === 0}
				<div class="flex flex-col items-center justify-center py-8 text-center">
					<BellIcon class="h-8 w-8 text-muted-foreground mb-2" />
					<div class="text-sm text-muted-foreground">No notifications yet</div>
					<div class="text-xs text-muted-foreground">You'll see messages and updates here</div>
				</div>
			{:else}
				{#each $notificationState.notifications as notification (notification.id)}
					{@const IconComponent = getNotificationIcon(notification.type)}
					<DropdownMenu.Item 
						class="flex items-start space-x-3 p-3 cursor-pointer hover:bg-muted/50 {!notification.read ? 'bg-muted/30' : ''}"
						onclick={() => handleNotificationClick(notification.id)}
					>
						<div class="flex-shrink-0 mt-0.5">
							<IconComponent class="h-4 w-4 {!notification.read ? 'text-primary' : 'text-muted-foreground'}" />
						</div>
						
						<div class="flex-1 min-w-0">
							<div class="flex items-start justify-between">
								<p class="text-sm font-medium leading-none {!notification.read ? 'text-foreground' : 'text-muted-foreground'}">
									{notification.title}
								</p>
								{#if !notification.read}
									<div class="h-2 w-2 bg-primary rounded-full ml-2 mt-1 flex-shrink-0"></div>
								{/if}
							</div>
							
							<p class="text-sm text-muted-foreground mt-1 line-clamp-2">
								{notification.message}
							</p>
							
							<p class="text-xs text-muted-foreground mt-1">
								{formatTimestamp(notification.timestamp)}
							</p>
						</div>
					</DropdownMenu.Item>
				{/each}
			{/if}
		</div>

		{#if $notificationState.notifications.length > 0}
			<Separator />
			<DropdownMenu.Item class="justify-center text-center py-2">
				<Button variant="ghost" size="sm" class="text-xs" onclick={() => {}}>
					View all messages
				</Button>
			</DropdownMenu.Item>
		{/if}
	</DropdownMenu.Content>
</DropdownMenu.Root>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style> 