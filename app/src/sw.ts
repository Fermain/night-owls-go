/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />
import { cleanupOutdatedCaches, precacheAndRoute } from 'workbox-precaching';

interface SyncEvent extends ExtendableEvent {
	tag: string;
	lastChance: boolean;
}

declare let self: ServiceWorkerGlobalScope & {
	__WB_MANIFEST: (string | PrecacheEntry)[];
};

interface PrecacheEntry {
	url: string;
	revision?: string;
}

interface ExtendedNotificationOptions extends NotificationOptions {
	actions?: NotificationAction[];
	title?: string;
	timestamp?: number;
	vibrate?: number[];
}

interface NotificationAction {
	action: string;
	title: string;
	icon?: string;
}

interface PushNotificationData {
	type?: string;
	title?: string;
	body?: string;
	booking_id?: number;
	actions?: NotificationAction[];
	requireInteraction?: boolean;
	silent?: boolean;
	tag?: string;
	vibrate?: number[];
}

// Clean up old caches
cleanupOutdatedCaches();

// Precache and route
precacheAndRoute(self.__WB_MANIFEST);

// Push notification event handler
self.addEventListener('push', (event: PushEvent) => {
	console.log('Push event received:', event);

	if (!event.data) {
		console.log('Push event has no data');
		return;
	}

	try {
		const data: PushNotificationData = event.data.json();
		console.log('Push data:', data);

		const options: ExtendedNotificationOptions = {
			body: data.body || 'You have a new notification',
			icon: '/logo.png',
			badge: '/logo.png',
			data: data,
			actions: data.actions || [],
			requireInteraction: data.requireInteraction || false,
			silent: data.silent || false,
			tag: data.tag || 'default',
			timestamp: Date.now(),
			vibrate: data.vibrate || [200, 100, 200]
		};

		// Handle different notification types
		switch (data.type) {
			case 'shift_reminder':
				options.title = 'Upcoming Shift Reminder';
				options.body = data.body || 'You have a shift starting soon';
				options.icon = '/logo.png';
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
				options.icon = '/logo.png';
				options.tag = 'broadcast';
				break;
			case 'shift_assignment':
				options.title = 'New Shift Assignment';
				options.body = data.body || 'You have been assigned a new shift';
				options.icon = '/logo.png';
				options.tag = 'shift_assignment';
				options.requireInteraction = true;
				break;
			default:
				options.title = data.title || 'Night Owls';
		}

		event.waitUntil(self.registration.showNotification(options.title || 'Night Owls', options));
	} catch (error) {
		console.error('Error processing push notification:', error);

		// Fallback notification
		event.waitUntil(
			self.registration.showNotification('Night Owls', {
				body: 'You have a new notification',
				icon: '/logo.png',
				tag: 'fallback'
			})
		);
	}
});

// Notification click event handler
self.addEventListener('notificationclick', (_event: NotificationEvent) => {
	console.log('Notification clicked:', _event);

	_event.notification.close();

	const data = _event.notification.data as PushNotificationData;
	const action = _event.action;

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
	_event.waitUntil(
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
self.addEventListener('sync', (event: Event) => {
	const syncEvent = event as SyncEvent;
	console.log('Background sync:', syncEvent.tag);

	if (syncEvent.tag === 'background-sync') {
		syncEvent.waitUntil(
			// Handle offline actions when back online
			Promise.resolve()
		);
	}
});

// Install event
self.addEventListener('install', (event: ExtendableEvent) => {
	console.log('Service worker installing');
	self.skipWaiting();
});

// Activate event
self.addEventListener('activate', (event: ExtendableEvent) => {
	console.log('Service worker activating');
	event.waitUntil(self.clients.claim());
});
