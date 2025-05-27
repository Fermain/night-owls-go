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
	import {
		UserApiService,
		type AvailableShiftSlot,
		type CreateBookingRequest
	} from '$lib/services/api/user';
	import { toast } from 'svelte-sonner';

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
			is_active: isActive,
			location: 'Main Street Area' // TODO: Add location to schedule data
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
		if (!currentUser.isAuthenticated) {
			toast.error('Please log in to book shifts');
			return;
		}

		const request: CreateBookingRequest = {
			schedule_id: shift.schedule_id,
			start_time: shift.start_time
		};

		$bookingMutation.mutate(request);
	}
</script>

<svelte:head>
	<title>Night Owls</title>
</svelte:head>

<div class="min-h-screen bg-background">
	{#if currentUser.isAuthenticated}
		<!-- Authenticated Dashboard -->
		<!-- Header -->
		<div class="bg-card border-b">
			<div class="px-4 py-3">
				<div class="flex items-center justify-between">
					<div>
						<h1 class="text-lg font-semibold">
							{#if currentUser.name}
								Evening, {currentUser.name.split(' ')[0]}
							{:else}
								Night Owls Patrol
							{/if}
						</h1>
						<p class="text-sm text-muted-foreground">Ready for patrol</p>
					</div>
					<Button variant="destructive" size="sm" onclick={handleEmergency}>Emergency</Button>
				</div>
			</div>
		</div>

		<div class="p-4">
			<!-- My Next Shift -->
			{#if nextShift}
				<Card.Root>
					<Card.Header class="pb-3">
						<div class="flex items-center justify-between">
							<Card.Title class="text-base">My next shift</Card.Title>
							<Badge variant={nextShift.is_active ? 'default' : 'secondary'}>
								{nextShift.is_active ? 'Active' : getTimeUntil(nextShift.start_time)}
							</Badge>
						</div>
					</Card.Header>
					<Card.Content class="pt-0">
						<div class="text-sm text-muted-foreground mb-3">
							{nextShift.schedule_name} • {nextShift.location}
						</div>

						<div class="flex items-center text-sm mb-4">
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
							{:else}
								<Button disabled class="flex-1">
									<ClockIcon class="h-4 w-4 mr-2" />
									Too Early
								</Button>
							{/if}
						</div>
					</Card.Content>
				</Card.Root>
			{:else}
				<Card.Root>
					<Card.Content class="text-center py-8">
						<CalendarIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
						<h3 class="text-sm font-medium mb-1">No upcoming shifts</h3>
						<p class="text-xs text-muted-foreground">Check available shifts below</p>
					</Card.Content>
				</Card.Root>
			{/if}

			<!-- Quick Actions -->
			<div class="grid grid-cols-2 gap-2 my-4">
				<a
					href="/shifts"
					class="inline-flex items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium ring-offset-background transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
				>
					<CalendarIcon class="h-4 w-4 mr-2" />
					Browse Shifts
				</a>
				<a
					href="/report"
					class="inline-flex items-center justify-center rounded-md border border-input bg-background px-4 py-2 text-sm font-medium ring-offset-background transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
				>
					<AlertTriangleIcon class="h-4 w-4 mr-2" />
					Report Incident
				</a>
			</div>

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
						{#each unfillableShifts as shift, i}
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
								<a href="/shifts" class="text-sm text-primary hover:underline">
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
		</div>
	{:else}
		<!-- Unauthenticated Welcome Page -->
		<div class="min-h-screen flex flex-col bg-patrol-gradient">
			<!-- Enhanced Header with better branding -->
			<header class="border-b border-border/50 bg-card/95 backdrop-blur-sm">
				<div class="container mx-auto px-4 py-5 flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div
							class="bg-primary text-primary-foreground flex size-10 items-center justify-center rounded-xl shadow-lg"
						>
							<AlertTriangleIcon class="size-5" />
						</div>
						<div>
							<span class="font-bold text-xl text-foreground">Night Owls Patrol</span>
							<p class="text-xs text-muted-foreground">Community Watch</p>
						</div>
					</div>
					<div class="flex gap-3">
						<Button variant="ghost" href="/login" class="interactive-scale">Sign In</Button>
						<Button href="/register" class="interactive-scale">Join Us</Button>
					</div>
				</div>
			</header>

			<!-- Enhanced Hero Section -->
			<main class="flex-1 flex items-center justify-center px-4 py-16">
				<div class="text-center max-w-4xl animate-in">
					<div class="mb-6">
						<div class="bg-primary/10 p-4 rounded-2xl w-fit mx-auto mb-6">
							<AlertTriangleIcon class="h-16 w-16 text-primary" />
						</div>
					</div>
					<h1
						class="text-5xl md:text-6xl font-bold tracking-tight mb-6 bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text"
					>
						Protecting Our Community Together
					</h1>
					<p
						class="text-xl md:text-2xl text-muted-foreground mb-10 leading-relaxed max-w-3xl mx-auto"
					>
						Join your neighbors in keeping our community safe through coordinated patrols, real-time
						communication, and shared vigilance.
					</p>
					<div class="flex flex-col sm:flex-row gap-4 justify-center items-center">
						<Button size="lg" href="/register" class="text-lg px-8 py-6 interactive-scale">
							Join the Watch
						</Button>
						<Button
							variant="outline"
							size="lg"
							href="/login"
							class="text-lg px-8 py-6 interactive-scale"
						>
							Sign In
						</Button>
					</div>
				</div>
			</main>

			<!-- Enhanced Features Section -->
			<section class="py-20 bg-card/30 border-t border-border/50">
				<div class="container mx-auto px-4">
					<div class="text-center mb-16">
						<h2 class="text-3xl md:text-4xl font-bold mb-4">How We Keep Our Community Safe</h2>
						<p class="text-xl text-muted-foreground max-w-2xl mx-auto">
							Our coordinated approach ensures comprehensive neighborhood protection
						</p>
					</div>
					<div class="grid md:grid-cols-3 gap-8 lg:gap-12">
						<div class="text-center group">
							<div
								class="bg-primary/10 p-6 rounded-2xl w-fit mx-auto mb-6 group-hover:bg-primary/20 transition-colors"
							>
								<CalendarIcon class="h-12 w-12 text-primary" />
							</div>
							<h3 class="text-xl font-bold mb-3">Coordinate Patrols</h3>
							<p class="text-muted-foreground leading-relaxed">
								Schedule and join community patrol shifts with your neighbors for comprehensive
								coverage
							</p>
						</div>
						<div class="text-center group">
							<div
								class="bg-primary/10 p-6 rounded-2xl w-fit mx-auto mb-6 group-hover:bg-primary/20 transition-colors"
							>
								<AlertTriangleIcon class="h-12 w-12 text-primary" />
							</div>
							<h3 class="text-xl font-bold mb-3">Report Incidents</h3>
							<p class="text-muted-foreground leading-relaxed">
								Quickly report and track community incidents with real-time alerts to all members
							</p>
						</div>
						<div class="text-center group">
							<div
								class="bg-primary/10 p-6 rounded-2xl w-fit mx-auto mb-6 group-hover:bg-primary/20 transition-colors"
							>
								<ClockIcon class="h-12 w-12 text-primary" />
							</div>
							<h3 class="text-xl font-bold mb-3">Real-time Updates</h3>
							<p class="text-muted-foreground leading-relaxed">
								Stay informed with instant notifications and coordinate responses effectively
							</p>
						</div>
					</div>
				</div>
			</section>

			<!-- Call to Action Section -->
			<section class="py-16 bg-primary/5 border-t border-border/50">
				<div class="container mx-auto px-4 text-center">
					<h2 class="text-3xl md:text-4xl font-bold mb-4">Ready to Make a Difference?</h2>
					<p class="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
						Join hundreds of neighbors who are already making our community safer
					</p>
					<Button size="lg" href="/register" class="text-lg px-8 py-6 interactive-scale">
						Start Your Watch Today
					</Button>
				</div>
			</section>
		</div>
	{/if}
</div>
