<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Avatar from '$lib/components/ui/avatar';
	import { Badge } from '$lib/components/ui/badge';
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import TrophyIcon from '@lucide/svelte/icons/trophy';
	import type { UserShiftDistribution } from '$lib/utils/userProcessing';

	let { volunteers }: { volunteers: UserShiftDistribution[] } = $props();

	function getUserInitials(name: string): string {
		if (!name) return '?';
		return name
			.split(' ')
			.map((n) => n[0])
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	function getRoleColor(role: string) {
		switch (role) {
			case 'admin':
				return 'destructive';
			case 'owl':
				return 'default';
			case 'guest':
				return 'secondary';
			default:
				return 'outline';
		}
	}

	function getRankIcon(index: number) {
		if (index === 0) return '🥇';
		if (index === 1) return '🥈';
		if (index === 2) return '🥉';
		return `#${index + 1}`;
	}
</script>

<Card.Root>
	<Card.Header class="pb-4">
		<div class="flex items-center gap-2">
			<TrophyIcon class="h-5 w-5 text-yellow-600" />
			<Card.Title>Top Volunteers</Card.Title>
		</div>
		<Card.Description>Most active volunteers by shift count</Card.Description>
	</Card.Header>
	<Card.Content>
		{#if volunteers.length === 0}
			<div class="text-center py-8 text-muted-foreground">No volunteers with shifts yet</div>
		{:else}
			<div class="space-y-4">
				{#each volunteers as volunteer, index (volunteer.userId)}
					<div class="flex items-center gap-4">
						<div
							class="flex items-center justify-center w-8 h-8 rounded-full bg-muted text-sm font-medium"
						>
							{getRankIcon(index)}
						</div>
						<Avatar.Root class="h-10 w-10">
							<Avatar.Fallback class="text-sm">
								{getUserInitials(volunteer.userName)}
							</Avatar.Fallback>
						</Avatar.Root>
						<div class="flex-1 space-y-1">
							<div class="flex items-center gap-3">
								<p class="text-sm font-medium leading-none">
									{volunteer.userName}
								</p>
								<Badge variant={getRoleColor(volunteer.userRole)} class="text-xs">
									{#if volunteer.userRole === 'admin'}
										<ShieldIcon class="h-3 w-3 mr-1" />
									{:else if volunteer.userRole === 'owl'}
										<StarIcon class="h-3 w-3 mr-1" />
									{:else}
										<UserIcon class="h-3 w-3 mr-1" />
									{/if}
									{volunteer.userRole}
								</Badge>
							</div>
							<div class="flex items-center justify-between">
								<p class="text-xs text-muted-foreground">
									{volunteer.shiftCount} shifts ({volunteer.percentage}% of total)
								</p>
								<div class="w-24 bg-secondary rounded-full h-2">
									<div
										class="bg-primary h-2 rounded-full transition-all duration-300"
										style="width: {Math.min(volunteer.percentage, 100)}%"
									></div>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</Card.Content>
</Card.Root>
