<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import { selectedUserForForm, type UserData } from '$lib/stores/userEditingStore';
	import { Button } from '$lib/components/ui/button';
	import { goto } from '$app/navigation';

	// No need to fetch individual user data here anymore, 
	// it comes from the layout via the store if a user is selected for editing.

	// To clear the selection and show dashboard, the layout now handles setting the store to undefined.
	// This page just reflects the store's content.

	let currentUserForForm: UserData | undefined = undefined;
	const unsubscribe = selectedUserForForm.subscribe(value => {
		currentUserForForm = value;
	});

	// Ensure to unsubscribe when the component is destroyed
	import { onDestroy } from 'svelte';
	onDestroy(unsubscribe);

</script>

{#if currentUserForForm}
	<UserForm user={currentUserForForm} />
{:else}
	<div class="p-4 text-center">
		<h1 class="text-2xl font-semibold mb-4">Manage Users</h1>
		<p class="mb-6 text-muted-foreground">
			Select a user from the list to view or edit their details, or create a new user.
		</p>
		<Button onclick={() => goto('/admin/users/new')}>Create New User</Button>
		<!-- 
			Alternative: if 'Create New User' in sidebar already clears selectedUserForForm 
			and navigates to /admin/users/new, this button might be redundant 
			or could directly call selectedUserForForm.set(undefined) and then goto(...)
		-->
	</div>
{/if}
