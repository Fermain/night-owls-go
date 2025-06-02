<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { pushNotificationService } from '$lib/services/pushNotificationService';
	import { onMount } from 'svelte';

	let status = $state({
		supported: false,
		permission: 'default' as NotificationPermission,
		subscribed: false
	});

	let debugInfo = $state<string[]>([]);
	let isInitialized = $state(false);

	onMount(async () => {
		await checkStatus();
		try {
			const initialized = await pushNotificationService.initialize();
			isInitialized = initialized;
			addDebugInfo(`Initialization: ${initialized ? 'SUCCESS' : 'FAILED'}`);
			await checkStatus();
		} catch (error) {
			addDebugInfo(`Initialization error: ${error}`);
		}
	});

	async function checkStatus() {
		try {
			status = await pushNotificationService.getStatus();
			addDebugInfo(`Status check - Supported: ${status.supported}, Permission: ${status.permission}, Subscribed: ${status.subscribed}`);
		} catch (error) {
			addDebugInfo(`Status check error: ${error}`);
		}
	}

	async function requestPermission() {
		try {
			const result = await pushNotificationService.subscribe();
			addDebugInfo(`Subscription attempt: ${result ? 'SUCCESS' : 'FAILED'}`);
			await checkStatus();
		} catch (error) {
			addDebugInfo(`Subscription error: ${error}`);
		}
	}

	async function testNotification() {
		try {
			await pushNotificationService.testNotification();
			addDebugInfo('Test notification sent');
		} catch (error) {
			addDebugInfo(`Test notification error: ${error}`);
		}
	}

	async function testVAPIDEndpoint() {
		try {
			const response = await fetch('/push/vapid-public');
			if (response.ok) {
				const data = await response.json();
				addDebugInfo(`VAPID endpoint: SUCCESS - Key length: ${data.vapid_public?.length || 'undefined'}`);
			} else {
				addDebugInfo(`VAPID endpoint: FAILED - ${response.status} ${response.statusText}`);
			}
		} catch (error) {
			addDebugInfo(`VAPID endpoint error: ${error}`);
		}
	}

	function addDebugInfo(info: string) {
		const timestamp = new Date().toLocaleTimeString();
		debugInfo = [
			`[${timestamp}] ${info}`,
			...debugInfo.slice(0, 19) // Keep last 20 entries
		];
	}

	function clearDebugInfo() {
		debugInfo = [];
	}

	function getStatusBadge(value: boolean | string) {
		if (typeof value === 'boolean') {
			return value ? 'default' : 'destructive';
		}
		return value === 'granted' ? 'default' : 'destructive';
	}

	function getStatusText(value: boolean | string) {
		if (typeof value === 'boolean') {
			return value ? 'Yes' : 'No';
		}
		return value;
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Push Notification Debug</Card.Title>
		<Card.Description>Test and troubleshoot push notification functionality</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<!-- Status Overview -->
		<div class="grid grid-cols-2 gap-4">
			<div>
				<p class="text-sm font-medium mb-1">Browser Support</p>
				<Badge variant={getStatusBadge(status.supported)}>
					{getStatusText(status.supported)}
				</Badge>
			</div>
			<div>
				<p class="text-sm font-medium mb-1">Permission</p>
				<Badge variant={getStatusBadge(status.permission)}>
					{getStatusText(status.permission)}
				</Badge>
			</div>
			<div>
				<p class="text-sm font-medium mb-1">Subscribed</p>
				<Badge variant={getStatusBadge(status.subscribed)}>
					{getStatusText(status.subscribed)}
				</Badge>
			</div>
			<div>
				<p class="text-sm font-medium mb-1">Service Initialized</p>
				<Badge variant={getStatusBadge(isInitialized)}>
					{getStatusText(isInitialized)}
				</Badge>
			</div>
		</div>

		<!-- Action Buttons -->
		<div class="flex flex-wrap gap-2">
			<Button size="sm" onclick={checkStatus}>Refresh Status</Button>
			<Button size="sm" onclick={testVAPIDEndpoint}>Test VAPID Endpoint</Button>
			{#if status.permission !== 'granted' || !status.subscribed}
				<Button size="sm" onclick={requestPermission} variant="default">
					Subscribe to Push
				</Button>
			{/if}
			{#if status.subscribed}
				<Button size="sm" onclick={testNotification} variant="secondary">
					Test Notification
				</Button>
			{/if}
		</div>

		<!-- Debug Info -->
		<div class="space-y-2">
			<div class="flex justify-between items-center">
				<h4 class="text-sm font-medium">Debug Log</h4>
				<Button size="sm" variant="outline" onclick={clearDebugInfo}>Clear</Button>
			</div>
			<div class="bg-muted p-3 rounded-md text-xs font-mono max-h-40 overflow-y-auto">
				{#if debugInfo.length === 0}
					<p class="text-muted-foreground">No debug information yet...</p>
				{:else}
					{#each debugInfo as info (info)}
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
	</Card.Content>
</Card.Root> 