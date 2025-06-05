<script lang="ts">
	import { onMount } from 'svelte';
	import { pushNotificationService } from '$lib/services/pushNotificationService';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Switch } from '$lib/components/ui/switch';

	let status = $state({
		supported: false,
		permission: 'default' as NotificationPermission,
		subscribed: false
	});

	let isLoading = $state(false);

	onMount(async () => {
		await checkStatus();
	});

	async function checkStatus() {
		try {
			status = await pushNotificationService.getStatus();
		} catch (error) {
			console.error('Failed to get push notification status:', error);
		}
	}

	async function togglePushNotifications() {
		isLoading = true;
		try {
			if (status.subscribed) {
				await pushNotificationService.unsubscribe();
			} else {
				await pushNotificationService.subscribe();
			}
			await checkStatus();
		} catch (error) {
			console.error('Failed to toggle push notifications:', error);
		} finally {
			isLoading = false;
		}
	}

	async function testNotification() {
		try {
			await pushNotificationService.testNotification();
		} catch (error) {
			console.error('Failed to send test notification:', error);
		}
	}

	const supported = $derived(status.supported);
	const enabled = $derived(status.subscribed && status.permission === 'granted');
	const canEnable = $derived(status.supported && status.permission !== 'denied');
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="flex items-center justify-between">
			Push Notifications
			<Badge variant={supported ? 'default' : 'secondary'}>
				{supported ? 'Supported' : 'Not Supported'}
			</Badge>
		</Card.Title>
	</Card.Header>

	<Card.Content class="space-y-4">
		{#if supported}
			<div class="flex items-center justify-between">
				<div class="space-y-1">
					<p class="text-sm font-medium">Enable push notifications</p>
					<p class="text-xs text-muted-foreground">
						Receive important security alerts and shift reminders
					</p>
				</div>
				<Switch
					checked={enabled}
					disabled={!canEnable || isLoading}
					onCheckedChange={togglePushNotifications}
				/>
			</div>

			{#if enabled}
				<div class="pt-4 border-t">
					<Button size="sm" variant="outline" onclick={testNotification}>
						Send Test Notification
					</Button>
				</div>
			{/if}

			{#if status.permission === 'denied'}
				<div class="text-xs text-muted-foreground p-3 bg-muted rounded-md">
					<p class="font-medium mb-1">Permission Denied</p>
					<p>
						Push notifications are blocked. Please enable them in your browser settings and refresh
						the page.
					</p>
				</div>
			{/if}
		{:else}
			<div class="text-xs text-muted-foreground">
				<p class="mb-2"><strong>Push notifications are not supported in this browser.</strong></p>
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
