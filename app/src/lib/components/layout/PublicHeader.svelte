<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import NotificationDropdown from '$lib/components/ui/notifications/NotificationDropdown.svelte';
	import UserIcon from '@lucide/svelte/icons/user';
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
			<Button variant="ghost" size="sm" class="h-9 w-9 p-0">
				<UserIcon class="h-4 w-4" />
				<span class="sr-only">User menu</span>
			</Button>

			<!-- Mobile menu (optional) -->
			<Button variant="ghost" size="sm" class="h-9 w-9 p-0 md:hidden">
				<MenuIcon class="h-4 w-4" />
				<span class="sr-only">Menu</span>
			</Button>
		</div>
	</div>
</header> 