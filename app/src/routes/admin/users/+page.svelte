<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import UsersDashboard from '$lib/components/admin/users/UsersDashboard.svelte';
	import { selectedUserForForm } from '$lib/stores/userEditingStore';
	import { createUsersQuery } from '$lib/queries/admin/users/usersQuery';
	import type { UserData } from '$lib/schemas/user';

	// currentUserForForm is derived from the store, which is synced with URL by the layout
	let currentUserForForm = $state<UserData | undefined>(undefined);
	selectedUserForForm.subscribe((value) => {
		currentUserForForm = value;
	});

	// Create users query for dashboard
	const usersQuery = $derived(createUsersQuery());
</script>

{#if currentUserForForm}
	{#key currentUserForForm.id}
		<UserForm user={currentUserForForm} />
	{/key}
{:else}
	<!-- Users Dashboard -->
	<UsersDashboard 
		isLoading={$usersQuery.isLoading}
		isError={$usersQuery.isError}
		error={$usersQuery.error || undefined}
		users={$usersQuery.data}
	/>
{/if}
