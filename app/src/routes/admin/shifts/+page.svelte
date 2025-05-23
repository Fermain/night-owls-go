<script lang="ts">
	import { useQueryClient } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import type { AdminShiftSlot, Schedule } from '$lib/types';
	import ShiftsDashboard from '$lib/components/dashboard/ShiftsDashboard.svelte';
	import ShiftBookingForm from '$lib/components/admin/shifts/ShiftBookingForm.svelte';
	import { createDashboardShiftsQuery } from '$lib/queries/admin/shifts/dashboardShiftsQuery';
	import { createShiftDetailsQuery } from '$lib/queries/admin/shifts/shiftDetailsQuery';
	import { createSchedulesQuery } from '$lib/queries/admin/schedules/schedulesQuery';
	import { calculateDashboardMetrics } from '$lib/utils/shiftProcessing';

	// State
	let selectedShift = $state<AdminShiftSlot | null>(null);
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));

	// Schedule dialog state
	let showScheduleDialog = $state(false);
	let selectedScheduleForEdit = $state<Schedule | null>(null);
	let scheduleDialogMode = $state<'create' | 'edit'>('create');

	const queryClient = useQueryClient();

	// Queries using new centralized query hooks
	const shiftDetailsQuery = $derived(createShiftDetailsQuery(shiftStartTimeFromUrl));
	const dashboardShiftsQuery = $derived(createDashboardShiftsQuery(!shiftStartTimeFromUrl));
	const schedulesQuery = $derived(createSchedulesQuery());

	// Dashboard metrics using new utility function
	const dashboardMetrics = $derived.by(() => {
		const shifts = $dashboardShiftsQuery.data ?? [];
		return calculateDashboardMetrics(shifts);
	});

	// Schedule management functions
	function openCreateScheduleDialog() {
		selectedScheduleForEdit = null;
		scheduleDialogMode = 'create';
		showScheduleDialog = true;
	}

	function openEditScheduleDialog(schedule: Schedule) {
		selectedScheduleForEdit = schedule;
		scheduleDialogMode = 'edit';
		showScheduleDialog = true;
	}

	function closeScheduleDialog() {
		showScheduleDialog = false;
		selectedScheduleForEdit = null;
	}

	function handleBookingSuccess() {
		// Refresh shift details after successful booking
		queryClient.invalidateQueries({ queryKey: ['shiftDetails'] });
	}

	// Effects
	$effect(() => {
		if ($shiftDetailsQuery.data) {
			selectedShift = $shiftDetailsQuery.data;
		} else if (!shiftStartTimeFromUrl) {
			selectedShift = null;
		}
	});
</script>

<svelte:head>
	<title
		>Admin - {selectedShift ? `Shift: ${selectedShift.schedule_name}` : 'Shifts Dashboard'}</title
	>
</svelte:head>

{#if shiftStartTimeFromUrl}
	<!-- Individual Shift View -->
	<div class="p-6">
		<div class="max-w-4xl mx-auto">
			{#if $shiftDetailsQuery.isLoading}
				<div class="flex justify-center items-center h-64">
					<div class="text-muted-foreground">Loading shift details...</div>
				</div>
			{:else if $shiftDetailsQuery.isError}
				<div class="text-center">
					<p class="text-destructive text-lg mb-2">Error Loading Shift</p>
					<p class="text-muted-foreground">
						{$shiftDetailsQuery.error.message}
					</p>
				</div>
			{:else if !selectedShift}
				<div class="text-center">
					<p class="text-muted-foreground text-lg">Shift not found</p>
					<p class="text-sm text-muted-foreground">
						The requested shift could not be found or may have been deleted.
					</p>
				</div>
			{:else}
				<!-- Shift Booking Form Component -->
				<ShiftBookingForm {selectedShift} onBookingSuccess={handleBookingSuccess} />
			{/if}
		</div>
	</div>
{:else}
	<!-- Dashboard View using refactored component -->
	<ShiftsDashboard
		isLoading={$dashboardShiftsQuery.isLoading}
		isError={$dashboardShiftsQuery.isError}
		error={$dashboardShiftsQuery.error || undefined}
		metrics={dashboardMetrics}
	/>
{/if}
