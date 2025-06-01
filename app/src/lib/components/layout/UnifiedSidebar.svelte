<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { contextualNavigation } from '$lib/stores/navigation';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { logout } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import type { ComponentProps, Snippet } from 'svelte';

	// Icons
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import ChevronsUpDown from '@lucide/svelte/icons/chevrons-up-down';
	import LogOut from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';

	// Props for customization
	let {
		ref = $bindable(null),
		mode = 'adaptive', // 'admin' | 'public' | 'adaptive'
		showSecondSidebar = false,
		listContent,
		title,
		searchTerm = $bindable(''),
		...restProps
	}: Omit<ComponentProps<typeof Sidebar.Root>, 'children'> & {
		mode?: 'admin' | 'public' | 'adaptive';
		showSecondSidebar?: boolean;
		listContent?: Snippet;
		title?: string;
		searchTerm?: string;
	} = $props();

	// Safe sidebar context access
	function getSidebarContext() {
		try {
			return useSidebar();
		} catch {
			return null;
		}
	}

	const navigation = $derived($contextualNavigation);

	// Determine which navigation items to show based on mode and user role
	const navItems = $derived.by(() => {
		if (mode === 'admin' || (mode === 'adaptive' && page.url.pathname.startsWith('/admin'))) {
			return navigation.admin;
		}
		return navigation.public;
	});

	// Determine if we should show the dual sidebar layout
	const isDualSidebar = $derived.by(() => {
		if (mode === 'admin') return true;
		if (mode === 'public') return false;
		// Adaptive mode: dual sidebar for admin routes
		return page.url.pathname.startsWith('/admin');
	});

	// Get user initials for avatar
	const userInitials = $derived.by(() => {
		if (!$currentUser?.name) return 'U';
		return $currentUser.name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	});

	// Get role-specific styling and icons
	const roleInfo = $derived.by(() => {
		const role = $currentUser?.role;
		switch (role) {
			case 'admin':
				return {
					label: 'Administrator',
					icon: ShieldIcon,
					color: 'text-red-600'
				};
			case 'owl':
				return {
					label: 'Night Owl',
					icon: StarIcon,
					color: 'text-yellow-600'
				};
			default:
				return {
					label: 'User',
					icon: UserIcon,
					color: 'text-gray-600'
				};
		}
	});

	// Handle logout
	function handleLogout() {
		toast.success('Logged out successfully');
		logout();
	}

	// Get home URL based on context
	const homeUrl = $derived.by(() => {
		if (mode === 'admin' || (mode === 'adaptive' && page.url.pathname.startsWith('/admin'))) {
			return '/admin';
		}
		return '/';
	});

	// Get app title based on context
	const appTitle = $derived.by(() => {
		if (mode === 'admin' || (mode === 'adaptive' && page.url.pathname.startsWith('/admin'))) {
			return { main: 'Mount Moreland Night Owls', sub: 'Admin' };
		}
		return { main: 'Mount Moreland Night Owls', sub: 'Community' };
	});
</script>

