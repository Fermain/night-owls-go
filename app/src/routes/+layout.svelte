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
	import { notificationStore } from '$lib/services/notificationService';
	import { userSession } from '$lib/stores/authStore';
	import { pwaInstallPrompt } from '$lib/stores/onboardingStore';

	let { children } = $props();

	// Create QueryClient
	let queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 5 * 60 * 1000, // 5 minutes
				gcTime: 10 * 60 * 1000 // 10 minutes
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
	let storesInitialized = $state(false);
	let currentUserSession = $state({ isAuthenticated: false });

	// Initialize everything after component is mounted to avoid ALL lifecycle errors
	onMount(() => {
		// Initialize notification service
		notificationStore.init();

		// Subscribe to user session changes
		const userSessionUnsubscribe = userSession.subscribe((session) => {
			currentUserSession = session;

			// Only fetch notifications if user is authenticated
			if (session.isAuthenticated) {
				notificationStore.fetchNotifications();
			}
		});

		// Apply theme based on store
		const themeUnsubscribe = actualTheme.subscribe((theme) => {
			themeActions.applyTheme(theme);
		});

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
		storesInitialized = true;

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
