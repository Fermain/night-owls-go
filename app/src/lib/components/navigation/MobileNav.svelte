<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import HomeIcon from '@lucide/svelte/icons/home';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import UserIcon from '@lucide/svelte/icons/user';
	import MenuIcon from '@lucide/svelte/icons/menu';
	import XIcon from '@lucide/svelte/icons/x';
	import ShieldIcon from '@lucide/svelte/icons/shield';

	let { isOpen = $bindable(false) }: { isOpen?: boolean } = $props();

	// Navigation items for authenticated users
	const navItems = [
		{
			href: '/',
			icon: HomeIcon,
			label: 'Home',
			badge: null
		},
		{
			href: '/bookings',
			icon: CalendarIcon,
			label: 'My Shifts',
			badge: null
		},
		{
			href: '/broadcasts',
			icon: MessageSquareIcon,
			label: 'Messages',
			badge: 2 // TODO: Connect to real notification count
		},
		{
			href: '/report',
			icon: AlertTriangleIcon,
			label: 'Report',
			badge: null
		}
	];

	// Get current route for active state
	const currentPath = $derived(page.url.pathname);

	function isActiveRoute(href: string): boolean {
		if (href === '/') return currentPath === '/';
		return currentPath.startsWith(href);
	}

	function closeMenu() {
		isOpen = false;
	}
</script>

{#if $isAuthenticated}
	<!-- Bottom Tab Navigation for Authenticated Users -->
	<nav class="fixed bottom-0 left-0 right-0 z-40 bg-background/95 backdrop-blur-sm border-t border-border md:hidden">
		<div class="flex items-center justify-around px-2 py-2">
			{#each navItems as item}
				{@const IconComponent = item.icon}
				{@const isActive = isActiveRoute(item.href)}
				<a
					href={item.href}
					class="flex flex-col items-center justify-center min-w-0 flex-1 px-1 py-2 text-xs transition-colors
						{isActive 
							? 'text-primary' 
							: 'text-muted-foreground hover:text-foreground'}"
				>
					<div class="relative">
						<IconComponent class="h-5 w-5 mb-1" />
						{#if item.badge}
							<Badge 
								variant="destructive" 
								class="absolute -top-2 -right-2 h-4 w-4 p-0 text-xs flex items-center justify-center"
							>
								{item.badge}
							</Badge>
						{/if}
					</div>
					<span class="truncate max-w-full">{item.label}</span>
				</a>
			{/each}
		</div>
	</nav>
{:else}
	<!-- Mobile Menu Button for Unauthenticated Users -->
	<div class="md:hidden">
		<Button variant="ghost" size="icon" onclick={() => isOpen = !isOpen}>
			<MenuIcon class="h-5 w-5" />
		</Button>
	</div>

	<!-- Slide-out Menu for Unauthenticated Users -->
	{#if isOpen}
		<div class="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm md:hidden">
			<div class="fixed inset-y-0 left-0 z-50 w-3/4 max-w-sm border-r bg-background">
				<!-- Header -->
				<div class="flex items-center justify-between p-4 border-b">
					<div class="flex items-center gap-2">
						<div class="h-8 w-8 bg-gradient-to-br from-primary to-primary/80 rounded flex items-center justify-center">
							<span class="text-primary-foreground text-sm font-bold">NO</span>
						</div>
						<span class="font-semibold">Night Owls</span>
					</div>
					<Button variant="ghost" size="icon" onclick={closeMenu}>
						<XIcon class="h-5 w-5" />
					</Button>
				</div>

				<!-- Navigation Links -->
				<div class="p-4 space-y-2">
					<a 
						href="/login" 
						class="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
						onclick={closeMenu}
					>
						<UserIcon class="h-5 w-5" />
						<span>Sign In</span>
					</a>
					<a 
						href="/register" 
						class="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
						onclick={closeMenu}
					>
						<ShieldIcon class="h-5 w-5" />
						<span>Join Community</span>
					</a>
					<a 
						href="/bookings" 
						class="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
						onclick={closeMenu}
					>
						<CalendarIcon class="h-5 w-5" />
						<span>View Shifts</span>
					</a>
				</div>

				<!-- Footer -->
				<div class="absolute bottom-4 left-4 right-4 text-center">
					<p class="text-xs text-muted-foreground">
						Community Watch Platform
					</p>
				</div>
			</div>
		</div>
	{/if}
{/if} 