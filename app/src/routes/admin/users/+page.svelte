<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import { selectedUserForForm, type UserData } from '$lib/stores/userEditingStore';
	import { Button } from '$lib/components/ui/button';
	import { goto } from '$app/navigation';
	import { page } from '$app/state'; // For reading URL params
	import { Skeleton } from '$lib/components/ui/skeleton/index.js'; // For dashboard placeholders

	// currentUserForForm is derived from the store, which is synced with URL by the layout
	let currentUserForForm = $state<UserData | undefined>(undefined);
	selectedUserForForm.subscribe((value) => {
		currentUserForForm = value;
	});

	// No need for onDestroy(unsubscribe) with auto-unsubscription for stores in Svelte 5 components,
	// but if selectedUserForForm were a raw Svelte store used directly in markup ($selectedUserForForm),
	// then manual sub/unsub like this would be for reacting and setting local $state.
	// Given currentUserForForm is $state, the subscription is to update this local reactive state.
	// This pattern is fine.

	let isDashboardView = $derived(page.url.searchParams.get('view') === 'dashboard');

</script>

{#if isDashboardView}
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
{:else if currentUserForForm}
	{#key currentUserForForm.id}
		<UserForm user={currentUserForForm} />
	{/key}
{/if}
