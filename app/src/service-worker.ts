/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

declare const self: ServiceWorkerGlobalScope;

// App Shell Cache
const CACHE = `cache-${version}`;
const ASSETS = [
	...build, // the app itself
	...files // everything in static
];

// Install event - cache app shell
self.addEventListener('install', (event) => {
	async function addFilesToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	}

	event.waitUntil(addFilesToCache());
});

// Activate event - cleanup old caches
self.addEventListener('activate', (event) => {
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	event.waitUntil(deleteOldCaches());
});

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
	// ignore POST requests etc
	if (event.request.method !== 'GET') return;

	async function respond() {
		const url = new URL(event.request.url);
		const cache = await caches.open(CACHE);

		// serve build files from cache
		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(url.pathname);
			if (response) {
				return response;
			}
		}

		// try the network first
		try {
			const response = await fetch(event.request);

			// if we're offline, we might get a failed response
			if (!(response instanceof Response)) {
				throw new Error('invalid response from fetch');
			}

			if (response.status === 200) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch (err) {
			// fall back to cache
			const response = await cache.match(event.request);
			if (response) {
				return response;
			}

			// if there's no cache, we're out of luck
			throw err;
		}
	}

	event.respondWith(respond());
});

// **PUSH NOTIFICATION HANDLING - SvelteKit Native**

interface PushData {
	title?: string;
	body?: string;
	icon?: string;
	badge?: string;
	tag?: string;
	data?: Record<string, unknown>;
	requireInteraction?: boolean;
}

// Push event - handle incoming push messages
self.addEventListener('push', (event) => {
	console.log('[SW] Push event received:', event);

	let data: PushData = {};

	// Parse push data
	if (event.data) {
		try {
			data = event.data.json();
			console.log('[SW] Push data:', data);
		} catch (error) {
			console.warn('[SW] Failed to parse push data as JSON:', error);
			data = {
				title: 'Night Owls',
				body: event.data.text() || 'New notification'
			};
		}
	} else {
		console.warn('[SW] Push event had no data');
		data = {
			title: 'Night Owls',
			body: 'New notification'
		};
	}

	// Prepare notification options
	const options: NotificationOptions = {
		body: data.body || 'You have a new notification',
		icon: data.icon || '/icons/icon-192x192.png',
		badge: data.badge || '/icons/icon-96x96.png',
		tag: data.tag || 'night-owls-notification',
		data: data.data || {},
		requireInteraction: data.requireInteraction || false,
		silent: false
	};

	// Show notification
	const notificationPromise = self.registration.showNotification(
		data.title || 'Night Owls',
		options
	);

	// Send message to all clients about the push
	const messagePromise = self.clients.matchAll({ type: 'window' }).then((clients) => {
		clients.forEach((client) => {
			client.postMessage({
				type: 'PUSH_RECEIVED',
				title: data.title,
				body: data.body,
				data: data.data,
				notificationType: data.data?.type || 'broadcast'
			});
		});
	});

	event.waitUntil(Promise.all([notificationPromise, messagePromise]));
});

// Notification click event - handle notification interactions
self.addEventListener('notificationclick', (event) => {
	console.log('[SW] Notification click received:', event);

	event.notification.close();

	// Handle action button clicks
	if (event.action) {
		console.log('[SW] Notification action clicked:', event.action);

		switch (event.action) {
			case 'view':
				event.waitUntil(openAppPage('/'));
				break;
			case 'dismiss':
				// Just close the notification (already done above)
				break;
			default:
				console.log('[SW] Unknown action:', event.action);
				event.waitUntil(openAppPage('/'));
		}
		return;
	}

	// Handle main notification click
	const notificationData = event.notification.data;
	let targetUrl = '/';

	// Determine target URL based on notification type
	if (notificationData?.type) {
		switch (notificationData.type) {
			case 'shift_reminder':
				targetUrl = '/shifts';
				break;
			case 'emergency':
				targetUrl = '/emergency';
				break;
			case 'broadcast':
				targetUrl = '/broadcasts';
				break;
			case 'booking_confirmation':
				targetUrl = '/bookings/my';
				break;
			default:
				targetUrl = '/';
		}
	}

	event.waitUntil(openAppPage(targetUrl));
});

// Helper function to open app page
async function openAppPage(url: string) {
	const clients = await self.clients.matchAll({ type: 'window' });

	// Check if app is already open
	for (const client of clients) {
		if (client.url.includes(self.location.origin)) {
			// Focus existing window and navigate
			await client.focus();
			// Send message to navigate
			(client as WindowClient).postMessage({
				type: 'NAVIGATE',
				url: url
			});
			return;
		}
	}

	// Open new window
	await self.clients.openWindow(url);
}

// Message handling from main thread
self.addEventListener('message', (event) => {
	console.log('[SW] Message received:', event.data);

	if (event.data?.type === 'SKIP_WAITING') {
		self.skipWaiting();
	}
});

console.log('[SW] Service worker loaded');
