<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';

	// Icons
	import SearchIcon from '@lucide/svelte/icons/search';
	import UserCheckIcon from '@lucide/svelte/icons/user-check';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import UserIcon from '@lucide/svelte/icons/user';
	import CheckIcon from '@lucide/svelte/icons/check';

	import { createUsersQuery } from '$lib/queries/admin/users/usersQuery';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { UsersApiService } from '$lib/services/api';

	// Search state
	let searchTerm = $state('');
	let approvalInProgress = $state<Set<number>>(new Set());

	// Reactive queries
	const usersQuery = $derived(createUsersQuery());

	// Derived states
	const isLoading = $derived($usersQuery.isLoading);
	const isError = $derived($usersQuery.isError);
	const usersData = $derived($usersQuery.data || []);

	// Calculate user categories
	const userCategories = $derived.by(() => {
		const pendingGuests = usersData.filter((u) => u.role === 'guest');
		const admins = usersData.filter((u) => u.role === 'admin');
		const owls = usersData.filter((u) => u.role === 'owl');

		// Simple search filter
		const searchFiltered = searchTerm.trim()
			? usersData.filter(
					(u) =>
						u.name?.toLowerCase().includes(searchTerm.toLowerCase()) || u.phone.includes(searchTerm)
				)
			: [];

		return {
			pendingGuests,
			admins,
			owls,
			searchResults: searchFiltered,
			total: usersData.length
		};
	});

	// Quick approval action
	async function handleApproveUser(userId: number, userName?: string) {
		if (approvalInProgress.has(userId)) return;

		try {
			approvalInProgress.add(userId);
			approvalInProgress = new Set(approvalInProgress);

			// Get current user data first since update requires all fields
			const currentUser = await UsersApiService.getById(userId);

			// Update user with new role while keeping existing data
			await UsersApiService.update(userId, {
				name: currentUser.name || '',
				phone: currentUser.phone,
				role: 'owl'
			});

			toast.success(`${userName ?? 'User'} approved as Night Owl volunteer!`);

			// Refresh the data
			$usersQuery.refetch();
		} catch (error) {
			toast.error('Failed to approve user');
			console.error('Approval error:', error);
		} finally {
			approvalInProgress.delete(userId);
			approvalInProgress = new Set(approvalInProgress);
		}
	}

	// Quick actions
	function handleViewUser(userId: number) {
		goto(`/admin/users?userId=${userId}`);
	}

	function handleContactUser(phone: string, name?: string) {
		const message = `Hello ${name || 'there'}, this is Mount Moreland Night Owls admin...`;
		const whatsappUrl = `https://wa.me/${phone.replace(/[^0-9]/g, '')}?text=${encodeURIComponent(message)}`;
		window.open(whatsappUrl, '_blank');
	}

	function getRoleIcon(role: string) {
		switch (role) {
			case 'admin':
				return ShieldIcon;
			case 'owl':
				return StarIcon;
			default:
				return UserIcon;
		}
	}

	function getRoleColor(role: string) {
		switch (role) {
			case 'admin':
				return 'text-red-600';
			case 'owl':
				return 'text-green-600';
			case 'guest':
				return 'text-amber-600';
			default:
				return 'text-gray-600';
		}
	}
</script>

