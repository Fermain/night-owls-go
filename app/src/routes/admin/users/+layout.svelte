<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import PlusIcon from '@lucide/svelte/icons/plus-circle';
	import UserIcon from '@lucide/svelte/icons/user';
	import ShieldUserIcon from '@lucide/svelte/icons/shield-user';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { selectedUserForForm } from '$lib/stores/userEditingStore';
	import type { UserData } from '$lib/schemas/user';
	import BulkActionsToolbar from '$lib/components/admin/users/BulkActionsToolbar.svelte';
	import { UsersApiService } from '$lib/services/api';
	import { filterUsers } from '$lib/utils/userProcessing';

	let searchTerm = $state('');

	// Bulk actions state
	let bulkMode = $state(false);
	let selectedUserIds = $state<Set<number>>(new Set());

	// Define specific navigation items for the users section
	const usersNavItems = [
		{
			title: 'Dashboard',
			url: '/admin/users',
			icon: LayoutDashboardIcon
		}
		// "Create User" is now a separate button at the bottom
	];

	// Create a query for users using our API service
	const usersQuery = $derived(
		createQuery<UserData[], Error, UserData[], [string, string]>({
			queryKey: ['adminUsers', searchTerm],
			queryFn: () => UsersApiService.getAll(),
			staleTime: 1000 * 60 * 5, // 5 minutes
			gcTime: 1000 * 60 * 10, // 10 minutes
			retry: 2
		})
	);

	// Filtered users for display in sidebar
	const filteredUsers = $derived.by(() => {
		const users = $usersQuery.data ?? [];
		return filterUsers(users, searchTerm);
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

	// Bulk actions functions
	function toggleBulkMode() {
		bulkMode = !bulkMode;
		if (!bulkMode) {
			selectedUserIds.clear();
		}
	}

	function toggleUserSelection(userId: number, checked: boolean) {
		if (checked) {
			selectedUserIds.add(userId);
		} else {
			selectedUserIds.delete(userId);
		}
		// Trigger reactivity
		selectedUserIds = new Set(selectedUserIds);
	}

	function toggleSelectAll() {
		if (!$usersQuery.data) return;

		const allUserIds = $usersQuery.data.map((u) => u.id);
		const allSelected = allUserIds.every((id) => selectedUserIds.has(id));

		if (allSelected) {
			selectedUserIds.clear();
		} else {
			allUserIds.forEach((id) => selectedUserIds.add(id));
		}
		// Trigger reactivity
		selectedUserIds = new Set(selectedUserIds);
	}

	function onExitBulkMode() {
		bulkMode = false;
		selectedUserIds.clear();
	}

	function onClearSelection() {
		selectedUserIds.clear();
		selectedUserIds = new Set(selectedUserIds);
	}

	// Computed values for bulk actions
	const selectedUsers = $derived(
		$usersQuery.data?.filter((user) => selectedUserIds.has(user.id)) || []
	);

	const allUsersSelected = $derived(
		Boolean(
			$usersQuery.data?.length &&
				$usersQuery.data?.length > 0 &&
				$usersQuery.data.every((user) => selectedUserIds.has(user.id))
		)
	);

	const someUsersSelected = $derived(selectedUserIds.size > 0 && !allUsersSelected);

	let { children } = $props();
</script>

{#snippet userListContent()}
	<div class="flex flex-col h-full">
		{#if bulkMode}
			<BulkActionsToolbar
				{selectedUsers}
				allUsers={$usersQuery.data || []}
				{onExitBulkMode}
				_onClearSelection={onClearSelection}
			/>
		{/if}

		<!-- Top static nav items (Dashboard) -->
		{#each usersNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={page.url.pathname === '/admin/users' && !currentSelectedUserIdInStore}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Bulk Mode Toggle -->
		<div class="p-3 border-b">
			<div class="flex items-center space-x-2">
				<Switch id="bulk-mode" bind:checked={bulkMode} />
				<Label for="bulk-mode" class="text-sm font-medium cursor-pointer">Bulk Actions</Label>
			</div>
		</div>

		<!-- Select All (in bulk mode) -->
		{#if bulkMode && filteredUsers && filteredUsers.length > 0}
			<div class="p-3 border-b bg-muted/50">
				<div class="space-y-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<Checkbox
							checked={allUsersSelected}
							indeterminate={someUsersSelected}
							onCheckedChange={() => toggleSelectAll()}
						/>
						<span class="text-sm font-medium">
							{#if allUsersSelected}
								Deselect All
							{:else}
								Select All
							{/if}
						</span>
					</label>
					{#if selectedUserIds.size > 0}
						<div class="text-xs text-muted-foreground pl-6">
							{selectedUserIds.size} of {filteredUsers.length} selected
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- User list (potentially scrollable) -->
		<div class="flex-grow overflow-y-auto">
			{#if $usersQuery.isLoading}
				<div class="p-4 text-sm">Loading users...</div>
			{:else if $usersQuery.isError}
				<div class="p-4 text-sm text-destructive">
					Error loading users: {$usersQuery.error.message}
				</div>
			{:else if filteredUsers && filteredUsers.length > 0}
				{#each filteredUsers as user (user.id)}
					<div
						class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0 {currentSelectedUserIdInStore ===
							user.id && !bulkMode
							? 'active'
							: ''} {bulkMode && selectedUserIds.has(user.id)
							? 'bg-primary/10 border-primary/20'
							: ''}"
					>
						{#if bulkMode}
							<label class="flex items-center gap-2 cursor-pointer w-full">
								<Checkbox
									checked={selectedUserIds.has(user.id)}
									onCheckedChange={(checked) => toggleUserSelection(user.id, checked)}
								/>
								<div class="flex items-center gap-2 flex-grow">
									{#if user?.role === 'admin'}
										<ShieldUserIcon class="h-4 w-4" />
									{:else}
										<UserIcon class="h-4 w-4" />
									{/if}
									<span class="truncate">{user.name || 'Unnamed User'}</span>
								</div>
							</label>
						{:else}
							<a
								href={`/admin/users?userId=${user.id}`}
								class="flex items-center gap-2 w-full"
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
								<span class="truncate">{user.name || 'Unnamed User'}</span>
							</a>
						{/if}
					</div>
				{/each}
			{:else if $usersQuery.data}
				<div class="p-4 text-sm text-muted-foreground">
					{searchTerm ? `No users found matching "${searchTerm}".` : 'No users found.'}
				</div>
			{/if}
		</div>

		<!-- Create User button at the bottom -->
		<div class="p-3 border-t mt-auto">
			<Button
				href="/admin/users/new"
				class="w-full"
				variant={page.url.pathname === '/admin/users/new' ? 'default' : 'outline'}
			>
				<PlusIcon class="h-4 w-4 mr-2" />
				Create User
			</Button>
		</div>
	</div>
{/snippet}

<SidebarPage listContent={userListContent} title="Users" bind:searchTerm>
	{@render children()}
</SidebarPage>
