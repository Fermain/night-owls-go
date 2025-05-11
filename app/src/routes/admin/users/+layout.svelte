<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import PlusIcon from '@lucide/svelte/icons/plus-circle';
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldUserIcon from '@lucide/svelte/icons/shield-user';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { selectedUserForForm, type UserData } from '$lib/stores/userEditingStore';

	let searchTerm = $state('');

	// Define specific navigation items for the users section
	const usersNavItems = [
		{
			title: 'Dashboard',
			url: '/admin/users',
			icon: LayoutDashboardIcon
		}
		// "Create User" is now a separate button at the bottom
	];

	// Function to fetch users
	const fetchUsers = async (currentSearchTerm: string) => {
		let url = '/api/admin/users';
		if (currentSearchTerm) {
			url += `?search=${encodeURIComponent(currentSearchTerm)}`;
		}
		const response = await fetch(url);
		if (!response.ok) {
			throw new Error('Failed to fetch users');
		}
		return response.json() as Promise<UserData[]>;
	};

	// Create a query for users
	const usersQuery = $derived(
		createQuery<UserData[], Error, UserData[], [string, string]>({
			queryKey: ['adminUsers', searchTerm],
			queryFn: () => fetchUsers(searchTerm)
		})
	);

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
				// User ID in URL not found in list, clear selection
				if ($selectedUserForForm !== undefined) {
					selectedUserForForm.set(undefined);
				}
			}
		} else if (!userIdFromUrl) {
			// No userId in URL (could be dashboard, new, or default listing)
			// Clear selected user for form if any is set
			if ($selectedUserForForm !== undefined) {
				selectedUserForForm.set(undefined);
			}
		}
	});

	let { children } = $props();
</script>

{#snippet userListContent()}
	<div class="flex flex-col h-full">
		<!-- Top static nav items (Dashboard) -->
		{#each usersNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={page.url.pathname === '/admin/users' &&
					page.url.searchParams.get('view') === 'dashboard' &&
					!currentSelectedUserIdInStore}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- User list (potentially scrollable) -->
		<div class="flex-grow overflow-y-auto">
			{#if $usersQuery.isLoading}
				<div class="p-4 text-sm">Loading users...</div>
			{:else if $usersQuery.isError}
				<div class="p-4 text-sm text-destructive">
					Error loading users: {$usersQuery.error.message}
				</div>
			{:else if $usersQuery.data && $usersQuery.data.length > 0}
				{#each $usersQuery.data as user (user.id)}
					<a
						href={`/admin/users?userId=${user.id}`}
						class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0"
						class:active={currentSelectedUserIdInStore === user.id}
						onclick={(event) => {
							event.preventDefault();
							selectUserForEditing(user);
						}}
					>
						{#if user?.role === 'admin'}
							<ShieldUserIcon class="h-4 w-4" />
						{:else}
							<UserIcon class="h-4 w-4" />
						{/if}
						<span>{user.name || 'Unnamed User'}</span>
					</a>
				{/each}
			{:else if $usersQuery.data && $usersQuery.data.length === 0}
				<div class="p-4 text-sm text-muted-foreground">No users found.</div>
			{/if}
		</div>

		<!-- Create User button at the bottom -->
		<div class="p-3 border-t mt-auto">
			<Button
				href="/admin/users/new"
				class="w-full"
				variant={page.url.pathname === '/admin/users/new' ? 'default' : 'outline'}
			>
				<PlusIcon />
				Create User
			</Button>
		</div>
	</div>
{/snippet}

<SidebarPage listContent={userListContent} title="Users" bind:searchTerm>
	{@render children()}
</SidebarPage>
