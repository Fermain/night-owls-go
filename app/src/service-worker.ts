/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

declare const self: ServiceWorkerGlobalScope;

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

// Create a unique cache name for this deployment
const CACHE = `cache-${version}`;

const ASSETS = [
	...build, // the app itself
	...files // everything in static
];

// Install event
self.addEventListener('install', (event) => {
	console.log('SvelteKit service worker installing');

	// Create a new cache and add all files to it
	async function addFilesToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	}

	event.waitUntil(addFilesToCache());
});

// Activate event
self.addEventListener('activate', (event) => {
	console.log('SvelteKit service worker activating');

	// Remove previous cached data from disk
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	event.waitUntil(deleteOldCaches());
});

// Fetch event
self.addEventListener('fetch', (event) => {
	// ignore POST requests etc
	if (event.request.method !== 'GET') return;

	async function respond() {
		const url = new URL(event.request.url);
		const cache = await caches.open(CACHE);

		// `build`/`files` can always be served from the cache
		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(url.pathname);

			if (response) {
				return response;
			}
		}

		// for everything else, try the network first, but
		// fall back to the cache if we're offline
		try {
			const response = await fetch(event.request);

			// if we're offline, fetch can return a value that is not a Response
			// instead of throwing - and we can't pass this non-Response to respondWith
			if (!(response instanceof Response)) {
				throw new Error('invalid response from fetch');
			}

			if (response.status === 200) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch (err) {
			const response = await cache.match(event.request);

			if (response) {
				return response;
			}

			// if there's no cache, then just error out
			// as there is nothing we can do to respond to this request
			throw err;
		}
	}

	event.respondWith(respond());
});

// Push notification event handler
self.addEventListener('push', (event) => {
	console.log('Push event received:', event);

	if (!event.data) {
		console.log('Push event has no data');
		return;
	}

	let data: PushNotificationData;
	try {
		data = event.data.json();
	} catch (error) {
		console.error('Failed to parse push data:', error);
		data = {
			title: 'Night Owls',
			body: 'You have a new notification',
			type: 'default'
		};
	}

	console.log('Push data:', data);

	const options: ExtendedNotificationOptions = {
		body: data.body || 'You have a new notification',
		icon: '/icons/icon-192x192.png',
		badge: '/icons/icon-192x192.png',
		data: data,
		requireInteraction: data.requireInteraction || false,
		silent: data.silent || false,
		tag: data.tag || 'default',
		vibrate: data.vibrate || [200, 100, 200]
	};

	// Handle different notification types
	switch (data.type) {
		case 'shift_reminder':
			options.title = 'Upcoming Shift Reminder';
			options.body = data.body || 'You have a shift starting soon';
			options.tag = 'shift_reminder';
			options.requireInteraction = true;
			options.actions = [
				{
					action: 'view_shift',
					title: 'View Shift'
				},
				{
					action: 'dismiss',
					title: 'Dismiss'
				}
			];
			break;
		case 'broadcast':
			options.title = data.title || 'New Message';
			options.body = data.body || 'You have a new message from coordinators';
			options.tag = 'broadcast';
			break;
		case 'shift_assignment':
			options.title = 'New Shift Assignment';
			options.body = data.body || 'You have been assigned a new shift';
			options.tag = 'shift_assignment';
			options.requireInteraction = true;
			break;
		default:
			options.title = data.title || 'Night Owls';
	}

	event.waitUntil(
		self.registration.showNotification(options.title!, options).then(() => {
			// Notify all clients about the new push notification
			return self.clients.matchAll().then((clients) => {
				clients.forEach((client) => {
					client.postMessage({
						type: 'PUSH_RECEIVED',
						notificationType: data.type,
						title: options.title,
						body: options.body,
						data: data
					});
				});
			});
		})
	);
});

// Notification click event handler
self.addEventListener('notificationclick', (event) => {
	console.log('Notification clicked:', event);

	event.notification.close();

	const data = event.notification.data;
	const action = event.action;

	let url = '/';

	// Handle different actions and notification types
	if (action === 'dismiss') {
		return;
	}

	switch (data?.type) {
		case 'shift_reminder':
			if (action === 'view_shift') {
				url = data.booking_id ? `/bookings` : '/shifts';
			}
			break;
		case 'broadcast':
			url = '/broadcasts';
			break;
		case 'shift_assignment':
			url = '/shifts';
			break;
		default:
			url = '/';
	}

	// Open or focus the app window
	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clients) => {
			// Check if there's already a window open
			for (const client of clients) {
				if (client.url.includes(url) && 'focus' in client) {
					return client.focus();
				}
			}

			// Open new window if none found
			if (self.clients.openWindow) {
				return self.clients.openWindow(url);
			}
		})
	);
});

// Background sync for offline actions (future enhancement)
self.addEventListener('sync', (event) => {
	const syncEvent = event as SyncEvent;
	console.log('Background sync:', syncEvent.tag);

	if (syncEvent.tag === 'background-sync') {
		syncEvent.waitUntil(
			// Handle offline actions when back online
			Promise.resolve()
		);
	}
});

// Message handler for communication with main thread
self.addEventListener('message', (event) => {
	console.log('Service worker received message:', event.data);

	if (event.data.type === 'TEST_MESSAGE') {
		console.log('Test message received:', event.data.payload);

		// Send response back to client
		if (event.source) {
			event.source.postMessage({
				type: 'TEST_RESPONSE',
				message: 'Service worker received and processed your test message!',
				timestamp: new Date().toISOString()
			});
		}
	}
});
