<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
	import AdminShiftCalendar from '$lib/components/admin/calendar/AdminShiftCalendar.svelte';
	import { createAdminShiftsQuery } from '$lib/queries/admin/shifts/adminShiftsQuery';
	import { LoadingState, ErrorState } from '$lib/components/ui';

	// Create the admin shifts query for the calendar
	const adminShiftsQuery = $derived(createAdminShiftsQuery('14')); // 2 weeks default

	const isLoading = $derived($adminShiftsQuery.isLoading);
	const isError = $derived($adminShiftsQuery.isError);
	const error = $derived($adminShiftsQuery.error || undefined);
	const shiftsData = $derived($adminShiftsQuery.data || []);

	// Handle shift updates (refresh calendar after assignment changes)
	function handleShiftUpdate() {
		$adminShiftsQuery.refetch();
	}
</script>

<svelte:head>
	<title>Admin Dashboard - Mount Moreland Night Owls</title>
</svelte:head>

<SidebarPage title="Upcoming Shifts">
	{#snippet listContent()}
		<UpcomingShifts maxItems={8} />
	{/snippet}

	<div class="p-6 space-y-6">
		<!-- Admin Calendar Content -->
		{#if isLoading}
			<LoadingState isLoading={true} loadingText="Loading shifts calendar..." className="py-16" />
		{:else if isError}
			<ErrorState
				error={error || null}
				title="Failed to load shifts"
				showRetry={true}
				onRetry={() => $adminShiftsQuery.refetch()}
				className="py-16"
			/>
		{:else}
			<AdminShiftCalendar
				shifts={shiftsData}
				selectedDayRange="14"
				onShiftUpdate={handleShiftUpdate}
			/>
		{/if}
	</div>
</SidebarPage>
