<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { formatDistanceToNow } from 'date-fns';
	import type { AdminShiftSlot } from '$lib/types';

	let {
		shift,
		isSelected = false,
		onSelect
	}: {
		shift: AdminShiftSlot;
		isSelected?: boolean;
		onSelect: (shift: AdminShiftSlot) => void;
	} = $props();

	function formatShiftTitleCondensed(startTimeIso: string, endTimeIso: string): string {
		if (!startTimeIso || !endTimeIso) return 'N/A';
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);
			const startDay = startDate.toLocaleDateString(undefined, { weekday: 'short' }).toUpperCase();

			const formatHourWithAmPm = (date: Date) => {
				let h = date.getHours();
				const m = date.getMinutes();
				const ampm = h >= 12 ? 'PM' : 'AM';
				h = h % 12;
				h = h ? h : 12;
				return h + (m === 0 ? '' : `:${m.toString().padStart(2, '0')}`);
			};

			const startHourStr = formatHourWithAmPm(startDate);
			const endHourStr = formatHourWithAmPm(endDate);
			const endAmPm = endDate.getHours() >= 12 ? 'PM' : 'AM';

			return `${startDay} ${startHourStr}-${endHourStr}${endAmPm}`;
		} catch (e) {
			console.error('Error formatting shift title condensed:', e);
			return 'Invalid Time';
		}
	}
</script>

<Sidebar.MenuItem>
	<Sidebar.MenuButton
		onclick={() => onSelect(shift)}
		isActive={isSelected}
		class="flex flex-col items-start h-auto py-2 w-full text-left"
	>
		<span class="font-semibold text-sm">
			{formatShiftTitleCondensed(shift.start_time, shift.end_time)}
		</span>
		<span class="text-xs text-muted-foreground">{shift.schedule_name ?? 'Unknown Schedule'}</span>
		<span class="text-xs text-muted-foreground">
			{formatDistanceToNow(new Date(shift.start_time), { addSuffix: true })}
		</span>
	</Sidebar.MenuButton>
</Sidebar.MenuItem>
