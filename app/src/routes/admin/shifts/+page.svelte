<script lang="ts">
	import { useQueryClient } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import type { AdminShiftSlot } from '$lib/types';
	import SimplifiedShiftsDashboard from '$lib/components/dashboard/SimplifiedShiftsDashboard.svelte';
	import ShiftBookingForm from '$lib/components/admin/shifts/ShiftBookingForm.svelte';
	import { createShiftDetailsQuery } from '$lib/queries/admin/shifts/shiftDetailsQuery';
	import { createShiftsAnalyticsQuery } from '$lib/queries/admin/shifts/shiftsAnalyticsQuery';

	// State
	let selectedShift = $state<AdminShiftSlot | null>(null);
	let shiftStartTimeFromUrl = $derived(page.url.searchParams.get('shiftStartTime'));

	const queryClient = useQueryClient();

	// Queries using analytics query for simplified dashboard
	const shiftDetailsQuery = $derived(createShiftDetailsQuery(shiftStartTimeFromUrl));
	const shiftsAnalyticsQuery = $derived(createShiftsAnalyticsQuery(30)); // 30 days of analytics

	function handleBookingSuccess() {
		// Refresh shift details and analytics after successful booking
		queryClient.invalidateQueries({ queryKey: ['shiftDetails'] });
		queryClient.invalidateQueries({ queryKey: ['shifts', 'analytics'] });
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
	<!-- Simplified Dashboard View -->
	<SimplifiedShiftsDashboard
		isLoading={$shiftsAnalyticsQuery.isLoading}
		isError={$shiftsAnalyticsQuery.isError}
		error={$shiftsAnalyticsQuery.error || undefined}
		analytics={$shiftsAnalyticsQuery.data}
	/>
{/if}
