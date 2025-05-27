<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { page } from '$app/state';
	
	let { children } = $props();
	
	// Check if this page handles its own sidebar layout
	const hasOwnSidebarLayout = $derived(
		page.url.pathname === '/admin' || // Dashboard
		page.url.pathname.startsWith('/admin/shifts') || // Shifts section
		page.url.pathname.startsWith('/admin/broadcasts') || // Broadcasts section
		page.url.pathname.startsWith('/admin/schedules') || // Schedules section
		page.url.pathname.startsWith('/admin/reports') || // Reports section
		page.url.pathname.startsWith('/admin/users') || // Users section
		page.url.pathname.startsWith('/admin/emergency-contacts') // Emergency contacts section
	);
</script>

{#if hasOwnSidebarLayout}
	<!-- These pages handle their own sidebar layout -->
	{@render children?.()}
{:else}
	<!-- Fallback for any other admin pages -->
	<SidebarPage>
		{#snippet children()}
			{@render children?.()}
		{/snippet}
	</SidebarPage>
{/if}
