<script lang="ts">
	import SidebarPage from "$lib/components/sidebar-page.svelte";

	let { children } = $props();
</script>

<SidebarPage>
	{@render children?.()}
</SidebarPage>
