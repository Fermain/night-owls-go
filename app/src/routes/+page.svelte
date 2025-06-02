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
	import FilterIcon from '@lucide/svelte/icons/filter';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import { userSession } from '$lib/stores/authStore';
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
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import { onMount } from 'svelte';

	// Get current user from auth store
	const currentUser = $derived($userSession);

	// State for booking confirmation dialog
	let showBookingDialog = $state(false);
	let selectedShift = $state<AvailableShiftSlot | null>(null);

	// State for cancellation confirmation dialog
	let showCancelDialog = $state(false);
	let shiftToCancel = $state<{ id: number; details: string } | null>(null);

	// State for bulk assign dialog
	let showBulkAssignDialog = $state(false);

	// Date filter state - simplified to preset periods
	let selectedDayRange = $state<string>('7'); // Default to 7 days
	let showDateFilters = $state(false);
	let shiftLimit = $state(10);

	// Day range options
	const dayRangeOptions = [
		{ value: '7', label: 'Next 7 days' },
		{ value: '14', label: 'Next 14 days' },
		{ value: '30', label: 'Next 30 days' },
		{ value: '60', label: 'Next 60 days' }
	];

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

	// Helper function to get date range for shifts
	function getShiftDateRange() {
		const days = parseInt(selectedDayRange);
		const from = new Date().toISOString();
		const to = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();

		return { from, to };
	}

	// Initialize queries after component is mounted to avoid lifecycle errors
	onMount(() => {
		// Query for available shifts with dynamic date range
		availableShiftsQuery = createQuery({
			queryKey: ['available-shifts', selectedDayRange, shiftLimit],
			queryFn: () => {
				const { from, to } = getShiftDateRange();
				return UserApiService.getAvailableShifts({ from, to, limit: shiftLimit });
			}
		});

		// Query for user's bookings (only if authenticated)
		userBookingsQuery = createQuery({
			queryKey: ['user-bookings'],
			queryFn: () => {
				if (!$userSession.isAuthenticated) {
					throw new Error('User not authenticated');
				}
				return UserApiService.getMyBookings();
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

	// Calculate how many shifts to display based on shiftLimit, but always show at least 5
	const displayLimit = $derived(Math.max(5, Math.min(shiftLimit, availableShifts.length)));
	const displayedShifts = $derived(availableShifts.slice(0, displayLimit));

	// Date filter handlers
	function handleDayRangeChange(value: string) {
		selectedDayRange = value;
		// Reset shift limit when changing date range to avoid confusion
		shiftLimit = 10;
		// Query will automatically refetch due to reactive dependencies
	}

	function handleShowMoreShifts() {
		shiftLimit = shiftLimit + 10; // Increase limit by 10
	}

	function handleResetFilters() {
		selectedDayRange = '7';
		shiftLimit = 10;
		showDateFilters = false;
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
		console.log('Checking in to shift...');
		toast.success('Checked in successfully!');
	}

	function handleCheckOut() {
		console.log('Checking out of shift...');
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

	// Check if any filters are active
	const hasActiveFilters = $derived.by(() => {
		return selectedDayRange !== '7' || shiftLimit > 10;
	});

	// Get current day range label for display
	const currentDayRangeLabel = $derived.by(() => {
		return dayRangeOptions.find((opt) => opt.value === selectedDayRange)?.label || 'Next 7 days';
	});

	// Check if we're loading more shifts (when shiftLimit increases)
	const isLoadingMore = $derived($availableShiftsQuery?.isFetching && displayedShifts.length > 0);

	// Check if there are more shifts to load
	const hasMoreShifts = $derived(
		availableShifts.length > displayedShifts.length ||
			($availableShiftsQuery?.data?.length === shiftLimit && shiftLimit < 100)
	); // Assume more might be available if we hit the limit
</script>

<svelte:head>
	<title>Mount Moreland Night Owls</title>
</svelte:head>

<div class="bg-background flex-1">
	{#if currentUser.isAuthenticated}
		<!-- Authenticated Dashboard -->
		<div class="p-4 space-y-4">
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

			<!-- Available Shifts with Filters -->
			{#if !availableShiftsQuery}
				<!-- Loading state while queries are being initialized -->
				<Card.Root>
					<Card.Content class="text-center py-8">
						<div
							class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
						></div>
						<p class="text-sm text-muted-foreground">Initializing...</p>
					</Card.Content>
				</Card.Root>
			{:else if $availableShiftsQuery?.isLoading}
				<Card.Root>
					<Card.Content class="text-center py-8">
						<div
							class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
						></div>
						<p class="text-sm text-muted-foreground">Loading available shifts...</p>
					</Card.Content>
				</Card.Root>
			{:else if $availableShiftsQuery?.isError}
				<Card.Root>
					<Card.Content class="text-center py-8">
						<AlertTriangleIcon class="h-8 w-8 mx-auto mb-2 text-destructive" />
						<h3 class="text-sm font-medium mb-1">Error loading shifts</h3>
						<p class="text-xs text-muted-foreground">{$availableShiftsQuery?.error?.message}</p>
					</Card.Content>
				</Card.Root>
			{:else}
				<Card.Root>
					<Card.Header class="pb-3">
						<div class="space-y-3">
							<!-- Header with filter toggle and bulk assign -->
							<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
								<Card.Title class="text-base">Available shifts</Card.Title>
								<div class="flex items-center gap-2 flex-wrap">
									{#if availableShifts.length > 1}
										<Button
											variant="outline"
											size="sm"
											onclick={() => (showBulkAssignDialog = true)}
											class="h-8 text-xs sm:text-sm"
											disabled={$availableShiftsQuery?.isFetching}
										>
											<ListChecksIcon class="h-3 w-3 mr-1 sm:mr-2" />
											<span class="hidden sm:inline">Bulk Assign ({availableShifts.length})</span>
											<span class="sm:hidden">Bulk ({availableShifts.length})</span>
										</Button>
									{/if}
									<Button
										variant="outline"
										size="sm"
										onclick={() => (showDateFilters = !showDateFilters)}
										class="h-8"
									>
										<FilterIcon class="h-3 w-3 mr-1 sm:mr-2" />
										{showDateFilters ? 'Hide' : 'Filter'}
										<ChevronDownIcon
											class="h-3 w-3 ml-1 transition-transform {showDateFilters
												? 'rotate-180'
												: ''}"
										/>
									</Button>
								</div>
							</div>

							<!-- Date filter controls -->
							{#if showDateFilters}
								<div class="space-y-3 p-3 bg-muted/30 rounded-lg border">
									<div class="space-y-2">
										<div class="flex items-center gap-2">
											<Label class="text-sm font-medium">Time Period</Label>
											{#if $availableShiftsQuery?.isFetching}
												<div
													class="animate-spin rounded-full h-3 w-3 border-b-2 border-primary"
												></div>
											{/if}
										</div>
										<Select.Root
											type="single"
											value={selectedDayRange}
											onValueChange={handleDayRangeChange}
										>
											<Select.Trigger class="w-full" disabled={$availableShiftsQuery?.isFetching}>
												{currentDayRangeLabel}
											</Select.Trigger>
											<Select.Content>
												{#each dayRangeOptions as option (option.value)}
													<Select.Item value={option.value} label={option.label}
														>{option.label}</Select.Item
													>
												{/each}
											</Select.Content>
										</Select.Root>
									</div>

									{#if hasActiveFilters}
										<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
											<span class="text-xs text-muted-foreground min-w-0 truncate">
												Showing shifts for {currentDayRangeLabel.toLowerCase()}
											</span>
											<Button
												variant="ghost"
												size="sm"
												onclick={handleResetFilters}
												class="h-6 px-2 text-xs whitespace-nowrap self-start sm:self-auto"
											>
												Reset
											</Button>
										</div>
									{/if}
								</div>
							{/if}

							<!-- Results summary -->
							<div
								class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between text-sm text-muted-foreground"
							>
								<span class="min-w-0 truncate">
									{#if $availableShiftsQuery?.isFetching}
										Loading shifts...
									{:else if availableShifts.length === 0}
										No shifts found
									{:else if displayedShifts.length < availableShifts.length}
										Showing {displayedShifts.length} of {availableShifts.length} shifts
									{:else}
										{availableShifts.length} shift{availableShifts.length === 1 ? '' : 's'} available
									{/if}
								</span>
								{#if hasMoreShifts && !$availableShiftsQuery?.isFetching}
									<Button
										variant="ghost"
										size="sm"
										onclick={handleShowMoreShifts}
										class="h-6 px-2 text-xs whitespace-nowrap"
									>
										Show more
									</Button>
								{/if}
							</div>
						</div>
					</Card.Header>

					{#if displayedShifts.length > 0}
						<Card.Content class="pt-0">
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
						</Card.Content>
					{:else if $availableShiftsQuery?.isFetching}
						<!-- Loading state when fetching new date range -->
						<Card.Content class="pt-0">
							<div class="text-center py-8">
								<div
									class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
								></div>
								<p class="text-sm text-muted-foreground">
									Loading shifts for {currentDayRangeLabel.toLowerCase()}...
								</p>
							</div>
						</Card.Content>
					{:else}
						<!-- No shifts available -->
						<Card.Content class="pt-0">
							<div class="text-center py-8">
								<CalendarIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
								<h3 class="text-sm font-medium mb-1">No shifts available</h3>
								<p class="text-xs text-muted-foreground">
									{hasActiveFilters
										? 'Try adjusting your time period or clearing filters'
										: 'Check back later for new opportunities'}
								</p>
								{#if hasActiveFilters}
									<Button variant="outline" size="sm" onclick={handleResetFilters} class="mt-2">
										Clear filters
									</Button>
								{/if}
							</div>
						</Card.Content>
					{/if}
				</Card.Root>
			{/if}

			<!-- Shift Calendar -->
			<ShiftCalendar
				shifts={availableShifts}
				userBookings={($userBookingsQuery?.data as UserBooking[]) ?? []}
				{selectedDayRange}
				onShiftSelect={handleBookShift}
			/>
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
