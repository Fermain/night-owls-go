<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import { createQuery, createMutation } from '@tanstack/svelte-query';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import PlayIcon from '@lucide/svelte/icons/play';
	import SquareIcon from '@lucide/svelte/icons/square';
	import { userSession } from '$lib/stores/authStore';
	import EmergencyContacts from '$lib/components/emergency/EmergencyContacts.svelte';
	import {
		UserApiService,
		type AvailableShiftSlot,
		type CreateBookingRequest
	} from '$lib/services/api/user';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';

	// Get current user from auth store
	const currentUser = $derived($userSession);

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
		retry: false // Don't retry if user is not authenticated
	});

	// Derived data
	const availableShifts = $derived($availableShiftsQuery.data ?? []);
	const unfillableShifts = $derived(availableShifts.slice(0, 3)); // Show first 3 as examples

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

	// Mutations for booking
	const bookingMutation = createMutation({
		mutationFn: (request: CreateBookingRequest) => UserApiService.createBooking(request),
		onSuccess: () => {
			toast.success('Shift booked successfully!');
			$availableShiftsQuery.refetch();
		},
		onError: (error) => {
			toast.error(`Failed to book shift: ${error.message}`);
		}
	});

	// Helper functions
	function formatTime(timeString: string) {
		return new Date(timeString).toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getTimeUntil(timeString: string) {
		const now = new Date();
		const time = new Date(timeString);
		const diffMs = time.getTime() - now.getTime();
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

		if (diffMs < 0) return 'Started';
		if (diffHours > 0) return `${diffHours}h ${diffMins}m`;
		return `${diffMins}m`;
	}

	function handleCheckIn() {
		// TODO: Implement with real booking attendance API
		console.log('Checking in to shift...');
		toast.success('Checked in successfully!');
	}

	function handleCheckOut() {
		// TODO: Implement with real booking attendance API
		console.log('Checking out of shift...');
		toast.success('Checked out successfully!');
	}

	function handleQuickReport() {
		// TODO: Navigate to report page with booking context
		window.location.href = '/report';
	}

	function handleEmergency() {
		if (confirm('Call emergency services?')) {
			window.location.href = 'tel:999';
		}
	}

	function handleBookShift(shift: AvailableShiftSlot) {
		goto(`/bookings?scheduleId=${shift.schedule_id}&startTime=${shift.start_time}`);
	}

	function handleViewMyBookings() {
		goto('/bookings/my');
	}
</script>

<svelte:head>
	<title>Mount Moreland Night Owls</title>
</svelte:head>

<div class="bg-background">
	{#if currentUser.isAuthenticated}
		<!-- Authenticated Dashboard -->
		<div class="p-4 space-y-4">
			<!-- My Next Shift -->
			{#if nextShift}
				<Card.Root class="border-l-4 border-l-primary">
					<Card.Header class="pb-3">
						<div class="flex items-center justify-between">
							<Card.Title class="text-base">My next shift</Card.Title>
							<Badge variant={nextShift.is_active ? 'default' : 'secondary'}>
								{nextShift.is_active ? 'Active' : getTimeUntil(nextShift.start_time)}
							</Badge>
						</div>
					</Card.Header>
					<Card.Content class="pt-0 space-y-3">
						<div class="text-sm text-muted-foreground">
							{nextShift.schedule_name}
						</div>

						<div class="flex items-center text-sm">
							<ClockIcon class="h-4 w-4 mr-2 text-muted-foreground" />
							<span>{formatTime(nextShift.start_time)} - {formatTime(nextShift.end_time)}</span>
						</div>

						<div class="flex gap-2">
							{#if nextShift.is_active}
								<Button onclick={handleCheckOut} variant="destructive" class="flex-1">
									<SquareIcon class="h-4 w-4 mr-2" />
									Check Out
								</Button>
								<Button onclick={handleQuickReport} variant="outline">
									<AlertTriangleIcon class="h-4 w-4 mr-2" />
									Report
								</Button>
							{:else if nextShift.can_checkin}
								<Button onclick={handleCheckIn} class="flex-1">
									<PlayIcon class="h-4 w-4 mr-2" />
									Check In
								</Button>
							{/if}
						</div>
					</Card.Content>
				</Card.Root>
			{:else}
				<Card.Root>
					<Card.Content class="text-center py-6">
						<CalendarIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
						<h3 class="text-sm font-medium mb-1">No upcoming shifts</h3>
						<p class="text-xs text-muted-foreground mb-3">Check available shifts below</p>
						<Button size="sm" href="/bookings" variant="outline">Browse Shifts</Button>
					</Card.Content>
				</Card.Root>
			{/if}

			<!-- Unfilled Shifts -->
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
						<Card.Title class="text-base">Available shifts</Card.Title>
					</Card.Header>
					<Card.Content class="pt-0">
						{#each unfillableShifts as shift, i (`${shift.schedule_id}-${shift.start_time}`)}
							<div class="flex items-center justify-between py-3">
								<div class="flex-1">
									<div class="text-sm font-medium">{shift.schedule_name}</div>
									<div class="text-xs text-muted-foreground">
										{formatTime(shift.start_time)} - {formatTime(shift.end_time)}
									</div>
									<div class="text-xs text-orange-600 dark:text-orange-400">Available now</div>
								</div>
								<Button
									size="sm"
									onclick={() => handleBookShift(shift)}
									disabled={$bookingMutation.isPending}
								>
									{$bookingMutation.isPending ? 'Booking...' : 'Book'}
								</Button>
							</div>
							{#if i < unfillableShifts.length - 1}
								<Separator class="my-2" />
							{/if}
						{/each}
						{#if availableShifts.length > 3}
							<div class="mt-4 text-center">
								<a href="/bookings" class="text-sm text-primary hover:underline">
									View all {availableShifts.length} available shifts →
								</a>
							</div>
						{/if}
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

			<!-- Emergency Contacts -->
			<EmergencyContacts />
		</div>
	{:else}
		<!-- Unauthenticated Welcome Page -->
		<div class="flex flex-col">
			<!-- Hero Section -->
			<main class="flex-1 flex items-center justify-center px-4 py-16">
				<div class="text-center max-w-4xl">
					<div class="mb-8">
						<div class="bg-primary/10 p-6 rounded-2xl w-fit mx-auto mb-8">
							<div class="h-20 w-20 flex items-center justify-center">
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
