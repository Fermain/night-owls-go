<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import MobileUsersDashboard from '$lib/components/admin/users/MobileUsersDashboard.svelte';
	import { selectedUserForForm } from '$lib/stores/userEditingStore';
	import type { User } from '$lib/types/domain';
	import UsersIcon from '@lucide/svelte/icons/users';

	// currentUserForForm is derived from the store, which is synced with URL by the layout
	let currentUserForForm = $state<User | undefined>(undefined);
	selectedUserForForm.subscribe((value) => {
		if (value) {
			// Convert UserData from store to our domain User type
			currentUserForForm = {
				id: value.id,
				name: value.name ?? '',
				phone: value.phone,
				role: value.role,
				createdAt: value.created_at,
				isActive: true // Default value for new domain field
			};
		} else {
			currentUserForForm = undefined;
		}
	});
</script>

<svelte:head>
	<title>User Management - Mount Moreland Night Owls</title>
</svelte:head>

{#if currentUserForForm}
	{#key currentUserForForm.id}
		<UserForm user={currentUserForForm} />
	{/key}
{:else}
	<div class="p-4 md:p-6 space-y-6">
		<!-- Page Header -->
		<div class="border-b pb-4">
			<div class="flex items-center gap-3 mb-2">
				<UsersIcon class="h-8 w-8 text-primary" />
				<h1 class="text-2xl md:text-3xl font-bold tracking-tight">User Management</h1>
			</div>
		</div>

		<!-- Mobile-First User Dashboard -->
		<MobileUsersDashboard />
	</div>
{/if}
