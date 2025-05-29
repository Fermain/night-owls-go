<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import type { BookingWithScheduleResponse } from '$lib/utils/bookings';
	import {
		formatDateTime,
		getShiftStatus,
		canCancelBooking,
		BOOKING_STATUS
	} from '$lib/utils/bookings';

	let {
		booking,
		onCheckIn,
		onCancel,
		isCheckingIn = false,
		isCancelling = false
	}: {
		booking: BookingWithScheduleResponse;
		onCheckIn?: (bookingId: number) => void;
		onCancel?: (bookingId: number) => void;
		isCheckingIn?: boolean;
		isCancelling?: boolean;
	} = $props();

	const status = $derived(
		getShiftStatus(
			booking.shift_start ?? '',
			booking.shift_end ?? '',
			booking.checked_in_at ?? undefined
		)
	);

	const canCancel = $derived(canCancelBooking(booking.shift_start ?? ''));

	function getBadgeVariant(status: string) {
		switch (status) {
			case BOOKING_STATUS.COMPLETED:
				return 'default';
			case BOOKING_STATUS.ACTIVE:
				return 'destructive';
			case BOOKING_STATUS.UPCOMING:
				return 'secondary';
			case BOOKING_STATUS.MISSED:
				return 'destructive';
			default:
				return 'outline';
		}
	}
</script>

<Card.Root>
	<Card.Header>
		<div class="flex items-center justify-between">
			<div>
				<Card.Title class="text-lg">{booking.schedule_name ?? 'Unknown Schedule'}</Card.Title>
				<Card.Description>
					Booking #{booking.booking_id}
				</Card.Description>
			</div>
			<Badge variant={getBadgeVariant(status)}>
				{status.charAt(0).toUpperCase() + status.slice(1)}
			</Badge>
		</div>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="flex items-center gap-4 text-sm">
			<div class="flex items-center gap-1">
				<CalendarIcon class="h-4 w-4" />
				<span>{formatDateTime(booking.shift_start ?? '')}</span>
			</div>
			<span class="text-muted-foreground">to</span>
			<div class="flex items-center gap-1">
				<ClockIcon class="h-4 w-4" />
				<span>{formatDateTime(booking.shift_end ?? '')}</span>
			</div>
		</div>

		{#if booking.buddy_name}
			<div class="text-sm">
				<span class="font-medium">Buddy:</span>
				{booking.buddy_name}
			</div>
		{/if}

		{#if status === BOOKING_STATUS.PENDING}
			<Separator />
			<div class="flex items-center justify-between">
				<p class="text-sm text-muted-foreground">Check in for this completed shift</p>
				<div class="flex gap-2">
					<Button
						size="sm"
						variant="outline"
						disabled={isCheckingIn}
						onclick={() => onCheckIn?.(booking.booking_id ?? 0)}
					>
						<CheckCircleIcon class="h-4 w-4 mr-1" />
						Check In
					</Button>
				</div>
			</div>
		{:else if status === BOOKING_STATUS.ACTIVE}
			<Separator />
			<div class="flex items-center justify-between">
				<p class="text-sm font-medium text-primary">Shift is currently active</p>
				<div class="flex gap-2">
					<Button
						size="sm"
						variant="outline"
						disabled={isCheckingIn}
						onclick={() => onCheckIn?.(booking.booking_id ?? 0)}
					>
						<CheckCircleIcon class="h-4 w-4 mr-1" />
						Check In
					</Button>
					<Button size="sm" href="/report?bookingId={booking.booking_id}">Report Incident</Button>
				</div>
			</div>
		{:else if status === BOOKING_STATUS.UPCOMING}
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
							disabled={isCancelling}
							onclick={() => onCancel?.(booking.booking_id ?? 0)}
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
