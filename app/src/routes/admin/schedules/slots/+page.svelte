<script lang="ts">
	import {
		createQuery,
		useQueryClient,
		createMutation
	} from '@tanstack/svelte-query';
	// import * as Table from '$lib/components/ui/table/index.js'; // Not currently used
	import { toast } from 'svelte-sonner';
	import { formatDistanceToNow } from 'date-fns';
	import { page } from '$app/stores'; // Correct import source
	// import * as Sidebar from '$lib/components/ui/sidebar/index.js'; // No longer directly used here for list rendering
	// import SidebarPage from '$lib/components/sidebar-page.svelte'; // Removed
	import { Button } from '$lib/components/ui/button'; // Assuming this is the correct path from other files
	import { Skeleton } from '$lib/components/ui/skeleton'; // Assuming path
	import { Calendar } from '$lib/components/ui/calendar';
	import {
		CalendarDate,
		today,
		getLocalTimeZone,
		startOfMonth,
		endOfMonth
	} from '@internationalized/date';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import * as Select from '$lib/components/ui/select'; // For User Select
	// For form validation
	import { authenticatedFetch } from '$lib/utils/api';
	import type { AdminShiftSlot } from '$lib/types'; // Import AdminShiftSlot
	import type { UserData } from '$lib/schemas/user'; // Import UserData for users list

	// --- Types ---
	// type AdminShiftSlot = { // REMOVED
	// 	schedule_id: number;
	// 	schedule_name: string;
	// 	start_time: string; // ISO date string
	// 	end_time: string; // ISO date string
	// 	timezone?: string | null;
	// 	is_booked: boolean;
	// 	booking_id?: number | null;
	// 	user_name?: string | null;
	// 	user_phone?: string | null;
	// };

	// User type for the dropdown - REPLACED with UserData
	// type User = { // REMOVED
	// 	id: number;
	// 	name: string | null;
	// 	phone: string;
	// 	role: string; 
	// };

	// --- State for selected shift ---
	let selectedShift = $state<AdminShiftSlot | null>(null);
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));

	// --- Booking Form State ---
	let selectedUserIdForBooking = $state<string | undefined>(undefined); // Store the string value from select
	let bookingFormError = $state<string | null>(null);
	const queryClient = useQueryClient();

	// --- Calendar State (new) ---
	let currentDisplayMonth = $state(today(getLocalTimeZone()));

	// Derived state for enabling usersQuery
	let isBookingFormEnabled = $derived(!!selectedShift && !selectedShift.is_booked);

	// --- Utility Functions (formatTimeSlot, formatRelativeTime, getAkaDescription are still useful) ---
	function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
		if (!startTimeIso || !endTimeIso) return 'N/A';
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);
			const options: Intl.DateTimeFormatOptions = {
				weekday: 'short',
				month: 'short',
				day: 'numeric',
				hour: 'numeric',
				minute: '2-digit',
				hour12: true
			};
			const startFormatted = startDate.toLocaleString(undefined, options);
			const endFormatted = endDate.toLocaleTimeString(undefined, {
				hour: 'numeric',
				minute: '2-digit',
				hour12: true
			});
			if (startDate.toDateString() === endDate.toDateString()) {
				return `${startFormatted.replace(startDate.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit', hour12: true }), '').trim()} - ${endFormatted}`;
			} else {
				const endDayFormatted = endDate.toLocaleString(undefined, {
					weekday: 'short',
					month: 'short',
					day: 'numeric'
				});
				return `${startFormatted} - ${endDayFormatted}, ${endFormatted}`;
			}
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

	// --- Data Fetching for all slots in a range (to find the selected one by ID/startTime) ---
	// This page still fetches a broad range of slots to be able to find the specific one selected from the sidebar.
	async function fetchAllSlotsData(from?: Date, to?: Date) {
		let isLoadingSlots = true;
		try {
			const params = new URLSearchParams();
			if (from) params.append('from', from.toISOString().split('T')[0]);
			if (to) params.append('to', to.toISOString().split('T')[0]);

			const response = await authenticatedFetch(`/api/admin/schedules/all-slots?${params.toString()}`);
			if (!response.ok) {
				let errorMsg = `HTTP error ${response.status}`;
				try {
					const errorData = await response.json();
					errorMsg = errorData.message || errorData.error || errorMsg;
				} catch (e) {
					/* ignore */
				}
				toast.error(`Failed to fetch shift slots for detail lookup: ${errorMsg}`);
				throw new Error(errorMsg);
			}
			return response.json() as Promise<AdminShiftSlot[]>;
		} finally {
			isLoadingSlots = false;
		}
	}

	const allSlotsForDetailLookupQuery = $derived.by(() => {
		const currentStartTime = shiftStartTimeFromUrl;
		return createQuery<AdminShiftSlot[], Error>({
			queryKey: ['allAdminShiftSlotsForSlotDetailPageLookup', currentStartTime],
			queryFn: () => fetchAllSlotsData(currentStartTime ? new Date(currentStartTime) : undefined),
			enabled: !!currentStartTime
		});
	});

	let shiftListForDetail = $derived($allSlotsForDetailLookupQuery.data ?? []);

	$effect(() => {
		const allSlots = $allSlotsForDetailLookupQuery.data;
		if (shiftStartTimeFromUrl && allSlots && allSlots.length > 0) {
			selectedShift =
				allSlots.find((s: AdminShiftSlot) => s.start_time === shiftStartTimeFromUrl) || null;
			selectedUserIdForBooking = undefined;
			bookingFormError = null;
		} else if (!shiftStartTimeFromUrl) {
			selectedShift = null; // Clear selection if URL param is removed
		}
		// If shiftStartTimeFromUrl is present but not found, selectedShift will be null (handled by find)
		// If query is loading, selectedShift might be null temporarily, which is fine.
	});

	// --- Data Fetching for Calendar View (new) ---
	const fetchShiftSlotsForCalendarMonth = async (
		monthDate: CalendarDate
	): Promise<AdminShiftSlot[]> => {
		const from = startOfMonth(monthDate).toDate(getLocalTimeZone()).toISOString();
		const to = endOfMonth(monthDate).toDate(getLocalTimeZone()).toISOString();
		// const response = await fetch(`/api/admin/schedules/all-slots?from=${from}&to=${to}`); // Should use authenticatedFetch
		const response = await authenticatedFetch(`/api/admin/schedules/all-slots?from=${from}&to=${to}`);
		if (!response.ok) {
			throw new Error('Failed to fetch shift slots for calendar');
		}
		return response.json();
	};

	const calendarMonthSlotsQuery = $derived.by(() => {
		const shouldBeEnabled = !shiftStartTimeFromUrl;
		return createQuery<AdminShiftSlot[], Error, AdminShiftSlot[], [string, string]>({
			queryKey: ['adminCalendarMonthSlots', currentDisplayMonth.toString()],
			queryFn: () => fetchShiftSlotsForCalendarMonth(currentDisplayMonth),
			enabled: shouldBeEnabled
		});
	});

	const shiftDatesForCalendarDisplay = $derived.by(() => {
		const slots = $calendarMonthSlotsQuery.data;
		if (!slots) return [];
		const uniqueDates = new Set<string>();
		slots.forEach((slot) => {
			const dateStr = slot.start_time.split('T')[0];
			uniqueDates.add(dateStr);
		});
		return Array.from(uniqueDates).map((dateStr) => {
			const [year, month, day] = dateStr.split('-').map(Number);
			return new CalendarDate(year, month, day);
		});
	});

	// Re-adding prevMonth and nextMonth functions
	function prevMonth() {
		currentDisplayMonth = currentDisplayMonth.add({ months: -1 });
	}

	function nextMonth() {
		currentDisplayMonth = currentDisplayMonth.add({ months: 1 });
	}

	// --- Data Fetching for Users (for select dropdown) ---
	async function fetchUsers() {
		try {
			const response = await authenticatedFetch('/api/admin/users');
			if (!response.ok) {
				throw new Error('Failed to fetch users');
			}
			return response.json() as Promise<UserData[]>; // Ensure return type matches UserData[]
		} catch (error) {
			toast.error('Failed to load users');
			console.error('Fetch users error:', error);
			throw error;
		}
	}
	const usersQuery = $derived.by(() => {
		const enabled = isBookingFormEnabled;
		return createQuery<UserData[], Error>({
			queryKey: ['allAdminUsersForBooking'],
			queryFn: fetchUsers,
			enabled: enabled
		});
	});

	// --- Booking Mutation ---
	type AdminBookingPayload = {
		schedule_id: number;
		start_time: string;
		user_id: number;
	};

	const bookShiftMutation = createMutation<any, Error, AdminBookingPayload>({
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
			queryClient.invalidateQueries({
				queryKey: ['allAdminShiftSlotsForSlotDetailPageLookup', shiftStartTimeFromUrl]
			});
			queryClient.invalidateQueries({ queryKey: ['adminCalendarMonthSlots'] });
			selectedUserIdForBooking = undefined;
		},
		onError: (error: Error) => {
			toast.error(`Assignment failed: ${error.message}`);
			bookingFormError = error.message;
		}
	});

	function handleBookShift(event: SubmitEvent) {
		// Added event type
		event.preventDefault(); // Manual preventDefault
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
</script>

<svelte:head>
	<title
		>Admin - {selectedShift
			? `Shift Details - ${selectedShift.schedule_name} @ ${new Date(selectedShift.start_time).toLocaleDateString()}`
			: 'Shifts Calendar View'}</title
	>
</svelte:head>

<!-- Main content area for selected shift details. No SidebarPage wrapper. -->
<div class="p-4 md:p-8">
	{#if shiftStartTimeFromUrl && $allSlotsForDetailLookupQuery.isLoading}
		<p>Loading shift details...</p>
	{:else if selectedShift}
		<div class="border rounded-lg shadow-sm mb-6">
			<div class="p-6">
				<h2 class="text-xl font-semibold mb-1">{selectedShift.schedule_name}</h2>
				<p class="text-sm text-muted-foreground mb-4">
					{formatTimeSlot(selectedShift.start_time, selectedShift.end_time)} ({formatRelativeTime(
						selectedShift.start_time
					)})
				</p>
				<div class="space-y-2">
					<p>
						<strong>Status:</strong>
						{#if selectedShift.is_booked}
							<span class="text-orange-600 font-semibold">Taken</span>
							{#if selectedShift.user_name || selectedShift.user_phone}
								<span class="text-xs text-muted-foreground ml-1">
									by: {selectedShift.user_name ?? 'N/A'}
									{#if selectedShift.user_phone}({selectedShift.user_phone}){/if}
								</span>
							{/if}
						{:else}
							<span class="text-green-600 font-semibold">Available</span>
						{/if}
					</p>
					{#if selectedShift.is_booked && (selectedShift.user_name || selectedShift.user_phone)}
						<div class="mt-1">
							<p class="text-sm">
								<strong class="font-medium">Assigned to: </strong>
								<span class="text-gray-700">{selectedShift.user_name ?? 'N/A'}</span>
								{#if selectedShift.user_phone}
									<span class="text-xs text-muted-foreground ml-1"
										>({selectedShift.user_phone})</span
									>
								{/if}
							</p>
							<!-- Future: Add Unassign/Reassign button here -->
						</div>
					{/if}
					<p>
						<strong>Timezone:</strong>
						{selectedShift.timezone || 'Not specified (defaults to schedule timezone)'}
					</p>
					<p><strong>AKA:</strong> {getAkaDescription(selectedShift.start_time) || 'N/A'}</p>
				</div>
			</div>
		</div>

		{#if !selectedShift.is_booked}
			<div class="border rounded-lg shadow-sm p-6">
				<h3 class="text-lg font-medium mb-4">Book Shift for User</h3>
				{#if $usersQuery.isLoading}
					<p>Loading users...</p>
				{:else if $usersQuery.isError}
					<p class="text-destructive">Error loading users: {$usersQuery.error.message}</p>
				{:else if $usersQuery.data}
					<form onsubmit={handleBookShift} class="space-y-4">
						<div>
							<Select.Root
								type="single"
								value={selectedUserIdForBooking}
								onValueChange={(val?: string) => {
									selectedUserIdForBooking = val;
									bookingFormError = null;
								}}
							>
								<Select.Trigger
									name="userIdForBookingSelect"
									class="w-full md:w-[280px]"
									placeholder="Select a user"
								/>
								<Select.Content>
									{#each $usersQuery.data as user (user.id)}
										<Select.Item value={user.id.toString()} label={user.name || user.phone}
											>{user.name || user.phone}</Select.Item
										>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						{#if bookingFormError}
							<p class="text-sm text-destructive">{bookingFormError}</p>
						{/if}
						<Button type="submit" disabled={$bookShiftMutation.isPending}>
							{#if $bookShiftMutation.isPending}Assigning...{:else}Assign Selected User{/if}
						</Button>
					</form>
				{/if}
			</div>
		{/if}
	{:else if shiftStartTimeFromUrl && $allSlotsForDetailLookupQuery.isError}
		<p class="text-destructive">
			Error loading data to find shift: {$allSlotsForDetailLookupQuery.error.message}
		</p>
	{:else if shiftStartTimeFromUrl && !$allSlotsForDetailLookupQuery.isLoading && !selectedShift}
		<p>Shift with start time {shiftStartTimeFromUrl} not found.</p>
	{:else}
		<div class="flex flex-col items-center">
			<h1 class="text-2xl font-semibold mb-6">All Shift Slots Calendar</h1>
			{#if $calendarMonthSlotsQuery.isLoading}
				<div class="flex justify-center items-center h-64">
					<Skeleton class="h-48 w-full max-w-md" />
				</div>
			{:else if $calendarMonthSlotsQuery.isError}
				<p class="text-destructive">
					Error loading shifts for calendar: {$calendarMonthSlotsQuery.error.message}
				</p>
			{:else}
				<div class="flex items-center gap-4 mb-4">
					<Button variant="outline" size="icon" onclick={prevMonth} aria-label="Previous month"
						><ChevronLeftIcon class="h-4 w-4" /></Button
					>
					<h2 class="text-xl font-medium">
						{currentDisplayMonth
							.toDate(getLocalTimeZone())
							.toLocaleString('default', { month: 'long', year: 'numeric' })}
					</h2>
					<Button variant="outline" size="icon" onclick={nextMonth} aria-label="Next month"
						><ChevronRightIcon class="h-4 w-4" /></Button
					>
				</div>
				<Calendar
					class="p-0 rounded-md border w-full max-w-3xl"
					type="multiple"
					value={shiftDatesForCalendarDisplay}
					bind:placeholder={currentDisplayMonth}
					weekdayFormat="long"
					readonly
				/>
			{/if}
		</div>
	{/if}
</div>
