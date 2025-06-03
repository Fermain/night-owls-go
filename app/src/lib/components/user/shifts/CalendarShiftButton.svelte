<script lang="ts">
	import { formatTime } from '$lib/utils/shiftFormatting';
	import type { AvailableShiftSlot, UserBooking } from '$lib/services/api/user';

	let {
		shift,
		userShifts,
		isPast,
		onShiftSelect
	}: {
		shift: AvailableShiftSlot;
		userShifts: UserBooking[];
		isPast: boolean;
		onShiftSelect: (shift: AvailableShiftSlot) => void;
	} = $props();

	// Check if user is booked for this shift
	const isBooked = $derived(
		userShifts.some((booking) => {
			const shiftStart = new Date(shift.start_time).getTime();
			const bookingStart = new Date(booking.shift_start).getTime();
			return Math.abs(shiftStart - bookingStart) < 60000; // Within 1 minute tolerance
		})
	);

	// Check if this shift is currently active
	const isActive = $derived.by(() => {
		const now = new Date();
		return userShifts.some((booking) => {
			const shiftStart = new Date(shift.start_time).getTime();
			const bookingStart = new Date(booking.shift_start).getTime();
			const bookingEnd = new Date(booking.shift_end);
			return (
				Math.abs(shiftStart - bookingStart) < 60000 &&
				now >= new Date(booking.shift_start) &&
				now <= bookingEnd
			);
		});
	});
</script>

<button
	class="text-xs px-1 py-0.5 rounded transition-colors flex items-center justify-between
		{isBooked
		? isActive
			? 'bg-green-600 text-white font-bold'
			: 'bg-blue-100 text-blue-800 border border-blue-300'
		: 'bg-primary/10 hover:bg-primary/20 text-primary border border-primary/30'}
	"
	onclick={() => !isBooked && onShiftSelect(shift)}
	disabled={isBooked || isPast}
	title={isBooked
		? 'Already booked'
		: `${formatTime(shift.start_time)} - ${formatTime(shift.end_time)}`}
>
	<span class="truncate flex-1 text-left leading-none">
		{formatTime(shift.start_time)}
	</span>
	{#if isBooked}
		<span class="ml-1 leading-none">🦉</span>
	{/if}
</button>
