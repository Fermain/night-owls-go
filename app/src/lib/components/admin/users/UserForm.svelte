<script lang="ts">
	// === IMPORTS ===
	// UI Components (centralized imports)
	import { Label, Input, Button, PhoneInput, LoadingState, ErrorState } from '$lib/components/ui';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';

	// Icons
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import UserIcon from '@lucide/svelte/icons/user';
	import { CalendarIcon, ClockIcon } from 'lucide-svelte';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';

	// Utilities with new patterns
	import { apiPost, apiPut, apiDelete, apiGet } from '$lib/utils/api';
	import { classifyError, getErrorMessage } from '$lib/utils/errors';
	import { formatTimeSlot } from '$lib/utils/datetime';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { format, isPast } from 'date-fns';

	// Types using our new domain types and API mappings
	import type { User, Booking, CreateUserData, UpdateUserData } from '$lib/types/domain';
	import type { BaseComponentProps } from '$lib/types/ui';
	import type { components } from '$lib/types/api';
	import {
		mapAPIUserToDomain,
		mapAPIBookingArrayToDomain,
		mapCreateUserToAPIRequest,
		mapUpdateUserToAPIRequest
	} from '$lib/types/api-mappings';

	// Legacy imports for dialogs (will migrate these later)
	import UserDeleteConfirmDialog from '$lib/components/admin/dialogs/UserDeleteConfirmDialog.svelte';
	import UserRoleChangeDialog from '$lib/components/admin/dialogs/UserRoleChangeDialog.svelte';

	// Phone input types
	import type { E164Number } from 'svelte-tel-input/types';

	// Validation
	import { userSchema, type UserFormValues } from '$lib/schemas/user';

	// === COMPONENT PROPS ===
	interface UserFormProps extends BaseComponentProps {
		user?: User;
		onSuccess?: (user: User) => void;
	}

	let { user, onSuccess, className, id, 'data-testid': testId, ...props }: UserFormProps = $props();

	// === STATE MANAGEMENT ===
	const queryClient = useQueryClient();

	// State for phone input validity
	let phoneInputValid = $state(true);

	// Local Svelte state for form data, initialized with user prop data if available
	let formData = $state<UserFormValues>({
		phone: (user?.phone as E164Number) || '',
		name: user?.name || null,
		role: user?.role || 'guest'
	});

	const roleDisplayValues = {
		admin: 'Admin',
		owl: 'Owl',
		guest: 'Guest'
	};

	let showRoleChangeDialog = $state(false);
	let showDeleteConfirm = $state(false);

	// State for Zod validation errors
	let zodErrors = $state<Partial<Record<keyof UserFormValues, string>>>({});

	// === EFFECTS ===
	$effect(() => {
		// When the user prop changes (e.g., selecting a different user to edit),
		// reset formData to the new user's data.
		if (user) {
			formData = {
				phone: (user.phone as E164Number) || '',
				name: user.name || null,
				role: user.role || 'guest'
			};
		} else {
			formData = {
				phone: '',
				name: null,
				role: 'guest'
			};
		}
		zodErrors = {};
	});

	// === SERVICE FUNCTIONS ===
	async function createUser(userData: CreateUserData): Promise<User> {
		const requestData = mapCreateUserToAPIRequest(userData);
		const apiResponse = await apiPost<
			typeof requestData,
			components['schemas']['api.UserAPIResponse']
		>('api/admin/users', requestData);
		return mapAPIUserToDomain(apiResponse);
	}

	async function updateUser(userId: number, userData: UpdateUserData): Promise<User> {
		// Ensure role is always defined for the API mapping
		const dataWithRole = {
			...userData,
			role: userData.role || ('guest' as const)
		};
		const requestData = mapUpdateUserToAPIRequest(dataWithRole);
		const apiResponse = await apiPut<
			typeof requestData,
			components['schemas']['api.UserAPIResponse']
		>(`api/admin/users/${userId}`, requestData);
		return mapAPIUserToDomain(apiResponse);
	}

	async function deleteUser(userId: number): Promise<void> {
		await apiDelete(`api/admin/users/${userId}`);
	}

	async function fetchUserBookings(userId: number): Promise<Booking[]> {
		const apiBookings = await apiGet<components['schemas']['api.BookingWithScheduleResponse'][]>(
			`api/admin/users/${userId}/bookings`
		);
		return mapAPIBookingArrayToDomain(apiBookings);
	}

	// === MUTATIONS ===
	// Save user mutation using our new service functions
	const saveUserMutation = createMutation({
		mutationFn: async (data: { payload: UserFormValues; userId?: number }) => {
			const { payload, userId } = data;

			const userData = {
				name: payload.name || '',
				phone: payload.phone as string,
				role: payload.role
			};

			if (userId) {
				return await updateUser(userId, { ...userData, id: userId });
			} else {
				return await createUser(userData);
			}
		},
		onSuccess: (result) => {
			toast.success(user?.id ? 'User updated successfully' : 'User created successfully');
			queryClient.invalidateQueries({ queryKey: ['adminUsers'] });

			if (onSuccess) {
				onSuccess(result);
			} else if (!user?.id) {
				// Reset form for new user creation
				formData = {
					phone: '',
					name: null,
					role: 'guest'
				};
				zodErrors = {};
			}
		},
		onError: (error: Error) => {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
		}
	});

	// Delete user mutation using our new service functions
	const deleteUserMutation = createMutation({
		mutationFn: deleteUser,
		onSuccess: () => {
			toast.success('User deleted successfully');
			queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
			showDeleteConfirm = false;
			goto('/admin/users');
		},
		onError: (error: Error) => {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
			showDeleteConfirm = false;
		}
	});

	function openRoleDialog() {
		showRoleChangeDialog = true;
	}

	function handleRoleConfirm(newRole: 'admin' | 'owl' | 'guest') {
		formData.role = newRole;
	}

	function validateForm(): boolean {
		const result = userSchema.safeParse(formData);
		if (!result.success) {
			const newErrors: Partial<Record<keyof UserFormValues, string>> = {};
			for (const issue of result.error.issues) {
				if (issue.path.length > 0) {
					newErrors[issue.path[0] as keyof UserFormValues] = issue.message;
				}
			}
			zodErrors = newErrors;
		} else {
			zodErrors = {};
		}

		if (!phoneInputValid && !zodErrors.phone) {
			zodErrors.phone = 'Invalid phone number format.';
		}

		return result.success && phoneInputValid;
	}

	function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		const currentUserIdFromProp = user?.id;

		if (formData.phone === '' || !phoneInputValid) {
			toast.error('Phone number is invalid or empty.');
			return;
		}

		const payloadForSubmit = {
			phone: formData.phone as E164Number,
			name: formData.name?.trim() === '' ? null : formData.name,
			role: formData.role
		};

		$saveUserMutation.mutate({
			payload: payloadForSubmit,
			userId: currentUserIdFromProp
		});
	}

	function handleDeleteClick() {
		if (user?.id) {
			showDeleteConfirm = true;
		}
	}

	// Fetch user bookings using our new service function
	const userBookingsQuery = $derived.by(() => {
		if (!user?.id) return null;

		return createQuery<Booking[], Error>({
			queryKey: ['userBookings', user.id],
			queryFn: () => fetchUserBookings(user.id!),
			staleTime: 1000 * 60 * 5, // 5 minutes
			retry: 2
		});
	});

	// Separate past and upcoming bookings
	const { pastBookings, upcomingBookings } = $derived.by(() => {
		if (!userBookingsQuery) {
			return { pastBookings: [], upcomingBookings: [] };
		}

		const queryData = $userBookingsQuery;
		if (!queryData || !queryData.data) {
			return { pastBookings: [], upcomingBookings: [] };
		}

		const bookings = queryData.data;
		const _now = new Date();

		const past = bookings.filter((booking: Booking) => isPast(new Date(booking.shiftStart)));
		const upcoming = bookings.filter((booking: Booking) => !isPast(new Date(booking.shiftStart)));

		// Sort past bookings by date (newest first)
		past.sort(
			(a: Booking, b: Booking) =>
				new Date(b.shiftStart).getTime() - new Date(a.shiftStart).getTime()
		);

		// Sort upcoming bookings by date (earliest first)
		upcoming.sort(
			(a: Booking, b: Booking) =>
				new Date(a.shiftStart).getTime() - new Date(b.shiftStart).getTime()
		);

		return { pastBookings: past, upcomingBookings: upcoming };
	});

	// Helper to safely access query properties
	const queryState = $derived.by(() => {
		if (!userBookingsQuery) {
			return { isLoading: false, isError: false, error: null, data: null };
		}
		const query = $userBookingsQuery;
		return {
			isLoading: query?.isLoading || false,
			isError: query?.isError || false,
			error: query?.error,
			data: query?.data || null
		};
	});

	function formatShiftDate(dateString: string): string {
		return format(new Date(dateString), 'EEEE, MMMM d, yyyy');
	}