<!-- Shared logo component -->
{#snippet logoSection()}
	<div class="flex aspect-square size-8 items-center justify-center">
		<img src="/logo.png" alt="Mount Moreland Night Owls" class="h-6 w-6 object-contain" />
	</div>
{/snippet}

<!-- Shared header content -->
{#snippet headerContent()}
	<Sidebar.Header>
		<Sidebar.Menu>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton size="lg" class="md:h-8 md:p-0">
					{#snippet child({ props })}
						<a href={homeUrl} {...props}>
							{@render logoSection()}
							<div class="grid flex-1 text-left text-sm leading-tight">
								<span class="truncate font-semibold">{appTitle.main}</span>
								<span class="truncate text-xs">{appTitle.sub}</span>
							</div>
						</a>
					{/snippet}
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Header>
{/snippet}

<!-- Shared navigation content -->
{#snippet navigationContent()}
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupContent class="px-1.5 md:px-0">
				<Sidebar.Menu>
					{#each navItems as item (item.title)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton
								tooltipContentProps={{
									hidden: false
								}}
								onclick={() => goto(item.url)}
								isActive={page.url.pathname === item.url ||
									(item.url !== '/' && page.url.pathname.startsWith(item.url))}
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
{/snippet}

<!-- Shared user menu content -->
{#snippet userMenuContent()}
	<Sidebar.Footer>
		{#if $isAuthenticated}
			<Sidebar.Menu>
				<Sidebar.MenuItem>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Sidebar.MenuButton
									{...props}
									size="lg"
									class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground md:h-8 md:p-0"
								>
									<Avatar.Root class="h-8 w-8 rounded-lg">
										<Avatar.Fallback class="rounded-lg text-xs">
											{userInitials}
										</Avatar.Fallback>
									</Avatar.Root>
									<div class="grid flex-1 text-left text-sm leading-tight">
										<span class="truncate font-semibold">{$currentUser?.name || 'User'}</span>
										<span class="truncate text-xs">{$currentUser?.phone || ''}</span>
									</div>
									<ChevronsUpDown class="ml-auto size-4" />
								</Sidebar.MenuButton>
							{/snippet}
						</DropdownMenu.Trigger>

						<DropdownMenu.Content
							class="w-[var(--bits-dropdown-menu-anchor-width)] min-w-56 rounded-lg"
							side={getSidebarContext()?.isMobile ? 'bottom' : 'right'}
							align="end"
							sideOffset={4}
						>
							<DropdownMenu.Label class="p-0 font-normal">
								<div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
									<Avatar.Root class="h-8 w-8 rounded-lg">
										<Avatar.Fallback class="rounded-lg text-xs">
											{userInitials}
										</Avatar.Fallback>
									</Avatar.Root>
									<div class="grid flex-1 text-left text-sm leading-tight">
										<span class="truncate font-semibold">{$currentUser?.name || 'User'}</span>
										{#if $currentUser?.role}
											{@const IconComponent = roleInfo.icon}
											<div class="flex items-center gap-1 mt-1">
												<IconComponent class="h-3 w-3 {roleInfo.color}" />
												<span class="text-xs text-muted-foreground">
													{roleInfo.label}
												</span>
											</div>
										{/if}
									</div>
								</div>
							</DropdownMenu.Label>

							<DropdownMenu.Separator />

							<DropdownMenu.Group>
								<DropdownMenu.Item class="cursor-pointer">
									<SettingsIcon class="mr-2 h-4 w-4" />
									Settings
								</DropdownMenu.Item>
							</DropdownMenu.Group>

							<DropdownMenu.Separator />

							<DropdownMenu.Item
								class="cursor-pointer text-destructive focus:text-destructive"
								onclick={handleLogout}
							>
								<LogOut class="mr-2 h-4 w-4" />
								Log out
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Sidebar.MenuItem>
			</Sidebar.Menu>
		{:else}
			<!-- Login prompt for unauthenticated users -->
			<Sidebar.Menu>
				<Sidebar.MenuItem>
					<Sidebar.MenuButton onclick={() => goto('/login')} size="lg" class="md:h-8 md:p-0">
						<Avatar.Root class="h-8 w-8 rounded-lg">
							<Avatar.Fallback class="rounded-lg text-xs">
								<UserIcon class="h-4 w-4" />
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="grid flex-1 text-left text-sm leading-tight">
							<span class="truncate font-semibold">Sign In</span>
							<span class="truncate text-xs">Access your account</span>
						</div>
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			</Sidebar.Menu>
		{/if}
	</Sidebar.Footer>
{/snippet}

<!-- Secondary sidebar content -->
{#snippet secondarySidebarContent()}
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
{/snippet}

{#if isDualSidebar}
	<!-- Dual Sidebar Layout (Admin) -->
	<Sidebar.Root
		bind:ref
		collapsible="icon"
		class="overflow-hidden [&>[data-sidebar=sidebar]]:flex-row"
		{...restProps}
	>
		<!-- Primary Navigation Sidebar -->
		<Sidebar.Root collapsible="none" class="!w-[calc(var(--sidebar-width-icon)_+_1px)] border-r">
			{@render headerContent()}
			{@render navigationContent()}
			{@render userMenuContent()}
		</Sidebar.Root>

		<!-- Secondary Content Sidebar (Admin only) -->
		{#if showSecondSidebar}
			{@render secondarySidebarContent()}
		{/if}
	</Sidebar.Root>
{:else}
	<!-- Single Sidebar Layout (Public) -->
	<Sidebar.Root bind:ref collapsible="icon" class="border-r" {...restProps}>
		{@render headerContent()}
		{@render navigationContent()}
		{@render userMenuContent()}
	</Sidebar.Root>
{/if}
