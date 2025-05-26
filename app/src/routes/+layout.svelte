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
		{#if isAdminRoute}
			<!-- Admin layout (existing admin pages) -->
			{@render children()}
		{:else}
			<!-- End-user layout (mobile-first) -->
			{#await import('$lib/components/layout/PublicHeader.svelte') then { default: PublicHeader }}
				<PublicHeader />
			{/await}
			<main class="pb-16 md:pb-0">
				{@render children()}
			</main>

			<!-- Mobile bottom navigation (only for end-users) -->
			<nav class="fixed bottom-0 left-0 right-0 bg-background border-t border-border md:hidden">
				<div class="flex items-center justify-around h-16">
					<a
						href="/"
						class="flex flex-col items-center gap-1 px-2 py-1 text-xs
						{page.url.pathname === '/' ? 'text-primary' : 'text-muted-foreground'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
							></path>
						</svg>
						<span>Home</span>
					</a>

					<a
						href="/shifts"
						class="flex flex-col items-center gap-1 px-2 py-1 text-xs
						{page.url.pathname.startsWith('/shifts') ? 'text-primary' : 'text-muted-foreground'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
							></path>
						</svg>
						<span>Shifts</span>
					</a>

					<a
						href="/report"
						class="flex flex-col items-center gap-1 px-2 py-1 text-xs
						{page.url.pathname.startsWith('/report') ? 'text-primary' : 'text-muted-foreground'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z"
							></path>
						</svg>
						<span>Report</span>
					</a>

					<a
						href="/broadcasts"
						class="flex flex-col items-center gap-1 px-2 py-1 text-xs
						{page.url.pathname.startsWith('/broadcasts') && !page.url.pathname.startsWith('/admin')
							? 'text-primary'
							: 'text-muted-foreground'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
							></path>
						</svg>
						<span>Messages</span>
					</a>

					<a
						href="/bookings"
						class="flex flex-col items-center gap-1 px-2 py-1 text-xs
						{page.url.pathname.startsWith('/bookings') ? 'text-primary' : 'text-muted-foreground'}"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
							></path>
						</svg>
						<span>Bookings</span>
					</a>
				</div>
			</nav>
		{/if}
	</div>

	<Toaster position="top-center" />
</QueryClientProvider>
