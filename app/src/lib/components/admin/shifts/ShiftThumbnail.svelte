<script lang="ts">
	import type { AdminShiftSlot } from '$lib/types';
	import { Badge } from '$lib/components/ui/badge';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import UserCheckIcon from '@lucide/svelte/icons/user-check';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
	import {
		getRelativeTime,
		formatShiftTimeRangeLocal,
		getShiftBookingStatus
	} from '$lib/utils/shifts';
	import { formatDayNight } from '$lib/utils/shiftFormatting';

	let {
		shift,
		onSelect
	}: {
		shift: AdminShiftSlot;
		onSelect: (shift: AdminShiftSlot) => void;
	} = $props();

	const bookingStatus = $derived(getShiftBookingStatus(shift));
</script>

<div
	class="p-3 border-b hover:bg-muted/50 transition-all duration-200 cursor-pointer group"
	onclick={() => onSelect(shift)}
	onkeydown={(e) => {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onSelect(shift);
		}
	}}
	role="button"
	tabindex="0"
	aria-label={`View details for ${shift.schedule_name} shift on ${getRelativeTime(shift.start_time)}`}
>
	<!-- Day/night name and time -->
	<div class="flex items-start justify-between gap-2 mb-2">
		<div class="min-w-0 flex-1">
			<div class="flex items-center gap-2">
				<h4 class="font-medium text-sm truncate">{formatDayNight(shift.start_time)}</h4>
				<ExternalLinkIcon
					class="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
				/>
			</div>
			<div class="flex items-center gap-1 text-xs text-muted-foreground mt-1">
				<ClockIcon class="h-3 w-3" />
				<span>{formatShiftTimeRangeLocal(shift.start_time, shift.end_time)}</span>
			</div>
		</div>
		<Badge variant={shift.is_booked ? 'default' : 'secondary'} class="text-xs">
			{getRelativeTime(shift.start_time)}
		</Badge>
	</div>

	<!-- Booking status -->
	<div class="flex items-center justify-between gap-2">
		{#if bookingStatus.status === 'available'}
			<div class="flex items-center gap-1 text-xs {bookingStatus.color}">
				<UserIcon class="h-3 w-3" />
				<span>{bookingStatus.label}</span>
			</div>
		{:else}
			<div class="flex items-center gap-1 text-xs {bookingStatus.color} min-w-0 flex-1">
				<UserCheckIcon class="h-3 w-3 flex-shrink-0" />
				<div class="min-w-0 flex-1">
					{#if bookingStatus.status === 'buddy'}
						<!-- Two-person assignment - stack vertically -->
						<div class="space-y-0.5">
							<div class="truncate font-medium">{bookingStatus.primary}</div>
							<div class="truncate text-muted-foreground">+ {bookingStatus.buddy}</div>
						</div>
					{:else}
						<!-- Single assignment -->
						<span class="truncate">{bookingStatus.label}</span>
					{/if}
				</div>
			</div>
		{/if}

		<div class="flex-shrink-0">
			{#if shift.is_recurring_reservation}
				<Badge variant="outline" class="text-xs">Recurring</Badge>
			{/if}
		</div>
	</div>
</div>
