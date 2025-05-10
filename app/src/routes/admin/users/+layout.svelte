<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import UsersIcon from '@lucide/svelte/icons/users';
	import PlusIcon from '@lucide/svelte/icons/plus-circle';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { selectedUserForForm, type UserData } from '$lib/stores/userEditingStore';

	// Define specific navigation items for the users section
	const usersNavItems = [
		{
			title: 'All Users',
			url: '/admin/users',
			icon: UsersIcon
		},
		{
			title: 'Create User',
			url: '/admin/users/new',
			icon: PlusIcon
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

	// Handle click for static nav items
	const handleStaticNavClick = (url: string) => {
		if (url === '/admin/users/new' || url === '/admin/users') {
			selectedUserForForm.set(undefined);
		}
		goto(url);
	};

	// Handle selecting a user from the dynamic list
	const selectUserForEditing = (user: UserData) => {
		selectedUserForForm.set(user);
		goto('/admin/users');
	};

	// Reactive variable to check if a user is selected for active highlighting
	let currentSelectedUserIdInStore = $state<number | undefined>(undefined);
	$effect(() => {
		const unsub = selectedUserForForm.subscribe(value => {
			currentSelectedUserIdInStore = value?.id;
		});
		return unsub;
	});

  let { children } = $props();
</script>

{#snippet userListContent()}
	<Sidebar.Group>
		<Sidebar.GroupContent>
			<Sidebar.Menu>
				{#each usersNavItems as item (item.title)}
					<Sidebar.MenuItem>
						<Sidebar.MenuButton onclick={() => handleStaticNavClick(item.url || '')}>
							<a
								href={item.url || undefined}
								class="flex items-center w-full h-full"
								class:active={page.url.pathname === item.url && !currentSelectedUserIdInStore}
							>
								<item.icon />
								<span>{item.title}</span>
							</a>
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				{/each}
				{#if $usersQuery.isLoading}
					<Sidebar.MenuItem>Loading users...</Sidebar.MenuItem>
				{:else if $usersQuery.isError}
					<Sidebar.MenuItem>Error loading users: {$usersQuery.error.message}</Sidebar.MenuItem>
				{:else if $usersQuery.data}
					{#each $usersQuery.data as user (user.id)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton onclick={() => selectUserForEditing(user)}>
								<a
									href={'/admin/users'} 
									class="flex items-center w-full h-full"
									class:active={currentSelectedUserIdInStore === user.id}
								>
									<span>{user.name || user.phone || `User ${user.id}`}</span>
								</a>
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				{/if}
			</Sidebar.Menu>
		</Sidebar.GroupContent>
	</Sidebar.Group>
{/snippet}

{#snippet pageSlotContent()}
	<div class="p-4">
		{@render children()}
	</div>
{/snippet}

<SidebarPage listContent={userListContent} children={pageSlotContent} /> 