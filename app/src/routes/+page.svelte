<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { createQuery, createMutation } from '@tanstack/svelte-query';
	import { canCancelBooking } from '$lib/utils/bookings';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import XIcon from '@lucide/svelte/icons/x';
	import { userSession } from '$lib/stores/authStore';
	import {
		UserApiService,
		type AvailableShiftSlot,
		type CreateBookingRequest
	} from '$lib/services/api/user';
	import { toast } from 'svelte-sonner';
	import CompactShiftCard from '$lib/components/user/shifts/CompactShiftCard.svelte';
	import BookingConfirmationDialog from '$lib/components/user/bookings/BookingConfirmationDialog.svelte';
	import CancellationConfirmationDialog from '$lib/components/user/bookings/CancellationConfirmationDialog.svelte';

	// Get current user from auth store
	const currentUser = $derived($userSession);

	// State for booking confirmation dialog
	let showBookingDialog = $state(false);
	let selectedShift = $state<AvailableShiftSlot | null>(null);

	// State for cancellation confirmation dialog
	let showCancelDialog = $state(false);
	let shiftToCancel = $state<{id: number, details: string} | null>(null);

	// Query for available shifts (next 7 days)
	const availableShiftsQuery = createQuery({
		queryKey: ['available-shifts'],
		queryFn: () => {
			const from = new Date().toISOString();
			const to = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString();
			return UserApiService.getAvailableShifts({ from, to, limit: 10 });
		}
	});

	// Query for user's bookings (only if authenticated)
	const userBookingsQuery = createQuery({
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

	// Derived data
	const availableShifts = $derived($availableShiftsQuery.data ?? []);
	const unfillableShifts = $derived(availableShifts.slice(0, 5)); // Show first 5

	// Find next shift from user bookings
	const nextShift = $derived.by(() => {
		if (!$userBookingsQuery.data) return null;

		const now = new Date();
		const upcomingBookings = $userBookingsQuery.data
			.filter((booking) => new Date(booking.shift_start) > now)
			.sort((a, b) => new Date(a.shift_start).getTime() - new Date(b.shift_start).getTime());

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
		if (!$userBookingsQuery.data) return [];

		const now = new Date();
		const upcomingBookings = $userBookingsQuery.data
			.filter((booking) => new Date(booking.shift_start) > now)
			.sort((a, b) => new Date(a.shift_start).getTime() - new Date(b.shift_start).getTime());

		// Return all upcoming shifts except the first one (which is the "next shift")
		return upcomingBookings.slice(1).slice(0, 3); // Show up to 3 additional shifts
	});

	// Mutations for booking
	const bookingMutation = createMutation({
		mutationFn: (request: CreateBookingRequest) => UserApiService.createBooking(request),
		onSuccess: () => {
			toast.success('Shift committed successfully!');
			$availableShiftsQuery.refetch();
			$userBookingsQuery.refetch();
			showBookingDialog = false;
			selectedShift = null;
		},
		onError: (error) => {
			toast.error(`Failed to commit to shift: ${error.message}`);
		}
	});

	// Mutation for canceling booking
	const cancelBookingMutation = createMutation({
		mutationFn: (bookingId: number) => UserApiService.cancelBooking(bookingId),
		onSuccess: () => {
			toast.success('Shift cancelled successfully!');
			$userBookingsQuery.refetch();
			$availableShiftsQuery.refetch();
			showCancelDialog = false;
			shiftToCancel = null;
		},
		onError: (error) => {
			toast.error(`Failed to cancel shift: ${error.message}`);
			showCancelDialog = false;
			shiftToCancel = null;
		}
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
			const additionalShift = additionalShifts.find(shift => shift.booking_id === shiftId);
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
		$bookingMutation.mutate(request);
	}

	function handleBookingCancel() {
		showBookingDialog = false;
		selectedShift = null;
	}

	function handleCancellationConfirm() {
		if (shiftToCancel) {
			$cancelBookingMutation.mutate(shiftToCancel.id);
		}
	}

	function handleCancellationCancel() {
		showCancelDialog = false;
		shiftToCancel = null;
	}

	function formatShiftTimeFromBooking(shift: any) {
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

	function formatShiftTimeCompact(booking: any) {
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
					isLoading={$cancelBookingMutation.isPending}
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
									disabled={$cancelBookingMutation.isPending}
								>
									<XIcon class="h-3 w-3 mr-1" />
									Cancel
								</Button>
							{/if}
						</div>
					{/each}
				</div>
			{/if}

			<!-- Available Shifts -->
			{#if $availableShiftsQuery.isLoading}
				<Card.Root>
					<Card.Content class="text-center py-8">
						<div
							class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"
						></div>
						<p class="text-sm text-muted-foreground">Loading available shifts...</p>
					</Card.Content>
				</Card.Root>
			{:else if $availableShiftsQuery.isError}
				<Card.Root>
					<Card.Content class="text-center py-8">
						<AlertTriangleIcon class="h-8 w-8 mx-auto mb-2 text-destructive" />
						<h3 class="text-sm font-medium mb-1">Error loading shifts</h3>
						<p class="text-xs text-muted-foreground">{$availableShiftsQuery.error?.message}</p>
					</Card.Content>
				</Card.Root>
			{:else if unfillableShifts.length > 0}
				<Card.Root>
					<Card.Header class="pb-3">
						<div class="flex items-center justify-between">
							<Card.Title class="text-base">Available shifts</Card.Title>
							{#if availableShifts.length > 5}
								<span class="text-sm text-muted-foreground">
									Showing 5 of {availableShifts.length}
								</span>
							{/if}
						</div>
					</Card.Header>
					<Card.Content class="pt-0">
						{#each unfillableShifts as shift (`${shift.schedule_id}-${shift.start_time}`)}
							<CompactShiftCard
								{shift}
								type="available"
								onBook={handleBookShift}
								isLoading={$bookingMutation.isPending}
							/>
						{/each}
					</Card.Content>
				</Card.Root>
			{:else}
				<Card.Root>
					<Card.Content class="text-center py-8">
						<CalendarIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
						<h3 class="text-sm font-medium mb-1">No shifts available</h3>
						<p class="text-xs text-muted-foreground">Check back later for new opportunities</p>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>
	{:else}
		<!-- Unauthenticated Welcome Page -->
		<div class="flex flex-col">
			<!-- Hero Section -->
			<main class="flex-1 flex items-center justify-center px-4 py-16">
				<div class="text-center max-w-4xl">
					<div class="mb-8">
						<div class="bg-primary/10 p-6 rounded-2xl w-fit mx-auto mb-8">
							<div class="h-40 w-40 flex items-center justify-center">
								<img src="/logo.png" alt="Mount Moreland Night Owls" class="object-contain" />
							</div>
						</div>
					</div>

					<h1 class="text-5xl md:text-6xl font-bold tracking-tight mb-4">
						Mount Moreland Night Owls
					</h1>

					<h2 class="text-3xl md:text-4xl font-semibold text-primary mb-6">
						Digital Control Centre
					</h2>

					<p
						class="text-xl md:text-2xl text-muted-foreground mb-12 leading-relaxed max-w-3xl mx-auto"
					>
						View and book shifts, send emergency alerts and help keep our community secure
					</p>

					<div class="flex flex-col sm:flex-row gap-6 justify-center items-center">
						<Button size="lg" href="/register" class="text-lg px-8 py-6">Become an Owl</Button>
						<Button variant="outline" size="lg" href="/login" class="text-lg px-8 py-6">
							Sign in
						</Button>
					</div>
				</div>
			</main>
		</div>
	{/if}
</div>

<!-- Booking Confirmation Dialog -->
<BookingConfirmationDialog 
	bind:open={showBookingDialog}
	bind:shift={selectedShift}
	isLoading={$bookingMutation.isPending}
	onConfirm={handleBookingConfirm}
	onCancel={handleBookingCancel}
/>

<!-- Cancellation Confirmation Dialog -->
<CancellationConfirmationDialog 
	bind:open={showCancelDialog}
	shiftDetails={shiftToCancel?.details || ''}
	isLoading={$cancelBookingMutation.isPending}
	onConfirm={handleCancellationConfirm}
	onCancel={handleCancellationCancel}
/>
