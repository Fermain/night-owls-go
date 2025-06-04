<script lang="ts">
	import UnifiedSidebar from '$lib/components/layout/UnifiedSidebar.svelte';
	import UnifiedHeader from '$lib/components/layout/UnifiedHeader.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import type { Snippet } from 'svelte';

	let {
		children,
		listContent,
		title,
		searchTerm = $bindable('')
	}: { children?: Snippet; listContent?: Snippet; title?: string; searchTerm?: string } = $props();
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
		<UnifiedHeader />
		<main class="flex-1 overflow-auto">
			{@render mainContent()}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
