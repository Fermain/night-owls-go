<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Badge } from '$lib/components/ui/badge';
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import { formatRelativeTime } from '$lib/utils/dateFormatting';
	import type { UserData } from '$lib/schemas/user';

	let { users }: { users: UserData[] } = $props();

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

	function getUserInitials(name: string | null | undefined): string {
		if (!name) return '?';
		return name.split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase();
	}
</script>

<Card.Root>
	<Card.Header class="pb-4">
		<Card.Title>Recent Registrations</Card.Title>
		<Card.Description>Latest users to join the platform</Card.Description>
	</Card.Header>
	<Card.Content>
		{#if users.length === 0}
			<div class="text-center py-8 text-muted-foreground">
				No recent registrations
			</div>
		{:else}
			<div class="space-y-6">
				{#each users as user (user.id)}
					<div class="flex items-center gap-4">
						<Avatar.Root class="h-10 w-10">
							<Avatar.Fallback class="text-sm">
								{getUserInitials(user.name)}
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="flex-1 space-y-2">
							<div class="flex items-center gap-3">
								<p class="text-sm font-medium leading-none">
									{user.name || 'Unnamed User'}
								</p>
								<Badge variant={getRoleColor(user.role)} class="text-xs">
									{#if user.role === 'admin'}
										<ShieldIcon class="h-3 w-3 mr-1" />
									{:else if user.role === 'owl'}
										<StarIcon class="h-3 w-3 mr-1" />
									{:else}
										<UserIcon class="h-3 w-3 mr-1" />
									{/if}
									{user.role}
								</Badge>
							</div>
							<div class="flex items-center justify-between">
								<p class="text-xs text-muted-foreground">
									{user.phone}
								</p>
								{#if user.created_at}
									<p class="text-xs text-muted-foreground">
										{formatRelativeTime(user.created_at)}
									</p>
								{/if}
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</Card.Content>
</Card.Root> 