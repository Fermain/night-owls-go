<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';
	import { Separator } from '$lib/components/ui/separator';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import NotificationDropdown from '$lib/components/ui/notifications/NotificationDropdown.svelte';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { logout } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import UserIcon from '@lucide/svelte/icons/user';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import MenuIcon from '@lucide/svelte/icons/menu';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';

	// Props for customization
	let {
		showBreadcrumbs = false,
		showMobileMenu = false,
		customTitle = null
	}: {
		showBreadcrumbs?: boolean;
		showMobileMenu?: boolean;
		customTitle?: string | null;
	} = $props();

	// Determine if we're in admin area
	const isAdminRoute = $derived(page.url.pathname.startsWith('/admin'));

	// Get current page title based on route and context
	const pageTitle = $derived.by(() => {
		if (customTitle) return customTitle;
		
		const pathname = page.url.pathname;
		
		// Admin routes
		if (pathname.startsWith('/admin')) {
			if (pathname === '/admin') return 'Dashboard';
			if (pathname.startsWith('/admin/users')) return 'Users';
			if (pathname.startsWith('/admin/shifts')) return 'Shifts';
			if (pathname.startsWith('/admin/schedules')) return 'Schedules';
			if (pathname.startsWith('/admin/broadcasts')) return 'Broadcasts';
			if (pathname.startsWith('/admin/reports')) return 'Reports';
			return 'Admin';
		}
		
		// Public routes
		if (pathname === '/') return 'Night Owls';
		if (pathname === '/bookings') return 'My Shifts';
		if (pathname === '/broadcasts') return 'Messages';
		if (pathname === '/report') return 'Report Incident';
		if (pathname === '/login') return 'Sign In';
		if (pathname === '/register') return 'Join Community';
		
		return 'Night Owls';
	});

	// Generate breadcrumbs for admin pages
	const breadcrumbs = $derived.by(() => {
		if (!showBreadcrumbs || !isAdminRoute) return [];
		
		const pathSegments = page.url.pathname.split('/').filter(Boolean);
		return pathSegments.map((segment, index) => {
			const href = '/' + pathSegments.slice(0, index + 1).join('/');
			const label = segment
				.replace(/-/g, ' ')
				.replace(/\b\w/g, (char) => char.toUpperCase());
			return { label, href };
		});
	});

	// Handle logout
	function handleLogout() {
		toast.success('Logged out successfully');
		logout();
	}

	// Get user initials for avatar
	const userInitials = $derived.by(() => {
		if (!$currentUser?.name) return 'U';
		return $currentUser.name
			.split(' ')
			.map(n => n[0])
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
</script>

<header class="flex h-14 shrink-0 items-center gap-2 border-b px-4">
	<!-- Sidebar trigger for admin routes -->
	{#if isAdminRoute}
		<Sidebar.Trigger class="-ml-1" />
		<Separator orientation="vertical" class="mr-2 h-4" />
	{/if}

	<!-- Left side: Logo/Title and Breadcrumbs -->
	<div class="mr-4 flex items-center gap-4">
		<!-- Logo and Title (only for public routes) -->
		{#if !isAdminRoute}
			<a href="/" class="flex items-center space-x-2">
				<div class="h-6 w-6 bg-gradient-to-br from-primary to-primary/80 rounded flex items-center justify-center">
					<span class="text-primary-foreground text-xs font-bold">NO</span>
				</div>
				<span class="hidden font-bold sm:inline-block">
					{pageTitle}
				</span>
			</a>
		{/if}

		<!-- Breadcrumbs for admin pages -->
		{#if showBreadcrumbs && breadcrumbs.length > 0}
			<Breadcrumb.Root>
				<Breadcrumb.List>
					{#each breadcrumbs as crumb, i (crumb.href)}
						<Breadcrumb.Item class="hidden md:block">
							<Breadcrumb.Link href={crumb.href}>{crumb.label}</Breadcrumb.Link>
						</Breadcrumb.Item>
						{#if i < breadcrumbs.length - 1}
							<Breadcrumb.Separator class="hidden md:block" />
						{/if}
					{/each}
				</Breadcrumb.List>
			</Breadcrumb.Root>
		{/if}
	</div>

	<!-- Right side: User actions -->
	<div class="flex flex-1 items-center justify-end space-x-2">
			<!-- Notifications (only for authenticated users) -->
			{#if $isAuthenticated}
				<NotificationDropdown />
			{/if}

			<!-- User Menu -->
			{#if $isAuthenticated}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						<Button variant="ghost" size="sm" class="h-9 w-9 p-0">
							<Avatar class="h-7 w-7">
								<AvatarFallback class="text-xs">
									{userInitials}
								</AvatarFallback>
							</Avatar>
							<span class="sr-only">User menu</span>
						</Button>
					</DropdownMenu.Trigger>

					<DropdownMenu.Content class="w-56" align="end">
						<div class="flex items-center justify-start gap-2 p-2">
							<Avatar class="h-8 w-8">
								<AvatarFallback class="text-sm">
									{userInitials}
								</AvatarFallback>
							</Avatar>
							<div class="flex flex-col space-y-1 leading-none">
								{#if $currentUser?.name}
									<p class="font-medium">{$currentUser.name}</p>
								{/if}
								{#if $currentUser?.phone}
									<p class="w-[180px] truncate text-sm text-muted-foreground">
										{$currentUser.phone}
									</p>
								{/if}
								{#if $currentUser?.role}
									{@const IconComponent = roleInfo.icon}
									<div class="flex items-center gap-1">
										<IconComponent class="h-3 w-3 {roleInfo.color}" />
										<p class="text-xs text-muted-foreground">
											{roleInfo.label}
										</p>
									</div>
								{/if}
							</div>
						</div>
						
						<Separator />
						
						<!-- Role-specific menu items -->
						{#if $currentUser?.role === 'admin'}
							<DropdownMenu.Item class="cursor-pointer" onclick={() => window.location.href = '/admin'}>
								<ShieldIcon class="mr-2 h-4 w-4" />
								<span>Admin Dashboard</span>
							</DropdownMenu.Item>
						{/if}
						
						<DropdownMenu.Item class="cursor-pointer">
							<SettingsIcon class="mr-2 h-4 w-4" />
							<span>Settings</span>
						</DropdownMenu.Item>
						
						<Separator />
						
						<DropdownMenu.Item class="cursor-pointer text-red-600 focus:text-red-600" onclick={handleLogout}>
							<LogOutIcon class="mr-2 h-4 w-4" />
							<span>Log out</span>
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			{:else}
				<!-- Show login button if not authenticated -->
				<Button variant="ghost" size="sm" onclick={() => window.location.href = '/login'}>
					<UserIcon class="h-4 w-4 mr-2" />
					<span class="hidden sm:inline">Sign in</span>
				</Button>
			{/if}

			<!-- Mobile menu button (optional) -->
			{#if showMobileMenu}
				<Button variant="ghost" size="sm" class="h-9 w-9 p-0 md:hidden">
					<MenuIcon class="h-4 w-4" />
					<span class="sr-only">Menu</span>
				</Button>
			{/if}
		</div>
</header> 