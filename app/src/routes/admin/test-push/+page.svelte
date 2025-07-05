<script lang="ts">
	import { onMount } from 'svelte';
	import { pushNotificationService } from '$lib/services/pushNotificationService';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';

	let status = $state({
		supported: false,
		permission: 'default' as NotificationPermission,
		subscribed: false
	});

	let debugInfo = $state<string[]>([]);

	onMount(async () => {
		await checkStatus();
		await initializeService();
	});

	async function checkStatus() {
		try {
			status = await pushNotificationService.getStatus();
			addDebugInfo(
				`Status check - Supported: ${status.supported}, Permission: ${status.permission}, Subscribed: ${status.subscribed}`
			);
		} catch (error) {
			addDebugInfo(`Status check error: ${error}`);
		}
	}

	async function initializeService() {
		addDebugInfo('Initializing push notification service...');

		try {
			const success = await pushNotificationService.initialize();
			addDebugInfo(`Initialization: ${success ? 'SUCCESS' : 'FAILED'}`);
			await checkStatus();
		} catch (error) {
			addDebugInfo(`Initialization error: ${error}`);
		}
	}

	async function subscribe() {
		addDebugInfo('Attempting to subscribe to push notifications...');
		try {
			const success = await pushNotificationService.subscribe();
			addDebugInfo(`Subscription attempt: ${success ? 'SUCCESS' : 'FAILED'}`);
			await checkStatus();
		} catch (error) {
			addDebugInfo(`Subscription error: ${error}`);
		}
	}

	async function unsubscribe() {
		addDebugInfo('Attempting to unsubscribe from push notifications...');
		try {
			const success = await pushNotificationService.unsubscribe();
			addDebugInfo(`Unsubscription attempt: ${success ? 'SUCCESS' : 'FAILED'}`);
			await checkStatus();
		} catch (error) {
			addDebugInfo(`Unsubscription error: ${error}`);
		}
	}

	async function testNotification() {
		addDebugInfo('Testing local notification...');
		try {
			await pushNotificationService.testNotification();
			addDebugInfo('Test notification sent');
		} catch (error) {
			addDebugInfo(`Test notification error: ${error}`);
		}
	}

	async function testVAPIDEndpoint() {
		addDebugInfo('Testing VAPID endpoint...');
		try {
			const response = await fetch('/api/push/vapid-public');
			if (response.ok) {
				const data = await response.json();
				addDebugInfo(`VAPID endpoint: SUCCESS - Key length: ${data.key?.length || 'undefined'}`);
			} else {
				addDebugInfo(`VAPID endpoint: FAILED - ${response.status} ${response.statusText}`);
			}
		} catch (error) {
			addDebugInfo(`VAPID endpoint error: ${error}`);
		}
	}

	async function testBackendPush() {
		addDebugInfo('Testing backend push notification...');
		try {
			const response = await fetch('/api/admin/debug/test-push', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (response.ok) {
				const data = await response.json();
				addDebugInfo(`Backend push test: SUCCESS - ${data.message}`);
			} else {
				const errorText = await response.text();
				addDebugInfo(`Backend push test: FAILED - ${response.status} ${errorText}`);
			}
		} catch (error) {
			addDebugInfo(`Backend push test error: ${error}`);
		}
	}

	function testDirectNotification() {
		addDebugInfo('Testing direct browser notification...');
		try {
			if (Notification.permission !== 'granted') {
				addDebugInfo('Direct notification: Permission not granted');
				return;
			}

			const notification = new Notification('Direct Test Notification', {
				body: 'This notification was created directly by the browser (not service worker)',
				icon: '/icons/icon-192x192.png',
				tag: 'direct-test'
			});

			notification.onclick = () => {
				addDebugInfo('Direct notification clicked');
				notification.close();
			};

			notification.onshow = () => {
				addDebugInfo('Direct notification shown successfully');
			};

			notification.onerror = (error) => {
				addDebugInfo(`Direct notification error: ${error}`);
			};

			// Auto-close after 5 seconds
			setTimeout(() => {
				notification.close();
				addDebugInfo('Direct notification auto-closed');
			}, 5000);
		} catch (error) {
			addDebugInfo(`Direct notification failed: ${error}`);
		}
	}

	function addDebugInfo(message: string) {
		const timestamp = new Date().toLocaleTimeString();
		debugInfo = [`[${timestamp}] ${message}`, ...debugInfo.slice(0, 19)];
	}

	function clearDebugInfo() {
		debugInfo = [];
	}

	function getVariant(value: boolean | string): 'default' | 'destructive' {
		if (typeof value === 'boolean') {
			return value ? 'default' : 'destructive';
		}
		return value === 'granted' ? 'default' : 'destructive';
	}
</script>

<div class="container mx-auto p-6">
	<Card>
		<CardHeader>
			<CardTitle>🔔 Push Notification Testing</CardTitle>
			<CardDescription>Rock-solid push notification testing for community security</CardDescription>
		</CardHeader>

		<CardContent class="space-y-6">
			<!-- Status Overview -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<div>
					<p class="text-sm font-medium mb-1">Browser Support</p>
					<Badge variant={getVariant(status.supported)}>
						{status.supported ? 'Supported' : 'Not Supported'}
					</Badge>
				</div>
				<div>
					<p class="text-sm font-medium mb-1">Permission Status</p>
					<Badge variant={getVariant(status.permission)}>
						{status.permission}
					</Badge>
				</div>
				<div>
					<p class="text-sm font-medium mb-1">Subscription Status</p>
					<Badge variant={getVariant(status.subscribed)}>
						{status.subscribed ? 'Subscribed' : 'Not Subscribed'}
					</Badge>
				</div>
			</div>

			<!-- Action Buttons -->
			<div class="flex flex-wrap gap-2">
				<Button size="sm" onclick={checkStatus}>Refresh Status</Button>
				<Button size="sm" onclick={testVAPIDEndpoint}>Test VAPID Endpoint</Button>
				{#if status.permission !== 'granted' || !status.subscribed}
					<Button size="sm" onclick={subscribe} variant="default">Subscribe to Push</Button>
				{/if}
				{#if status.subscribed}
					<Button size="sm" onclick={testNotification} variant="secondary">Test Notification</Button
					>
					<Button size="sm" onclick={testBackendPush} variant="outline">Test Backend Push</Button>
					<Button size="sm" onclick={testDirectNotification} variant="outline"
						>Test Direct Notification</Button
					>
					<Button size="sm" onclick={unsubscribe} variant="destructive">Unsubscribe</Button>
				{/if}
			</div>

			<!-- Debug Log -->
			<div class="space-y-2">
				<div class="flex justify-between items-center">
					<h4 class="text-sm font-medium">Debug Log</h4>
					<Button size="sm" variant="outline" onclick={clearDebugInfo}>Clear</Button>
				</div>
				<div class="bg-muted p-3 rounded-md text-xs font-mono max-h-40 overflow-y-auto">
					{#if debugInfo.length === 0}
						<p class="text-muted-foreground">No debug information yet...</p>
					{:else}
						{#each debugInfo as info, index (index)}
							<div class="mb-1">{info}</div>
						{/each}
					{/if}
				</div>
			</div>

			<!-- Instructions -->
			<div class="bg-blue-50 dark:bg-blue-950/30 p-3 rounded-md">
				<h4 class="text-sm font-medium mb-2">Rock-Solid Testing Instructions:</h4>
				<ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
					<li>Verify all status indicators show green/success states</li>
					<li>Test VAPID endpoint connectivity to ensure server communication</li>
					<li>Subscribe to push notifications (this enables community security alerts)</li>
					<li>Test notification display to verify the system works correctly</li>
					<li>For production: Test emergency broadcasts to ensure critical alerts work</li>
				</ol>
			</div>

			<!-- Security Notice -->
			<div
				class="bg-amber-50 dark:bg-amber-950/30 p-3 rounded-md border border-amber-200 dark:border-amber-800"
			>
				<h4 class="text-sm font-medium mb-2 text-amber-800 dark:text-amber-200">
					⚠️ Community Security Notice
				</h4>
				<p class="text-sm text-amber-700 dark:text-amber-300">
					Push notifications are critical for community safety. Emergency alerts, incident reports,
					and shift reminders depend on this system working reliably. Always test thoroughly before
					deployment.
				</p>
			</div>
		</CardContent>
	</Card>
</div>
