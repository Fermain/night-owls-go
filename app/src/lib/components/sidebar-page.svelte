<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
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
	<AppSidebar {listContent} bind:searchTerm />
	<Sidebar.Inset>
		{@render mainContent()}
	</Sidebar.Inset>
</Sidebar.Provider>