<!-- Minimalist Mobile-First User Management -->
<div class="space-y-4">
	{#if isLoading}
		<!-- Loading state -->
		<div class="space-y-3">
			<Skeleton class="h-12 w-full" />
			<Skeleton class="h-8 w-full" />
			<Skeleton class="h-24 w-full" />
		</div>
	{:else if isError}
		<Card.Root class="p-4 border-red-200 bg-red-50 dark:bg-red-950/20">
			<div class="flex items-center gap-3">
				<AlertTriangleIcon class="h-5 w-5 text-red-600" />
				<div>
					<p class="font-medium text-red-700">Error Loading Users</p>
					<p class="text-sm text-red-600">Failed to load user data</p>
				</div>
			</div>
		</Card.Root>
	{:else}
		<!-- Compact Stats -->
		<div class="flex gap-4 text-center">
			<div class="flex-1">
				<p class="text-lg font-bold">{userCategories.total}</p>
				<p class="text-xs text-muted-foreground">Total</p>
			</div>
			<div class="flex-1">
				<p class="text-lg font-bold text-green-600">{userCategories.owls.length}</p>
				<p class="text-xs text-muted-foreground">Owls</p>
			</div>
			<div class="flex-1">
				<p class="text-lg font-bold text-amber-600">{userCategories.pendingGuests.length}</p>
				<p class="text-xs text-muted-foreground">Pending</p>
			</div>
			<div class="flex-1">
				<p class="text-lg font-bold text-red-600">{userCategories.admins.length}</p>
				<p class="text-xs text-muted-foreground">Admins</p>
			</div>
		</div>

		<!-- Search -->
		<div class="relative">
			<SearchIcon class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
			<Input placeholder="Search users by name or phone..." bind:value={searchTerm} class="pl-10" />
		</div>

		<!-- Search Results -->
		{#if searchTerm.trim()}
			{#if userCategories.searchResults.length > 0}
				<Card.Root>
					<Card.Header class="pb-3">
						<Card.Title class="text-base flex items-center gap-2">
							Search Results
							<Badge variant="secondary">{userCategories.searchResults.length}</Badge>
						</Card.Title>
					</Card.Header>
					<Card.Content class="space-y-2">
						{#each userCategories.searchResults as user (user.id)}
							{@const IconComponent = getRoleIcon(user.role)}
							<div class="flex items-center justify-between p-3 rounded border">
								<div class="flex items-center gap-3">
									<IconComponent class="h-4 w-4 {getRoleColor(user.role)}" />
									<div>
										<p class="font-medium text-sm">{user.name || 'Unnamed User'}</p>
										<p class="text-xs text-muted-foreground">{user.phone}</p>
									</div>
								</div>
								<div class="flex gap-1">
									<Button
										size="sm"
										variant="outline"
										onclick={() => handleViewUser(user.id)}
										class="h-8 text-xs"
									>
										Edit
									</Button>
									<Button
										size="sm"
										variant="outline"
										onclick={() => handleContactUser(user.phone, user.name ?? undefined)}
										class="h-8 w-8 p-0"
									>
										<PhoneIcon class="h-3 w-3" />
									</Button>
								</div>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			{:else}
				<div class="text-center py-8 text-muted-foreground">
					<SearchIcon class="h-6 w-6 mx-auto mb-2" />
					<p class="text-sm">No users found matching "{searchTerm}"</p>
				</div>
			{/if}
		{:else}
			<!-- Pending Approvals (Priority) -->
			{#if userCategories.pendingGuests.length > 0}
				<Card.Root class="border-amber-200 bg-amber-50 dark:bg-amber-950/20">
					<Card.Header class="pb-3">
						<Card.Title class="text-base flex items-center gap-2">
							<UserCheckIcon class="h-4 w-4 text-amber-600" />
							Pending Approvals
							<Badge class="bg-amber-600 text-white">{userCategories.pendingGuests.length}</Badge>
						</Card.Title>
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each userCategories.pendingGuests as guest (guest.id)}
							<div
								class="flex items-center justify-between p-3 bg-white dark:bg-amber-900/10 rounded border"
							>
								<div class="flex items-center gap-3">
									<div
										class="h-8 w-8 bg-amber-100 dark:bg-amber-800 rounded-full flex items-center justify-center"
									>
										<UserIcon class="h-4 w-4 text-amber-600" />
									</div>
									<div>
										<p class="font-medium text-sm">{guest.name || 'Unnamed User'}</p>
										<p class="text-xs text-muted-foreground">{guest.phone}</p>
									</div>
								</div>
								<div class="flex gap-1">
									<Button
										size="sm"
										onclick={() => handleApproveUser(guest.id, guest.name ?? undefined)}
										disabled={approvalInProgress.has(guest.id)}
										class="h-8 bg-green-600 hover:bg-green-700 text-white"
									>
										{#if approvalInProgress.has(guest.id)}
											<div
												class="h-3 w-3 border-2 border-white border-t-transparent rounded-full animate-spin"
											></div>
										{:else}
											<CheckIcon class="h-3 w-3" />
										{/if}
									</Button>
									<Button
										size="sm"
										variant="outline"
										onclick={() => handleViewUser(guest.id)}
										class="h-8 text-xs"
									>
										Edit
									</Button>
									<Button
										size="sm"
										variant="outline"
										onclick={() => handleContactUser(guest.phone, guest.name ?? undefined)}
										class="h-8 w-8 p-0"
									>
										<PhoneIcon class="h-3 w-3" />
									</Button>
								</div>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			{:else}
				<!-- No Pending Approvals -->
				<div class="text-center py-6 text-muted-foreground">
					<UserCheckIcon class="h-6 w-6 mx-auto mb-2 text-green-600" />
					<p class="text-sm font-medium text-green-700 dark:text-green-400">All Caught Up!</p>
					<p class="text-xs text-green-600 dark:text-green-500">No pending user approvals</p>
				</div>
			{/if}
		{/if}
	{/if}
</div>
