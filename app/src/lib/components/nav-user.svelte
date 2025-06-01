<script lang="ts">
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import ChevronsUpDown from '@lucide/svelte/icons/chevrons-up-down';
	import LogOut from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';

	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import { userStore, logout } from '$lib/stores/authStore';

	// Safe sidebar context access
	function getSidebarContext() {
		try {
			return useSidebar();
		} catch {
			return null;
		}
	}

	// Get current user from auth store
	const currentUser = $derived($userStore);

	function getUserInitials(name: string | null | undefined): string {
		if (!name) return '?';
		return name
			.split(' ')
			.map((n) => n[0])
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	function getRoleLabel(role: string | null) {
		switch (role) {
			case 'admin':
				return 'Administrator';
			case 'owl':
				return 'Night Owl';
			case 'guest':
				return 'Guest';
			default:
				return 'User';
		}
	}

	function _getRoleIcon(role: string | null) {
		switch (role) {
			case 'admin':
				return ShieldIcon;
			case 'owl':
				return StarIcon;
			default:
				return UserIcon;
		}
	}

	async function handleLogout() {
		logout();
	}

	// Fallback user data if not authenticated
	const displayUser = $derived(
		currentUser?.isAuthenticated
			? currentUser
			: {
					name: 'Guest User',
					phone: 'Not logged in',
					role: 'guest'
				}
	);
</script>

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
								{getUserInitials(displayUser.name)}
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="grid flex-1 text-left text-sm leading-tight">
							<span class="truncate font-semibold">{displayUser.name || 'Unnamed User'}</span>
							<span class="truncate text-xs">{displayUser.phone || 'No phone'}</span>
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
								{getUserInitials(displayUser.name)}
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="grid flex-1 text-left text-sm leading-tight">
							<span class="truncate font-semibold">{displayUser.name || 'Unnamed User'}</span>
							<div class="flex items-center gap-1 mt-1">
								<span class="text-xs text-muted-foreground">
									{#if displayUser.role === 'admin'}
										<ShieldIcon class="h-3 w-3 inline mr-1" />
									{:else if displayUser.role === 'owl'}
										<StarIcon class="h-3 w-3 inline mr-1" />
									{:else}
										<UserIcon class="h-3 w-3 inline mr-1" />
									{/if}
									{getRoleLabel(displayUser.role)}
								</span>
							</div>
						</div>
					</div>
				</DropdownMenu.Label>
				<DropdownMenu.Separator />

				{#if currentUser?.isAuthenticated}
					<DropdownMenu.Group>
						<DropdownMenu.Item class="cursor-pointer">
							<UserIcon class="mr-2 h-4 w-4" />
							My Profile
						</DropdownMenu.Item>
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
				{:else}
					<DropdownMenu.Item
						class="cursor-pointer"
						onclick={() => (window.location.href = '/login')}
					>
						<UserIcon class="mr-2 h-4 w-4" />
						Sign In
					</DropdownMenu.Item>
				{/if}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
