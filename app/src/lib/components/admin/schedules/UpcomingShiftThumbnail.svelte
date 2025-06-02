<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { formatDistanceToNow } from 'date-fns';
	import type { AdminShiftSlot } from '$lib/types';
	import { format } from 'date-fns';
	import { formatShiftTitleCondensed } from '$lib/utils/shiftFormatting';

	let {
		shift,
		isSelected = false,
		onSelect
	}: {
		shift: AdminShiftSlot;
		isSelected?: boolean;
		onSelect: (shift: AdminShiftSlot) => void;
	} = $props();

	const _timeOnly = $derived.by(() => {
		const date = new Date(shift.start_time);
		return format(date, 'HH:mm');
	});
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
