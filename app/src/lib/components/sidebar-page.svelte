<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';

	let { children, listContent }: { children?: Snippet; listContent?: Snippet } = $props();

	const breadcrumbs = $derived(
		page.url.pathname
			.split('/')
			.slice(1)
			.map((crumb, index) => ({
				label: crumb.replace(/-/g, ' ').replace(/ \w/g, (char) => char.toUpperCase()),
				href: `${index === 0 ? '/' : '/' + page.url.pathname.split(crumb)[0]}${crumb}`
			}))
	);
</script>

{#snippet mainContentWithHeader()}
	<header class="bg-background sticky top-0 z-10 flex shrink-0 items-center gap-2 border-b p-4">
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
	{#if children}
		{@render children()}
	{/if}
{/snippet}

<Sidebar.Provider style="--sidebar-width: 350px;">
	<AppSidebar listContent={listContent} />
	<Sidebar.Inset>
		{@render mainContentWithHeader()}
	</Sidebar.Inset>
</Sidebar.Provider>
