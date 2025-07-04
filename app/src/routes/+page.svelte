<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import {
		createQuery,
		createMutation,
		type CreateQueryResult,
		type CreateMutationResult
	} from '@tanstack/svelte-query';
	import { canCancelBooking } from '$lib/utils/bookings';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import XIcon from '@lucide/svelte/icons/x';
	import { userSession } from '$lib/stores/authStore';
	import { selectedDayRange, getShiftDateRange } from '$lib/stores/shiftFilterStore';
	import {
		UserApiService,
		type AvailableShiftSlot,
		type CreateBookingRequest,
		type UserBooking
	} from '$lib/services/api/user';
	import { toast } from 'svelte-sonner';
	import CompactShiftCard from '$lib/components/user/shifts/CompactShiftCard.svelte';
	import BookingConfirmationDialog from '$lib/components/user/bookings/BookingConfirmationDialog.svelte';
	import CancellationConfirmationDialog from '$lib/components/user/bookings/CancellationConfirmationDialog.svelte';
	import BulkAssignDialog from '$lib/components/user/bookings/BulkAssignDialog.svelte';
	import ShiftCalendar from '$lib/components/user/shifts/ShiftCalendar.svelte';
	import { onMount } from 'svelte';
	import MyReportsWidget from '$lib/components/user/dashboard/MyReportsWidget.svelte';
	import { getPageOpenGraph } from '$lib/utils/opengraph';

	// OpenGraph tags for this page
	const ogTags = getPageOpenGraph('home');

	// Get current user from auth store
	const currentUser = $derived($userSession);

	// Get day range from persistent store
	const dayRange = $derived($selectedDayRange);

	// State for booking confirmation dialog
	let showBookingDialog = $state(false);
	let selectedShift = $state<AvailableShiftSlot | null>(null);

	// State for cancellation confirmation dialog
	let showCancelDialog = $state(false);
	let shiftToCancel = $state<{ id: number; details: string } | null>(null);

	// State for bulk assign dialog
	let showBulkAssignDialog = $state(false);

	// Shift limit for pagination in the list view only (not API limit)
	let displayShiftLimit = $state(10);

	// Query states - will be initialized in onMount with proper types
	let availableShiftsQuery = $state<CreateQueryResult<AvailableShiftSlot[], Error> | null>(null);
	let userBookingsQuery = $state<CreateQueryResult<UserBooking[], Error> | null>(null);
	let bookingMutation = $state<CreateMutationResult<
		UserBooking,
		Error,
		CreateBookingRequest,
		unknown
	> | null>(null);
	let cancelBookingMutation = $state<CreateMutationResult<void, Error, number, unknown> | null>(
		null
	);

	// State to track if component is mounted to prevent Dialog lifecycle errors
	let mounted = $state(false);

	// Initialize queries after component is mounted to avoid lifecycle errors
	onMount(() => {
		// Query for available shifts - reactive to dateRange changes
		availableShiftsQuery = createQuery({
			queryKey: ['available-shifts', dayRange],
			queryFn: async () => {
				const { from, to } = getShiftDateRange(dayRange);
				const result = await UserApiService.getAvailableShifts({ from, to });
				return result;
			}
		});

		// Query for user's bookings (only if authenticated)
		userBookingsQuery = createQuery({
			queryKey: ['user-bookings'],
			queryFn: async () => {
				if (!$userSession.isAuthenticated) {
					throw new Error('User not authenticated');
				}
				const result = await UserApiService.getMyBookings();
				return result;
			},
			enabled: $userSession.isAuthenticated,
			retry: false
		});

		// Mutations for booking
		bookingMutation = createMutation({
			mutationFn: (request: CreateBookingRequest) => UserApiService.createBooking(request),
			onSuccess: () => {
				toast.success('Shift committed successfully!');
				$availableShiftsQuery?.refetch();
				$userBookingsQuery?.refetch();
				showBookingDialog = false;
				selectedShift = null;
			},
			onError: (error: Error) => {
				toast.error(`Failed to commit to shift: ${error.message}`);
			}
		});

		// Mutation for canceling booking
		cancelBookingMutation = createMutation({
			mutationFn: (bookingId: number) => UserApiService.cancelBooking(bookingId),
			onSuccess: () => {
				toast.success('Shift cancelled successfully!');
				$userBookingsQuery?.refetch();
				$availableShiftsQuery?.refetch();
				showCancelDialog = false;
				shiftToCancel = null;
			},
			onError: (error: Error) => {
				toast.error(`Failed to cancel shift: ${error.message}`);
				showCancelDialog = false;
				shiftToCancel = null;
			}
		});

		// Mark component as mounted to allow Dialog components to be rendered
		mounted = true;
	});

	// Derived data - with null checks since queries are initialized in onMount
	const availableShifts = $derived(($availableShiftsQuery?.data as AvailableShiftSlot[]) ?? []);
	const userBookings = $derived(($userBookingsQuery?.data as UserBooking[]) ?? []);

	// Calculate how many shifts to display based on shiftLimit, but always show at least 5
	const displayLimit = $derived(Math.max(5, Math.min(displayShiftLimit, availableShifts.length)));
	const displayedShifts = $derived(availableShifts.slice(0, displayLimit));

	// Event handlers
	function handleShowMoreShifts() {
		displayShiftLimit = displayShiftLimit + 10; // Increase limit by 10
	}

	// Bulk assign handlers
	function handleBulkAssignSuccess() {
		$availableShiftsQuery?.refetch();
		$userBookingsQuery?.refetch();
	}

	function handleBulkAssignCancel() {
		// No specific action needed
	}

	// Find next shift from user bookings
	const nextShift = $derived.by(() => {
		if (!$userBookingsQuery?.data) return null;

		const now = new Date();
		const upcomingBookings = ($userBookingsQuery.data as UserBooking[])
			.filter((booking: UserBooking) => new Date(booking.shift_start) > now)
			.sort(
				(a: UserBooking, b: UserBooking) =>
					new Date(a.shift_start).getTime() - new Date(b.shift_start).getTime()
			);

		if (upcomingBookings.length === 0) return null;

		const booking = upcomingBookings[0];
		const startTime = new Date(booking.shift_start);
		const endTime = new Date(booking.shift_end);
		const canCheckin = startTime.getTime() - now.getTime() <= 30 * 60 * 1000; // 30 min before
		const isActive = now >= startTime && now <= endTime;

		return {
			id: booking.booking_id,
			start_time: booking.shift_start,
			end_time: booking.shift_end,
			buddy_name: booking.buddy_name,
			schedule_name: booking.schedule_name,
			can_checkin: canCheckin,
			is_active: isActive
		};
	});

	// Find additional upcoming shifts (after the next one)
	const additionalShifts = $derived.by(() => {
		if (!$userBookingsQuery?.data) return [];

		const now = new Date();
		const upcomingBookings = ($userBookingsQuery.data as UserBooking[])
			.filter((booking: UserBooking) => new Date(booking.shift_start) > now)
			.sort(
				(a: UserBooking, b: UserBooking) =>
					new Date(a.shift_start).getTime() - new Date(b.shift_start).getTime()
			);

		// Return all upcoming shifts except the first one (which is the "next shift")
		return upcomingBookings.slice(1).slice(0, 3); // Show up to 3 additional shifts
	});

	// Event handlers
	function handleCheckIn() {
		toast.success('Checked in successfully!');
	}

	function handleCheckOut() {
		toast.success('Checked out successfully!');
	}

	function handleCancelShift(shiftId: number) {
		// Find the shift details for the confirmation dialog
		let shiftDetails = '';

		if (nextShift && nextShift.id === shiftId) {
			shiftDetails = formatShiftTimeFromBooking(nextShift);
		} else {
			const additionalShift = additionalShifts.find(
				(shift: UserBooking) => shift.booking_id === shiftId
			);
			if (additionalShift) {
				shiftDetails = formatShiftTimeCompact(additionalShift);
			}
		}

		shiftToCancel = { id: shiftId, details: shiftDetails };
		showCancelDialog = true;
	}

	function handleBookShift(shift: AvailableShiftSlot) {
		selectedShift = shift;
		showBookingDialog = true;
	}

	function handleBookingConfirm(request: CreateBookingRequest) {
		$bookingMutation?.mutate(request);
	}

	function handleBookingCancel() {
		showBookingDialog = false;
		selectedShift = null;
	}

	function handleCancellationConfirm() {
		if (shiftToCancel) {
			$cancelBookingMutation?.mutate(shiftToCancel.id);
		}
	}

	function handleCancellationCancel() {
		showCancelDialog = false;
		shiftToCancel = null;
	}

	function formatShiftTimeFromBooking(shift: { start_time: string; end_time: string }) {
		const start = new Date(shift.start_time);
		const end = new Date(shift.end_time);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);

		let dateLabel = '';
		if (start.toDateString() === today.toDateString()) {
			dateLabel = 'Today';
		} else if (start.toDateString() === tomorrow.toDateString()) {
			dateLabel = 'Tomorrow';
		} else {
			dateLabel = start.toLocaleDateString('en-GB', {
				weekday: 'short',
				month: 'short',
				day: 'numeric'
			});
		}

		const timeRange = `${start.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		})} - ${end.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		})}`;

		return `${dateLabel} • ${timeRange}`;
	}

	function formatShiftTimeCompact(booking: { shift_start: string; shift_end: string }) {
		const start = new Date(booking.shift_start);
		const end = new Date(booking.shift_end);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);
		const dayAfterTomorrow = new Date(today);
		dayAfterTomorrow.setDate(today.getDate() + 2);

		let dateLabel = '';
		if (start.toDateString() === today.toDateString()) {
			dateLabel = 'Today';
		} else if (start.toDateString() === tomorrow.toDateString()) {
			dateLabel = 'Tomorrow';
		} else if (start.toDateString() === dayAfterTomorrow.toDateString()) {
			dateLabel = start.toLocaleDateString('en-GB', { weekday: 'short' });
		} else {
			dateLabel = start.toLocaleDateString('en-GB', {
				month: 'short',
				day: 'numeric'
			});
		}

		const timeRange = `${start.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		})} - ${end.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		})}`;

		return `${dateLabel} • ${timeRange}`;
	}

	// Check if we're loading more shifts (when shiftLimit increases)
	const isLoadingMore = $derived($availableShiftsQuery?.isFetching && displayedShifts.length > 0);

	// Check if there are more shifts to load
	const hasMoreShifts = $derived(
		availableShifts.length > displayedShifts.length ||
			($availableShiftsQuery?.data?.length === displayShiftLimit && displayShiftLimit < 100)
	); // Assume more might be available if we hit the limit
