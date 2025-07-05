<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import type { AdminShiftSlot } from '$lib/types';
	import { formatShiftTitleCondensed } from '$lib/utils/shiftFormatting';
	import { generateThumbnailDateInfo } from '$lib/utils/adminDialogs';

	let {
		shift,
		isSelected = false,
		onSelect
	}: {
		shift: AdminShiftSlot;
		isSelected?: boolean;
		onSelect: (shift: AdminShiftSlot) => void;
	} = $props();

	const thumbnailDate = $derived(generateThumbnailDateInfo(shift));
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
		<div class="flex items-center gap-2 text-xs text-muted-foreground mt-1 w-full">
			<span class="font-mono">{thumbnailDate.shortDate}</span>
			<span class="text-muted-foreground/60">•</span>
			<span
				class={`font-medium ${thumbnailDate.isToday ? 'text-red-600' : thumbnailDate.isSoon ? 'text-orange-600' : ''}`}
			>
				{thumbnailDate.relativeTime}
			</span>
		</div>
		<span class="text-xs text-muted-foreground">{shift.schedule_name ?? 'Unknown Schedule'}</span>
	</Sidebar.MenuButton>
</Sidebar.MenuItem>
