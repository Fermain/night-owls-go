<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
	import AdminShiftCalendar from '$lib/components/admin/calendar/AdminShiftCalendar.svelte';
	import { createAdminShiftsQuery } from '$lib/queries/admin/shifts/adminShiftsQuery';
	import { LoadingState, ErrorState } from '$lib/components/ui';
	import { getPageOpenGraph } from '$lib/utils/opengraph';

	// OpenGraph tags for admin dashboard
	const ogTags = getPageOpenGraph('admin');

	// Import centralized default to avoid divergence with calendar component
	import { DEFAULT_DAY_RANGE } from '$lib/utils/adminCalendar';

	// Create the admin shifts query for the calendar with longer range
	const adminShiftsQuery = $derived(createAdminShiftsQuery(DEFAULT_DAY_RANGE));

	const isLoading = $derived($adminShiftsQuery.isLoading);
	const isError = $derived($adminShiftsQuery.isError);
	const error = $derived($adminShiftsQuery.error || undefined);
	const shiftsData = $derived($adminShiftsQuery.data || []);

	// Handle shift updates (refresh calendar after assignment changes)
	function handleShiftUpdate() {
		$adminShiftsQuery.refetch();
	}

	// Performance optimization: Memoize shift count for header
	const shiftsSummary = $derived.by(() => {
		const total = shiftsData.length;
		const filled = shiftsData.filter((shift) => shift.is_booked && shift.user_name).length;
		const unfilled = total - filled;
		const fillRate = total > 0 ? Math.round((filled / total) * 100) : 0;

		return { total, filled, unfilled, fillRate };
	});
</script>

<svelte:head>
	<title>{ogTags.title}</title>
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.description}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImageAlt}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogType}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogSiteName}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterCard}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImageAlt}
</svelte:head>

<SidebarPage title="Upcoming Shifts">
	{#snippet listContent()}
		<UpcomingShifts maxItems={8} />
	{/snippet}

	<div class="p-6 space-y-6">
		<!-- Page Header with Performance Summary -->
		<div class="border-b pb-4">
			<h1 class="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mt-2">
				<p class="text-lg text-muted-foreground">
					Complete view of all shifts - filled and unfilled
				</p>
				{#if shiftsSummary.total > 0}
					<div class="text-sm text-muted-foreground">
						<span class="font-medium">{shiftsSummary.total}</span> shifts ·
						<span class="text-green-600 font-medium">{shiftsSummary.filled}</span> filled ·
						<span class="text-red-600 font-medium">{shiftsSummary.unfilled}</span> unfilled ({shiftsSummary.fillRate}%
						filled)
					</div>
				{/if}
			</div>
		</div>

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
				selectedDayRange={DEFAULT_DAY_RANGE}
				onShiftUpdate={handleShiftUpdate}
			/>
		{/if}
	</div>
</SidebarPage>
