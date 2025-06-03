<script lang="ts">
	import { formatTime, formatTimeCompact } from '$lib/utils/shiftFormatting';
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
			const timeDiff = Math.abs(shiftStart - bookingStart);
			const isMatch = timeDiff < 60000; // Within 1 minute tolerance

			return isMatch;
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

	// Compute button classes based on state
	const buttonClasses = $derived(() => {
		const baseClasses =
			'text-center text-xs px-1 py-0.5 rounded transition-colors flex items-center justify-center';

		if (isBooked) {
			if (isActive) {
				return `${baseClasses} bg-green-600 text-white font-bold`;
			} else {
				return `${baseClasses} bg-green-100 text-green-800 border border-green-300 dark:bg-green-900 dark:text-green-100 dark:border-green-700`;
			}
		} else {
			return `${baseClasses} bg-gray-100 text-gray-900 border border-gray-300 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-100 dark:border-green-600 dark:hover:bg-gray-700`;
		}
	});
</script>

<button
	class={buttonClasses}
	onclick={() => !isBooked && onShiftSelect(shift)}
	disabled={isBooked || isPast}
	title={isBooked
		? 'Already booked'
		: `${formatTime(shift.start_time)} - ${formatTime(shift.end_time)}`}
>
	{#if isBooked}
		<span class="w-full leading-none text-[8px] text-left sm:text-xs sm:text-center"
			>ME: {formatTimeCompact(shift.start_time)}</span
		>
	{:else}
		<!-- Mobile: compact time format only -->
		<span class="truncate flex-1 leading-none text-[8px] text-left sm:hidden">
			{formatTimeCompact(shift.start_time)}
		</span>
		<!-- Desktop: time range format -->
		<span class="truncate flex-1 leading-none text-xs text-center hidden sm:block">
			{formatTimeCompact(shift.start_time)} - {formatTimeCompact(shift.end_time)}
		</span>
	{/if}
</button>
