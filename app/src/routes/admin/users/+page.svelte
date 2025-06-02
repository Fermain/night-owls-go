<script lang="ts">
	import UserForm from '$lib/components/admin/users/UserForm.svelte';
	import IntelligentDashboard from '$lib/components/admin/IntelligentDashboard.svelte';
	import { selectedUserForForm } from '$lib/stores/userEditingStore';
	import type { UserData } from '$lib/schemas/user';
	import UsersIcon from '@lucide/svelte/icons/users';

	// currentUserForForm is derived from the store, which is synced with URL by the layout
	let currentUserForForm = $state<UserData | undefined>(undefined);
	selectedUserForForm.subscribe((value) => {
		currentUserForForm = value;
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
				<h1 class="text-2xl md:text-3xl font-bold tracking-tight">User Intelligence Center</h1>
			</div>
			<p class="text-base md:text-lg text-muted-foreground">
				Smart insights and quick actions for managing Night Owls community members
			</p>
		</div>

		<!-- Intelligent Dashboard for Users -->
		<IntelligentDashboard />
	</div>
{/if}
