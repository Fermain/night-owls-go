// Push notification handlers for Night Owls PWA

// Debug flag - environment-based, defaults to false for production
const DEBUG = self.location.hostname === 'localhost' || self.location.hostname === '127.0.0.1';

// Push event handler
self.addEventListener('push', (event) => {
	if (DEBUG) console.log('[SW] Push event received:', event);

	let data = {};

	// Parse push data
	if (event.data) {
		try {
			data = event.data.json();
			if (DEBUG) console.log('[SW] Push data parsed:', data);
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
	const options = {
		body: data.body || 'You have a new notification',
		icon: data.icon || '/icons/icon-192x192.png',
		badge: data.badge || '/icons/icon-96x96.png',
		tag: data.tag || 'night-owls-notification',
		data: data.data || {},
		requireInteraction: data.requireInteraction || false,
		silent: false,
		vibrate: [200, 100, 200] // Add vibration pattern
	};

	if (DEBUG) {
		console.log('[SW] Attempting to show notification with options:', {
			title: data.title || 'Night Owls',
			options: options
		});
	}

	// Show notification with error handling
	const notificationPromise = self.registration
		.showNotification(data.title || 'Night Owls', options)
		.then(() => {
			if (DEBUG) console.log('[SW] Notification shown successfully');
		})
		.catch((error) => {
			console.error('[SW] Failed to show notification:', error);
		});

	// Send message to all clients about the push
	const messagePromise = self.clients.matchAll({ type: 'window' }).then((clients) => {
		if (DEBUG) console.log('[SW] Sending push message to', clients.length, 'clients');
		clients.forEach((client) => {
			client.postMessage({
				type: 'PUSH_RECEIVED',
				title: data.title,
				body: data.body,
				data: data.data,
				notificationType: (data.data && data.data.type) || 'broadcast'
			});
		});
	});

	event.waitUntil(Promise.all([notificationPromise, messagePromise]));
});

// Notification click event handler
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
	if (notificationData && notificationData.type) {
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
async function openAppPage(url) {
	const clients = await self.clients.matchAll({ type: 'window' });

	// Check if app is already open
	for (const client of clients) {
		if (client.url.includes(self.location.origin)) {
			// Focus existing window and navigate
			await client.focus();
			// Send message to navigate
			client.postMessage({
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

	if (event.data && event.data.type === 'SKIP_WAITING') {
		self.skipWaiting();
	}
});

console.log('[SW] Push notification handlers loaded');
