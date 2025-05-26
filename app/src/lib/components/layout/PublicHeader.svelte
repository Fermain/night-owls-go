<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Separator } from '$lib/components/ui/separator';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import NotificationDropdown from '$lib/components/ui/notifications/NotificationDropdown.svelte';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { logout } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import UserIcon from '@lucide/svelte/icons/user';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import MenuIcon from '@lucide/svelte/icons/menu';

	// Get current page title based on route
	const pageTitle = $derived.by(() => {
		const pathname = page.url.pathname;
		
		if (pathname === '/') return 'Night Owls';
		if (pathname.startsWith('/shifts')) return 'Shifts';
		if (pathname.startsWith('/broadcasts')) return 'Messages';
		if (pathname.startsWith('/report')) return 'Report Incident';
		if (pathname.startsWith('/bookings')) return 'My Bookings';
		
		return 'Night Owls';
	});

	// Handle logout
	function handleLogout() {
		toast.success('Logged out successfully');
		logout(); // This will clear session and navigate to login page
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
</script>

<header class="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
	<div class="container flex h-14 max-w-screen-2xl items-center">
		<!-- Left side: Logo/Title -->
		<div class="mr-4 flex">
			<a href="/" class="mr-6 flex items-center space-x-2">
				<div class="h-6 w-6 bg-gradient-to-br from-primary to-primary/80 rounded flex items-center justify-center">
					<span class="text-primary-foreground text-xs font-bold">NO</span>
				</div>
				<span class="hidden font-bold sm:inline-block">
					{pageTitle}
				</span>
			</a>
		</div>

		<!-- Right side: User actions -->
		<div class="flex flex-1 items-center justify-end space-x-2">
			<!-- Notifications -->
			<NotificationDropdown />

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
									<p class="text-xs text-muted-foreground capitalize">
										{$currentUser.role}
									</p>
								{/if}
							</div>
						</div>
						
						<Separator />
						
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

			<!-- Mobile menu (optional) -->
			<Button variant="ghost" size="sm" class="h-9 w-9 p-0 md:hidden">
				<MenuIcon class="h-4 w-4" />
				<span class="sr-only">Menu</span>
			</Button>
		</div>
	</div>
</header> 