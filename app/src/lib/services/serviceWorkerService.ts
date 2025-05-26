import { browser } from '$app/environment';

class ServiceWorkerService {
	private registration: ServiceWorkerRegistration | null = null;
	private isRegistered = false;

	/**
	 * Register the service worker
	 */
	async register(): Promise<boolean> {
		if (!browser) {
			console.log('Not in browser, skipping service worker registration');
			return false;
		}

		if (!('serviceWorker' in navigator)) {
			console.warn('Service workers not supported');
			return false;
		}

		try {
			// Try to register the service worker from different possible locations
			const possiblePaths = ['/sw.js', '/service-worker.js'];
			
			for (const swPath of possiblePaths) {
				try {
					console.log(`Attempting to register service worker at: ${swPath}`);
					
					this.registration = await navigator.serviceWorker.register(swPath, {
						scope: '/'
					});

					console.log('✅ Service worker registered successfully:', this.registration);
					this.isRegistered = true;

					// Listen for updates
					this.registration.addEventListener('updatefound', () => {
						console.log('🔄 Service worker update found');
						const newWorker = this.registration?.installing;
						if (newWorker) {
							newWorker.addEventListener('statechange', () => {
								if (newWorker.state === 'installed') {
									console.log('🆕 New service worker installed');
									// Optionally notify user about update
								}
							});
						}
					});

					return true;
				} catch (error) {
					console.log(`Failed to register SW at ${swPath}:`, error);
					continue;
				}
			}

			throw new Error('No valid service worker found at any path');
		} catch (error) {
			console.error('❌ Service worker registration failed:', error);
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