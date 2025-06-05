/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

declare const self: ServiceWorkerGlobalScope & {
	__WB_MANIFEST: Array<{ url: string; revision: string | null }>;
};

// Workbox precaching manifest injection point
const WB_MANIFEST = self.__WB_MANIFEST || [];

// Additional files to precache from Workbox

const precacheManifest = self.__WB_MANIFEST;

// Extend NotificationOptions to include missing properties
interface NotificationAction {
	action: string;
	title: string;
	icon?: string;
}

interface ExtendedNotificationOptions extends NotificationOptions {
	title?: string;
	actions?: NotificationAction[];
	vibrate?: number[];
}

// Type for sync events
interface SyncEvent extends ExtendableEvent {
	tag: string;
}

// Type for push notification data
interface PushNotificationData {
	title?: string;
	body?: string;
	type?: string;
	actions?: NotificationAction[];
	requireInteraction?: boolean;
	silent?: boolean;
	tag?: string;
	vibrate?: number[];
	booking_id?: string;
}

// Create a unique cache name for this version
const CACHE = `cache-${version}`;

// Assets to cache: build files + static files
const ASSETS = [
	...build, // the app itself
	...files // everything in `static`
];

// Install event - cache assets
self.addEventListener('install', (event) => {
	async function addFilesToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
		console.log('[SW] Assets cached successfully');
	}

	event.waitUntil(addFilesToCache());
	// Immediately take control
	self.skipWaiting();
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) {
				await caches.delete(key);
				console.log(`[SW] Deleted old cache: ${key}`);
			}
		}
		console.log('[SW] Service worker activated');
	}

	event.waitUntil(deleteOldCaches().then(() => self.clients.claim()));
});

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
	// Only handle GET requests
	if (event.request.method !== 'GET') return;

	async function respond() {
		const url = new URL(event.request.url);
		const cache = await caches.open(CACHE);

		// Always serve assets from cache
		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(url.pathname);
			if (response) {
				return response;
			}
		}

		// For everything else, try network first, then cache
		try {
			const response = await fetch(event.request);

			if (!(response instanceof Response)) {
				throw new Error('Invalid response from fetch');
			}

			// Cache successful responses
			if (response.status === 200) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch (err) {
			const response = await cache.match(event.request);
			if (response) {
				return response;
			}
			throw err;
		}
	}

	event.respondWith(respond());
});

// Push notification event handler - CRITICAL FOR COMMUNITY SECURITY
self.addEventListener('push', (event) => {
	console.log('[SW] Push notification received');

	if (!event.data) {
		console.warn('[SW] Push event has no data');
		return;
	}

	let data;
	try {
		data = event.data.json();
	} catch (error) {
		console.error('[SW] Failed to parse push data:', error);
		// Fallback to generic notification
		data = {
			title: 'Night Owls Alert',
			body: 'You have a new security notification',
			type: 'emergency'
		};
	}

	console.log('[SW] Push data:', data);

	// Base notification options
	const options = {
		body: data.body || 'You have a new notification',
		icon: '/icons/icon-192x192.png',
		badge: '/icons/icon-192x192.png',
		data: data,
		requireInteraction: false,
		silent: false,
		tag: data.tag || 'night-owls',
		timestamp: Date.now(),
		vibrate: [200, 100, 200, 100, 200]
	};

	// Customize based on notification type - CRITICAL TYPES GET PRIORITY
	let title = data.title || 'Night Owls';

	switch (data.type) {
		case 'emergency':
		case 'incident':
			title = '🚨 EMERGENCY ALERT';
			options.requireInteraction = true;
			options.vibrate = [500, 200, 500, 200, 500];
			options.tag = 'emergency';
			break;

		case 'shift_reminder':
			title = '🦉 Shift Reminder';
			options.requireInteraction = true;
			options.tag = 'shift_reminder';
			break;

		case 'broadcast':
			title = '📢 ' + (data.title || 'Community Message');
			options.tag = 'broadcast';
			break;

		case 'shift_assignment':
			title = '🔔 New Shift Assignment';
			options.requireInteraction = true;
			options.tag = 'shift_assignment';
			break;

		default:
			title = data.title || 'Night Owls';
	}

	event.waitUntil(
		self.registration
			.showNotification(title, options)
			.then(() => {
				console.log('[SW] Notification displayed successfully');

				// Notify all clients about the push notification
				return self.clients.matchAll().then((clients) => {
					clients.forEach((client) => {
						client.postMessage({
							type: 'PUSH_RECEIVED',
							notificationType: data.type,
							title: title,
							body: options.body,
							data: data
						});
					});
				});
			})
			.catch((error) => {
				console.error('[SW] Failed to show notification:', error);
			})
	);
});

// Notification click event handler
self.addEventListener('notificationclick', (event) => {
	console.log('[SW] Notification clicked:', event.notification.tag);

	event.notification.close();

	// Handle different notification types
	const data = event.notification.data;
	let urlToOpen = '/';

	switch (data?.type) {
		case 'emergency':
		case 'incident':
			urlToOpen = '/emergency';
			break;
		case 'shift_reminder':
		case 'shift_assignment':
			urlToOpen = '/shifts';
			break;
		case 'broadcast':
			urlToOpen = '/broadcasts';
			break;
		default:
			urlToOpen = '/';
	}

	event.waitUntil(
		self.clients.matchAll().then((clients) => {
			// Check if we already have a window open
			for (const client of clients) {
				if (client.url === new URL(urlToOpen, self.location.origin).href && 'focus' in client) {
					return (client as WindowClient).focus();
				}
			}
			// If no window is open, open a new one
			if (self.clients.openWindow) {
				return self.clients.openWindow(urlToOpen);
			}
		})
	);
});

// Message event handler for communication with main thread
self.addEventListener('message', (event) => {
	if (event.data?.type === 'SKIP_WAITING') {
		self.skipWaiting();
	}
});

console.log('[SW] Service worker loaded and ready for community security notifications');
