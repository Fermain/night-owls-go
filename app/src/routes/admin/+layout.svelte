<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';

	let { children } = $props();
</script>

{@render children?.()}
