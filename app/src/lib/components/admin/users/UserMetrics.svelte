<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import type { UserMetrics } from '$lib/utils/userProcessing';

	let { metrics }: { metrics: UserMetrics } = $props();
</script>

<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
	<!-- Total Users -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Total Users</Card.Title>
			<UsersIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.totalUsers}</div>
			<p class="text-xs text-muted-foreground mt-1">Registered users in the system</p>
		</Card.Content>
	</Card.Root>

	<!-- Admin Users -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Administrators</Card.Title>
			<ShieldIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.adminUsers}</div>
			<p class="text-xs text-muted-foreground mt-1">System administrators</p>
		</Card.Content>
	</Card.Root>

	<!-- Owl Users -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Night Owls</Card.Title>
			<StarIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.owlUsers}</div>
			<p class="text-xs text-muted-foreground mt-1">Active volunteers</p>
		</Card.Content>
	</Card.Root>

	<!-- Recent Users -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Recent Signups</Card.Title>
			<UserPlusIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.recentUsers}</div>
			<p class="text-xs text-muted-foreground mt-1">Last 30 days</p>
		</Card.Content>
	</Card.Root>
</div>
