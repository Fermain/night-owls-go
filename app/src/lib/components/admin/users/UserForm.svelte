<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import UserIcon from '@lucide/svelte/icons/user';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { PhoneInput } from '$lib/components/ui/phone-input';
	import type { E164Number } from 'svelte-tel-input/types';
	import { createSaveUserMutation } from '$lib/queries/admin/users/saveUserMutation';
	import { createDeleteUserMutation } from '$lib/queries/admin/users/deleteUserMutation';
	import { userSchema, type UserFormValues, type UserData } from '$lib/schemas/user';
	import UserDeleteConfirmDialog from '$lib/components/admin/dialogs/UserDeleteConfirmDialog.svelte';
	import UserRoleChangeDialog from '$lib/components/admin/dialogs/UserRoleChangeDialog.svelte';
	import { BookingsApiService, type BookingResponse } from '$lib/services/api/bookings';
	import { createQuery } from '@tanstack/svelte-query';
	import { CalendarIcon, ClockIcon } from 'lucide-svelte';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import { format, isPast } from 'date-fns';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';

	// Use $props() for Svelte 5 runes mode
	let { user }: { user?: UserData } = $props();

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

	$effect(() => {
		// When the user prop changes (e.g., selecting a different user to edit),
		// reset roleInDialog to the new user's current role.
		formData.role = user?.role || 'guest';
	});

	function openRoleDialog() {
		showRoleChangeDialog = true;
	}

	function handleRoleConfirm(newRole: 'admin' | 'owl' | 'guest') {
		formData.role = newRole;
	}

	// State for Zod validation errors
	let zodErrors = $state<Partial<Record<keyof UserFormValues, string>>>({});

	// State for controlling delete confirmation dialog
	let showDeleteConfirm = $state(false);

	const mutation = createSaveUserMutation();

	const deleteUserMutation = createDeleteUserMutation(() => {
		showDeleteConfirm = false;
	});

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
		console.log('UserForm handleSubmit - User object:', user);
		console.log('UserForm handleSubmit - User ID from prop:', currentUserIdFromProp);

		if (formData.phone === '' || !phoneInputValid) {
			toast.error('Phone number is invalid or empty.');
			return;
		}

		const payloadForSubmit = {
			phone: formData.phone as E164Number,
			name: formData.name?.trim() === '' ? null : formData.name,
			role: formData.role
		};

		console.log('UserForm handleSubmit - Payload:', payloadForSubmit);
		console.log('UserForm handleSubmit - UserId for mutation:', currentUserIdFromProp);

		$mutation.mutate({
			payload: payloadForSubmit,
			userId: currentUserIdFromProp
		});
	}

	function handleDeleteClick() {
		if (user?.id) {
			showDeleteConfirm = true;
		}
	}

	// Fetch user bookings for display (only when editing an existing user)
	const userBookingsQuery = $derived.by(() => {
		if (!user?.id) return null;

		return createQuery<BookingResponse[], Error>({
			queryKey: ['userBookings', user.id],
			queryFn: () => BookingsApiService.getUserBookings(user.id),
			staleTime: 1000 * 60 * 5 // 5 minutes
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

		const past = bookings.filter((booking: BookingResponse) =>
			isPast(new Date(booking.shift_start))
		);
		const upcoming = bookings.filter(
			(booking: BookingResponse) => !isPast(new Date(booking.shift_start))
		);

		// Sort past bookings by date (newest first)
		past.sort(
			(a: BookingResponse, b: BookingResponse) =>
				new Date(b.shift_start).getTime() - new Date(a.shift_start).getTime()
		);

		// Sort upcoming bookings by date (earliest first)
		upcoming.sort(
			(a: BookingResponse, b: BookingResponse) =>
				new Date(a.shift_start).getTime() - new Date(b.shift_start).getTime()
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

	function formatShiftTime(startTime: string, endTime: string): string {
		const start = format(new Date(startTime), 'HH:mm');
		const end = format(new Date(endTime), 'HH:mm');
		return `${start} - ${end}`;
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
			: 'Add a new user to the community watch system'}
	/>

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
					{new Date(user.created_at).toLocaleString()}
				</time>
			</div>
		{/if}

		<div class="flex gap-4">
			<Button type="submit" disabled={$mutation.isPending} class="flex-1">
				{#if $mutation.isPending}
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
</div>

<!-- Booked Shifts Section (only shown for existing users) -->
{#if user?.id}
	<div class="container mr-auto p-4 border-t">
		<h2 class="text-xl font-bold mb-4 flex items-center gap-2">
			<CalendarIcon class="h-5 w-5" />
			Booked Shifts
		</h2>

		{#if queryState.isLoading}
			<div class="text-sm text-muted-foreground">Loading bookings...</div>
		{:else if queryState.isError}
			<div class="text-sm text-destructive">
				Error loading bookings: {queryState.error?.message || 'Unknown error'}
			</div>
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
							{#each upcomingBookings as booking (booking.booking_id)}
								<div class="border rounded-lg p-4 bg-blue-50 border-blue-200">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<div class="font-medium text-sm">{booking.schedule_name}</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
												<CalendarIcon class="h-3 w-3" />
												{formatShiftDate(booking.shift_start)}
											</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1">
												<ClockIcon class="h-3 w-3" />
												{formatShiftTime(booking.shift_start, booking.shift_end)}
											</div>
											{#if booking.buddy_name}
												<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
													<UserIcon class="h-3 w-3" />
													Buddy: {booking.buddy_name}
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
							{#each pastBookings.slice(0, 10) as booking (booking.booking_id)}
								<div class="border rounded-lg p-4 bg-gray-50 border-gray-200">
									<div class="flex items-start justify-between">
										<div class="flex-1">
											<div class="font-medium text-sm">{booking.schedule_name}</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
												<CalendarIcon class="h-3 w-3" />
												{formatShiftDate(booking.shift_start)}
											</div>
											<div class="text-sm text-muted-foreground flex items-center gap-1">
												<ClockIcon class="h-3 w-3" />
												{formatShiftTime(booking.shift_start, booking.shift_end)}
											</div>
											{#if booking.buddy_name}
												<div class="text-sm text-muted-foreground flex items-center gap-1 mt-1">
													<UserIcon class="h-3 w-3" />
													Buddy: {booking.buddy_name}
												</div>
											{/if}
										</div>
										<div class="flex items-center gap-2">
											{#if booking.checked_in_at}
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
						{#if pastBookings.some((b: BookingResponse) => b.checked_in_at)}
							| <strong>Checked In:</strong>
							{pastBookings.filter((b: BookingResponse) => b.checked_in_at).length}
						{/if}
						{#if pastBookings.some((b: BookingResponse) => !b.checked_in_at)}
							| <strong>No Check-ins:</strong>
							{pastBookings.filter((b: BookingResponse) => !b.checked_in_at).length}
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
