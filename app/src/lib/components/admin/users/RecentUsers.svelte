<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import UsersIcon from '@lucide/svelte/icons/users';
	import type { UserData } from '$lib/schemas/user';

	let {
		isLoading = false,
		_users = []
	}: {
		isLoading?: boolean;
		_users?: UserData[];
	} = $props();
</script>

<Card.Root class="p-6">
	<div class="flex items-center gap-2 mb-4">
		<UsersIcon class="h-5 w-5 text-muted-foreground" />
		<h2 class="text-lg font-semibold">Recent Users</h2>
	</div>
	{#if isLoading}
		<Skeleton class="h-32 w-full" />
	{:else}
		<div class="text-center py-8 text-muted-foreground">
			<p class="text-sm">Recent users will be displayed here</p>
		</div>
	{/if}
</Card.Root>
