import { browser } from '$app/environment';

class ServiceWorkerService {
	private registration: ServiceWorkerRegistration | null = null;
	private isRegistered = false;

	/**
	 * Check if service worker is registered (Vite PWA handles registration automatically)
	 */
	async register(): Promise<boolean> {
		if (!browser) {
			console.log('Not in browser, skipping service worker check');
			return false;
		}

		if (!('serviceWorker' in navigator)) {
			console.warn('Service workers not supported');
			return false;
		}

		try {
			// With Vite PWA, the service worker should already be registered
			// We just need to wait for it to be ready
			this.registration = await navigator.serviceWorker.ready;

			if (this.registration) {
				console.log('✅ Service worker found (managed by Vite PWA):', this.registration);
				this.isRegistered = true;
				return true;
			}

			return false;
		} catch (error) {
			console.error('❌ Service worker check failed:', error);
			return false;
		}
	}

	/**
	 * Unregister the service worker
	 */
	async unregister(): Promise<boolean> {
		if (!browser || !('serviceWorker' in navigator)) {
			return false;
		}

		try {
			const registrations = await navigator.serviceWorker.getRegistrations();

			for (const registration of registrations) {
				await registration.unregister();
			}

			console.log('✅ Service worker unregistered');
			this.registration = null;
			this.isRegistered = false;
			return true;
		} catch (error) {
			console.error('❌ Service worker unregistration failed:', error);
			return false;
		}
	}

	/**
	 * Check if service worker is registered
	 */
	getRegistration(): ServiceWorkerRegistration | null {
		return this.registration;
	}

	/**
	 * Check if registered
	 */
	isServiceWorkerRegistered(): boolean {
		return this.isRegistered;
	}

	/**
	 * Get service worker status
	 */
	async getStatus(): Promise<{
		supported: boolean;
		registered: boolean;
		active: boolean;
		registration: ServiceWorkerRegistration | null;
	}> {
		const supported = browser && 'serviceWorker' in navigator;

		if (!supported) {
			return {
				supported: false,
				registered: false,
				active: false,
				registration: null
			};
		}

		// Get current registration
		const registration = await navigator.serviceWorker.getRegistration();
		const isActive = !!registration?.active;

		return {
			supported: true,
			registered: !!registration,
			active: isActive,
			registration: registration || null
		};
	}
}

// Export singleton instance
export const serviceWorkerService = new ServiceWorkerService();
export default serviceWorkerService;
