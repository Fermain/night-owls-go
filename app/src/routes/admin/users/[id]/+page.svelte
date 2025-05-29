<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import ShieldUserIcon from '@lucide/svelte/icons/shield-user';
	import { UsersApiService } from '$lib/services/api';
	import { BookingsApiService } from '$lib/services/api/bookings';

	const userId = $derived(parseInt(page.params.id));

	// Query for user details
	const userQuery = createQuery({
		queryKey: ['admin-user', () => userId],
		queryFn: () => UsersApiService.getById(userId),
		enabled: () => !isNaN(userId)
	});

	// Query for user's bookings
	const userBookingsQuery = createQuery({
		queryKey: ['admin-user-bookings', () => userId],
		queryFn: () => BookingsApiService.getUserBookings(userId),
		enabled: () => !isNaN(userId)
	});

	const user = $derived($userQuery.data);
	const bookings = $derived($userBookingsQuery.data ?? []);
	const isLoading = $derived($userQuery.isLoading || $userBookingsQuery.isLoading);
	const error = $derived($userQuery.error || $userBookingsQuery.error);

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
		if (checkedInAt) return 'completed'; // User checked in
		if (now > end) return 'missed'; // Past shift, no check-in
		return 'pending'; // Past shift, attendance not marked
	}
</script>

<svelte:head>
	<title>User Details - Night Owls Admin</title>
</svelte:head>

<div class="container mx-auto p-6 space-y-6">
	{#if isLoading}
		<div class="space-y-4">
			<Card.Root>
				<Card.Content class="pt-6">
					<div class="animate-pulse space-y-4">
						<div class="h-6 bg-muted rounded w-1/4"></div>
						<div class="h-4 bg-muted rounded w-1/2"></div>
						<div class="h-4 bg-muted rounded w-1/3"></div>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	{:else if error}
		<Card.Root>
			<Card.Content class="pt-6">
				<p class="text-destructive text-center">
					Error loading user details: {error.message}
				</p>
			</Card.Content>
		</Card.Root>
	{:else if user}
		<!-- User Information -->
		<Card.Root>
			<Card.Header>
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						{#if user.role === 'admin'}
							<ShieldUserIcon class="h-6 w-6" />
						{:else}
							<UserIcon class="h-6 w-6" />
						{/if}
						<div>
							<Card.Title class="text-xl">{user.name || 'Unnamed User'}</Card.Title>
							<Card.Description>User ID: {user.id}</Card.Description>
						</div>
					</div>
					<Badge variant={user.role === 'admin' ? 'default' : 'secondary'}>
						{user.role?.toUpperCase() || 'USER'}
					</Badge>
				</div>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="flex items-center gap-2">
						<PhoneIcon class="h-4 w-4 text-muted-foreground" />
						<span class="font-mono">{user.phone}</span>
					</div>
					{#if user.created_at}
						<div class="flex items-center gap-2">
							<CalendarIcon class="h-4 w-4 text-muted-foreground" />
							<span class="text-sm">
								Joined: {new Date(user.created_at).toLocaleDateString()}
							</span>
						</div>
					{/if}
				</div>

				<div class="flex gap-2">
					<Button href="/admin/users?userId={user.id}" variant="outline" size="sm">
						Edit User
					</Button>
					<Button href="/admin/bookings/assign?userId={user.id}" variant="outline" size="sm">
						Assign Shift
					</Button>
				</div>
			</Card.Content>
		</Card.Root>

		<!-- User's Bookings -->
		<Card.Root>
			<Card.Header>
				<Card.Title>Bookings ({bookings.length})</Card.Title>
				<Card.Description>
					{user.name}'s shift bookings and attendance history
				</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if bookings.length === 0}
					<div class="text-center py-8 text-muted-foreground">
						<CalendarIcon class="h-12 w-12 mx-auto mb-4" />
						<p>No bookings found for this user.</p>
					</div>
				{:else}
					<div class="space-y-4">
						{#each bookings as booking (booking.booking_id)}
							{@const status = getShiftStatus(
								booking.shift_start,
								booking.shift_end,
								booking.checked_in_at
							)}
							<div class="border rounded-lg p-4 space-y-3">
								<div class="flex items-center justify-between">
									<div>
										<h4 class="font-semibold">{booking.schedule_name}</h4>
										<p class="text-sm text-muted-foreground">
											Booking #{booking.booking_id}
										</p>
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

								<div class="text-xs text-muted-foreground">
									Booked: {new Date(booking.created_at).toLocaleString()}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}
</div>
