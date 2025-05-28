<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import { authenticatedFetch } from '$lib/utils/api';

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
				return data as Array<{
					report_id: number;
					severity: number;
					created_at: string;
					message: string;
					user_name: string;
					schedule_name: string;
					[key: string]: any;
				}>;
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
				getSeverityLabel(report.severity),
				`#${report.report_id}`
			].join(' ').toLowerCase();
			
			return searchableText.includes(query);
		});
	});

	// Handle selecting a report from the dynamic list
	const selectReportForViewing = (report: any) => {
		goto(`/admin/reports?reportId=${report.report_id}`);
	};

	// Get current selected report ID from URL
	const currentSelectedReportId = $derived.by(() => {
		const reportIdFromUrl = page.url.searchParams.get('reportId');
		return reportIdFromUrl ? parseInt(reportIdFromUrl, 10) : undefined;
	});

	function getSeverityIcon(severity: number) {
		switch (severity) {
			case 0:
				return InfoIcon;
			case 1:
				return AlertTriangleIcon;
			case 2:
				return ShieldAlertIcon;
			default:
				return InfoIcon;
		}
	}

	function getSeverityColor(severity: number) {
		switch (severity) {
			case 0:
				return 'text-blue-600';
			case 1:
				return 'text-orange-600';
			case 2:
				return 'text-red-600';
			default:
				return 'text-gray-600';
		}
	}

	function getSeverityLabel(severity: number) {
		switch (severity) {
			case 0:
				return 'Normal';
			case 1:
				return 'Suspicion';
			case 2:
				return 'Incident';
			default:
				return 'Unknown';
		}
	}

	function formatRelativeTime(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));

		if (diffInHours < 1) return 'Just now';
		if (diffInHours < 24) return `${diffInHours}h ago`;
		if (diffInHours < 48) return 'Yesterday';
		return `${Math.floor(diffInHours / 24)}d ago`;
	}
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
					{@const SeverityIcon = getSeverityIcon(report.severity)}
					<div
						class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0 {currentSelectedReportId === report.report_id ? 'active' : ''}"
					>
						<a
							href={`/admin/reports?reportId=${report.report_id}`}
							class="flex items-center gap-2 w-full"
							onclick={(event) => {
								event.preventDefault();
								selectReportForViewing(report);
							}}
						>
							<div class="p-1 rounded {getSeverityColor(report.severity)} bg-opacity-10">
								<SeverityIcon class="h-3 w-3 {getSeverityColor(report.severity)}" />
							</div>
							<div class="flex-1 min-w-0">
								<div class="font-medium truncate">Report #{report.report_id}</div>
								<div class="text-xs text-muted-foreground truncate">
									{report.user_name} • {formatRelativeTime(report.created_at)}
								</div>
							</div>
						</a>
					</div>
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
