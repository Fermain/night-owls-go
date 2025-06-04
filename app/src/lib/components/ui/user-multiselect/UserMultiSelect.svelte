<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { UsersApiService } from '$lib/services/api';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Popover from '$lib/components/ui/popover';
	import { Users, X, Search, Phone, User as UserIcon } from 'lucide-svelte';
	import type { UserData } from '$lib/schemas/user';

	// Props
	let {
		selectedUserIds = $bindable([]),
		placeholder = 'Select users...',
		maxDisplayItems = 3,
		variant = 'default'
	}: {
		selectedUserIds: (number | string)[];
		placeholder?: string;
		maxDisplayItems?: number;
		variant?: 'default' | 'compact';
	} = $props();

	// State
	let open = $state(false);
	let searchTerm = $state('');

	// Query users
	const usersQuery = createQuery<UserData[], Error>({
		queryKey: ['adminUsers'],
		queryFn: () => UsersApiService.getAll(),
		staleTime: 1000 * 60 * 5, // 5 minutes
		retry: 2
	});

	// Filtered users for search
	const filteredUsers = $derived.by(() => {
		const users = $usersQuery.data || [];
		if (!searchTerm.trim()) return users;

		const term = searchTerm.toLowerCase();
		return users.filter(
			(user) => user.name?.toLowerCase().includes(term) || user.phone.toLowerCase().includes(term)
		);
	});

	// Selected users data
	const selectedUsers = $derived.by(() => {
		const users = $usersQuery.data || [];
		return users.filter(
			(user) => selectedUserIds.includes(user.id) || selectedUserIds.includes(user.phone)
		);
	});

	// Toggle user selection
	function toggleUser(user: UserData) {
		const isSelected = selectedUserIds.includes(user.id) || selectedUserIds.includes(user.phone);

		if (isSelected) {
			// Remove user
			selectedUserIds = selectedUserIds.filter((id) => id !== user.id && id !== user.phone);
		} else {
			// Add user ID (preferred) or phone if no ID
			selectedUserIds = [...selectedUserIds, user.id || user.phone];
		}
	}

	// Remove selected user
	function removeUser(user: UserData) {
		selectedUserIds = selectedUserIds.filter((id) => id !== user.id && id !== user.phone);
	}

	// Clear all selections
	function clearAll() {
		selectedUserIds = [];
	}

	// Display text for trigger
	const displayText = $derived.by(() => {
		if (selectedUsers.length === 0) return placeholder;
		if (selectedUsers.length <= maxDisplayItems) {
			return selectedUsers.map((u) => u.name || u.phone).join(', ');
		}
		return `${selectedUsers.length} users selected`;
	});

	// Check if user is selected
	function isUserSelected(user: UserData): boolean {
		return selectedUserIds.includes(user.id) || selectedUserIds.includes(user.phone);
	}
</script>

<div class="space-y-2">
	<!-- Selected users display (for compact variant) -->
	{#if variant === 'compact' && selectedUsers.length > 0}
		<div class="flex flex-wrap gap-1">
			{#each selectedUsers.slice(0, maxDisplayItems) as user (user.id)}
				<Badge variant="secondary" class="text-xs">
					{user.name || user.phone}
					<button
						type="button"
						onclick={() => removeUser(user)}
						class="ml-1 hover:bg-destructive/20 rounded-full p-0.5"
					>
						<X class="h-2 w-2" />
					</button>
				</Badge>
			{/each}
			{#if selectedUsers.length > maxDisplayItems}
				<Badge variant="outline" class="text-xs">
					+{selectedUsers.length - maxDisplayItems} more
				</Badge>
			{/if}
		</div>
	{/if}

	<!-- Popover trigger -->
	<Popover.Root bind:open>
		<Popover.Trigger>
			{#snippet child({ props })}
				<Button
					{...props}
					variant="outline"
					role="combobox"
					aria-expanded={open}
					class="w-full justify-between text-left font-normal {variant === 'compact'
						? 'h-8 text-xs'
						: 'h-9 text-sm'}"
				>
					<div class="flex items-center gap-2 min-w-0">
						<Users class="{variant === 'compact' ? 'h-3 w-3' : 'h-4 w-4'} flex-shrink-0" />
						{#if variant === 'default' && selectedUsers.length > 0}
							<div class="flex flex-wrap gap-1 overflow-hidden">
								{#each selectedUsers.slice(0, maxDisplayItems) as user (user.id)}
									<Badge variant="secondary" class="text-xs">
										{user.name || user.phone}
										<button
											type="button"
											onclick={(e) => {
												e.stopPropagation();
												removeUser(user);
											}}
											class="ml-1 hover:bg-destructive/20 rounded-full p-0.5"
										>
											<X class="h-2 w-2" />
										</button>
									</Badge>
								{/each}
								{#if selectedUsers.length > maxDisplayItems}
									<Badge variant="outline" class="text-xs">
										+{selectedUsers.length - maxDisplayItems}
									</Badge>
								{/if}
							</div>
						{:else}
							<span class="truncate text-muted-foreground">
								{displayText}
							</span>
						{/if}
					</div>
				</Button>
			{/snippet}
		</Popover.Trigger>

		<Popover.Content class="w-[400px] p-0" align="start">
			<!-- Search header -->
			<div class="p-3 border-b">
				<div class="relative">
					<Search class="absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input placeholder="Search users..." bind:value={searchTerm} class="pl-8" />
				</div>

				{#if selectedUsers.length > 0}
					<div class="flex items-center justify-between mt-2">
						<span class="text-xs text-muted-foreground">
							{selectedUsers.length} selected
						</span>
						<Button onclick={clearAll} variant="ghost" size="sm" class="h-6 text-xs">
							Clear all
						</Button>
					</div>
				{/if}
			</div>

			<!-- Users list -->
			<div class="max-h-60 overflow-y-auto">
				{#if $usersQuery.isLoading}
					<div class="p-4 text-center text-sm text-muted-foreground">Loading users...</div>
				{:else if $usersQuery.isError}
					<div class="p-4 text-center text-sm text-destructive">
						Error loading users: {$usersQuery.error?.message}
					</div>
				{:else if filteredUsers.length === 0}
					<div class="p-4 text-center text-sm text-muted-foreground">
						{searchTerm ? 'No users found' : 'No users available'}
					</div>
				{:else}
					{#each filteredUsers as user (user.id)}
						<label class="flex items-center gap-3 p-3 hover:bg-accent cursor-pointer">
							<Checkbox checked={isUserSelected(user)} onCheckedChange={() => toggleUser(user)} />
							<div class="flex items-center gap-2 flex-1 min-w-0">
								{#if user.phone.startsWith('+')}
									<Phone class="h-3 w-3 text-muted-foreground flex-shrink-0" />
								{:else}
									<UserIcon class="h-3 w-3 text-muted-foreground flex-shrink-0" />
								{/if}
								<div class="min-w-0 flex-1">
									<div class="text-sm font-medium truncate">
										{user.name || 'Unnamed User'}
									</div>
									<div class="text-xs text-muted-foreground truncate">
										{user.phone}
									</div>
								</div>
								{#if user.role === 'admin'}
									<Badge variant="destructive" class="text-xs">Admin</Badge>
								{:else if user.role === 'owl'}
									<Badge variant="secondary" class="text-xs">Owl</Badge>
								{/if}
							</div>
						</label>
					{/each}
				{/if}
			</div>
		</Popover.Content>
	</Popover.Root>
</div>
