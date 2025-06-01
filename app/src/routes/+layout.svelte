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
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { notificationStore } from '$lib/services/notificationService';
	import { userSession } from '$lib/stores/authStore';
	import { pwaInstallPrompt } from '$lib/stores/onboardingStore';

	let { children } = $props();

	let queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 5 * 60 * 1000, // 5 minutes
				gcTime: 10 * 60 * 1000 // 10 minutes
			}
		}
	});

	// Check if we're in admin area (defensive check for SSR/hydration)
	const isAdminRoute = $derived(page?.url?.pathname?.startsWith('/admin') ?? false);

	// Initialize notification service and theme on app startup
	onMount(() => {
		// Initialize notification service
		notificationStore.init();

		// Only fetch notifications if user is authenticated
		if ($userSession.isAuthenticated) {
			notificationStore.fetchNotifications();
		}

		// Apply theme based on store
		const unsubscribe = actualTheme.subscribe((theme) => {
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

		return unsubscribe;
	});
</script>

<QueryClientProvider client={queryClient}>
	<Sidebar.Provider>
		<div class="min-h-screen bg-background text-foreground">
			{#if isAdminRoute}
				<!-- Admin layout needs Sidebar context for UnifiedHeader Sidebar.Trigger -->
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
	</Sidebar.Provider>

	<Toaster position="top-center" />
</QueryClientProvider>
