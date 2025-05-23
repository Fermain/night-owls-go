<script lang="ts">
	import { createQuery, useQueryClient, createMutation } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { formatDistanceToNow } from 'date-fns';
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Calendar } from '$lib/components/ui/calendar';
	import {
		CalendarDate,
		today,
		getLocalTimeZone
	} from '@internationalized/date';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import * as Select from '$lib/components/ui/select';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { AdminShiftSlot } from '$lib/types';
	import type { UserData } from '$lib/schemas/user';
	import { Label } from '$lib/components/ui/label';
	import CheckIcon from '@lucide/svelte/icons/check';
	import XIcon from '@lucide/svelte/icons/x';

	// State
	let selectedShift = $state<AdminShiftSlot | null>(null);
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));
	let selectedUserIdForBooking = $state<string | undefined>(undefined);
	let bookingFormError = $state<string | null>(null);
	const queryClient = useQueryClient();
	let currentDisplayMonth = $state(today(getLocalTimeZone()));
	let isBookingFormEnabled = $derived(!!selectedShift && !selectedShift.is_booked);

	// Utility Functions
	function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
		if (!startTimeIso || !endTimeIso) return 'N/A';
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);
			
			const startFormatted = startDate.toLocaleString('en-ZA', {
				weekday: 'short',
				day: 'numeric',
				month: 'short',
				hour: '2-digit',
				minute: '2-digit',
				hour12: false
			});
			
			const endFormatted = endDate.toLocaleTimeString('en-ZA', {
				hour: '2-digit',
				minute: '2-digit',
				hour12: false
			});
			
			return `${startFormatted} - ${endFormatted}`;
		} catch (e) {
			return 'Invalid Date Range';
		}
	}

	function formatRelativeTime(timeIso: string): string {
		if (!timeIso) return 'N/A';
		try {
			return formatDistanceToNow(new Date(timeIso), { addSuffix: true });
		} catch (e) {
			return 'Invalid Date';
		}
	}

	function getAkaDescription(startTimeIso: string): string {
		if (!startTimeIso) return '';
		try {
			const startDate = new Date(startTimeIso);
			const day = startDate.getDay();
			const hour = startDate.getHours();
			if (day === 6 && hour >= 0 && hour < 5) {
				return 'Friday Night';
			}
			if (day === 0 && hour >= 0 && hour < 5) {
				return 'Saturday Night';
			}
			return '';
		} catch (e) {
			return '';
		}
	}

	// Data Fetching
	async function fetchShiftDetails(startTime: string) {
		try {
			const response = await authenticatedFetch('/api/admin/schedules/all-slots');
			if (!response.ok) {
				throw new Error('Failed to fetch shift details');
			}
			const allShifts = (await response.json()) as AdminShiftSlot[];
			return allShifts.find((shift) => shift.start_time === startTime) || null;
		} catch (error) {
			console.error('Error fetching shift details:', error);
			throw error;
		}
	}

	async function fetchCalendarShifts() {
		try {
			const now = new Date();
			const futureDate = new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000);

			const params = new URLSearchParams();
			params.append('from', now.toISOString());
			params.append('to', futureDate.toISOString());

			const response = await authenticatedFetch(
				`/api/admin/schedules/all-slots?${params.toString()}`
			);
			if (!response.ok) {
				throw new Error('Failed to fetch calendar shifts');
			}
			return response.json() as Promise<AdminShiftSlot[]>;
		} catch (error) {
			console.error('Error fetching calendar shifts:', error);
			throw error;
		}
	}

	async function fetchUsers() {
		try {
			const response = await authenticatedFetch('/api/admin/users');
			if (!response.ok) {
				throw new Error('Failed to fetch users');
			}
			return response.json() as Promise<UserData[]>;
		} catch (error) {
			toast.error('Failed to load users');
			console.error('Fetch users error:', error);
			throw error;
		}
	}

	// Queries
	const shiftDetailsQuery = $derived(
		createQuery<AdminShiftSlot | null, Error>({
			queryKey: ['shiftDetails', shiftStartTimeFromUrl || ''],
			queryFn: () => fetchShiftDetails(shiftStartTimeFromUrl!),
			enabled: !!shiftStartTimeFromUrl
		})
	);

	const calendarShiftsQuery = $derived(
		createQuery<AdminShiftSlot[], Error>({
			queryKey: ['calendarShifts'],
			queryFn: fetchCalendarShifts,
			enabled: !shiftStartTimeFromUrl
		})
	);

	const usersQuery = $derived.by(() => {
		return createQuery<UserData[], Error>({
			queryKey: ['allAdminUsersForBooking'],
			queryFn: fetchUsers,
			enabled: isBookingFormEnabled
		});
	});

	// Calendar functions
	function prevMonth() {
		currentDisplayMonth = currentDisplayMonth.add({ months: -1 });
	}

	function nextMonth() {
		currentDisplayMonth = currentDisplayMonth.add({ months: 1 });
	}

	const shiftDatesForCalendar = $derived.by(() => {
		const shifts = $calendarShiftsQuery.data ?? [];
		const uniqueDates = new Set<string>();
		shifts.forEach((shift: AdminShiftSlot) => {
			const dateStr = shift.start_time.split('T')[0];
			uniqueDates.add(dateStr);
		});
		return Array.from(uniqueDates).map((dateStr) => {
			const [year, month, day] = dateStr.split('-').map(Number);
			return new CalendarDate(year, month, day);
		});
	});

	// Booking Mutation
	type AdminBookingPayload = {
		schedule_id: number;
		start_time: string;
		user_id: number;
	};

	type MutationResponse = {
		success: boolean;
		message?: string;
	};

	const bookShiftMutation = createMutation<MutationResponse, Error, AdminBookingPayload>({
		mutationFn: async (bookingData: AdminBookingPayload) => {
			const response = await authenticatedFetch('/api/admin/bookings/assign', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(bookingData)
			});
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: 'Failed to assign shift' }));
				throw new Error(errorData.message || `HTTP error ${response.status}`);
			}
			return response.json();
		},
		onSuccess: () => {
			toast.success('Shift assigned to user successfully!');
			queryClient.invalidateQueries({ queryKey: ['shiftDetails'] });
			selectedUserIdForBooking = undefined;
		},
		onError: (error: Error) => {
			toast.error(`Assignment failed: ${error.message}`);
			bookingFormError = error.message;
		}
	});

	function handleBookShift(event: SubmitEvent) {
		event.preventDefault();
		bookingFormError = null;
		const userIdToBook = selectedUserIdForBooking ? parseInt(selectedUserIdForBooking) : undefined;

		if (!selectedShift || !userIdToBook) {
			bookingFormError = 'Please select a user.';
			return;
		}
		$bookShiftMutation.mutate({
			schedule_id: selectedShift.schedule_id,
			start_time: selectedShift.start_time,
			user_id: userIdToBook
		});
	}

	// Effects
	$effect(() => {
		if ($shiftDetailsQuery.data) {
			selectedShift = $shiftDetailsQuery.data;
		} else if (!shiftStartTimeFromUrl) {
			selectedShift = null;
		}
	});
