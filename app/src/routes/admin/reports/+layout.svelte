<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import ReportThumbnail from '$lib/components/admin/reports/ReportThumbnail.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import { authenticatedFetch } from '$lib/utils/api';
	import { getSeverityLabel } from '$lib/utils/reports';
	import type { components } from '$lib/types/api';

	type ReportResponse = components['schemas']['api.ReportResponse'] & {
		user_name: string;
		schedule_name: string;
	};

	let searchTerm = $state('');
	let { children } = $props();

	// Define navigation items for the reports section
	const reportsNavItems = [
		{
			title: 'All Reports',
			url: '/admin/reports',
			icon: FileTextIcon
		},
		{
			title: 'Analytics',
			url: '/admin/reports/analytics',
			icon: TrendingUpIcon
		}
	];

	// Fetch shift reports from the real API
	const reportsQuery = $derived(
		createQuery({
			queryKey: ['adminReportsForLayout'],
			queryFn: async () => {
				const response = await authenticatedFetch('/api/admin/reports');
				if (!response.ok) {
					throw new Error(`Failed to fetch reports: ${response.status}`);
				}
				const data = await response.json();
				return data as ReportResponse[];
			}
		})
	);

	// Filtered reports for display in sidebar
	const filteredReports = $derived.by(() => {
		const reports = $reportsQuery.data ?? [];
		if (!searchTerm.trim()) return reports;

		const query = searchTerm.toLowerCase();
		return reports.filter((report) => {
			const searchableText = [
				report.message,
				report.user_name,
				report.schedule_name,
				getSeverityLabel(report.severity ?? 0),
				`#${report.report_id}`
			]
				.join(' ')
				.toLowerCase();

			return searchableText.includes(query);
		});
	});

	// Handle selecting a report from the dynamic list
	const selectReportForViewing = (report: ReportResponse) => {
		goto(`/admin/reports?reportId=${report.report_id}`);
	};

	// Get current selected report ID from URL
	const currentSelectedReportId = $derived.by(() => {
		const reportIdFromUrl = page.url.searchParams.get('reportId');
		return reportIdFromUrl ? parseInt(reportIdFromUrl, 10) : undefined;
	});
</script>

{#snippet reportsListContent()}
	<div class="flex flex-col h-full">
		<!-- Top static nav items -->
		{#each reportsNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={page.url.pathname === item.url && !currentSelectedReportId}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Reports list (potentially scrollable) -->
		<div class="flex-grow overflow-y-auto">
			{#if $reportsQuery.isLoading}
				<div class="p-4 text-sm">Loading reports...</div>
			{:else if $reportsQuery.isError}
				<div class="p-4 text-sm text-destructive">
					Error loading reports: {$reportsQuery.error.message}
				</div>
			{:else if filteredReports && filteredReports.length > 0}
				{#each filteredReports as report (report.report_id)}
					<ReportThumbnail
						{report}
						isSelected={currentSelectedReportId === report.report_id}
						onSelect={selectReportForViewing}
					/>
				{/each}
			{:else if $reportsQuery.data}
				<div class="p-4 text-sm text-muted-foreground">
					{searchTerm ? `No reports found matching "${searchTerm}".` : 'No reports found.'}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={reportsListContent} title="Reports" bind:searchTerm>
	{@render children()}
</SidebarPage>
