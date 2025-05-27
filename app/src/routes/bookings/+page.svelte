<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import { UserApiService } from '$lib/services/api/user';
	import { userSession } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';

	const queryClient = useQueryClient();

	// Query for user's bookings
	const userBookingsQuery = createQuery({
		queryKey: ['user-bookings'],
		queryFn: () => UserApiService.getMyBookings(),
		enabled: $userSession.isAuthenticated
	});

	// Mutation for marking attendance
	const markAttendanceMutation = createMutation({
		mutationFn: ({ bookingId, attended }: { bookingId: number; attended: boolean }) =>
			UserApiService.markAttendance(bookingId, attended),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['user-bookings'] });
			toast.success('Attendance updated successfully!');
		},
		onError: (error) => {
			toast.error(`Failed to update attendance: ${error.message}`);
		}
	});

	const bookings = $derived($userBookingsQuery.data ?? []);
	const isLoading = $derived($userBookingsQuery.isLoading);
	const error = $derived($userBookingsQuery.error);

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

	function getShiftStatus(startTime: string, endTime: string, attended?: boolean) {
		const now = new Date();
		const start = new Date(startTime);
		const end = new Date(endTime);

		if (now < start) return 'upcoming';
		if (now >= start && now <= end) return 'active';
		if (attended === true) return 'completed';
		if (attended === false) return 'missed';
		return 'pending'; // Past shift, attendance not marked
	}

	function handleMarkAttendance(bookingId: number, attended: boolean) {
		$markAttendanceMutation.mutate({ bookingId, attended });
	}
</script>

<svelte:head>
	<title>My Bookings - Night Owls</title>
</svelte:head>

	<div class="container mx-auto p-4 pb-20 md:pb-6 space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">My Bookings</h1>
		<Button href="/shifts" variant="outline">
			<CalendarIcon class="h-4 w-4 mr-2" />
			Book New Shift
		</Button>
	</div>

	{#if !$userSession.isAuthenticated}
		<Card.Root>
			<Card.Content class="pt-6">
				<p class="text-center text-muted-foreground">
					Please <a href="/login" class="text-primary hover:underline">sign in</a> to view your bookings.
				</p>
			</Card.Content>
		</Card.Root>
	{:else if isLoading}
		<div class="space-y-4">
			{#each Array(3) as _}
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
	{:else if error}
		<Card.Root>
			<Card.Content class="pt-6">
				<p class="text-destructive text-center">
					Error loading bookings: {error.message}
				</p>
			</Card.Content>
		</Card.Root>
	{:else if bookings.length === 0}
		<Card.Root>
			<Card.Content class="pt-6 text-center space-y-4">
				<CalendarIcon class="h-12 w-12 mx-auto text-muted-foreground" />
				<div>
					<h3 class="font-semibold">No bookings yet</h3>
					<p class="text-muted-foreground">
						You haven't booked any shifts. Check out available shifts to get started.
					</p>
				</div>
				<Button href="/shifts">Browse Available Shifts</Button>
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="space-y-4">
			{#each bookings as booking (booking.booking_id)}
				{@const status = getShiftStatus(booking.shift_start, booking.shift_end, booking.attended)}
				<Card.Root>
					<Card.Header>
						<div class="flex items-center justify-between">
							<div>
								<Card.Title class="text-lg">{booking.schedule_name}</Card.Title>
								<Card.Description>
									Booking #{booking.booking_id}
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
								<p class="text-sm text-muted-foreground">Mark your attendance for this shift</p>
								<div class="flex gap-2">
									<Button
										size="sm"
										variant="outline"
										disabled={$markAttendanceMutation.isPending}
										onclick={() => handleMarkAttendance(booking.booking_id, true)}
									>
										<CheckCircleIcon class="h-4 w-4 mr-1" />
										Attended
									</Button>
									<Button
										size="sm"
										variant="outline"
										disabled={$markAttendanceMutation.isPending}
										onclick={() => handleMarkAttendance(booking.booking_id, false)}
									>
										<XCircleIcon class="h-4 w-4 mr-1" />
										Missed
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
										disabled={$markAttendanceMutation.isPending}
										onclick={() => handleMarkAttendance(booking.booking_id, true)}
									>
										<CheckCircleIcon class="h-4 w-4 mr-1" />
										Check In
									</Button>
									<Button size="sm" href="/report?bookingId={booking.booking_id}">
										Report Incident
									</Button>
								</div>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
