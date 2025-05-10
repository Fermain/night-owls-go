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

	$effect(() => {
		// Log the data when it's available or changes
		if ($usersQuery.data) {
			console.log('Admin Users Data:', $usersQuery.data);
		}
	});

	// Handle selecting a user from the dynamic list
	const selectUserForEditing = (user: UserData) => {
		goto(`/admin/users?userId=${user.id}`);
	};

	// Reactive variable to check if a user is selected for active highlighting
	let currentSelectedUserIdInStore = $state<number | undefined>(undefined);
	$effect(() => {
		const unsub = selectedUserForForm.subscribe(value => {
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
			const userFromUrl = users.find(u => u.id === userIdNum);

			const currentStoreUserId = $selectedUserForForm?.id;

			if (userFromUrl) {
				if (currentStoreUserId !== userIdNum) {
					selectedUserForForm.set(userFromUrl);
				}
			} else {
				// User ID in URL but not found (e.g. invalid ID, or list not fully loaded yet for a deep link)
				if ($selectedUserForForm !== undefined) {
					selectedUserForForm.set(undefined);
				}
				// Optional: if truly invalid and not just a loading race, clear URL
				// if (page.url.pathname === '/admin/users') { // only if on the main users page
				// goto('/admin/users', { replaceState: true, noScroll: true });
				// }
			}
		} else if (!userIdFromUrl) {
			// No userId in URL (e.g. /admin/users or /admin/users/new)
			if ($selectedUserForForm !== undefined) {
				selectedUserForForm.set(undefined);
			}
		}
		// This effect depends on page.url and $usersQuery.data
		// Access them to ensure reactivity if not already done: page.url; $usersQuery.data;
	});

  let { children } = $props();
</script>

{#snippet userListContent()}
	<Sidebar.Group class="p-0">
		<Sidebar.GroupContent>
			<Sidebar.Menu class="gap-0">
				{#each usersNavItems as item (item.title)}
					<!-- <Sidebar.MenuItem> -->
						<!-- <Sidebar.MenuButton onclick={() => handleStaticNavClick(item.url || '')}> -->
							<a
								href={item.url || undefined}
								class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex flex-col items-start gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0"
								class:active={page.url.pathname === item.url && !currentSelectedUserIdInStore && !page.url.searchParams.has('userId')}
							>
								<!-- <item.icon /> -->
								<span>{item.title}</span>
							</a>
						<!-- </Sidebar.MenuButton> -->
					<!-- </Sidebar.MenuItem> -->
				{/each}
				{#if $usersQuery.isLoading}
					<Sidebar.MenuItem>Loading users...</Sidebar.MenuItem>
				{:else if $usersQuery.isError}
					<Sidebar.MenuItem>Error loading users: {$usersQuery.error.message}</Sidebar.MenuItem>
				{:else if $usersQuery.data}
					{#each $usersQuery.data as user (user.id)}
						<!-- <Sidebar.MenuItem> -->
							<!-- <Sidebar.MenuButton onclick={() => selectUserForEditing(user)}> -->
								<a
									href={`/admin/users?userId=${user.id}`}
									class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex flex-col items-start gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0"
									class:active={currentSelectedUserIdInStore === user.id}
                  onclick={() => selectUserForEditing(user)}
								>
									<span>{user.name} [{user.phone}]</span>
								</a>
							<!-- </Sidebar.MenuButton> -->
						<!-- </Sidebar.MenuItem> -->
					{/each}
				{/if}
			</Sidebar.Menu>
		</Sidebar.GroupContent>
	</Sidebar.Group>
{/snippet}

<SidebarPage listContent={userListContent} title="Users">
  {@render children()}
</SidebarPage>