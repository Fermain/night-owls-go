<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';

	let swStatus = $state<{
		supported: boolean;
		registered: boolean;
		active: boolean;
		registration: ServiceWorkerRegistration | null;
	}>({
		supported: false,
		registered: false,
		active: false,
		registration: null
	});

	let isLoading = $state(true);

	async function checkServiceWorkerStatus() {
		isLoading = true;
		
		try {
			const { serviceWorkerService } = await import('$lib/services/serviceWorkerService');
			swStatus = await serviceWorkerService.getStatus();
		} catch (error) {
			console.error('Failed to check service worker status:', error);
		} finally {
			isLoading = false;
		}
	}

	async function registerServiceWorker() {
		try {
			const { serviceWorkerService } = await import('$lib/services/serviceWorkerService');
			const success = await serviceWorkerService.register();
			
			if (success) {
				console.log('✅ Service worker registered');
				await checkServiceWorkerStatus();
			} else {
				console.error('❌ Service worker registration failed');
			}
		} catch (error) {
			console.error('Service worker registration error:', error);
		}
	}

	async function unregisterServiceWorker() {
		try {
			const { serviceWorkerService } = await import('$lib/services/serviceWorkerService');
			const success = await serviceWorkerService.unregister();
			
			if (success) {
				console.log('✅ Service worker unregistered');
				await checkServiceWorkerStatus();
			} else {
				console.error('❌ Service worker unregistration failed');
			}
		} catch (error) {
			console.error('Service worker unregistration error:', error);
		}
	}

	async function testServiceWorker() {
		try {
			const registration = await navigator.serviceWorker.getRegistration();
			
			if (!registration || !registration.active) {
				console.error('❌ No active service worker found');
				return;
			}

			// Test 1: Send a message to SW
			registration.active.postMessage({
				type: 'TEST_MESSAGE',
				payload: 'Hello from main thread!'
			});

			// Test 2: Test notification (if permission granted)
			if (Notification.permission === 'granted') {
				registration.showNotification('Test Notification', {
					body: 'This is a test notification from the service worker!',
					icon: '/logo.png',
					tag: 'sw-test'
				});
				console.log('✅ Test notification sent');
			} else {
				console.log('⚠️ Notification permission not granted, requesting...');
				const permission = await Notification.requestPermission();
				if (permission === 'granted') {
					registration.showNotification('Test Notification', {
						body: 'This is a test notification from the service worker!',
						icon: '/logo.png',
						tag: 'sw-test'
					});
					console.log('✅ Test notification sent');
				}
			}

			console.log('✅ Service worker tests completed');
		} catch (error) {
			console.error('❌ Service worker test failed:', error);
		}
	}

	onMount(() => {
		checkServiceWorkerStatus();
	});
</script>

<Card.Root class="w-full max-w-md">
	<Card.Header>
		<Card.Title class="flex items-center gap-2">
			🔧 Service Worker Debug
		</Card.Title>
		<Card.Description>
			Debug information for the service worker registration
		</Card.Description>
	</Card.Header>

	<Card.Content class="space-y-4">
		{#if isLoading}
			<div class="text-center text-muted-foreground">Checking status...</div>
		{:else}
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<span class="text-sm">Supported:</span>
					<Badge variant={swStatus.supported ? 'default' : 'destructive'}>
						{swStatus.supported ? 'Yes' : 'No'}
					</Badge>
				</div>

				<div class="flex items-center justify-between">
					<span class="text-sm">Registered:</span>
					<Badge variant={swStatus.registered ? 'default' : 'secondary'}>
						{swStatus.registered ? 'Yes' : 'No'}
					</Badge>
				</div>

				<div class="flex items-center justify-between">
					<span class="text-sm">Active:</span>
					<Badge variant={swStatus.active ? 'default' : 'secondary'}>
						{swStatus.active ? 'Yes' : 'No'}
					</Badge>
				</div>

				{#if swStatus.registration}
					<div class="text-xs text-muted-foreground border-t pt-2">
						<div>Scope: {swStatus.registration.scope}</div>
						{#if swStatus.registration.active}
							<div>Script URL: {swStatus.registration.active.scriptURL}</div>
							<div>State: {swStatus.registration.active.state}</div>
						{/if}
					</div>
				{/if}
			</div>

			<div class="flex gap-2 flex-wrap">
				<Button 
					variant="outline" 
					size="sm" 
					onclick={registerServiceWorker}
					disabled={swStatus.registered}
				>
					Register
				</Button>
				
				<Button 
					variant="outline" 
					size="sm" 
					onclick={unregisterServiceWorker}
					disabled={!swStatus.registered}
				>
					Unregister
				</Button>
				
				<Button 
					variant="ghost" 
					size="sm" 
					onclick={checkServiceWorkerStatus}
				>
					Refresh
				</Button>

				<Button 
					variant="default" 
					size="sm" 
					onclick={testServiceWorker}
					disabled={!swStatus.active}
				>
					🧪 Test
				</Button>
			</div>
		{/if}
	</Card.Content>
</Card.Root> 