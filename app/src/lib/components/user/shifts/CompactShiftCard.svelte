<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { canCancelBooking } from '$lib/utils/bookings';
	import { formatTime, formatDayNight } from '$lib/utils/shiftFormatting';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PlayIcon from '@lucide/svelte/icons/play';
	import SquareIcon from '@lucide/svelte/icons/square';
	import XIcon from '@lucide/svelte/icons/x';
	import type { AvailableShiftSlot } from '$lib/services/api/user';

	// Define flexible shift interface that works with different data structures
	type Shift = Partial<AvailableShiftSlot> & {
		// Required timing fields (from either structure)
		start_time?: string;
		end_time?: string;
		// Additional optional fields for booking/shift data
		id?: number;
		booking_id?: number;
		shift_start?: string;
		shift_end?: string;
		buddy_name?: string;
		can_checkin?: boolean;
		is_active?: boolean;
		user_name?: string;
		user_phone?: string;
	};

	let {
		shift,
		type = 'available', // 'available' | 'next' | 'active'
		onBook,
		onCheckIn,
		onCheckOut,
		onCancel,
		isLoading = false
	}: {
		shift: Shift;
		type?: 'available' | 'next' | 'active';
		onBook?: (shift: AvailableShiftSlot) => void;
		onCheckIn?: () => void;
		onCheckOut?: () => void;
		onCancel?: (shiftId: number) => void;
		isLoading?: boolean;
	} = $props();

	// Check if cancellation is allowed (2 hours before start time)
	const canCancel = $derived(
		(type === 'next' || type === 'active') &&
			onCancel &&
			canCancelBooking(shift.start_time || shift.shift_start || '')
	);

	// Get the correct booking ID regardless of data structure
	const bookingId = $derived(
		('id' in shift ? shift.id : undefined) ||
			('booking_id' in shift ? shift.booking_id : undefined) ||
			0
	);

	// Helper functions with local time support (API already returns SAST)
	function formatDate(timeString: string | undefined): string {
		if (!timeString) return 'Unknown date';
		const date = new Date(timeString);

		// Always show the literal date in dd/mm/yy format
		return date.toLocaleDateString('en-GB', {
			day: '2-digit',
			month: '2-digit',
			year: '2-digit'
		});
	}

	function getTimeUntil(timeString: string | undefined): string {
		if (!timeString) return 'Unknown';
		const now = new Date();
		const time = new Date(timeString);
		const diffMs = time.getTime() - now.getTime();
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

		if (diffMs < 0) return 'Started';
		if (diffHours > 24) {
			const days = Math.floor(diffHours / 24);
			return `in ${days}d`;
		}
		if (diffHours > 0) return `in ${diffHours}h ${diffMins}m`;
		return `in ${diffMins}m`;
	}
</script>

{#if type === 'available'}
	<!-- Available Shift - Roster Style -->
	<div class="border rounded-lg p-3 hover:bg-muted/50 transition-colors">
		<!-- Date & Time Header -->
		<div class="flex items-center justify-between mb-2">
			<div class="flex items-center gap-2 text-sm">
				<span class="font-medium">{formatDayNight(shift.start_time)}</span>
				<span class="text-muted-foreground">
					{formatTime(shift.start_time)} - {formatTime(shift.end_time)}
				</span>
			</div>
			<div class="text-xs text-orange-600 dark:text-orange-400">
				{getTimeUntil(shift.start_time)}
			</div>
		</div>

		<!-- Roster Information & Action -->
		<div class="flex items-center justify-between">
			<div class="flex-1 min-w-0">
				{#if shift.is_booked}
					<!-- Show who's working -->
					<div class="text-sm text-green-700 dark:text-green-400">
						<span class="font-medium">{shift.user_name || 'Someone'}</span>
						{#if shift.buddy_name}
							<span class="text-muted-foreground"> + {shift.buddy_name}</span>
						{/if}
					</div>
				{:else}
					<!-- Open shift -->
					<div class="text-sm text-muted-foreground">
						<span class="italic">Open shift</span>
					</div>
				{/if}
			</div>

			{#if !shift.is_booked}
				<!-- Available for booking -->
				<Button
					size="sm"
					onclick={() => onBook?.(shift as AvailableShiftSlot)}
					disabled={isLoading}
					class="ml-3 shrink-0"
				>
					<PlusIcon class="h-3 w-3 mr-1" />
					{isLoading ? 'Booking...' : 'Commit'}
				</Button>
			{:else}
				<!-- Already booked -->
				<div class="text-xs text-muted-foreground ml-3">Assigned</div>
			{/if}
		</div>
	</div>
{:else if type === 'next'}
	<!-- Next Shift - Compact Card -->
	<div class="border-l-4 border-l-primary bg-primary/5 rounded-r-lg p-4">
		<div class="flex items-center justify-between mb-2">
			<span class="text-sm font-medium">My next shift</span>
			<div class="flex flex-col items-end gap-1">
				<Badge variant="secondary" class="text-xs">
					{getTimeUntil(shift.start_time)}
				</Badge>
				{#if canCancel}
					<Button
						onclick={() => onCancel?.(bookingId)}
						variant="outline"
						size="sm"
						class="text-xs px-2 py-1 h-6 text-muted-foreground hover:text-destructive hover:border-destructive"
						disabled={isLoading}
					>
						<XIcon class="h-3 w-3 mr-1" />
						Cancel
					</Button>
				{/if}
			</div>
		</div>

		<div class="space-y-2">
			<div class="flex items-center gap-2 text-sm">
				<ClockIcon class="h-3 w-3 text-muted-foreground" />
				<span
					>{formatDayNight(shift.start_time)} • {formatTime(shift.start_time)} - {formatTime(
						shift.end_time
					)}</span
				>
			</div>

			{#if 'can_checkin' in shift && shift.can_checkin}
				<Button onclick={onCheckIn} size="sm" class="w-full">
					<PlayIcon class="h-3 w-3 mr-1" />
					Check In
				</Button>
			{/if}
		</div>
	</div>
{:else if type === 'active'}
	<!-- Active Shift - Highlighted Card -->
	<div class="border-l-4 border-l-green-500 bg-green-50 dark:bg-green-950 rounded-r-lg p-4">
		<div class="flex items-center justify-between mb-2">
			<span class="text-sm font-medium">Active shift</span>
			<Badge variant="default" class="text-xs bg-green-600">On Duty</Badge>
		</div>

		<div class="space-y-3">
			<div class="flex items-center gap-2 text-sm">
				<ClockIcon class="h-3 w-3 text-muted-foreground" />
				<span>{formatTime(shift.start_time)} - {formatTime(shift.end_time)}</span>
			</div>

			<div class="flex gap-2">
				<Button onclick={onCheckOut} variant="destructive" size="sm" class="flex-1">
					<SquareIcon class="h-3 w-3 mr-1" />
					Check Out
				</Button>
				{#if canCancel}
					<Button
						onclick={() => onCancel?.(bookingId)}
						variant="outline"
						size="sm"
						class="text-muted-foreground hover:text-destructive hover:border-destructive"
						disabled={isLoading}
					>
						<XIcon class="h-3 w-3 mr-1" />
						Cancel
					</Button>
				{/if}
			</div>
		</div>
	</div>
{/if}
