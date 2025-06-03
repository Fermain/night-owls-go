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

class PushNotificationService {
	private vapidPublicKey: string | null = null;
	private registration: ServiceWorkerRegistration | null = null;
	private subscription: PushSubscription | null = null;

	/**
	 * Initialize the push notification service
	 */
	async initialize(): Promise<boolean> {
		try {
			// Check if service workers and push notifications are supported
			if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
				console.warn('Push notifications not supported');
				return false;
			}

			// Check if there's already a registration
			const existingRegistration = await navigator.serviceWorker.getRegistration();

			if (existingRegistration) {
				this.registration = existingRegistration;
			} else {
				// Add timeout to prevent hanging
				const timeoutPromise = new Promise<ServiceWorkerRegistration>((_, reject) => {
					setTimeout(() => reject(new Error('Service worker ready timeout')), 10000);
				});

				try {
					this.registration = await Promise.race([navigator.serviceWorker.ready, timeoutPromise]);
				} catch (_error) {
					console.warn('Service worker ready timeout, attempting manual registration');
					// Fallback: try to register manually
					try {
						this.registration = await navigator.serviceWorker.register('/sw.js');
						await this.registration.update();
					} catch (regError) {
						console.error('Manual service worker registration failed:', regError);
						return false;
					}
				}
			}

			if (!this.registration) {
				console.error('No service worker registration available');
				return false;
			}

			// Get VAPID public key from server
			await this.fetchVAPIDPublicKey();

			// Check for existing subscription
			this.subscription = await this.registration.pushManager.getSubscription();

			return true;
		} catch (error) {
			console.error('Failed to initialize push notification service:', error);
			return false;
		}
	}

	/**
	 * Fetch VAPID public key from server
	 */
	private async fetchVAPIDPublicKey(): Promise<void> {
		try {
			const response = await fetch('/api/push/vapid-public');
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}

			const data: VAPIDKeyResponse = await response.json();
			this.vapidPublicKey = data.key;
		} catch (error) {
			console.error('Failed to fetch VAPID public key:', error);
			throw error;
		}
	}

	/**
	 * Request permission and subscribe to push notifications
	 */
	async subscribe(): Promise<boolean> {
		try {
			if (!this.registration || !this.vapidPublicKey) {
				console.error('Service not initialized properly');
				return false;
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

			// Set up listener for push messages
			this.setupPushMessageListener();

			toast.success('Push notifications enabled successfully!');
			return true;
		} catch (error) {
			console.error('Failed to subscribe to push notifications:', error);
			toast.error('Failed to enable push notifications');
			return false;
		}
	}

	/**
	 * Unsubscribe from push notifications
	 */
	async unsubscribe(): Promise<boolean> {
		try {
			if (!this.subscription) {
				return true; // Already unsubscribed
			}

			// Unsubscribe from browser
			await this.subscription.unsubscribe();

			// Remove subscription from server
			await this.removeSubscriptionFromServer();

			this.subscription = null;
			toast.success('Push notifications disabled');
			return true;
		} catch (error) {
			console.error('Failed to unsubscribe from push notifications:', error);
			toast.error('Failed to disable push notifications');
			return false;
		}
	}

	/**
	 * Check if subscribed
	 */
	isSubscribed(): boolean {
		return this.subscription !== null;
	}

	/**
	 * Get current subscription status and permission
	 */
	async getStatus(): Promise<{
		supported: boolean;
		permission: NotificationPermission;
		subscribed: boolean;
	}> {
		const supported = 'serviceWorker' in navigator && 'PushManager' in window;
		const permission = supported ? Notification.permission : 'denied';

		return {
			supported,
			permission,
			subscribed: this.isSubscribed()
		};
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
	}

	/**
	 * Remove subscription from server
	 */
	private async removeSubscriptionFromServer(): Promise<void> {
		if (!this.subscription) {
			return;
		}

		const endpoint = encodeURIComponent(this.subscription.endpoint);
		const response = await authenticatedFetch(`/api/push/subscribe/${endpoint}`, {
			method: 'DELETE'
		});

		if (!response.ok && response.status !== 404) {
			throw new Error(`Failed to remove subscription from server: ${response.statusText}`);
		}
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
	 * Set up listener for push messages to integrate with notification store
	 */
	private setupPushMessageListener(): void {
		if (!('serviceWorker' in navigator)) return;

		// Listen for messages from service worker
		navigator.serviceWorker.addEventListener('message', (event) => {
			if (event.data.type === 'PUSH_RECEIVED') {
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
	 * Test push notification (for development/testing)
	 */
	async testNotification(): Promise<void> {
		if (!this.registration) {
			toast.error('Service worker not ready');
			return;
		}

		try {
			await this.registration.showNotification('Test Notification', {
				body: 'This is a test notification from Night Owls',
				icon: '/icons/icon-192x192.png',
				badge: '/icons/icon-192x192.png',
				tag: 'test',
				requireInteraction: false
			});
		} catch (error) {
			console.error('Failed to show test notification:', error);
			toast.error('Failed to show test notification');
		}
	}
}

// Export singleton instance
export const pushNotificationService = new PushNotificationService();
export default pushNotificationService;
