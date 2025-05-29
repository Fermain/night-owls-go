<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import type { AvailableShiftSlot } from '$lib/utils/bookings';
	import { formatDateTime } from '$lib/utils/bookings';

	let {
		shift,
		onBook,
		isBooking = false
	}: {
		shift: AvailableShiftSlot;
		onBook?: (shift: AvailableShiftSlot) => void;
		isBooking?: boolean;
	} = $props();
</script>

<Card.Root>
	<Card.Header>
		<div class="flex items-center justify-between">
			<div>
				<Card.Title class="text-lg">{shift.schedule_name ?? 'Unknown Schedule'}</Card.Title>
				<Card.Description>Available shift slot</Card.Description>
			</div>
			<Badge variant="secondary">Available</Badge>
		</div>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="flex items-center gap-4 text-sm">
			<div class="flex items-center gap-1">
				<CalendarIcon class="h-4 w-4" />
				<span>{formatDateTime(shift.start_time ?? '')}</span>
			</div>
			<span class="text-muted-foreground">to</span>
			<div class="flex items-center gap-1">
				<ClockIcon class="h-4 w-4" />
				<span>{formatDateTime(shift.end_time ?? '')}</span>
			</div>
		</div>

		<Separator />
		<div class="flex items-center justify-between">
			<p class="text-sm text-muted-foreground">Book this shift</p>
			<Button size="sm" disabled={isBooking} onclick={() => onBook?.(shift)}>
				<PlusIcon class="h-4 w-4 mr-1" />
				{isBooking ? 'Booking...' : 'Book Shift'}
			</Button>
		</div>
	</Card.Content>
</Card.Root>
