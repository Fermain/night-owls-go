<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import type { AvailableShiftSlot } from '$lib/utils/bookings';
	import { formatTime } from '$lib/utils/bookings';

	let {
		shifts,
		isLoading = false,
		error,
		onBookShift,
		isBooking = false
	}: {
		shifts: AvailableShiftSlot[];
		isLoading?: boolean;
		error?: Error | null;
		onBookShift?: (shift: AvailableShiftSlot) => void;
		isBooking?: boolean;
	} = $props();

	const previewShifts = $derived(shifts.slice(0, 3));
</script>

{#if isLoading}
	<Card.Root>
		<Card.Content class="text-center py-8">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
			<p class="text-sm text-muted-foreground">Loading available shifts...</p>
		</Card.Content>
	</Card.Root>
{:else if error}
	<Card.Root>
		<Card.Content class="text-center py-8">
			<AlertTriangleIcon class="h-8 w-8 mx-auto mb-2 text-destructive" />
			<h3 class="text-sm font-medium mb-1">Error loading shifts</h3>
			<p class="text-xs text-muted-foreground">{error.message}</p>
		</Card.Content>
	</Card.Root>
{:else if previewShifts.length > 0}
	<Card.Root>
		<Card.Header class="pb-3">
			<Card.Title class="text-base">Available shifts</Card.Title>
		</Card.Header>
		<Card.Content class="pt-0">
			{#each previewShifts as shift, i (`${shift.schedule_id}-${shift.start_time}`)}
				<div class="flex items-center justify-between py-3">
					<div class="flex-1">
						<div class="text-sm font-medium">{shift.schedule_name ?? 'Unknown Schedule'}</div>
						<div class="text-xs text-muted-foreground">
							{formatTime(shift.start_time ?? '')} - {formatTime(shift.end_time ?? '')}
						</div>
						<div class="text-xs text-orange-600 dark:text-orange-400">Available now</div>
					</div>
					<Button size="sm" onclick={() => onBookShift?.(shift)} disabled={isBooking}>
						{isBooking ? 'Booking...' : 'Book'}
					</Button>
				</div>
				{#if i < previewShifts.length - 1}
					<Separator class="my-2" />
				{/if}
			{/each}
			{#if shifts.length > 3}
				<div class="mt-4 text-center">
					<a href="/bookings" class="text-sm text-primary hover:underline">
						View all {shifts.length} available shifts →
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
