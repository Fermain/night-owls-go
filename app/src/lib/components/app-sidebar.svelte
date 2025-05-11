<script lang="ts">
	import NavUser from '$lib/components/nav-user.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	// import { useSidebar } from '$lib/components/ui/sidebar/index.js'; // useSidebar was called but not used
	import type { ComponentProps, Snippet } from 'svelte';
	import { page } from '$app/state';

	import { navigation } from '$lib/stores/navigation';
	import { goto } from '$app/navigation';

	let {
		ref = $bindable(null),
		// children, // children prop was defined but not used by AppSidebar directly, nor passed by SidebarPage
		listContent,
		title,
		searchTerm = $bindable(''),
		...restProps
	}: Omit<ComponentProps<typeof Sidebar.Root>, 'children'> & {
		// Omit children from Sidebar.Root props if it conflicts
		listContent?: Snippet;
		// children?: Snippet; // Removed from explicit type as well
		title?: string;
		searchTerm?: string;
	} = $props();

	// Placeholder user data, ideally this would come from another store or context
	const user = {
		// Define user data here or import from a store
		name: 'owl',
		phone: '+2700000000',
		avatar: '' // Set to empty string to resolve 404 and type error
	};

	// const sidebar = useSidebar(); // Removed as sidebar variable was not used
</script>

<Sidebar.Root
	bind:ref
	collapsible="icon"
	class="overflow-hidden [&>[data-sidebar=sidebar]]:flex-row"
	{...restProps}
>
	<Sidebar.Root collapsible="none" class="!w-[calc(var(--sidebar-width-icon)_+_1px)] border-r">
		<Sidebar.Header>
			<Sidebar.Menu>
				<Sidebar.MenuItem>
					<Sidebar.MenuButton size="lg" class="md:h-8 md:p-0">
						{#snippet child({ props })}
							<a href="/admin" {...props}>
								<div
									class="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg"
								>
									<img src="/logo.png" alt="Night Owls" class="w-full h-full object-contain" />
								</div>
								<div class="grid flex-1 text-left text-sm leading-tight">
									<span class="truncate font-semibold">Night Owls</span>
									<span class="truncate text-xs">Admin</span>
								</div>
							</a>
						{/snippet}
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			</Sidebar.Menu>
		</Sidebar.Header>
		<Sidebar.Content>
			<Sidebar.Group>
				<Sidebar.GroupContent class="px-1.5 md:px-0">
					<Sidebar.Menu>
						{#each $navigation as item (item.title)}
							<Sidebar.MenuItem>
								<Sidebar.MenuButton
									tooltipContentProps={{
										hidden: false
									}}
									onclick={() => goto(item.url)}
									isActive={page.url.pathname === item.url}
									class="px-2.5 md:px-2"
								>
									{#snippet tooltipContent()}
										{item.title}
									{/snippet}
									<item.icon />
									<span>{item.title}</span>
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
						{/each}
					</Sidebar.Menu>
				</Sidebar.GroupContent>
			</Sidebar.Group>
		</Sidebar.Content>
		<Sidebar.Footer>
			<!-- // This will be a persistant svelte store -->
			<NavUser {user} />
		</Sidebar.Footer>
	</Sidebar.Root>

	<!-- This is the second sidebar -->
	<!-- We disable collapsible and let it fill remaining space -->
	<Sidebar.Root collapsible="none" class="hidden flex-1 md:flex">
		<Sidebar.Header class="gap-3.5 border-b p-4">
			{#if title}
				<div class="flex w-full items-center justify-between">
					<div class="text-foreground text-base font-medium">
						{title}
					</div>
				</div>
			{/if}
			<Sidebar.Input placeholder="Type to search..." bind:value={searchTerm} />
		</Sidebar.Header>
		<Sidebar.Content>
			<Sidebar.Group class="p-0">
				<Sidebar.GroupContent>
					{#if listContent}
						{@render listContent()}
					{/if}
				</Sidebar.GroupContent>
			</Sidebar.Group>
		</Sidebar.Content>
	</Sidebar.Root>
</Sidebar.Root>
