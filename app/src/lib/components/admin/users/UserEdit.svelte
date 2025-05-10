<script lang="ts">
	import UserForm from './UserForm.svelte';
	import type { UserData } from './UserForm.svelte';

	// The selectedUser prop may come from the parent component with a slightly different structure
	let { selectedUser }: { selectedUser: any | null } = $props();

	// Convert selectedUser to UserData format if needed
	$effect(() => {
		if (selectedUser) {
			// Ensure selectedUser has required UserData structure
			selectedUser = {
				user_id: selectedUser.user_id || selectedUser.id, // Handle either format
				phone: selectedUser.phone,
				name: selectedUser.name || null,
				created_at: selectedUser.created_at || new Date().toISOString()
			};
		}
	});
</script>

<div class="user-edit-container">
	{#if selectedUser}
		<UserForm user={selectedUser} />
	{:else}
		<p>Select a user to view details.</p>
	{/if}
</div>

<style>
	/* Add styling here if needed */
</style>
