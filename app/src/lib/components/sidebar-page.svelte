<script lang="ts">
	import UnifiedSidebar from '$lib/components/layout/UnifiedSidebar.svelte';
	import UnifiedHeader from '$lib/components/layout/UnifiedHeader.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';

	let {
		children,
		listContent,
		title,
		searchTerm = $bindable('')
	}: { children?: Snippet; listContent?: Snippet; title?: string; searchTerm?: string } = $props();

	// Check if we're in admin area for breadcrumbs
	const isAdminRoute = $derived(page.url.pathname.startsWith('/admin'));
</script>

{#snippet mainContent()}
	{#if children}
		{@render children()}
	{/if}
{/snippet}

<Sidebar.Provider style="--sidebar-width: 350px;">
	<UnifiedSidebar
		mode="admin"
		showSecondSidebar={!!listContent}
		{listContent}
		{title}
		bind:searchTerm
	/>
	<Sidebar.Inset>
		<!-- Header inside the inset to avoid overlap -->
		<UnifiedHeader showBreadcrumbs={isAdminRoute} showMobileMenu={false} />
		<main class="flex-1 overflow-auto">
			{@render mainContent()}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
