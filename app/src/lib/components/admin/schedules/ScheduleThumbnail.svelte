<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import type { Schedule } from '$lib/types';

	let {
		schedule,
		isSelected = false,
		onSelect
	}: {
		schedule: Schedule;
		isSelected?: boolean;
		onSelect: (schedule: Schedule) => void;
	} = $props();
</script>

<Sidebar.MenuItem>
	<Sidebar.MenuButton onclick={() => onSelect(schedule)} isActive={isSelected}>
		<CalendarDaysIcon class="mr-2 h-4 w-4 text-muted-foreground" />
		{schedule.name ?? 'Unnamed Schedule'}
	</Sidebar.MenuButton>
</Sidebar.MenuItem>
