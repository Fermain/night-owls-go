import { authenticatedFetch } from '$lib/utils/api';
import { toast } from 'svelte-sonner';
import { notificationStore } from './notificationService';

interface VAPIDKeyResponse {
	key: string;
}

interface PushSubscriptionData {
	endpoint: string;
	p256dh_key: string;
	auth_key: string;
	user_agent: string;
	platform: string;
}

/**
 * Rock-solid push notification service for community security app
 * Designed for maximum reliability with minimal complexity
 */
class PushNotificationService {
	private vapidPublicKey: string | null = null;
	private registration: ServiceWorkerRegistration | null = null;
	private subscription: PushSubscription | null = null;
	private isInitialized = false;

	/**
	 * Initialize the push notification service
	 * Simple, reliable initialization with proper error handling
	 */
	async initialize(): Promise<boolean> {
		try {
			// Basic support check
			if (!this.isSupported()) {
				console.warn('[PushService] Push notifications not supported');
				return false;
			}

			// Wait for service worker to be ready with timeout
			this.registration = await this.waitForServiceWorker();

			if (!this.registration) {
				console.error('[PushService] No service worker registration available');
				return false;
			}

			// Get VAPID key from server with retry logic
			await this.fetchVAPIDKey();

			// Check for existing subscription
			this.subscription = await this.registration.pushManager.getSubscription();

			// Set up message listener
			this.setupMessageListener();

			this.isInitialized = true;
			console.log('[PushService] Initialized successfully');
			return true;
		} catch (error) {
			console.error('[PushService] Initialization failed:', error);
			return false;
		}
	}

	/**
	 * Check if push notifications are supported
	 */
	private isSupported(): boolean {
		return 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;
	}

	/**
	 * Wait for service worker to be ready with timeout
	 */
	private async waitForServiceWorker(): Promise<ServiceWorkerRegistration | null> {
		if (!('serviceWorker' in navigator)) {
			return null;
		}

		try {
			// Wait for service worker to be ready with 10 second timeout
			const timeoutPromise = new Promise<never>((_, reject) => {
				setTimeout(() => reject(new Error('Service worker ready timeout')), 10000);
			});

			const registration = await Promise.race([navigator.serviceWorker.ready, timeoutPromise]);

			// Ensure the service worker is active
			if (registration.active) {
				return registration;
			}

			// If installing, wait for it to activate
			if (registration.installing) {
				await new Promise<void>((resolve) => {
					registration.installing!.addEventListener('statechange', () => {
						if (registration.installing!.state === 'activated') {
							resolve();
						}
					});
				});
				return registration;
			}

			return registration;
		} catch (error) {
			console.error('[PushService] Service worker ready failed:', error);
			return null;
		}
	}

	/**
	 * Fetch VAPID public key with retry logic
	 */
	private async fetchVAPIDKey(): Promise<void> {
		// Try to get from cache first
		const cached = sessionStorage.getItem('vapid-key');
		if (cached) {
			this.vapidPublicKey = cached;
			return;
		}

		// Fetch from server with retries
		let attempts = 0;
		const maxAttempts = 3;

		while (attempts < maxAttempts) {
			try {
				const response = await fetch('/api/push/vapid-public');
				if (!response.ok) {
					throw new Error(`HTTP ${response.status}: ${response.statusText}`);
				}

				const data: VAPIDKeyResponse = await response.json();
				this.vapidPublicKey = data.key;

				// Cache for session
				sessionStorage.setItem('vapid-key', data.key);
				console.log('[PushService] VAPID key fetched successfully');
				return;
			} catch (error) {
				attempts++;
				console.warn(`[PushService] VAPID key fetch attempt ${attempts} failed:`, error);

				if (attempts >= maxAttempts) {
					throw new Error(`Failed to fetch VAPID key after ${maxAttempts} attempts`);
				}

				// Wait before retry
				await new Promise((resolve) => setTimeout(resolve, 1000 * attempts));
			}
		}
	}

	/**
	 * Subscribe to push notifications
	 */
	async subscribe(): Promise<boolean> {
		try {
			if (!this.isInitialized) {
				const initialized = await this.initialize();
				if (!initialized) {
					throw new Error('Failed to initialize push service');
				}
			}

			if (!this.registration || !this.vapidPublicKey) {
				throw new Error('Service not properly initialized');
			}

			// Request permission
			const permission = await Notification.requestPermission();
			if (permission !== 'granted') {
				toast.error('Push notifications permission denied');
				return false;
			}

			// Subscribe to push notifications
			this.subscription = await this.registration.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: this.urlBase64ToUint8Array(this.vapidPublicKey)
			});

			// Send subscription to server
			await this.sendSubscriptionToServer();