</script>

<svelte:head>
	<title>Admin - {selectedShift ? `Shift: ${selectedShift.schedule_name}` : 'Shift Calendar'}</title>
</svelte:head>

{#if shiftStartTimeFromUrl}
	<!-- Shift Detail View -->
	{#if $shiftDetailsQuery.isLoading}
		<div class="p-6">
			<Skeleton class="h-8 w-64 mb-4" />
			<Skeleton class="h-4 w-48 mb-6" />
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div>
					<Skeleton class="h-6 w-32 mb-4" />
					<div class="space-y-2">
						<Skeleton class="h-4 w-full" />
						<Skeleton class="h-4 w-3/4" />
						<Skeleton class="h-4 w-1/2" />
					</div>
				</div>
				<div>
					<Skeleton class="h-6 w-32 mb-4" />
					<Skeleton class="h-10 w-full mb-2" />
					<Skeleton class="h-10 w-full" />
				</div>
			</div>
		</div>
	{:else if $shiftDetailsQuery.isError}
		<div class="p-6">
			<div class="text-center">
				<h1 class="text-xl font-semibold text-destructive mb-2">Error Loading Shift</h1>
				<p class="text-muted-foreground">
					{$shiftDetailsQuery.error.message}
				</p>
			</div>
		</div>
	{:else if !selectedShift}
		<div class="p-6">
			<div class="text-center">
				<h1 class="text-xl font-semibold mb-2">Shift Not Found</h1>
				<p class="text-muted-foreground">The requested shift could not be found.</p>
			</div>
		</div>
	{:else}
		<!-- Shift Details -->
		<div class="p-6">
			<div class="max-w-4xl mx-auto">
				<h1 class="text-2xl font-semibold mb-2">{selectedShift.schedule_name}</h1>
				<p class="text-muted-foreground mb-6">
					{formatTimeSlot(selectedShift.start_time, selectedShift.end_time)}
					({formatRelativeTime(selectedShift.start_time)})
				</p>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<!-- Shift Information -->
					<div class="space-y-4">
						<h2 class="text-lg font-medium">Shift Details</h2>
						<div class="space-y-3">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium">Status:</span>
								{#if selectedShift.is_booked}
									<span class="inline-flex items-center gap-1 text-green-600 font-semibold">
										<CheckIcon class="h-4 w-4" />
										Filled
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 text-orange-600 font-semibold">
										<XIcon class="h-4 w-4" />
										Available
									</span>
								{/if}
							</div>

							{#if selectedShift.is_booked && selectedShift.user_name}
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium">Assigned to:</span>
									<span class="text-sm">{selectedShift.user_name}</span>
									{#if selectedShift.user_phone}
										<span class="text-xs text-muted-foreground">({selectedShift.user_phone})</span>
									{/if}
								</div>
							{/if}

							{#if getAkaDescription(selectedShift.start_time)}
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium">AKA:</span>
									<span class="text-sm">{getAkaDescription(selectedShift.start_time)}</span>
								</div>
							{/if}
						</div>
					</div>

					<!-- Booking Form -->
					{#if !selectedShift.is_booked}
						<div class="space-y-4">
							<h2 class="text-lg font-medium">Assign Shift</h2>
							{#if $usersQuery.isLoading}
								<p class="text-sm text-muted-foreground">Loading users...</p>
							{:else if $usersQuery.isError}
								<p class="text-sm text-destructive">
									Error loading users: {$usersQuery.error.message}
								</p>
							{:else if $usersQuery.data}
								<form onsubmit={handleBookShift} class="space-y-4">
									<div>
										<Label for="user-select" class="text-sm font-medium">Select User</Label>
										<Select.Root
											type="single"
											value={selectedUserIdForBooking}
											onValueChange={(val?: string) => {
												selectedUserIdForBooking = val;
												bookingFormError = null;
											}}
										>
											<Select.Trigger
												class="w-full mt-1"
												placeholder="Choose a user"
											></Select.Trigger>
											<Select.Content>
												{#each $usersQuery.data as user (user.id)}
													<Select.Item value={user.id.toString()} label={user.name || user.phone}>
														{user.name || user.phone}
													</Select.Item>
												{/each}
											</Select.Content>
										</Select.Root>
									</div>

									{#if bookingFormError}
										<p class="text-sm text-destructive">{bookingFormError}</p>
									{/if}

									<Button type="submit" disabled={$bookShiftMutation.isPending} class="w-full">
										{#if $bookShiftMutation.isPending}
											Assigning...
										{:else}
											Assign User to Shift
										{/if}
									</Button>
								</form>
							{/if}
						</div>
					{:else}
						<div class="space-y-4">
							<h2 class="text-lg font-medium">Shift Management</h2>
							<p class="text-sm text-muted-foreground">
								This shift is currently assigned. Future options for reassignment or cancellation
								will appear here.
							</p>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
{:else}
	<!-- Calendar Dashboard View -->
	<div class="p-6">
		<div class="max-w-4xl mx-auto">
			<div class="text-center mb-6">
				<h1 class="text-2xl font-semibold mb-2">Shift Calendar Dashboard</h1>
				<p class="text-muted-foreground">
					View all scheduled shifts and select one from the sidebar for details
				</p>
			</div>

			{#if $calendarShiftsQuery.isLoading}
				<div class="flex justify-center items-center h-64">
					<Skeleton class="h-48 w-full max-w-md" />
				</div>
			{:else if $calendarShiftsQuery.isError}
				<div class="text-center">
					<p class="text-destructive">
						Error loading shifts for calendar: {$calendarShiftsQuery.error.message}
					</p>
				</div>
			{:else}
				<div class="flex justify-center">
					<div class="w-full max-w-3xl">
						<div class="flex items-center justify-center gap-4 mb-4">
							<Button variant="outline" size="icon" onclick={prevMonth} aria-label="Previous month">
								<ChevronLeftIcon class="h-4 w-4" />
							</Button>
							<h2 class="text-xl font-medium">
								{currentDisplayMonth
									.toDate(getLocalTimeZone())
									.toLocaleString('default', { month: 'long', year: 'numeric' })}
							</h2>
							<Button variant="outline" size="icon" onclick={nextMonth} aria-label="Next month">
								<ChevronRightIcon class="h-4 w-4" />
							</Button>
						</div>

						<Calendar
							class="p-0 rounded-md border w-full"
							type="multiple"
							value={shiftDatesForCalendar}
							bind:placeholder={currentDisplayMonth}
							weekdayFormat="long"
							readonly
						/>

						<div class="mt-4 text-center text-sm text-muted-foreground">
							Highlighted dates have scheduled shifts.
							{$calendarShiftsQuery.data?.length ?? 0} total shifts in the next 90 days.
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if} 