</script>

<svelte:head>
	<title>{ogTags.title}</title>
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.description}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImageAlt}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogType}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogSiteName}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterCard}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImageAlt}
</svelte:head>

<div class="bg-background flex-1">
	{#if currentUser.isAuthenticated}
		<!-- Authenticated Dashboard -->
		<div class="p-2 sm:p-4 space-y-4">
			<!-- My Next/Active Shift -->
			{#if nextShift}
				<CompactShiftCard
					shift={nextShift}
					type={nextShift.is_active ? 'active' : 'next'}
					onCheckIn={handleCheckIn}
					onCheckOut={handleCheckOut}
					onCancel={handleCancelShift}
					isLoading={$cancelBookingMutation?.isPending}
				/>
			{:else}
				<Card.Root>
					<Card.Content class="text-center py-6">
						<CalendarIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
						<h3 class="text-sm font-medium mb-1">No upcoming shifts</h3>
						<p class="text-xs text-muted-foreground">Commit to shifts below</p>
					</Card.Content>
				</Card.Root>
			{/if}

			<!-- Additional Upcoming Shifts -->
			{#if additionalShifts.length > 0}
				<div class="space-y-1">
					<h3 class="text-sm font-medium text-muted-foreground px-1 mb-2">Upcoming shifts</h3>
					{#each additionalShifts as shift (shift.booking_id)}
						{@const canCancel = canCancelBooking(shift.shift_start)}
						<div class="flex items-center justify-between p-2 bg-muted/30 rounded-lg border">
							<div class="flex-1 min-w-0">
								<div class="text-xs text-muted-foreground mt-0.5">
									{formatShiftTimeCompact(shift)}
								</div>
								{#if shift.buddy_name}
									<div class="text-xs text-muted-foreground mt-0.5">
										with {shift.buddy_name}
									</div>
								{/if}
							</div>
							{#if canCancel}
								<Button
									onclick={() => handleCancelShift(shift.booking_id)}
									variant="outline"
									size="sm"
									class="ml-3 text-muted-foreground hover:text-destructive hover:border-destructive"
									disabled={$cancelBookingMutation?.isPending}
								>
									<XIcon class="h-3 w-3 mr-1" />
									Cancel
								</Button>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			<!-- Shift Calendar (moved above shift list and outside of card) -->
			<ShiftCalendar
				shifts={availableShifts}
				{userBookings}
				selectedDayRange={dayRange}
				onShiftSelect={handleBookShift}
			/>

			<!-- Available Shifts (broken out of card layout) -->
			{#if !availableShiftsQuery}
				<!-- Loading state while queries are being initialized -->
				<div class="px-4">
					<div class="text-center py-8">
						<div
							class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
						></div>
						<p class="text-sm text-muted-foreground">Initializing...</p>
					</div>
				</div>
			{:else if $availableShiftsQuery?.isLoading}
				<div class="px-4">
					<div class="text-center py-8">
						<div
							class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
						></div>
						<p class="text-sm text-muted-foreground">Loading available shifts...</p>
					</div>
				</div>
			{:else if $availableShiftsQuery?.isError}
				<div class="px-4">
					<div class="text-center py-8">
						<AlertTriangleIcon class="h-8 w-8 mx-auto mb-2 text-destructive" />
						<h3 class="text-sm font-medium mb-1">Error loading shifts</h3>
						<p class="text-xs text-muted-foreground">{$availableShiftsQuery?.error?.message}</p>
					</div>
				</div>
			{:else}
				<div class="space-y-4">
					<!-- Header with bulk assign -->
					<div>
						<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
							<h2 class="text-base font-semibold">Shift Roster</h2>
						</div>

						<!-- Results summary -->
						<div
							class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between text-sm text-muted-foreground mt-2"
						>
							<span class="min-w-0 truncate">
								{#if $availableShiftsQuery?.isFetching}
									Loading roster...
								{:else if availableShifts.length === 0}
									No shifts scheduled
								{:else if displayedShifts.length < availableShifts.length}
									Showing {displayedShifts.length} of {availableShifts.length} shifts
								{:else}
									{availableShifts.length} shift{availableShifts.length === 1 ? '' : 's'} scheduled
								{/if}
							</span>
						</div>
					</div>

					{#if displayedShifts.length > 0}
						<div class="space-y-3">
							{#each displayedShifts as shift (`${shift.schedule_id}-${shift.start_time}`)}
								<CompactShiftCard
									{shift}
									type="available"
									onBook={handleBookShift}
									isLoading={$bookingMutation?.isPending ?? false}
								/>
							{/each}

							<!-- Show more section with loading states -->
							{#if hasMoreShifts}
								<div class="mt-4 text-center">
									{#if isLoadingMore}
										<!-- Loading more shifts -->
										<div class="flex items-center justify-center gap-2 py-3">
											<div
												class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"
											></div>
											<span class="text-sm text-muted-foreground">Loading more shifts...</span>
										</div>
									{:else}
										<!-- Load more button -->
										<Button
											variant="outline"
											onclick={handleShowMoreShifts}
											disabled={$availableShiftsQuery?.isFetching}
											class="text-sm"
										>
											{#if availableShifts.length > displayedShifts.length}
												<span class="hidden sm:inline"
													>Show {Math.min(10, availableShifts.length - displayedShifts.length)} more
													shifts</span
												>
												<span class="sm:hidden"
													>Show {Math.min(10, availableShifts.length - displayedShifts.length)} more</span
												>
											{:else}
												<span class="hidden sm:inline">Load more shifts</span>
												<span class="sm:hidden">Load more</span>
											{/if}
										</Button>
									{/if}
								</div>
							{/if}
						</div>
					{:else if $availableShiftsQuery?.isFetching}
						<!-- Loading state when fetching new date range -->
						<div class="px-4">
							<div class="text-center py-8">
								<div
									class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
								></div>
								<p class="text-sm text-muted-foreground">Loading shifts...</p>
							</div>
						</div>
					{:else}
						<!-- No shifts available -->
						<div class="px-4">
							<div class="text-center py-8">
								<CalendarIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
								<h3 class="text-sm font-medium mb-1">No shifts available</h3>
								<p class="text-xs text-muted-foreground">Check back later for new opportunities</p>
							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- My Reports Widget -->
			{#if $userSession.isAuthenticated}
				<MyReportsWidget className="mb-4" />
			{/if}
		</div>
	{:else}
		<!-- Unauthenticated Welcome Page -->
		<div class="flex flex-col">
			<!-- Hero Section -->
			<main class="flex-1 flex items-center justify-center px-4 py-16">
				<div class="text-center max-w-2xl lg:max-w-4xl">
					<div class="mb-8">
						<div class="bg-primary/10 p-4 sm:p-6 rounded-2xl w-fit mx-auto mb-6 sm:mb-8">
							<div class="h-32 w-32 sm:h-40 sm:w-40 flex items-center justify-center">
								<img src="/logo.png" alt="Mount Moreland Night Owls" class="object-contain" />
							</div>
						</div>
					</div>

					<h1 class="text-4xl sm:text-5xl md:text-6xl font-bold tracking-tight mb-4">
						Mount Moreland Night Owls
					</h1>

					<h2 class="text-2xl sm:text-3xl md:text-4xl font-semibold text-primary mb-6">
						Digital Control Centre
					</h2>

					<p
						class="text-lg sm:text-xl md:text-2xl text-muted-foreground mb-8 sm:mb-12 leading-relaxed max-w-2xl lg:max-w-3xl mx-auto px-2"
					>
						View and book shifts, send emergency alerts and help keep our community secure
					</p>

					<div class="flex flex-col sm:flex-row gap-4 sm:gap-6 justify-center items-center">
						<Button
							size="lg"
							href="/register"
							class="text-base sm:text-lg px-6 sm:px-8 py-4 sm:py-6 w-full sm:w-auto"
							>Become an Owl</Button
						>
						<Button
							variant="outline"
							size="lg"
							href="/login"
							class="text-base sm:text-lg px-6 sm:px-8 py-4 sm:py-6 w-full sm:w-auto"
						>
							Sign in
						</Button>
					</div>
				</div>
			</main>
		</div>
	{/if}
</div>

<!-- Booking Confirmation Dialog -->
{#if mounted}
	<BookingConfirmationDialog
		bind:open={showBookingDialog}
		bind:shift={selectedShift}
		isLoading={$bookingMutation?.isPending ?? false}
		onConfirm={handleBookingConfirm}
		onCancel={handleBookingCancel}
	/>

	<!-- Cancellation Confirmation Dialog -->
	<CancellationConfirmationDialog
		bind:open={showCancelDialog}
		shiftDetails={shiftToCancel?.details || ''}
		isLoading={$cancelBookingMutation?.isPending}
		onConfirm={handleCancellationConfirm}
		onCancel={handleCancellationCancel}
	/>

	<!-- Bulk Assign Dialog -->
	<BulkAssignDialog
		bind:open={showBulkAssignDialog}
		shifts={availableShifts}
		onSuccess={handleBulkAssignSuccess}
		onCancel={handleBulkAssignCancel}
	/>
{/if}
