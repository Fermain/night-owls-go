<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	// import * as Sidebar from '$lib/components/ui/sidebar/index.js'; // No longer used directly
	import UsersIcon from '@lucide/svelte/icons/users'; // Keep if icons might return, or remove if definitely not
	import PlusIcon from '@lucide/svelte/icons/plus-circle'; // Keep if icons might return, or remove if definitely not
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { selectedUserForForm, type UserData } from '$lib/stores/userEditingStore';

	// Define specific navigation items for the users section
	const usersNavItems = [
		{
			title: 'All Users',
			url: '/admin/users'
		},
		{
			title: 'Create User',
			url: '/admin/users/new'
		}
	];

	// Function to fetch users
	const fetchUsers = async () => {
		const response = await fetch('/api/admin/users');
		if (!response.ok) {
			throw new Error('Failed to fetch users');
		}
		return response.json() as Promise<UserData[]>;
	};

	// Create a query for users
	const usersQuery = createQuery<UserData[], Error, UserData[], string[]>({
		queryKey: ['adminUsers'],
		queryFn: fetchUsers
	});

	// Handle selecting a user from the dynamic list
	const selectUserForEditing = (user: UserData) => {
		goto(`/admin/users?userId=${user.id}`);
	};

	// Reactive variable to check if a user is selected for active highlighting
	let currentSelectedUserIdInStore = $state<number | undefined>(undefined);
	$effect(() => {
		const unsub = selectedUserForForm.subscribe((value) => {
			currentSelectedUserIdInStore = value?.id;
		});
		return unsub;
	});

	// Effect to synchronize the selectedUserForForm store with the userId URL query parameter
	$effect(() => {
		const userIdFromUrl = page.url.searchParams.get('userId');
		const users = $usersQuery.data;

		if (userIdFromUrl && users) {
			const userIdNum = parseInt(userIdFromUrl, 10);
			const userFromUrl = users.find((u) => u.id === userIdNum);
			const currentStoreUserId = $selectedUserForForm?.id;

			if (userFromUrl) {
				if (currentStoreUserId !== userIdNum) {
					selectedUserForForm.set(userFromUrl);
				}
			} else {
				if ($selectedUserForForm !== undefined) {
					selectedUserForForm.set(undefined);
				}
			}
		} else if (!userIdFromUrl) {
			if ($selectedUserForForm !== undefined) {
				selectedUserForForm.set(undefined);
			}
		}
	});

	let { children } = $props();
</script>

{#snippet userListContent()}
	<div class="flex flex-col">
		{#each usersNavItems as item (item.title)}
			<a
				href={item.url || undefined}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex flex-col items-start gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0"
				class:active={page.url.pathname === item.url &&
					!currentSelectedUserIdInStore &&
					!page.url.searchParams.has('userId')}
			>
				<span>{item.title}</span>
			</a>
		{/each}
		{#if $usersQuery.isLoading}
			<div class="p-4 text-sm">Loading users...</div>
		{:else if $usersQuery.isError}
			<div class="p-4 text-sm text-destructive">
				Error loading users: {$usersQuery.error.message}
			</div>
		{:else if $usersQuery.data}
			{#each $usersQuery.data as user (user.id)}
				<a
					href={`/admin/users?userId=${user.id}`}
					class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex flex-col items-start gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0"
					class:active={currentSelectedUserIdInStore === user.id}
					on:click={() => selectUserForEditing(user)}
				>
					<span>{user.name || 'Unnamed User'} [{user.phone}]</span>
				</a>
			{/each}
		{/if}
	</div>
{/snippet}

<SidebarPage listContent={userListContent} title="Users">
	{@render children()}
</SidebarPage>
