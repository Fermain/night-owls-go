<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Sheet from '$lib/components/ui/sheet';
	import { Badge } from '$lib/components/ui/badge';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	// Icons
	import MenuIcon from '@lucide/svelte/icons/menu';
	import HomeIcon from '@lucide/svelte/icons/home';
	import UsersIcon from '@lucide/svelte/icons/users';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import HistoryIcon from '@lucide/svelte/icons/history';
	import ShieldIcon from '@lucide/svelte/icons/shield';

	// State for mobile menu
	let mobileMenuOpen = $state(false);

	// Admin navigation items
	const adminNavItems = [
		{
			title: 'Dashboard',
			url: '/admin',
			icon: HomeIcon,
			description: 'Overview and quick actions'
		},
		{
			title: 'Users',
			url: '/admin/users',
			icon: UsersIcon,
			description: 'Manage community members'
		},
		{
			title: 'Shifts',
			url: '/admin/shifts',
			icon: CalendarIcon,
			description: 'Schedule and assignments'
		},
		{
			title: 'Reports',
			url: '/admin/reports',
			icon: BarChart3Icon,
			description: 'Incident reports and analytics'
		},
		{
			title: 'Broadcasts',
			url: '/admin/broadcasts',
			icon: MessageSquareIcon,
			description: 'Community notifications'
		},
		{
			title: 'Schedules',
			url: '/admin/schedules',
			icon: SettingsIcon,
			description: 'Manage shift schedules'
		},
		{
			title: 'Emergency Contacts',
			url: '/admin/emergency-contacts',
			icon: PhoneIcon,
			description: 'Emergency contact list'
		},
		{
			title: 'Audit History',
			url: '/admin/history',
			icon: HistoryIcon,
			description: 'System audit trail'
		}
	];

	// Get current page info
	const currentPage = $derived.by(() => {
		const currentPath = page.url.pathname;
		return (
			adminNavItems.find(
				(item) =>
					currentPath === item.url || (item.url !== '/admin' && currentPath.startsWith(item.url))
			) || adminNavItems[0]
		);
	});

	function handleNavigation(url: string) {
		mobileMenuOpen = false;
		goto(url);
	}
</script>

<!-- Mobile Admin Header (shown on small screens) -->
<header class="lg:hidden bg-background border-b border-border sticky top-0 z-50">
	<div class="flex items-center justify-between p-4">
		<!-- Current Page Title -->
		<div class="flex items-center gap-3">
			{#if currentPage.icon}
				{@const IconComponent = currentPage.icon}
				<IconComponent class="h-5 w-5 text-primary" />
			{/if}
			<div>
				<h1 class="font-semibold text-lg">{currentPage.title}</h1>
				<p class="text-xs text-muted-foreground">{currentPage.description}</p>
			</div>
		</div>

		<!-- Mobile Menu Trigger -->
		<Sheet.Root bind:open={mobileMenuOpen}>
			<Sheet.Trigger>
				<Button variant="ghost" size="sm" class="p-2">
					<MenuIcon class="h-5 w-5" />
				</Button>
			</Sheet.Trigger>

			<Sheet.Content side="right" class="w-80">
				<Sheet.Header>
					<Sheet.Title class="flex items-center gap-2">
						<ShieldIcon class="h-5 w-5 text-primary" />
						Admin Menu
					</Sheet.Title>
					<Sheet.Description>Navigate to different admin sections</Sheet.Description>
				</Sheet.Header>

				<div class="mt-6">
					<nav class="space-y-2">
						{#each adminNavItems as item (item.url)}
							{@const isActive =
								page.url.pathname === item.url ||
								(item.url !== '/admin' && page.url.pathname.startsWith(item.url))}

							<button
								onclick={() => handleNavigation(item.url)}
								class="w-full flex items-center gap-3 p-3 rounded-lg text-left transition-colors hover:bg-muted {isActive
									? 'bg-primary/10 text-primary border border-primary/20'
									: 'text-muted-foreground hover:text-foreground'}"
							>
								{#if item.icon}
									{@const IconComponent = item.icon}
									<IconComponent class="h-5 w-5" />
								{/if}
								<div class="flex-1">
									<div class="font-medium">{item.title}</div>
									<div class="text-xs text-muted-foreground">{item.description}</div>
								</div>
								{#if isActive}
									<Badge variant="secondary" class="text-xs">Current</Badge>
								{/if}
							</button>
						{/each}
					</nav>
				</div>

				<!-- Quick Actions in Mobile Menu -->
				<div class="mt-8 pt-6 border-t">
					<h3 class="font-medium mb-3">Quick Actions</h3>
					<div class="grid grid-cols-2 gap-2">
						<Button
							variant="outline"
							size="sm"
							onclick={() => handleNavigation('/admin/users/new')}
							class="h-12 flex-col gap-1"
						>
							<UsersIcon class="h-4 w-4" />
							<span class="text-xs">Add User</span>
						</Button>

						<Button
							variant="outline"
							size="sm"
							onclick={() => handleNavigation('/admin/broadcasts')}
							class="h-12 flex-col gap-1"
						>
							<MessageSquareIcon class="h-4 w-4" />
							<span class="text-xs">Broadcast</span>
						</Button>

						<Button
							variant="outline"
							size="sm"
							onclick={() => handleNavigation('/admin/shifts')}
							class="h-12 flex-col gap-1"
						>
							<CalendarIcon class="h-4 w-4" />
							<span class="text-xs">Fill Shifts</span>
						</Button>

						<Button
							variant="outline"
							size="sm"
							onclick={() => handleNavigation('/admin/reports')}
							class="h-12 flex-col gap-1"
						>
							<BarChart3Icon class="h-4 w-4" />
							<span class="text-xs">Reports</span>
						</Button>
					</div>
				</div>
			</Sheet.Content>
		</Sheet.Root>
	</div>
</header>
