<script lang="ts">
	import CalendarHeader from './AdminCalendarHeader.svelte';
	import CalendarDayNames from './AdminCalendarDayNames.svelte';
	import CalendarMonthGrid from './AdminCalendarMonthGrid.svelte';
	import AdminCalendarLegend from './AdminCalendarLegend.svelte';
	import AdminShiftDetailsDialog from './AdminShiftDetailsDialog.svelte';
	import AdminShiftAssignDialog from './AdminShiftAssignDialog.svelte';
	import type { AdminShiftSlot } from '$lib/types';
	import type {
		AdminCalendarDay,
		AdminCalendarCell,
		AdminMonthGrid
	} from './admin-calendar-types.js';

	let {
		shifts = [],
		selectedDayRange = '14', // Default to 2 weeks for admin view
		onShiftUpdate
	}: {
		shifts: AdminShiftSlot[];
		selectedDayRange: string;
		onShiftUpdate?: () => void;
	} = $props();

	// Dialog state management
	let detailsDialogOpen = $state(false);
	let assignDialogOpen = $state(false);
	let selectedShift = $state<AdminShiftSlot | null>(null);

	// Dialog handlers
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
		// Refresh the shift data
		onShiftUpdate?.();
	}

	// Helper function to format date as YYYY-MM-DD in local timezone
	function formatLocalDate(date: Date): string {
		return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
	}

	// Show calendar if there are shifts
	const shouldShowCalendar = $derived(shifts.length > 0);

	// Calculate how many months to show based on selected day range
	const monthsToShow = $derived.by(() => {
		const days = parseInt(selectedDayRange);
		if (days <= 7) return 1;
		if (days <= 31) return 2;
		if (days <= 62) return 3;
		return 3; // Max 3 months
	});

	// Calculate the date range we should show based on selectedDayRange
	const dateRange = $derived.by(() => {
		const days = parseInt(selectedDayRange);
		const startDate = new Date();
		const endDate = new Date(Date.now() + days * 24 * 60 * 60 * 1000);
		return { startDate, endDate };
	});

	// Get current date for calendar
	const today = new Date();

	// Generate separate month grids with month names in empty cells
	const calendarData = $derived.by(() => {
		const { startDate, endDate } = dateRange;
		const monthGrids: AdminMonthGrid[] = [];

		// Get month names for the first grid
		const firstMonthName = startDate.toLocaleDateString('en-US', {
			month: 'long',
			year: 'numeric'
		});

		// Calculate which months we need to show
		for (let monthOffset = 0; monthOffset < monthsToShow; monthOffset++) {
			const monthDate = new Date(startDate.getFullYear(), startDate.getMonth() + monthOffset, 1);
			const year = monthDate.getFullYear();
			const month = monthDate.getMonth();
			const monthName = monthDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });

			// Build cells for each day in the month
			const daysInMonth = new Date(year, month + 1, 0).getDate();
			const monthCells: AdminCalendarCell[] = [];

			// Add month name in the first cell
			if (monthOffset === 0) {
				monthCells.push({
					type: 'month-title',
					monthName: firstMonthName
				});
			} else {
				monthCells.push({
					type: 'month-title',
					monthName: monthName
				});
			}

			// Add empty cells to align with calendar grid
			const firstDayOfWeek = new Date(year, month, 1).getDay();
			for (let i = 1; i < firstDayOfWeek; i++) {
				monthCells.push({ type: 'empty' });
			}

			// Add actual day cells
			for (let day = 1; day <= daysInMonth; day++) {
				const date = new Date(year, month, day);
				const dateString = formatLocalDate(date);

				// Always create calendar cells to maintain grid structure
				// but only include shift data for days within our selected date range
				const isWithinRange = date >= startDate && date <= endDate;

				let dayShifts: AdminShiftSlot[] = [];

				if (isWithinRange) {
					// Find all shifts for this date using local date comparison
					dayShifts = shifts.filter((shift) => {
						const shiftDate = new Date(shift.start_time);
						return formatLocalDate(shiftDate) === dateString;
					});
				}

				const dayData: AdminCalendarDay = {
					day,
					date,
					dateString,
					shifts: dayShifts,
					isToday: date.toDateString() === today.toDateString(),
					isPast: date < today && date.toDateString() !== today.toDateString(),
					monthOffset,
					isWithinRange
				};

				monthCells.push({
					type: 'day',
					dayData
				});
			}

			// Group cells into weeks - show complete months to maintain grid alignment
			const weeks: AdminCalendarCell[][] = [];
			for (let i = 0; i < monthCells.length; i += 7) {
				const week = monthCells.slice(i, i + 7);
				// Pad the last week if necessary
				while (week.length < 7) {
					week.push({ type: 'empty' });
				}
				weeks.push(week);
			}

			// Always add the month grid to maintain proper calendar structure
			monthGrids.push({
				monthName,
				monthOffset,
				weeks
			});
		}

		return { monthGrids, firstMonthName };
	});
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
