<script lang="ts">
	import '../app.css'; // Import global styles (Tailwind)
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Toaster } from 'svelte-sonner';
	import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
	// SvelteKit environment module

	// For client-side only SPA as per SvelteKit SPA guide

	let queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 5 * 60 * 1000, // 5 minutes
				gcTime: 10 * 60 * 1000 // 10 minutes
			}
		}
	});

	// Check if we're in admin area
	const isAdminRoute = $derived(page.url.pathname.startsWith('/admin'));

	// Initialize notification service on app startup
	onMount(() => {
		notificationStore.init();
		// Only fetch notifications if user is authenticated
		if ($userSession.isAuthenticated) {
			notificationStore.fetchNotifications();
		}
	});

	// Import unified header and mobile navigation
	import UnifiedHeader from '$lib/components/layout/UnifiedHeader.svelte';
	import MobileNav from '$lib/components/navigation/MobileNav.svelte';
	import OfflineIndicator from '$lib/components/ui/offline/OfflineIndicator.svelte';
	import { notificationStore } from '$lib/services/notificationService';
	import { userSession } from '$lib/stores/authStore';
	import { pwaInstallPrompt } from '$lib/stores/onboardingStore';

	let { children } = $props();

	onMount(async () => {
		// Set dark mode preference
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		if (prefersDark) {
			document.documentElement.classList.add('dark');
		}

		// Listen for PWA install prompt
		window.addEventListener('beforeinstallprompt', (event) => {
			// Prevent the default prompt
			event.preventDefault();
			// Store the event for later use
			pwaInstallPrompt.set(event as any);
			console.log('PWA install prompt captured');
		});

		// Listen for app installed event
		window.addEventListener('appinstalled', () => {
			console.log('PWA was installed');
			pwaInstallPrompt.set(null);
		});
	});
</script>

<QueryClientProvider client={queryClient}>
	<div class="min-h-screen bg-background text-foreground">
		{#if isAdminRoute}
			<!-- Admin layout uses existing sidebar system -->
			{@render children()}
		{:else}
			<!-- Public layout with header + mobile nav -->
			<div class="flex flex-col min-h-screen">
				<UnifiedHeader showBreadcrumbs={false} />
				<!-- Main content area that fills remaining height -->
				<main class="flex-1 pb-16 md:pb-0 overflow-auto flex">
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
