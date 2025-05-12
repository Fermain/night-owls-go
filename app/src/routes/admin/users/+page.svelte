<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import { selectedUserForForm } from '$lib/stores/userEditingStore';
	import type { UserData } from '$lib/schemas/user';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js'; // For dashboard placeholders

	// currentUserForForm is derived from the store, which is synced with URL by the layout
	let currentUserForForm = $state<UserData | undefined>(undefined);
	selectedUserForForm.subscribe((value) => {
		currentUserForForm = value;
	});
</script>

{#if currentUserForForm}
	{#key currentUserForForm.id}
		<UserForm user={currentUserForForm} />
	{/key}
{:else}
	<div class="p-4 md:p-8">
		<h1 class="text-2xl font-semibold mb-6">Users Dashboard</h1>
		<p class="mb-6 text-muted-foreground">
			Analytics and statistics related to users will be displayed here.
		</p>
		<div class="space-y-4">
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each [1, 2, 3] as i (i)}
					<div class="p-4 border rounded-lg">
						<Skeleton class="h-6 w-3/4 mb-2" />
						<Skeleton class="h-10 w-1/2 mb-4" />
						<Skeleton class="h-4 w-full" />
						<Skeleton class="h-4 w-5/6 mt-1" />
					</div>
				{/each}
			</div>
			<div class="p-4 border rounded-lg">
				<Skeleton class="h-8 w-1/4 mb-4" />
				<Skeleton class="h-48 w-full" />
			</div>
		</div>
	</div>
{/if}
