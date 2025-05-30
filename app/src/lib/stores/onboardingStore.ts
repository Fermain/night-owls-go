import { persisted } from 'svelte-persisted-store';
import { writable } from 'svelte/store';

// Interface for PWA install prompt event
interface BeforeInstallPromptEvent extends Event {
	prompt(): Promise<void>;
	userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

export interface OnboardingState {
	isCompleted: boolean;
	hasCompletedPermissions: boolean;
	hasCompletedPWAPrompt: boolean;
	locationPermission: 'granted' | 'denied' | 'prompt' | 'unknown';
	notificationPermission: 'granted' | 'denied' | 'default' | 'unknown';
	pwaInstalled: boolean;
	pwaInstallPromptShown: boolean;
	lastOnboardingVersion: string;
}

const initialOnboardingState: OnboardingState = {
	isCompleted: false,
	hasCompletedPermissions: false,
	hasCompletedPWAPrompt: false,
	locationPermission: 'unknown',
	notificationPermission: 'unknown',
	pwaInstalled: false,
	pwaInstallPromptShown: false,
	lastOnboardingVersion: '1.0.0'
};

// Persisted store for onboarding state
export const onboardingState = persisted<OnboardingState>('onboarding-state', initialOnboardingState);

// Runtime store for PWA install prompt event
export const pwaInstallPrompt = writable<BeforeInstallPromptEvent | null>(null);

// Helper functions for onboarding management
export const onboardingActions = {
	// Mark onboarding as completed
	completeOnboarding() {
		onboardingState.update(state => ({
			...state,
			isCompleted: true,
			hasCompletedPermissions: true,
			hasCompletedPWAPrompt: true
		}));
	},

	// Mark permissions step as completed
	completePermissions() {
		onboardingState.update(state => ({
			...state,
			hasCompletedPermissions: true
		}));
	},

	// Mark PWA prompt as completed
	completePWAPrompt() {
		onboardingState.update(state => ({
			...state,
			hasCompletedPWAPrompt: true,
			pwaInstallPromptShown: true
		}));
	},

	// Update location permission status
	updateLocationPermission(permission: OnboardingState['locationPermission']) {
		onboardingState.update(state => ({
			...state,
			locationPermission: permission
		}));
	},

	// Update notification permission status
	updateNotificationPermission(permission: OnboardingState['notificationPermission']) {
		onboardingState.update(state => ({
			...state,
			notificationPermission: permission
		}));
	},

	// Mark PWA as installed
	markPWAInstalled() {
		onboardingState.update(state => ({
			...state,
			pwaInstalled: true
		}));
	},

	// Reset onboarding (for testing or new versions)
	resetOnboarding() {
		onboardingState.set(initialOnboardingState);
	},

	// Check if user needs onboarding
	needsOnboarding(currentState: OnboardingState): boolean {
		return !currentState.isCompleted || 
			   !currentState.hasCompletedPermissions || 
			   !currentState.hasCompletedPWAPrompt;
	},

	// Get onboarding progress percentage
	getProgress(currentState: OnboardingState): number {
		let completed = 0;
		const total = 3; // permissions, PWA, completion

		if (currentState.hasCompletedPermissions) completed++;
		if (currentState.hasCompletedPWAPrompt) completed++;
		if (currentState.isCompleted) completed++;

		return Math.round((completed / total) * 100);
	}
};

// Permission checking utilities
export const permissionUtils = {
	// Check current location permission status
	async checkLocationPermission(): Promise<OnboardingState['locationPermission']> {
		if (typeof navigator === 'undefined' || !navigator.permissions) {
			return 'unknown';
		}

		try {
			const result = await navigator.permissions.query({ name: 'geolocation' });
			return result.state as OnboardingState['locationPermission'];
		} catch (error) {
			console.warn('Could not check location permission:', error);
			return 'unknown';
		}
	},

	// Check current notification permission status
	checkNotificationPermission(): OnboardingState['notificationPermission'] {
		if (typeof Notification === 'undefined') {
			return 'unknown';
		}
		return Notification.permission as OnboardingState['notificationPermission'];
	},

	// Request location permission
	async requestLocationPermission(): Promise<OnboardingState['locationPermission']> {
		if (typeof navigator === 'undefined' || !navigator.geolocation) {
			return 'unknown';
		}

		return new Promise((resolve) => {
			navigator.geolocation.getCurrentPosition(
				() => {
					onboardingActions.updateLocationPermission('granted');
					resolve('granted');
				},
				(error) => {
					const permission = error.code === error.PERMISSION_DENIED ? 'denied' : 'unknown';
					onboardingActions.updateLocationPermission(permission);
					resolve(permission);
				},
				{ timeout: 10000 }
			);
		});
	},

	// Request notification permission
	async requestNotificationPermission(): Promise<OnboardingState['notificationPermission']> {
		if (typeof Notification === 'undefined') {
			return 'unknown';
		}

		try {
			const permission = await Notification.requestPermission();
			onboardingActions.updateNotificationPermission(permission);
			return permission;
		} catch (error) {
			console.warn('Could not request notification permission:', error);
			return 'unknown';
		}
	}
};

// PWA utilities
export const pwaUtils = {
	// Check if app is running as PWA
	isPWA(): boolean {
		if (typeof window === 'undefined') return false;
		
		return window.matchMedia('(display-mode: standalone)').matches ||
			   (window.navigator as { standalone?: boolean }).standalone === true ||
			   document.referrer.includes('android-app://');
	},

	// Check if PWA can be installed
	canInstallPWA(): boolean {
		return typeof window !== 'undefined' && 
			   'serviceWorker' in navigator && 
			   !pwaUtils.isPWA();
	},

	// Show PWA install prompt
	async showInstallPrompt(promptEvent: BeforeInstallPromptEvent | null): Promise<boolean> {
		if (!promptEvent || !('prompt' in promptEvent)) {
			return false;
		}

		try {
			// Show the install prompt
			await promptEvent.prompt();
			
			// Wait for user response
			const choiceResult = await promptEvent.userChoice;
			
			if (choiceResult.outcome === 'accepted') {
				onboardingActions.markPWAInstalled();
				return true;
			}
			
			return false;
		} catch (error) {
			console.warn('Error showing PWA install prompt:', error);
			return false;
		}
	}
}; 