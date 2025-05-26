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
	
	// Import unified header and mobile navigation
	import UnifiedHeader from '$lib/components/layout/UnifiedHeader.svelte';
	import MobileNav from '$lib/components/navigation/MobileNav.svelte';

	let { children } = $props();

	onMount(async () => {
		// Set dark mode preference
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		if (prefersDark) {
			document.documentElement.classList.add('dark');
		}

		// Register service worker
		try {
			const { serviceWorkerService } = await import('$lib/services/serviceWorkerService');
			const registered = await serviceWorkerService.register();
			
			if (registered) {
				console.log('🔧 Service worker registered successfully');
				
				// Listen for service worker messages
				navigator.serviceWorker.addEventListener('message', (event) => {
					console.log('📨 Message from service worker:', event.data);
					
					if (event.data.type === 'SW_ACTIVATED') {
						console.log('🎉 ' + event.data.message);
					}
				});
			} else {
				console.log('⚠️ Service worker registration failed');
			}
		} catch (error) {
			console.error('Service worker registration error:', error);
		}
	});
</script>

<QueryClientProvider client={queryClient}>
	<div class="min-h-screen bg-background text-foreground">
		<!-- Unified Header for all pages -->
		<UnifiedHeader 
			showBreadcrumbs={isAdminRoute}
			showMobileMenu={!isAdminRoute}
		/>

		{#if isAdminRoute}
			<!-- Admin layout (existing admin pages) -->
			{@render children()}
		{:else}
			<!-- End-user layout (mobile-first) -->
			<main class="pb-16 md:pb-0">
				{@render children()}
			</main>

			<!-- Mobile bottom navigation (only for end-users) -->
			<MobileNav />
		{/if}
	</div>

	<Toaster position="top-center" />
</QueryClientProvider>
