<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import BellIcon from '@lucide/svelte/icons/bell';
	import BellOffIcon from '@lucide/svelte/icons/bell-off';
	import TestTubeIcon from '@lucide/svelte/icons/test-tube';
	import { pushNotificationService } from '$lib/services/pushNotificationService';
	import { toast } from 'svelte-sonner';
	import ServiceWorkerDebug from '$lib/components/debug/ServiceWorkerDebug.svelte';

	// State
	let isLoading = $state(true);
	let supported = $state(false);
	let permission = $state<NotificationPermission>('default');
	let subscribed = $state(false);

	// Reactive status text
	const statusText = $derived.by(() => {
		if (!supported) return 'Not supported on this device';
		if (permission === 'denied') return 'Permission denied';
		if (permission === 'default') return 'Permission not requested';
		if (permission === 'granted' && subscribed) return 'Enabled';
		if (permission === 'granted' && !subscribed) return 'Ready to enable';
		return 'Unknown status';
	});

	const statusColor = $derived.by(() => {
		if (!supported || permission === 'denied') return 'text-destructive';
		if (permission === 'granted' && subscribed) return 'text-green-600 dark:text-green-400';
		return 'text-muted-foreground';
	});

	onMount(async () => {
		await refreshStatus();

		// Initialize the service
		await pushNotificationService.initialize();
		await refreshStatus();
	});

	async function refreshStatus() {
		try {
			const status = await pushNotificationService.getStatus();
			supported = status.supported;
			permission = status.permission;
			subscribed = status.subscribed;
		} catch (error) {
			console.error('Failed to get push notification status:', error);
		} finally {
			isLoading = false;
		}
	}

	async function handleToggle() {
		isLoading = true;

		try {
			if (subscribed) {
				await pushNotificationService.unsubscribe();
			} else {
				await pushNotificationService.subscribe();
			}
			await refreshStatus();
		} catch (error) {
			console.error('Failed to toggle push notifications:', error);
			toast.error('Failed to update notification settings');
		} finally {
			isLoading = false;
		}
	}

	async function handleTest() {
		try {
			await pushNotificationService.testNotification();
			toast.success('Test notification sent!');
		} catch (error) {
			console.error('Failed to send test notification:', error);
			toast.error('Failed to send test notification');
		}
	}

	function getPermissionHelp(): string {
		if (!supported) {
			return 'Push notifications are not supported on this browser or device.';
		}

		switch (permission) {
			case 'denied':
				return 'Push notifications are blocked. Please enable them in your browser settings and refresh the page.';
			case 'default':
				return 'Enable push notifications to receive alerts about upcoming shifts and important messages.';
			case 'granted':
				return subscribed
					? 'You will receive push notifications for shifts and messages.'
					: 'Permission granted. Toggle the switch to subscribe.';
			default:
				return '';
		}
	}
</script>

<Card.Root>
	<Card.Header>
		<div class="flex items-center gap-3">
			{#if subscribed}
				<div class="p-2 bg-green-100 dark:bg-green-900/20 rounded-lg">
					<BellIcon class="h-5 w-5 text-green-600 dark:text-green-400" />
				</div>
			{:else}
				<div class="p-2 bg-gray-100 dark:bg-gray-900/20 rounded-lg">
					<BellOffIcon class="h-5 w-5 text-gray-600 dark:text-gray-400" />
				</div>
			{/if}
			<div>
				<Card.Title class="text-lg">Push Notifications</Card.Title>
				<Card.Description>Receive alerts for shifts and important messages</Card.Description>
			</div>
		</div>
	</Card.Header>

	<Card.Content class="space-y-4">
		<!-- Status -->
		<div class="flex items-center justify-between p-3 bg-muted/50 rounded-lg">
			<div>
				<p class="text-sm font-medium">Status</p>
				<p class="text-sm {statusColor}">{statusText}</p>
			</div>
			{#if isLoading}
				<div
					class="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin"
				></div>
			{:else if supported && permission !== 'denied'}
				<Switch checked={subscribed} onCheckedChange={handleToggle} disabled={isLoading} />
			{/if}
		</div>

		<!-- Help text -->
		<div class="text-sm text-muted-foreground bg-muted/30 p-3 rounded-lg">
			{getPermissionHelp()}
		</div>

		<!-- Actions -->
		{#if supported && subscribed}
			<div class="flex gap-2">
				<Button variant="outline" size="sm" onclick={handleTest} class="flex items-center gap-2">
					<TestTubeIcon class="h-4 w-4" />
					Test Notification
				</Button>

				<Button variant="ghost" size="sm" onclick={() => refreshStatus()} disabled={isLoading}>
					Refresh Status
				</Button>
			</div>
		{/if}

		<!-- Browser support info -->
		{#if !supported}
			<div class="text-xs text-muted-foreground border-t pt-4">
				<p class="mb-1"><strong>Supported browsers:</strong></p>
				<ul class="list-disc list-inside space-y-1">
					<li>Chrome/Edge 50+</li>
					<li>Firefox 44+</li>
					<li>Safari 16+ (on macOS 13+)</li>
					<li>Most mobile browsers</li>
				</ul>
			</div>
		{/if}
	</Card.Content>
</Card.Root>

<!-- Debug Component (only in development) -->
{#if typeof window !== 'undefined' && window.location.hostname === 'localhost'}
	<div class="mt-6">
		<ServiceWorkerDebug />
	</div>
{/if}
