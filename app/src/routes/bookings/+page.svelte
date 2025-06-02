<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import {
		createQuery,
		createMutation,
		useQueryClient,
		type CreateQueryResult,
		type CreateMutationResult
	} from '@tanstack/svelte-query';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import {
		UserApiService,
		type AvailableShiftSlot,
		type UserBooking
	} from '$lib/services/api/user';
	import { userSession } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';

	const queryClient = useQueryClient();

	// Query states - will be initialized in onMount with proper types
	let userBookingsQuery = $state<CreateQueryResult<UserBooking[], Error> | null>(null);
	let availableShiftsQuery = $state<CreateQueryResult<AvailableShiftSlot[], Error> | null>(null);
	let checkInMutation = $state<CreateMutationResult<UserBooking, Error, number, unknown> | null>(
		null
	);
	let cancelMutation = $state<CreateMutationResult<void, Error, number, unknown> | null>(null);
	let bookShiftMutation = $state<CreateMutationResult<
		UserBooking,
		Error,
		{ schedule_id: number; start_time: string },
		unknown
	> | null>(null);

	// Initialize queries after component is mounted to avoid lifecycle errors
	onMount(() => {
		// Query for user's bookings
		userBookingsQuery = createQuery({
			queryKey: ['user-bookings'],
			queryFn: () => UserApiService.getMyBookings(),
			enabled: $userSession.isAuthenticated
		});

		// Query for available shifts
		availableShiftsQuery = createQuery({
			queryKey: ['available-shifts'],
			queryFn: () => {
				const now = new Date();
				const twoWeeksFromNow = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000);
				return UserApiService.getAvailableShifts({
					from: now.toISOString(),
					to: twoWeeksFromNow.toISOString(),
					limit: 50
				});
			}
		});

		// Mutation for checking in
		checkInMutation = createMutation({
			mutationFn: (bookingId: number) => UserApiService.markCheckIn(bookingId),
			onSuccess: () => {
				queryClient.invalidateQueries({ queryKey: ['user-bookings'] });
				toast.success('Checked in successfully!');
			},
			onError: (error: Error) => {
				toast.error(`Failed to check in: ${error.message}`);
			}
		});

		// Mutation for cancelling booking
		cancelMutation = createMutation({
			mutationFn: (bookingId: number) => UserApiService.cancelBooking(bookingId),
			onSuccess: () => {
				queryClient.invalidateQueries({ queryKey: ['user-bookings'] });
				queryClient.invalidateQueries({ queryKey: ['available-shifts'] });
				toast.success('Commitment cancelled successfully!');
			},
			onError: (error: Error) => {
				toast.error(`Failed to cancel commitment: ${error.message}`);
			}
		});

		// Mutation for booking a shift
		bookShiftMutation = createMutation({
			mutationFn: (params: { schedule_id: number; start_time: string }) =>
				UserApiService.createBooking(params),
			onSuccess: () => {
				queryClient.invalidateQueries({ queryKey: ['user-bookings'] });
				queryClient.invalidateQueries({ queryKey: ['available-shifts'] });
				toast.success('Shift committed successfully!');
			},
			onError: (error: Error) => {
				toast.error(`Failed to commit to shift: ${error.message}`);
			}
		});
	});

	const bookings = $derived(($userBookingsQuery?.data as UserBooking[]) ?? []);
	const availableShifts = $derived(($availableShiftsQuery?.data as AvailableShiftSlot[]) ?? []);
	const isLoadingBookings = $derived($userBookingsQuery?.isLoading ?? false);
	const isLoadingShifts = $derived($availableShiftsQuery?.isLoading ?? false);
	const bookingsError = $derived($userBookingsQuery?.error);
	const shiftsError = $derived($availableShiftsQuery?.error);

	function formatDateTime(dateString: string) {
		return new Date(dateString).toLocaleString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	function getShiftStatus(startTime: string, endTime: string, checkedInAt?: string) {
		const now = new Date();
		const start = new Date(startTime);
		const end = new Date(endTime);

		if (now < start) return 'upcoming';
		if (now >= start && now <= end) return 'active';
		if (checkedInAt) return 'completed';
		if (now > end) return 'missed';
		return 'pending';
	}

	function canCancelCommitment(startTime: string): boolean {
		const now = new Date();
		const start = new Date(startTime);
		const cancellationDeadline = new Date(start.getTime() - 2 * 60 * 60 * 1000); // 2 hours before
		return now < cancellationDeadline;
	}

	function handleCheckIn(bookingId: number) {
		$checkInMutation?.mutate(bookingId);
	}

	function handleCancelCommitment(bookingId: number) {
		if (confirm('Are you sure you want to cancel this commitment? This action cannot be undone.')) {
			$cancelMutation?.mutate(bookingId);
		}
	}

	function handleBookShift(shift: { schedule_id: number; start_time: string }) {
		$bookShiftMutation?.mutate({
			schedule_id: shift.schedule_id,
			start_time: shift.start_time
		});
	}
</script>

<svelte:head>
	<title>My Commitments - Night Owls</title>
</svelte:head>

<div class="container mx-auto p-4 pb-20 md:pb-6 space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">My Commitments</h1>
	</div>

	{#if !$userSession.isAuthenticated}
		<Card.Root>
			<Card.Content class="pt-6">
				<p class="text-center text-muted-foreground">
					Please <a href="/login" class="text-primary hover:underline">sign in</a> to view your shifts.
				</p>
			</Card.Content>
		</Card.Root>
	{:else}
		<!-- My Bookings Section -->
		<div class="space-y-4">
			{#if !userBookingsQuery}
				<!-- Loading state while queries are being initialized -->
				<div class="space-y-4">
					{#each Array(3) as _, i (i)}
						<Card.Root>
							<Card.Content class="pt-6">
								<div class="animate-pulse space-y-3">
									<div class="h-4 bg-muted rounded w-1/4"></div>
									<div class="h-4 bg-muted rounded w-1/2"></div>
									<div class="h-4 bg-muted rounded w-1/3"></div>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{:else if isLoadingBookings}
				<div class="space-y-4">
					{#each Array(3) as _, i (i)}
						<Card.Root>
							<Card.Content class="pt-6">
								<div class="animate-pulse space-y-3">
									<div class="h-4 bg-muted rounded w-1/4"></div>
									<div class="h-4 bg-muted rounded w-1/2"></div>
									<div class="h-4 bg-muted rounded w-1/3"></div>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{:else if bookingsError}
				<Card.Root>
					<Card.Content class="pt-6">
						<p class="text-destructive text-center">
							Error loading commitments: {bookingsError.message}
						</p>
					</Card.Content>
				</Card.Root>
			{:else if bookings.length === 0}
				<Card.Root>
					<Card.Content class="pt-6 text-center space-y-4">
						<CalendarIcon class="h-12 w-12 mx-auto text-muted-foreground" />
						<div>
							<h3 class="font-semibold">No commitments yet</h3>
							<p class="text-muted-foreground">
								You haven't committed to any shifts. Check out available shifts below to get
								started.
							</p>
						</div>
					</Card.Content>
				</Card.Root>
			{:else}
				<div class="space-y-4">
					{#each bookings as booking (booking.booking_id)}
						{@const status = getShiftStatus(
							booking.shift_start,
							booking.shift_end,
							booking.checked_in_at
						)}
						{@const canCancel = canCancelCommitment(booking.shift_start)}
						<Card.Root>
							<Card.Header>
								<div class="flex items-center justify-between">
									<div>
										<Card.Title class="text-lg">{booking.schedule_name}</Card.Title>
										<Card.Description>
											Commitment #{booking.booking_id}
										</Card.Description>
									</div>
									<Badge
										variant={status === 'completed'
											? 'default'
											: status === 'active'
												? 'destructive'
												: status === 'upcoming'
													? 'secondary'
													: status === 'missed'
														? 'destructive'
														: 'outline'}
									>
										{status.charAt(0).toUpperCase() + status.slice(1)}
									</Badge>
								</div>
							</Card.Header>
							<Card.Content class="space-y-4">
								<div class="flex items-center gap-4 text-sm">
									<div class="flex items-center gap-1">
										<CalendarIcon class="h-4 w-4" />
										<span>{formatDateTime(booking.shift_start)}</span>
									</div>
									<span class="text-muted-foreground">to</span>
									<div class="flex items-center gap-1">
										<ClockIcon class="h-4 w-4" />
										<span>{formatDateTime(booking.shift_end)}</span>
									</div>
								</div>

								{#if booking.buddy_name}
									<div class="text-sm">
										<span class="font-medium">Buddy:</span>
										{booking.buddy_name}
									</div>
								{/if}

								{#if status === 'pending'}
									<Separator />
									<div class="flex items-center justify-between">
										<p class="text-sm text-muted-foreground">Check in for this completed shift</p>
										<div class="flex gap-2">
											<Button
												size="sm"
												variant="outline"
												disabled={$checkInMutation?.isPending ?? false}
												onclick={() => handleCheckIn(booking.booking_id)}
											>
												<CheckCircleIcon class="h-4 w-4 mr-1" />
												Check In
											</Button>
										</div>
									</div>
								{:else if status === 'active'}
									<Separator />
									<div class="flex items-center justify-between">
										<p class="text-sm font-medium text-primary">Shift is currently active</p>
										<div class="flex gap-2">
											<Button
												size="sm"
												variant="outline"
												disabled={$checkInMutation?.isPending ?? false}
												onclick={() => handleCheckIn(booking.booking_id)}
											>
												<CheckCircleIcon class="h-4 w-4 mr-1" />
												Check In
											</Button>
											<Button size="sm" href="/report?bookingId={booking.booking_id}">
												Report Incident
											</Button>
										</div>
									</div>
								{:else if status === 'upcoming'}
									<Separator />
									<div class="flex items-center justify-between">
										<p class="text-sm text-muted-foreground">
											{canCancel
												? 'Upcoming shift'
												: 'Upcoming shift (cannot cancel - too close to start time)'}
										</p>
										<div class="flex gap-2">
											{#if canCancel}
												<Button
													size="sm"
													variant="outline"
													disabled={$cancelMutation?.isPending ?? false}
													onclick={() => handleCancelCommitment(booking.booking_id)}
												>
													<XCircleIcon class="h-4 w-4 mr-1" />
													Cancel
												</Button>
											{/if}
										</div>
									</div>
								{/if}
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{/if}
		</div>

		<Separator />

		<!-- Available Shifts Section -->
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<h2 class="text-xl font-semibold">Available Shifts</h2>
				<Badge variant="outline"
					>{availableShifts.filter((s: AvailableShiftSlot) => !s.is_booked).length} available</Badge
				>
			</div>

			{#if !availableShiftsQuery}
				<!-- Loading state while queries are being initialized -->
				<div class="space-y-4">
					{#each Array(3) as _, i (i)}
						<Card.Root>
							<Card.Content class="pt-6">
								<div class="animate-pulse space-y-3">
									<div class="h-4 bg-muted rounded w-1/4"></div>
									<div class="h-4 bg-muted rounded w-1/2"></div>
									<div class="h-4 bg-muted rounded w-1/3"></div>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{:else if isLoadingShifts}
				<div class="space-y-4">
					{#each Array(3) as _, i (i)}
						<Card.Root>
							<Card.Content class="pt-6">
								<div class="animate-pulse space-y-3">
									<div class="h-4 bg-muted rounded w-1/4"></div>
									<div class="h-4 bg-muted rounded w-1/2"></div>
									<div class="h-4 bg-muted rounded w-1/3"></div>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{:else if shiftsError}
				<Card.Root>
					<Card.Content class="pt-6">
						<p class="text-destructive text-center">
							Error loading available shifts: {shiftsError.message}
						</p>
					</Card.Content>
				</Card.Root>
			{:else if availableShifts.filter((s: AvailableShiftSlot) => !s.is_booked).length === 0}
				<Card.Root>
					<Card.Content class="pt-6 text-center space-y-4">
						<CalendarIcon class="h-12 w-12 mx-auto text-muted-foreground" />
						<div>
							<h3 class="font-semibold">No available shifts</h3>
							<p class="text-muted-foreground">
								All shifts are currently booked. Check back later for new opportunities.
							</p>
						</div>
					</Card.Content>
				</Card.Root>
			{:else}
				<div class="space-y-4">
					{#each availableShifts.filter((s: AvailableShiftSlot) => !s.is_booked) as shift (shift.schedule_id + shift.start_time)}
						<Card.Root>
							<Card.Header>
								<div class="flex items-center justify-between">
									<div>
										<Card.Title class="text-lg">{shift.schedule_name}</Card.Title>
										<Card.Description>Available shift slot</Card.Description>
									</div>
									<Badge variant="secondary">Available</Badge>
								</div>
							</Card.Header>
							<Card.Content class="space-y-4">
								<div class="flex items-center gap-4 text-sm">
									<div class="flex items-center gap-1">
										<CalendarIcon class="h-4 w-4" />
										<span>{formatDateTime(shift.start_time)}</span>
									</div>
									<span class="text-muted-foreground">to</span>
									<div class="flex items-center gap-1">
										<ClockIcon class="h-4 w-4" />
										<span>{formatDateTime(shift.end_time)}</span>
									</div>
								</div>

								<Separator />
								<div class="flex items-center justify-between">
									<p class="text-sm text-muted-foreground">Commit to this shift</p>
									<Button
										size="sm"
										disabled={$bookShiftMutation?.isPending ?? false}
										onclick={() => handleBookShift(shift)}
									>
										<PlusIcon class="h-4 w-4 mr-1" />
										Commit
									</Button>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
