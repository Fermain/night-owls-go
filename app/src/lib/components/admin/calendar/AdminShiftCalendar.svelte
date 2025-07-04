<script lang="ts">
	import CalendarHeader from './AdminCalendarHeader.svelte';
	import CalendarDayNames from './AdminCalendarDayNames.svelte';
	import CalendarMonthGrid from './AdminCalendarMonthGrid.svelte';
	import AdminCalendarLegend from './AdminCalendarLegend.svelte';
	import AdminShiftDetailsDialog from './AdminShiftDetailsDialog.svelte';
	import AdminShiftAssignDialog from './AdminShiftAssignDialog.svelte';
	import { generateAdminCalendarData, sanitizeDayRange } from '$lib/utils/adminCalendar';
	import type { AdminShiftSlot } from '$lib/types';

	let {
		shifts = [],
		selectedDayRange = '14', // Default to 2 weeks for admin view
		onShiftUpdate
	}: {
		shifts: AdminShiftSlot[];
		selectedDayRange: string;
		onShiftUpdate?: () => void;
	} = $props();

	// Sanitize input and generate calendar data using utilities
	const sanitizedDayRange = $derived(sanitizeDayRange(selectedDayRange));
	const calendarData = $derived(generateAdminCalendarData(shifts, sanitizedDayRange));
	const shouldShowCalendar = $derived(shifts.length > 0);

	// Simple dialog state management
	let detailsDialogOpen = $state(false);
	let assignDialogOpen = $state(false);
	let selectedShift = $state<AdminShiftSlot | null>(null);

	// Dialog event handlers
	function handleFilledShiftClick(shift: AdminShiftSlot) {
		selectedShift = shift;
		detailsDialogOpen = true;
	}

	function handleUnfilledShiftClick(shift: AdminShiftSlot) {
		selectedShift = shift;
		assignDialogOpen = true;
	}

	function handleReassignClick(shift: AdminShiftSlot) {
		detailsDialogOpen = false;
		selectedShift = shift;
		assignDialogOpen = true;
	}

	function handleAssignSuccess() {
		assignDialogOpen = false;
		detailsDialogOpen = false;
		selectedShift = null;
		onShiftUpdate?.();
	}
</script>

<!-- Only render if there are shifts -->
{#if shouldShowCalendar}
	<div class="space-y-4">
		<!-- Calendar Header -->
		<CalendarHeader firstMonthName={calendarData.firstMonthName} />

		<!-- Day Names Header -->
		<CalendarDayNames />

		<!-- Calendar Month Grids -->
		<div class="space-y-6">
			{#each calendarData.monthGrids as monthGrid (monthGrid.monthName)}
				<div>
					{#if monthGrid.monthOffset > 0}
						<h3 class="text-lg font-semibold mb-2 text-muted-foreground">{monthGrid.monthName}</h3>
					{/if}
					<CalendarMonthGrid
						{monthGrid}
						onFilledShiftClick={handleFilledShiftClick}
						onUnfilledShiftClick={handleUnfilledShiftClick}
					/>
				</div>
			{/each}
		</div>

		<!-- Calendar Legend -->
		<AdminCalendarLegend />
	</div>
{:else}
	<div class="text-center py-8 text-muted-foreground">
		<p>No shifts available for the selected time period.</p>
	</div>
{/if}

<!-- Dialogs -->
<AdminShiftDetailsDialog
	shift={selectedShift}
	bind:open={detailsDialogOpen}
	onReassignClick={handleReassignClick}
/>

<AdminShiftAssignDialog
	shift={selectedShift}
	bind:open={assignDialogOpen}
	onAssignSuccess={handleAssignSuccess}
/>
