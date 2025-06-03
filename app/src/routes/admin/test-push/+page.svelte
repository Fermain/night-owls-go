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

	let serviceWorkerState = $state({
		registered: false,
		active: false,
		installing: false,
		waiting: false
	});

	let debugInfo = $state<string[]>([]);

	onMount(async () => {
		await checkStatus();
		await initializeService();
	});

	async function checkStatus() {
		try {
			status = await pushNotificationService.getStatus();
			await checkServiceWorkerState();
			addDebugInfo(
				`Status check - Supported: ${status.supported}, Permission: ${status.permission}, Subscribed: ${status.subscribed}`
			);
		} catch (error) {
			addDebugInfo(`Status check error: ${error}`);
		}
	}

	async function checkServiceWorkerState() {
		if (!('serviceWorker' in navigator)) return;

		try {
			const registration = await navigator.serviceWorker.getRegistration();
			serviceWorkerState.registered = !!registration;

			if (registration) {
				serviceWorkerState.active = !!registration.active;
				serviceWorkerState.installing = !!registration.installing;
				serviceWorkerState.waiting = !!registration.waiting;
			}
		} catch (error) {
			addDebugInfo(`Service worker state check error: ${error}`);
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

	function getBooleanText(value: boolean | string): string {
		if (typeof value === 'boolean') {
			return value ? 'Yes' : 'No';
		}
		return value;
	}
</script>

<svelte:head>
	<title>Push Notification Testing</title>
</svelte:head>

<div class="space-y-6 p-4">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">Push Notification Testing</h1>
		<p class="text-muted-foreground">
			Debug and test push notification functionality with the live backend
		</p>
	</div>

	<Card>
		<CardHeader>
			<CardTitle>Push Notification Debug</CardTitle>
			<CardDescription>Test and troubleshoot push notification functionality</CardDescription>
		</CardHeader>
		<CardContent class="space-y-4">
			<!-- Status Grid -->
			<div class="grid grid-cols-2 gap-4">
				<div>
					<p class="text-sm font-medium mb-1">Browser Support</p>
					<Badge variant={getVariant(status.supported)}>{getBooleanText(status.supported)}</Badge>
				</div>
				<div>
					<p class="text-sm font-medium mb-1">Permission</p>
					<Badge variant={getVariant(status.permission)}>{getBooleanText(status.permission)}</Badge>
				</div>
				<div>
					<p class="text-sm font-medium mb-1">Subscribed</p>
					<Badge variant={getVariant(status.subscribed)}>{getBooleanText(status.subscribed)}</Badge>
				</div>
				<div>
					<p class="text-sm font-medium mb-1">Service Worker Active</p>
					<Badge variant={getVariant(serviceWorkerState.active)}
						>{getBooleanText(serviceWorkerState.active)}</Badge
					>
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
				<h4 class="text-sm font-medium mb-2">Testing Instructions:</h4>
				<ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
					<li>First, check that all status indicators are green</li>
					<li>Test the VAPID endpoint to ensure server connectivity</li>
					<li>Subscribe to push notifications if not already subscribed</li>
					<li>Test a notification to verify the system works</li>
					<li>Try sending a broadcast with push notifications enabled</li>
				</ol>
			</div>
		</CardContent>
	</Card>
</div>
