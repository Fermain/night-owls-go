<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { canCancelBooking } from '$lib/utils/bookings';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PlayIcon from '@lucide/svelte/icons/play';
	import SquareIcon from '@lucide/svelte/icons/square';
	import XIcon from '@lucide/svelte/icons/x';

	let {
		shift,
		type = 'available', // 'available' | 'next' | 'active'
		onBook,
		onCheckIn,
		onCheckOut,
		onCancel,
		isLoading = false
	}: {
		shift: any;
		type?: 'available' | 'next' | 'active';
		onBook?: (shift: any) => void;
		onCheckIn?: () => void;
		onCheckOut?: () => void;
		onCancel?: (shiftId: number) => void;
		isLoading?: boolean;
	} = $props();

	// Check if cancellation is allowed (2 hours before start time)
	const canCancel = $derived(
		(type === 'next' || type === 'active') &&
			onCancel &&
			canCancelBooking(shift.start_time || shift.shift_start)
	);

	// Get the correct booking ID regardless of data structure
	const bookingId = $derived(shift.id || shift.booking_id);

	// Helper functions
	function formatTime(timeString: string) {
		return new Date(timeString).toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatDate(timeString: string) {
		const date = new Date(timeString);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);

		if (date.toDateString() === today.toDateString()) {
			return 'Today';
		} else if (date.toDateString() === tomorrow.toDateString()) {
			return 'Tomorrow';
		} else {
			return date.toLocaleDateString('en-GB', {
				weekday: 'short',
				month: 'short',
				day: 'numeric'
			});
		}
	}

	function getTimeUntil(timeString: string) {
		const now = new Date();
		const time = new Date(timeString);
		const diffMs = time.getTime() - now.getTime();
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

		if (diffMs < 0) return 'Started';
		if (diffHours > 24) {
			const days = Math.floor(diffHours / 24);
			return `${days}d`;
		}
		if (diffHours > 0) return `${diffHours}h ${diffMins}m`;
		return `${diffMins}m`;
	}
</script>

{#if type === 'available'}
	<!-- Available Shift - Compact List Item -->
	<div class="flex items-center justify-between py-3 border-b last:border-b-0">
		<div class="flex-1 min-w-0">
			<div class="flex items-center gap-2 text-sm">
				<span class="font-medium">{formatDate(shift.start_time)}</span>
				<span class="text-muted-foreground">
					{formatTime(shift.start_time)} - {formatTime(shift.end_time)}
				</span>
			</div>
			<div class="text-xs text-orange-600 dark:text-orange-400 mt-1">
				{getTimeUntil(shift.start_time)} away
			</div>
		</div>
		<Button size="sm" onclick={() => onBook?.(shift)} disabled={isLoading} class="ml-4 shrink-0">
			<PlusIcon class="h-3 w-3 mr-1" />
			{isLoading ? 'Booking...' : 'Commit'}
		</Button>
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
					>{formatDate(shift.start_time)} • {formatTime(shift.start_time)} - {formatTime(
						shift.end_time
					)}</span
				>
			</div>

			{#if shift.can_checkin}
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