			toast.success('Push notifications enabled successfully!');
			console.log('[PushService] Subscription successful');
			return true;
		} catch (error) {
			console.error('[PushService] Subscription failed:', error);

			// Handle specific FCM/GCM errors
			const errorMessage = this.getErrorMessage(error);
			toast.error(errorMessage);
			return false;
		}
	}

	/**
	 * Unsubscribe from push notifications
	 */
	async unsubscribe(): Promise<boolean> {
		try {
			if (!this.subscription) {
				console.log('[PushService] No subscription to unsubscribe from');
				return true;
			}

			// Unsubscribe from push manager
			await this.subscription.unsubscribe();
			this.subscription = null;

			toast.success('Push notifications disabled');
			console.log('[PushService] Unsubscribed successfully');
			return true;
		} catch (error) {
			console.error('[PushService] Unsubscription failed:', error);
			toast.error('Failed to disable push notifications');
			return false;
		}
	}

	/**
	 * Get current subscription status
	 */
	async getStatus(): Promise<{
		supported: boolean;
		permission: NotificationPermission;
		subscribed: boolean;
	}> {
		const supported = this.isSupported();
		const permission = supported ? Notification.permission : 'denied';

		return {
			supported,
			permission,
			subscribed: this.isSubscribed()
		};
	}

	/**
	 * Check if currently subscribed
	 */
	isSubscribed(): boolean {
		return !!this.subscription;
	}

	/**
	 * Send subscription to server
	 */
	private async sendSubscriptionToServer(): Promise<void> {
		if (!this.subscription) {
			throw new Error('No subscription available');
		}

		const keys = this.subscription.getKey
			? {
					p256dh: this.subscription.getKey('p256dh'),
					auth: this.subscription.getKey('auth')
				}
			: { p256dh: null, auth: null };

		const subscriptionData: PushSubscriptionData = {
			endpoint: this.subscription.endpoint,
			p256dh_key: keys.p256dh ? this.arrayBufferToBase64(keys.p256dh) : '',
			auth_key: keys.auth ? this.arrayBufferToBase64(keys.auth) : '',
			user_agent: navigator.userAgent,
			platform: navigator.platform || 'unknown'
		};

		const response = await authenticatedFetch('/api/push/subscribe', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(subscriptionData)
		});

		if (!response.ok) {
			throw new Error(`Failed to send subscription to server: ${response.statusText}`);
		}

		console.log('[PushService] Subscription sent to server successfully');
	}

	/**
	 * Convert VAPID public key to Uint8Array
	 */
	private urlBase64ToUint8Array(base64String: string): Uint8Array {
		const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
		const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
		const rawData = window.atob(base64);
		const outputArray = new Uint8Array(rawData.length);
		for (let i = 0; i < rawData.length; ++i) {
			outputArray[i] = rawData.charCodeAt(i);
		}
		return outputArray;
	}

	/**
	 * Convert ArrayBuffer to base64
	 */
	private arrayBufferToBase64(buffer: ArrayBuffer): string {
		const bytes = new Uint8Array(buffer);
		let binary = '';
		for (let i = 0; i < bytes.byteLength; i++) {
			binary += String.fromCharCode(bytes[i]);
		}
		return window.btoa(binary);
	}

	/**
	 * Get user-friendly error message for push notification errors
	 */
	private getErrorMessage(error: unknown): string {
		if (error instanceof Error) {
			const message = error.message.toLowerCase();

			// FCM/GCM specific errors
			if (message.includes('registration failed')) {
				return 'Push service registration failed. Please check your internet connection.';
			}
			if (message.includes('vapid key')) {
				return 'Server configuration error. Please contact support.';
			}
			if (message.includes('not supported')) {
				return 'Push notifications are not supported on this device or browser.';
			}
			if (message.includes('permission')) {
				return 'Permission denied. Please enable notifications in browser settings.';
			}
			if (message.includes('quota exceeded')) {
				return 'Too many active subscriptions. Please try again later.';
			}
			if (message.includes('invalid vapid key') || message.includes('unauthorized')) {
				return 'Server authentication failed. Please contact support.';
			}
			if (message.includes('network') || message.includes('fetch')) {
				return 'Network error. Please check your connection and try again.';
			}

			// Generic error with specific message
			return `Push notification error: ${error.message}`;
		}

		return 'An unknown error occurred while setting up push notifications.';
	}

	/**
	 * Set up message listener for push notifications
	 */
	private setupMessageListener(): void {
		if (!('serviceWorker' in navigator)) return;

		navigator.serviceWorker.addEventListener('message', (event) => {
			if (event.data?.type === 'PUSH_RECEIVED') {
				// Add notification to store when push message is received
				notificationStore.addNotification({
					type: event.data.notificationType || 'broadcast',
					title: event.data.title || 'New Message',
					message: event.data.body || 'You have a new message',
					timestamp: new Date().toISOString(),
					read: false,
					data: event.data.data || {}
				});
			}
		});
	}

	/**
	 * Test notification (for development/testing)
	 */
	async testNotification(): Promise<void> {
		if (!this.registration) {
			toast.error('Service worker not ready');
			return;
		}

		try {
			await this.registration.showNotification('Night Owls Test', {
				body: 'Push notifications are working correctly!',
				icon: '/icons/icon-192x192.png',
				badge: '/icons/icon-192x192.png',
				tag: 'test',
				requireInteraction: false
			});
			console.log('[PushService] Test notification shown');
		} catch (error) {
			console.error('[PushService] Test notification failed:', error);
			toast.error('Failed to show test notification');
		}
	}
}

// Export singleton instance
export const pushNotificationService = new PushNotificationService();
