<script lang="ts">
	import { onMount } from 'svelte';

	let serviceWorkerStatus = 'Checking...';
	let registrationStatus = 'Not checked';
	let pushNotificationStatus = 'Not checked';
	let swRegistration: ServiceWorkerRegistration | null = null;

	onMount(async () => {
		// Check if service worker is supported
		if ('serviceWorker' in navigator) {
			serviceWorkerStatus = 'Service Worker API supported';

			try {
				// Check if there's already a registration
				const registration = await navigator.serviceWorker.getRegistration();
				if (registration) {
					registrationStatus = `✅ Service Worker registered: ${registration.scope}`;
					swRegistration = registration;

					// Check service worker state
					if (registration.active) {
						registrationStatus += ` (State: ${registration.active.state})`;
					}
				} else {
					registrationStatus = '❌ No service worker registered';
				}

				// Check push notification permission
				if ('Notification' in window) {
					const permission = Notification.permission;
					pushNotificationStatus = `Notification permission: ${permission}`;
				} else {
					pushNotificationStatus = 'Notifications not supported';
				}
			} catch (error) {
				registrationStatus = `❌ Error checking registration: ${error}`;
			}
		} else {
			serviceWorkerStatus = '❌ Service Worker not supported';
		}
	});

	async function testServiceWorkerManually() {
		try {
			const registration = await navigator.serviceWorker.register('/sw.js', { scope: '/' });
			registrationStatus = `✅ Manual registration successful: ${registration.scope}`;
			swRegistration = registration;
		} catch (error) {
			registrationStatus = `❌ Manual registration failed: ${error}`;
		}
	}

	async function testPushPermission() {
		try {
			const permission = await Notification.requestPermission();
			pushNotificationStatus = `Permission request result: ${permission}`;
		} catch (error) {
			pushNotificationStatus = `Permission request failed: ${error}`;
		}
	}
</script>

<svelte:head>
	<title>PWA & Service Worker Test - Night Owls</title>
	<meta name="description" content="Test page for PWA and service worker functionality" />
</svelte:head>

<div class="container mx-auto px-4 py-8 max-w-4xl">
	<h1 class="text-3xl font-bold mb-6">PWA & Service Worker Test</h1>

	<div class="space-y-6">
		<!-- Service Worker Status -->
		<div class="bg-white p-6 rounded-lg shadow border">
			<h2 class="text-xl font-semibold mb-4">Service Worker Status</h2>
			<div class="space-y-2">
				<p><strong>API Support:</strong> {serviceWorkerStatus}</p>
				<p><strong>Registration:</strong> {registrationStatus}</p>
			</div>
			<button
				on:click={testServiceWorkerManually}
				class="mt-4 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
			>
				Test Manual Registration
			</button>
		</div>

		<!-- Push Notifications -->
		<div class="bg-white p-6 rounded-lg shadow border">
			<h2 class="text-xl font-semibold mb-4">Push Notifications</h2>
			<div class="space-y-2">
				<p><strong>Status:</strong> {pushNotificationStatus}</p>
			</div>
			<button
				on:click={testPushPermission}
				class="mt-4 bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
			>
				Request Notification Permission
			</button>
		</div>

		<!-- PWA Features -->
		<div class="bg-white p-6 rounded-lg shadow border">
			<h2 class="text-xl font-semibold mb-4">PWA Features</h2>
			<div class="space-y-2">
				<p>
					<strong>Standalone Mode:</strong>
					{window.matchMedia('(display-mode: standalone)').matches ? '✅ Yes' : '❌ No'}
				</p>
				<p>
					<strong>Cache API:</strong>
					{'caches' in window ? '✅ Supported' : '❌ Not supported'}
				</p>
				<p>
					<strong>Push API:</strong>
					{'PushManager' in window ? '✅ Supported' : '❌ Not supported'}
				</p>
			</div>
		</div>

		<!-- Service Worker Details -->
		{#if swRegistration}
			<div class="bg-white p-6 rounded-lg shadow border">
				<h2 class="text-xl font-semibold mb-4">Service Worker Details</h2>
				<div class="space-y-2 text-sm font-mono">
					<p><strong>Scope:</strong> {swRegistration.scope}</p>
					<p><strong>Installing:</strong> {swRegistration.installing ? '✅ Yes' : '❌ No'}</p>
					<p><strong>Waiting:</strong> {swRegistration.waiting ? '✅ Yes' : '❌ No'}</p>
					<p><strong>Active:</strong> {swRegistration.active ? '✅ Yes' : '❌ No'}</p>
					{#if swRegistration.active}
						<p><strong>Active State:</strong> {swRegistration.active.state}</p>
						<p><strong>Script URL:</strong> {swRegistration.active.scriptURL}</p>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Console Output -->
		<div class="bg-gray-50 p-6 rounded-lg border">
			<h2 class="text-xl font-semibold mb-4">Instructions</h2>
			<div class="space-y-2 text-sm">
				<p>1. Open browser Developer Tools (F12)</p>
				<p>2. Go to <strong>Application</strong> tab → <strong>Service Workers</strong></p>
				<p>3. Check if service worker is registered and active</p>
				<p>
					4. Go to <strong>Application</strong> tab → <strong>Manifest</strong> to check PWA manifest
				</p>
				<p>
					5. Test push notifications in the <strong>Application</strong> tab →
					<strong>Storage</strong> section
				</p>
			</div>
		</div>
	</div>
</div>

<style>
	/* Add some basic styling */
	.container {
		min-height: 100vh;
		background-color: #f9fafb;
	}
</style>
