<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import ReportThumbnail from '$lib/components/admin/reports/ReportThumbnail.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';

	// Utilities with new patterns
	import { apiGet } from '$lib/utils/api';
	import { classifyError } from '$lib/utils/errors';

	// Types using our new domain types and API mappings
	import type { Report } from '$lib/types/domain';
	import type { components } from '$lib/types/api';
	import { mapAPIReportArrayToDomain } from '$lib/types/api-mappings';

	// Utilities
	import { getSeverityLabel } from '$lib/utils/reports';

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

	// Fetch shift reports using our new API utilities
	const reportsQuery = $derived(
		createQuery<Report[], Error>({
			queryKey: ['adminReportsForLayout'],
			queryFn: async () => {
				try {
					const apiReports =
						await apiGet<components['schemas']['api.AdminReportResponse'][]>('/api/admin/reports');
					return mapAPIReportArrayToDomain(apiReports);
				} catch (error) {
					throw classifyError(error);
				}
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
				report.userName || '',
				report.scheduleName || '',
				getSeverityLabel(report.severity ?? 0),
				`#${report.id}`
			]
				.join(' ')
				.toLowerCase();

			return searchableText.includes(query);
		});
	});

	// Handle selecting a report from the dynamic list
	const selectReportForViewing = (report: Report) => {
		goto(`/admin/reports?reportId=${report.id}`);
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
					Error loading reports: {$reportsQuery.error?.message || 'Unknown error'}
				</div>
			{:else if filteredReports && filteredReports.length > 0}
				{#each filteredReports as report (report.id)}
					<ReportThumbnail
						{report}
						isSelected={currentSelectedReportId === report.id}
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
