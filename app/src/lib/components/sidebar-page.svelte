<script lang="ts">
	import AppSidebar from "$lib/components/app-sidebar.svelte";
	import * as Breadcrumb from "$lib/components/ui/breadcrumb/index.js";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";

	let { children } = $props();

	const breadcrumbs = $derived(location.pathname.split("/").slice(1).map((crumb, index) => ({
		label: crumb.replace(/-/g, " ").replace(/\b\w/g, char => char.toUpperCase()),
		href: `${index === 0 ? "" : location.pathname.split(crumb).slice(0, -1).join("/")}${crumb}`
	})));
</script>

<Sidebar.Provider style="--sidebar-width: 350px;">
	<AppSidebar />
	<Sidebar.Inset>
		<header class="bg-background sticky top-0 flex shrink-0 items-center gap-2 border-b p-4">
			<Sidebar.Trigger class="-ml-1" />
			<Separator orientation="vertical" class="mr-2 h-4" />
			<Breadcrumb.Root>
				<Breadcrumb.List>
					{#each breadcrumbs as crumb, i}
						<Breadcrumb.Item class="hidden md:block">
							<Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
						</Breadcrumb.Item>
						{#if i < breadcrumbs.length - 1}
							<Breadcrumb.Separator class="hidden md:block" />
						{/if}
					{/each}
				</Breadcrumb.List>
			</Breadcrumb.Root>
		</header>
		{@render children?.()}
	</Sidebar.Inset>
</Sidebar.Provider>
