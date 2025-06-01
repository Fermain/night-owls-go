<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Separator } from '$lib/components/ui/separator';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import NotificationDropdown from '$lib/components/ui/notifications/NotificationDropdown.svelte';
	import EmergencyContactsDialog from '$lib/components/emergency/EmergencyContactsDialog.svelte';
	import ReportDialog from '$lib/components/user/report/ReportDialog.svelte';
	import UserSettingsDialog from '$lib/components/user/settings/UserSettingsDialog.svelte';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { logout } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';
	import UserIcon from '@lucide/svelte/icons/user';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import SettingsIcon from '@lucide/svelte/icons/settings';

	// State for dialogs
	let emergencyDialogOpen = $state(false);
	let reportDialogOpen = $state(false);
	let settingsDialogOpen = $state(false);

	// State to track if component is mounted to prevent Dialog lifecycle errors
	let mounted = $state(false);

	// Mount handler to prevent Dialog lifecycle errors
	onMount(() => {
		mounted = true;
	});

	// Determine if we're in admin area (defensive check for SSR/hydration)
	const isAdminRoute = $derived(page?.url?.pathname?.startsWith('/admin') ?? false);

	// Determine if we're on the report page
	const isReportPage = $derived(page.url.pathname === '/report');

	// Handle logout
	function handleLogout() {
		toast.success('Logged out successfully');
		logout();
	}

	// Handle emergency call
	function handleEmergency() {
		emergencyDialogOpen = true;
	}

	// Handle report dialog
	function handleReport() {
		reportDialogOpen = true;
	}

	// Handle settings dialog
	function handleSettings() {
		settingsDialogOpen = true;
	}

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
</script>

<header class="flex h-14 shrink-0 items-center gap-2 border-b px-4">
	<!-- Logo and Title -->
	<a href="/" class="flex items-center space-x-2">
		<div class="h-8 w-8 p-1 flex items-center justify-center">
			<img src="/logo.png" alt="Mount Moreland Night Owls" class="object-contain" />
		</div>
	</a>

	<!-- Right side: User actions -->
	<div class="flex flex-1 items-center justify-end space-x-2">
		<!-- Report button (only for authenticated users, not on report page) -->
		{#if $isAuthenticated && !isReportPage}
			<Button variant="outline" size="sm" onclick={handleReport} class="hidden sm:flex">
				<AlertTriangleIcon class="h-4 w-4 mr-2" />
				Report
			</Button>
			<Button variant="outline" size="sm" onclick={handleReport} class="sm:hidden h-9 w-9 p-0">
				<AlertTriangleIcon class="h-4 w-4" />
				<span class="sr-only">Report</span>
			</Button>
		{/if}

		<!-- Emergency button (only for authenticated users) -->
		{#if $isAuthenticated}
			<Button variant="destructive" size="sm" onclick={handleEmergency} class="hidden sm:flex">
				<PhoneIcon class="h-4 w-4 mr-2" />
				Emergency
			</Button>
			<Button
				variant="destructive"
				size="sm"
				onclick={handleEmergency}
				class="sm:hidden h-9 w-9 p-0"
			>
				<PhoneIcon class="h-4 w-4" />
				<span class="sr-only">Emergency</span>
			</Button>
		{/if}

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
						<DropdownMenu.Item
							class="cursor-pointer"
							onclick={() => (window.location.href = '/admin')}
						>
							<ShieldIcon class="mr-2 h-4 w-4" />
							<span>Admin Dashboard</span>
						</DropdownMenu.Item>

						<DropdownMenu.Item class="cursor-pointer" onclick={() => (window.location.href = '/')}>
							<UserIcon class="mr-2 h-4 w-4" />
							<span>User Dashboard</span>
						</DropdownMenu.Item>

						<Separator />
					{/if}

					<DropdownMenu.Item class="cursor-pointer" onclick={handleSettings}>
						<SettingsIcon class="mr-2 h-4 w-4" />
						<span>Settings</span>
					</DropdownMenu.Item>

					<DropdownMenu.Item
						class="cursor-pointer text-red-600 focus:text-red-600"
						onclick={handleLogout}
					>
						<LogOutIcon class="mr-2 h-4 w-4" />
						<span>Log out</span>
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{:else}
			<!-- Show register and login buttons if not authenticated -->
			<div class="flex items-center gap-2">
				<Button variant="ghost" size="sm" onclick={() => (window.location.href = '/login')}>
					<span class="hidden sm:inline">Sign in</span>
					<span class="sm:hidden">Sign in</span>
				</Button>
				<Button size="sm" onclick={() => (window.location.href = '/register')}>
					<span class="hidden sm:inline">Register</span>
					<span class="sm:hidden">Join</span>
				</Button>
			</div>
		{/if}
	</div>
</header>

<!-- Emergency Contacts Dialog -->
{#if mounted}
	<EmergencyContactsDialog bind:open={emergencyDialogOpen} />

	<!-- Report Dialog -->
	<ReportDialog bind:open={reportDialogOpen} />

	<!-- Settings Dialog -->
	<UserSettingsDialog bind:open={settingsDialogOpen} />
{/if}
