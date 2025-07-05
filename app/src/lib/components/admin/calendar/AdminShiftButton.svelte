<script lang="ts">
	import { formatTimeCompact } from '$lib/utils/shiftFormatting';
	import type { AdminShiftSlot } from '$lib/types';

	let {
		shift,
		isPast,
		onFilledShiftClick,
		onUnfilledShiftClick
	}: {
		shift: AdminShiftSlot;
		isPast: boolean;
		onFilledShiftClick?: (shift: AdminShiftSlot) => void;
		onUnfilledShiftClick?: (shift: AdminShiftSlot) => void;
	} = $props();

	// CORRECTED: Use proper $derived expressions (not functions)
	const isShiftFilled = $derived(shift.is_booked && shift.user_name);
	const isShiftAvailable = $derived(!shift.is_booked);
	const isShiftPartiallyFilled = $derived(shift.is_booked && !shift.user_name);

	// Use $derived.by for complex multi-line logic
	const buttonClasses = $derived.by(() => {
		const baseClasses =
			'text-center text-xs px-1 py-0.5 rounded transition-colors flex items-center justify-between cursor-pointer';

		if (isPast) {
			return `${baseClasses} bg-gray-400 text-gray-600 opacity-50`;
		}

		if (isShiftFilled) {
			// GREEN: Shift is properly filled with assigned user
			return `${baseClasses} bg-green-500 text-white border border-green-600 hover:bg-green-600 shadow-sm`;
		} else if (isShiftPartiallyFilled) {
			// YELLOW: Shift marked as booked but no user assigned (data inconsistency)
			return `${baseClasses} bg-yellow-500 text-white border border-yellow-600 hover:bg-yellow-600 shadow-sm`;
		} else {
			// RED: Shift needs attention - unfilled and urgent
			return `${baseClasses} bg-red-500 text-white border border-red-600 hover:bg-red-600 shadow-sm`;
		}
	});

	// Simple $derived expressions for display values
	const startTime = $derived(formatTimeCompact(shift.start_time));
	const endTime = $derived(formatTimeCompact(shift.end_time));

	// Use $derived.by for complex title generation
	const titleText = $derived.by(() => {
		let title = `${shift.schedule_name}: ${startTime} - ${endTime}`;
		if (isShiftFilled && shift.user_name) {
			title += `\nAssigned to: ${shift.user_name}`;
			if (shift.buddy_name) {
				title += ` + ${shift.buddy_name}`;
			}
			title += '\n\nClick to view details or reassign';
		} else if (isShiftPartiallyFilled) {
			title += '\nDATA ISSUE: Marked as booked but no user assigned';
			title += '\n\nClick to assign user';
		} else if (isShiftAvailable) {
			title += '\nUNFILLED SHIFT - Needs Assignment';
			title += '\n\nClick to assign team';
		}
		return title;
	});

	// Simple $derived expression for icon
	const shiftIcon = $derived(isShiftFilled ? '✓' : isShiftPartiallyFilled ? '?' : '!');

	// Simple $derived expression for status text
	const statusText = $derived(
		isShiftFilled && shift.user_name
			? shift.user_name
			: isShiftPartiallyFilled
				? 'Data Issue'
				: 'Needs Assignment'
	);

	// Handle click based on shift status
	function handleClick() {
		if (isPast) return; // No action for past shifts

		if (isShiftFilled) {
			// For filled shifts, show details dialog
			onFilledShiftClick?.(shift);
		} else {
			// For unfilled shifts (including partial), show assignment dialog
			onUnfilledShiftClick?.(shift);
		}
	}
</script>

<button class={buttonClasses} onclick={handleClick} disabled={isPast} title={titleText}>
	<div class="flex-1 text-left text-[10px] leading-tight min-w-0">
		<div class="font-medium">{startTime}</div>
		<div class="text-[8px] truncate text-white opacity-90">
			{statusText}
		</div>
	</div>

	<span class="flex-shrink-0 ml-1">
		<span class="text-[8px] font-bold">{shiftIcon}</span>
	</span>
</button>