</script>

<svelte:head>
	<title>{user?.id !== undefined ? 'Edit' : 'Create New'} User</title>
</svelte:head>

<div class="container mr-auto p-4">
	<AdminPageHeader
		icon={UserIcon}
		heading="{user?.id !== undefined ? 'Edit' : 'Create New'} User"
		subheading={user?.id !== undefined
			? 'Update user information and manage their account'
			: 'Add a new user to the Night Owls Control system'}
	/>

	{#if $saveUserMutation.isPending}
		<LoadingState isLoading={true} loadingText="Saving user..." />
	{:else if $saveUserMutation.isError}
		<ErrorState
			error={$saveUserMutation.error}
			title="Failed to save user"
			showRetry={true}
			onRetry={() => handleSubmit()}
		/>
	{:else}
		<form
			onsubmit={(event) => {
				event.preventDefault();
				handleSubmit();
			}}
			class="space-y-4"
		>
			<div>
				<Label for="phone" class="block mb-2">Phone Number</Label>
				<PhoneInput
					disabled={Boolean(user?.id)}
					readonly={Boolean(user?.id)}
					bind:value={formData.phone}
					bind:valid={phoneInputValid}
					required
				/>
				<p class="text-xs text-muted-foreground mt-1">
					We'll send verification codes to this number • Country: South Africa (ZA)
				</p>
				{#if zodErrors.phone}
					<p class="text-sm text-destructive mt-1">{zodErrors.phone}</p>
				{:else if !phoneInputValid && formData.phone !== ''}
					<p class="text-xs text-destructive mt-1">Please enter a valid phone number</p>
				{/if}
			</div>

			<div>
				<Label for="name" class="block mb-2">Name</Label>
				<Input
					id="name"
					type="text"
					bind:value={formData.name}
					class={zodErrors.name ? 'border-red-500' : ''}
				/>
				{#if zodErrors.name}
					<p class="text-sm text-destructive mt-1">{zodErrors.name}</p>
				{/if}
			</div>

			<div>
				<Label class="block mb-2">Role</Label>
				<div class="flex items-center gap-4">
					<Input disabled readonly value={roleDisplayValues[formData.role]} class="flex-grow" />
					<Button type="button" variant="outline" onclick={openRoleDialog}>Change Role</Button>
				</div>
				{#if zodErrors.role}
					<p class="text-sm text-destructive mt-1">{zodErrors.role}</p>
				{/if}
			</div>

			{#if user?.id !== undefined}
				<div class="text-sm text-muted-foreground">
					<Label>Created</Label>
					<time>
						{new Date(user.createdAt).toLocaleString()}
					</time>
				</div>
			{/if}

			<div class="flex gap-4">
				<Button type="submit" disabled={$saveUserMutation.isPending} class="flex-1">
					{#if $saveUserMutation.isPending}
						<Loader2Icon class="w-4 h-4 mr-2" />
						Saving...
					{:else}
						<UserPlusIcon class="w-4 h-4" />
						{user?.id !== undefined ? 'Update' : 'Create'} User
					{/if}
				</Button>
				<Button type="button" variant="outline" onclick={() => goto('/admin/users')}>Cancel</Button>
				{#if user?.id !== undefined}
					<Button
						type="button"
						variant="destructive"
						onclick={handleDeleteClick}
						disabled={$deleteUserMutation.isPending}
					>
						{#if $deleteUserMutation.isPending}Deleting...{:else}Delete User{/if}
					</Button>
				{/if}
			</div>
		</form>
	{/if}
</div>

<!-- Booked Shifts Section (only shown for existing users) -->
{#if user?.id}
	<div class="container mr-auto p-4 border-t">
		<h2 class="text-xl font-bold mb-4 flex items-center gap-2">
			<CalendarIcon class="h-5 w-5" />
			Booked Shifts
		</h2>

		{#if queryState.isLoading}
			<LoadingState isLoading={true} loadingText="Loading bookings..." />
		{:else if queryState.isError && queryState.error}
			<ErrorState
				error={queryState.error}
				title="Failed to load bookings"
				showRetry={true}
				onRetry={() => queryClient.invalidateQueries({ queryKey: ['userBookings', user.id] })}
			/>
		{:else if !queryState.data || queryState.data.length === 0}
			<div class="text-sm text-muted-foreground">No bookings found for this user.</div>
		{:else}
			<div class="space-y-6">
				<!-- Upcoming Bookings -->
				{#if upcomingBookings.length > 0}
					<div>
						<h3 class="text-lg font-semibold mb-3 text-blue-700">
							Upcoming Shifts ({upcomingBookings.length})
						</h3>
						<div class="grid gap-3">
							{#each upcomingBookings as booking (booking.id)}
								<div class="border rounded-lg p-4 bg-blue-50 border-blue-200">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<div class="font-medium text-sm">{booking.scheduleName}</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
												<CalendarIcon class="h-3 w-3" />
												{formatShiftDate(booking.shiftStart)}
											</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1">
												<ClockIcon class="h-3 w-3" />
												{formatTimeSlot(booking.shiftStart, booking.shiftEnd)}
											</div>
											{#if booking.buddyName}
												<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
													<UserIcon class="h-3 w-3" />
													Buddy: {booking.buddyName}
												</div>
											{/if}
										</div>
										<div class="text-xs text-blue-600 bg-blue-100 px-2 py-1 rounded">Upcoming</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Past Bookings -->
				{#if pastBookings.length > 0}
					<div>
						<h3 class="text-lg font-semibold mb-3 text-gray-700">
							Past Shifts ({pastBookings.length})
						</h3>
						<div class="grid gap-3">
							{#each pastBookings.slice(0, 10) as booking (booking.id)}
								<div class="border rounded-lg p-4 bg-gray-50 border-gray-200">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<div class="font-medium text-sm">{booking.scheduleName}</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
												<CalendarIcon class="h-3 w-3" />
												{formatShiftDate(booking.shiftStart)}
											</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1">
												<ClockIcon class="h-3 w-3" />
												{formatTimeSlot(booking.shiftStart, booking.shiftEnd)}
											</div>
											{#if booking.buddyName}
												<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
													<UserIcon class="h-3 w-3" />
													Buddy: {booking.buddyName}
												</div>
											{/if}
										</div>
										<div class="flex items-center gap-2">
											{#if booking.checkedInAt}
												<div
													class="flex items-center gap-1 text-xs text-green-600 bg-green-100 px-2 py-1 rounded"
												>
													<CheckCircleIcon class="h-3 w-3" />
													Checked In
												</div>
											{:else}
												<div class="text-xs text-gray-600 bg-gray-100 px-2 py-1 rounded">
													No Check-in
												</div>
											{/if}
										</div>
									</div>
								</div>
							{/each}
							{#if pastBookings.length > 10}
								<div class="text-sm text-muted-foreground text-center">
									Showing 10 of {pastBookings.length} past shifts
								</div>
							{/if}
						</div>
					</div>
				{/if}

				<!-- Summary -->
				<div class="bg-gray-50 rounded-lg p-4">
					<div class="text-sm text-muted-foreground">
						<strong>Total bookings:</strong>
						{queryState.data?.length || 0}
						| <strong>Upcoming:</strong>
						{upcomingBookings.length}
						| <strong>Past:</strong>
						{pastBookings.length}
						{#if pastBookings.some((b: Booking) => b.checkedInAt)}
							| <strong>Checked In:</strong>
							{pastBookings.filter((b: Booking) => b.checkedInAt).length}
						{/if}
						{#if pastBookings.some((b: Booking) => !b.checkedInAt)}
							| <strong>No Check-ins:</strong>
							{pastBookings.filter((b: Booking) => !b.checkedInAt).length}
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}

{#if showDeleteConfirm}
	<UserDeleteConfirmDialog
		bind:open={showDeleteConfirm}
		userName={user?.name ?? 'this user'}
		onConfirm={() => {
			if (user?.id) {
				$deleteUserMutation.mutate(user.id);
			}
		}}
		isLoading={$deleteUserMutation.isPending}
	/>
{/if}

{#if showRoleChangeDialog}
	<UserRoleChangeDialog
		bind:open={showRoleChangeDialog}
		userName={user?.name ?? 'this user'}
		bind:currentRole={formData.role}
		onConfirm={handleRoleConfirm}
	/>
{/if}
