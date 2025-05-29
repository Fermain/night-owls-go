<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import UsersDashboard from '$lib/components/admin/users/UsersDashboard.svelte';
	import { selectedUserForForm } from '$lib/stores/userEditingStore';
	import { createUsersQuery } from '$lib/queries/admin/users/usersQuery';
	import { createDashboardShiftsQuery } from '$lib/queries/admin/shifts/dashboardShiftsQuery';
	import type { UserData } from '$lib/schemas/user';

	// currentUserForForm is derived from the store, which is synced with URL by the layout
	let currentUserForForm = $state<UserData | undefined>(undefined);
	selectedUserForForm.subscribe((value) => {
		currentUserForForm = value;
	});

	// Create queries for dashboard
	const usersQuery = $derived(createUsersQuery());
	const shiftsQuery = $derived(createDashboardShiftsQuery());

	// Combined loading and error states
	const isLoading = $derived($usersQuery.isLoading || $shiftsQuery.isLoading);
	const isError = $derived($usersQuery.isError || $shiftsQuery.isError);
	const error = $derived($usersQuery.error || $shiftsQuery.error || undefined);
</script>

{#if currentUserForForm}
	{#key currentUserForForm.id}
		<UserForm user={currentUserForForm} />
	{/key}
{:else}
	<!-- Users Dashboard with Shift Data -->
	<UsersDashboard
		{isLoading}
		{isError}
		{error}
		users={$usersQuery.data}
		shifts={$shiftsQuery.data}
	/>
{/if}
