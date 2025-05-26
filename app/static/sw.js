// Night Owls Service Worker
// Basic service worker for push notifications and caching

const CACHE_NAME = 'night-owls-v1';
const CACHE_URLS = [
	'/'
];

// Install event
self.addEventListener('install', (event) => {
	console.log('🔧 Service worker installing');
	
	event.waitUntil(
		caches.open(CACHE_NAME)
			.then(async (cache) => {
				console.log('💾 Caching app shell');
				
				// Cache URLs individually with error handling
				const cachePromises = CACHE_URLS.map(async (url) => {
					try {
						await cache.add(url);
						console.log(`✅ Cached: ${url}`);
					} catch (error) {
						console.warn(`⚠️ Failed to cache ${url}:`, error.message);
						// Don't fail the whole installation for individual cache errors
					}
				});
				
				await Promise.allSettled(cachePromises);
				console.log('📦 Cache initialization completed');
			})
			.then(() => {
				console.log('⏩ Service worker skipping waiting');
				return self.skipWaiting();
			})
			.catch((error) => {
				console.error('❌ Service worker install failed:', error);
				throw error;
			})
	);
});

// Activate event
self.addEventListener('activate', (event) => {
	console.log('🚀 Service worker activating');
	
	event.waitUntil(
		caches.keys().then((cacheNames) => {
			return Promise.all(
				cacheNames
					.filter((cacheName) => {
						return cacheName !== CACHE_NAME;
					})
					.map((cacheName) => {
						console.log('🗑️ Deleting old cache:', cacheName);
						return caches.delete(cacheName);
					})
			);
		}).then(() => {
			console.log('📡 Service worker claiming clients');
			return self.clients.claim();
		}).then(() => {
			console.log('✅ Service worker is now ACTIVE and ready!');
			console.log('🎯 Service worker can now handle push notifications and caching');
			
			// Notify all clients that SW is ready
			return self.clients.matchAll().then(clients => {
				clients.forEach(client => {
					client.postMessage({
						type: 'SW_ACTIVATED',
						message: 'Service worker is now active and ready!'
					});
				});
			});
		})
	);
});

// Fetch event (basic caching strategy)
self.addEventListener('fetch', (event) => {
	// Skip non-GET requests
	if (event.request.method !== 'GET') {
		return;
	}

	// Skip API requests (let them go to network)
	if (event.request.url.includes('/api/')) {
		return;
	}

	event.respondWith(
		caches.match(event.request)
			.then((response) => {
				// Return cached version or fetch from network
				return response || fetch(event.request);
			})
			.catch(() => {
				// Fallback for offline
				if (event.request.mode === 'navigate') {
					return caches.match('/');
				}
			})
	);
});

// Push notification event handler
self.addEventListener('push', (event) => {
	console.log('Push event received:', event);

	if (!event.data) {
		console.log('Push event has no data');
		return;
	}

	let data;
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

	const options = {
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
		self.registration.showNotification(options.title, options).then(() => {
			// Notify all clients about the new push notification
			return self.clients.matchAll().then(clients => {
				clients.forEach(client => {
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
	console.log('Background sync:', event.tag);
	
	if (event.tag === 'background-sync') {
		event.waitUntil(
			// Handle offline actions when back online
			Promise.resolve()
		);
	}
});

// Message handler for communication with main thread
self.addEventListener('message', (event) => {
	console.log('📨 Service worker received message:', event.data);
	
	if (event.data.type === 'TEST_MESSAGE') {
		console.log('🧪 Test message received:', event.data.payload);
		
		// Send response back to client
		event.source.postMessage({
			type: 'TEST_RESPONSE',
			message: 'Service worker received and processed your test message!',
			timestamp: new Date().toISOString()
		});
	}
});

console.log('Service worker loaded and ready!'); 