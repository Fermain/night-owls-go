<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';

	// Icons
	import SearchIcon from '@lucide/svelte/icons/search';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import UserCheckIcon from '@lucide/svelte/icons/user-check';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import StarIcon from '@lucide/svelte/icons/star';
	import UserIcon from '@lucide/svelte/icons/user';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';

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

	function handleCreateUser() {
		goto('/admin/users/new');
	}

	function handleContactUser(phone: string, name?: string) {
		const message = `Hello ${name || 'there'}, this is Mount Moreland Night Owls admin...`;
		const whatsappUrl = `https://wa.me/${phone.replace(/[^0-9]/g, '')}?text=${encodeURIComponent(message)}`;
		window.open(whatsappUrl, '_blank');
	}

	function handleViewAnalytics() {
		goto('/admin/users/analytics');
	}
</script>

<!-- Mobile-First User Management Dashboard -->
<div class="space-y-4">
	{#if isLoading}
		<!-- Loading state -->
		<div class="space-y-4">
			<Skeleton class="h-16 w-full rounded-lg" />
			<div class="grid grid-cols-2 gap-3">
				{#each Array(4) as _, i (i)}
					<Skeleton class="h-20 rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-32 w-full rounded-lg" />
		</div>
	{:else if isError}
		<Card.Root class="p-4 border-destructive bg-destructive/5">
			<div class="flex items-center gap-3">
				<AlertTriangleIcon class="h-5 w-5 text-destructive" />
				<div>
					<p class="font-medium text-destructive">Error Loading Users</p>
					<p class="text-sm text-muted-foreground">Failed to load user data</p>
				</div>
			</div>
		</Card.Root>
	{:else}
		<!-- Quick Search -->
		<Card.Root class="p-4">
			<div class="relative">
				<SearchIcon class="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
				<Input
					placeholder="Search users by name or phone..."
					bind:value={searchTerm}
					class="pl-10"
				/>
			</div>
		</Card.Root>

		<!-- Search Results (if searching) -->
		{#if searchTerm.trim() && userCategories.searchResults.length > 0}
			<Card.Root>
				<Card.Header class="pb-3">
					<Card.Title class="flex items-center gap-2">
						<SearchIcon class="h-5 w-5" />
						Search Results
						<Badge variant="secondary">{userCategories.searchResults.length}</Badge>
					</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-2">
					{#each userCategories.searchResults as user (user.id)}
						<div
							class="flex items-center justify-between p-3 bg-muted/50 rounded-lg hover:bg-muted transition-colors"
						>
							<div class="flex items-center gap-3">
								{#if user.role === 'admin'}
									<ShieldIcon class="h-4 w-4 text-red-600" />
								{:else if user.role === 'owl'}
									<StarIcon class="h-4 w-4 text-yellow-600" />
								{:else}
									<UserIcon class="h-4 w-4 text-gray-600" />
								{/if}
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
									class="h-8"
								>
									View
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
		{:else if searchTerm.trim() && userCategories.searchResults.length === 0}
			<Card.Root class="p-6 text-center">
				<SearchIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
				<p class="text-muted-foreground">No users found matching "{searchTerm}"</p>
			</Card.Root>
		{/if}

		<!-- Quick Stats (when not searching) -->
		{#if !searchTerm.trim()}
			<div class="grid grid-cols-2 gap-3">
				<!-- Pending Approvals (Priority) -->
				<Card.Root
					class="p-4 cursor-pointer hover:bg-muted/50 transition-colors {userCategories
						.pendingGuests.length > 0
						? 'border-amber-500 bg-amber-50 dark:bg-amber-950/20'
						: ''}"
					onclick={() => {
						if (userCategories.pendingGuests.length > 0) {
							// Scroll to approvals section below
							document.getElementById('approvals-section')?.scrollIntoView({ behavior: 'smooth' });
						}
					}}
				>
					<div class="text-center">
						<div
							class="mx-auto w-10 h-10 bg-amber-100 dark:bg-amber-900/20 rounded-lg flex items-center justify-center mb-2"
						>
							<UserPlusIcon class="h-5 w-5 text-amber-600 dark:text-amber-400" />
						</div>
						<p class="text-2xl font-bold text-amber-600 dark:text-amber-400">
							{userCategories.pendingGuests.length}
						</p>
						<p class="text-xs text-muted-foreground">Pending Approvals</p>
					</div>
				</Card.Root>

				<!-- Active Volunteers -->
				<Card.Root class="p-4">
					<div class="text-center">
						<div
							class="mx-auto w-10 h-10 bg-green-100 dark:bg-green-900/20 rounded-lg flex items-center justify-center mb-2"
						>
							<StarIcon class="h-5 w-5 text-green-600 dark:text-green-400" />
						</div>
						<p class="text-2xl font-bold">
							{userCategories.owls.length}
						</p>
						<p class="text-xs text-muted-foreground">Active Volunteers</p>
					</div>
				</Card.Root>

				<!-- Admins -->
				<Card.Root class="p-4">
					<div class="text-center">
						<div
							class="mx-auto w-10 h-10 bg-red-100 dark:bg-red-900/20 rounded-lg flex items-center justify-center mb-2"
						>
							<ShieldIcon class="h-5 w-5 text-red-600 dark:text-red-400" />
						</div>
						<p class="text-2xl font-bold">
							{userCategories.admins.length}
						</p>
						<p class="text-xs text-muted-foreground">Administrators</p>
					</div>
				</Card.Root>

				<!-- Total Users -->
				<Card.Root class="p-4">
					<div class="text-center">
						<div
							class="mx-auto w-10 h-10 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center mb-2"
						>
							<UserIcon class="h-5 w-5 text-blue-600 dark:text-blue-400" />
						</div>
						<p class="text-2xl font-bold">
							{userCategories.total}
						</p>
						<p class="text-xs text-muted-foreground">Total Users</p>
					</div>
				</Card.Root>
			</div>

			<!-- Quick Actions -->
			<Card.Root class="p-4">
				<h3 class="font-medium mb-3">Quick Actions</h3>
				<div class="grid grid-cols-2 gap-3">
					<Button onclick={handleCreateUser} class="h-16 flex-col gap-1">
						<UserPlusIcon class="h-5 w-5" />
						<span class="text-xs">Add User</span>
					</Button>

					<Button variant="outline" onclick={handleViewAnalytics} class="h-16 flex-col gap-1">
						<BarChart3Icon class="h-5 w-5" />
						<span class="text-xs">View Analytics</span>
					</Button>

					<Button
						variant="outline"
						onclick={() => goto('/admin/users?role=owl')}
						class="h-16 flex-col gap-1"
					>
						<StarIcon class="h-5 w-5" />
						<span class="text-xs">Manage Volunteers</span>
					</Button>

					<Button
						variant="outline"
						onclick={() => goto('/admin/broadcasts')}
						class="h-16 flex-col gap-1"
					>
						<MessageSquareIcon class="h-5 w-5" />
						<span class="text-xs">Send Broadcast</span>
					</Button>
				</div>
			</Card.Root>

			<!-- PRIORITY: Pending User Approvals -->
			{#if userCategories.pendingGuests.length > 0}
				<Card.Root id="approvals-section" class="border-amber-500">
					<Card.Header class="pb-3">
						<Card.Title class="flex items-center gap-2">
							<UserPlusIcon class="h-5 w-5 text-amber-600" />
							🚨 Pending User Approvals
							<Badge variant="destructive">{userCategories.pendingGuests.length}</Badge>
						</Card.Title>
						<p class="text-sm text-muted-foreground">
							New users waiting for approval to become Night Owl volunteers
						</p>
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each userCategories.pendingGuests as guest (guest.id)}
							<div
								class="flex items-center justify-between p-3 bg-amber-50 dark:bg-amber-950/20 rounded-lg border border-amber-200 dark:border-amber-800"
							>
								<div class="flex items-center gap-3">
									<div
										class="h-10 w-10 bg-amber-200 dark:bg-amber-800 rounded-full flex items-center justify-center"
									>
										<UserIcon class="h-5 w-5 text-amber-700 dark:text-amber-300" />
									</div>
									<div>
										<p class="font-medium text-sm">{guest.name || 'Unnamed User'}</p>
										<p class="text-xs text-muted-foreground">{guest.phone}</p>
										{#if guest.created_at}
											<p class="text-xs text-muted-foreground">
												Registered: {new Date(guest.created_at).toLocaleDateString()}
											</p>
										{/if}
									</div>
								</div>
								<div class="flex gap-2">
									<Button
										size="sm"
										onclick={() => handleApproveUser(guest.id, guest.name ?? undefined)}
										disabled={approvalInProgress.has(guest.id)}
										class="h-8"
									>
										{#if approvalInProgress.has(guest.id)}
											Approving...
										{:else}
											✅ Approve
										{/if}
									</Button>
									<Button
										size="sm"
										variant="outline"
										onclick={() => handleViewUser(guest.id)}
										class="h-8"
									>
										View
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
				<Card.Root class="p-6 text-center border-green-200 bg-green-50 dark:bg-green-950/20">
					<UserCheckIcon class="h-8 w-8 mx-auto mb-2 text-green-600" />
					<p class="font-medium text-green-700 dark:text-green-300">All Caught Up!</p>
					<p class="text-sm text-green-600 dark:text-green-400">
						No pending user approvals at the moment
					</p>
				</Card.Root>
			{/if}
		{/if}
	{/if}
</div>
