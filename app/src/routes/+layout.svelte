<script lang="ts">
	import '../app.css'; // Import global styles (Tailwind)
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Toaster } from 'svelte-sonner';
	import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
	import { actualTheme, themeActions } from '$lib/stores/themeStore';
	import UnifiedHeader from '$lib/components/layout/UnifiedHeader.svelte';
	import MobileNav from '$lib/components/navigation/MobileNav.svelte';
	import OfflineIndicator from '$lib/components/ui/offline/OfflineIndicator.svelte';
	import PersistentTimeFilter from '$lib/components/layout/PersistentTimeFilter.svelte';
	import { notificationStore } from '$lib/services/notificationService';
	import { userSession } from '$lib/stores/authStore';
	import { pwaInstallPrompt } from '$lib/stores/onboardingStore';
	import { pushNotificationService } from '$lib/services/pushNotificationService';

	// Initialize background sync for offline forms
	import '$lib/utils/backgroundSync';

	let { children } = $props();

	// Create QueryClient with smart polling for community watch use case
	let queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 5 * 60 * 1000, // 5 minutes
				gcTime: 10 * 60 * 1000, // 10 minutes
				refetchOnWindowFocus: true,
				refetchOnReconnect: true,
				// For critical data (reports, emergency contacts)
				retry: 3,
				retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000)
			}
		}
	});

	// Check if we're in admin area (defensive check for SSR/hydration)
	// Use a getter function to avoid accessing page during initialization
	const isAdminRoute = $derived.by(() => {
		// During SSR or before hydration, page might not be available
		// Use optional chaining and provide a safe default
		return page?.url?.pathname?.startsWith('/admin') ?? false;
	});

	// State to track if stores are initialized
	let _storesInitialized = $state(false);
	let _currentUserSession = $state({ isAuthenticated: false });

	// Initialize everything after component is mounted to avoid ALL lifecycle errors
	onMount(() => {
		// Initialize notification service
		notificationStore.init();

		// Subscribe to user session changes
		const userSessionUnsubscribe = userSession.subscribe(async (session) => {
			_currentUserSession = session;

			// Only fetch notifications if user is authenticated
			if (session.isAuthenticated) {
				notificationStore.fetchNotifications();

				// Initialize push notification service for authenticated users
				try {
					const initialized = await pushNotificationService.initialize();
					if (initialized) {
						console.log('[App] Push notification service initialized successfully');
					} else {
						console.warn('[App] Push notification service failed to initialize');
					}
				} catch (error) {
					console.error('[App] Error initializing push notification service:', error);
				}
			}
		});

		// Apply theme based on store
		const themeUnsubscribe = actualTheme.subscribe((theme) => {
			themeActions.applyTheme(theme);
		});

		// Service Worker Navigation Handling (SvelteKit native approach)
		if ('serviceWorker' in navigator) {
			navigator.serviceWorker.addEventListener('message', (event) => {
				if (event.data?.type === 'NAVIGATE') {
					// Use SvelteKit's programmatic navigation
					import('$app/navigation').then(({ goto }) => {
						goto(event.data.url);
					});
				} else if (event.data?.type === 'PUSH_RECEIVED') {
					// Handle push notifications received while app is open
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

		// Listen for PWA install prompt
		window.addEventListener('beforeinstallprompt', (event) => {
			// Prevent the default prompt
			event.preventDefault();
			// Store the event for later use (cast to BeforeInstallPromptEvent)
			pwaInstallPrompt.set(
				event as Event & {
					prompt: () => Promise<void>;
					userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
				}
			);
		});

		// Listen for app installed event
		window.addEventListener('appinstalled', () => {
			console.log('PWA was installed');
			pwaInstallPrompt.set(null);
		});

		// Mark stores as initialized
		_storesInitialized = true;

		// Return cleanup function
		return () => {
			userSessionUnsubscribe();
			themeUnsubscribe();
		};
	});
</script>

<QueryClientProvider client={queryClient}>
	<div class="min-h-screen bg-background text-foreground">
		{#if isAdminRoute}
			<!-- Admin layout with existing sidebar system -->
			{@render children()}
		{:else}
			<!-- Public layout with header + mobile nav -->
			<div class="flex flex-col min-h-screen">
				<UnifiedHeader />
				<PersistentTimeFilter />
				<!-- Main content area that fills remaining height -->
				<main class="flex-1 overflow-auto flex">
					{@render children()}
				</main>
			</div>
			<MobileNav />
			<!-- Offline status indicator for public pages -->
			<OfflineIndicator />
		{/if}
	</div>

	<Toaster position="top-center" />
</QueryClientProvider>
