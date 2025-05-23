<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import { userStore } from '$lib/stores/authStore';
	import { logout } from '$lib/utils/auth';

	// Get current user from auth store
	const currentUser = $derived($userStore);

	function getUserInitials(name: string | null | undefined): string {
		if (!name) return '?';
		return name.split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase();
	}

	function getRoleIcon(role: string) {
		switch (role) {
			case 'admin': return ShieldIcon;
			case 'owl': return StarIcon;
			default: return UserIcon;
		}
	}

	function getRoleColor(role: string) {
		switch (role) {
			case 'admin': return 'destructive';
			case 'owl': return 'default';
			case 'guest': return 'secondary';
			default: return 'outline';
		}
	}

	function getRoleLabel(role: string) {
		switch (role) {
			case 'admin': return 'Administrator';
			case 'owl': return 'Night Owl';
			case 'guest': return 'Guest';
			default: return role;
		}
	}

	async function handleLogout() {
		await logout();
	}
</script>

{#if currentUser}
	<div class="p-3 border-t bg-muted/30">
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				<Button 
					class="w-full justify-start gap-3 h-auto p-3" 
					variant="ghost"
				>
					<Avatar.Root class="h-8 w-8">
						<Avatar.Fallback class="text-xs">
							{getUserInitials(currentUser.name)}
						</Avatar.Fallback>
					</Avatar.Root>
					<div class="flex-1 text-left">
						<div class="flex items-center gap-2 mb-1">
							<p class="text-sm font-medium leading-none truncate">
								{currentUser.name || 'Unnamed User'}
							</p>
						</div>
						<div class="flex items-center justify-between">
							<Badge variant={getRoleColor(currentUser.role)} class="text-xs">
								{#if currentUser.role === 'admin'}
									<ShieldIcon class="h-3 w-3 mr-1" />
								{:else if currentUser.role === 'owl'}
									<StarIcon class="h-3 w-3 mr-1" />
								{:else}
									<UserIcon class="h-3 w-3 mr-1" />
								{/if}
								{getRoleLabel(currentUser.role)}
							</Badge>
						</div>
					</div>
				</Button>
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="w-56" side="top" align="start">
				<DropdownMenu.Label>My Account</DropdownMenu.Label>
				<DropdownMenu.Separator />
				<DropdownMenu.Item class="cursor-pointer">
					<UserIcon class="mr-2 h-4 w-4" />
					<span>Profile</span>
				</DropdownMenu.Item>
				<DropdownMenu.Item class="cursor-pointer">
					<SettingsIcon class="mr-2 h-4 w-4" />
					<span>Settings</span>
				</DropdownMenu.Item>
				<DropdownMenu.Separator />
				<DropdownMenu.Item class="cursor-pointer text-destructive focus:text-destructive" onclick={handleLogout}>
					<LogOutIcon class="mr-2 h-4 w-4" />
					<span>Log out</span>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</div>
{/if} 