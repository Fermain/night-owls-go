<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import { authenticatedFetch } from '$lib/utils/api';

	let searchTerm = $state('');
	let { children } = $props();

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

	function formatRelativeTime(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));

		if (diffInHours < 1) return 'Just now';
		if (diffInHours < 24) return `${diffInHours}h ago`;
		if (diffInHours < 48) return 'Yesterday';
		return `${Math.floor(diffInHours / 24)}d ago`;
	}

	// Filter reports based on search term
	const filteredReports = $derived.by(() => {
		const reports = $reportsQuery.data ?? [];
		if (!searchTerm) return reports;

		return reports.filter(
			(report) =>
				report.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
				report.user_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
				report.schedule_name.toLowerCase().includes(searchTerm.toLowerCase())
		);
	});
</script>

{#snippet reportsListContent()}
	<div class="flex flex-col h-full">
		<!-- Header -->
		<div class="p-3 border-b bg-muted/50">
			<div class="flex items-center gap-2">
				<FileTextIcon class="h-4 w-4" />
				<span class="text-sm font-medium">Recent Reports</span>
			</div>
			<p class="text-xs text-muted-foreground">Latest incident reports</p>
		</div>

		<!-- Reports List -->
		<div class="flex-grow overflow-y-auto">
			{#if $reportsQuery.isLoading}
				<div class="p-3 space-y-3">
					{#each Array(4) as _, i (i)}
						<div class="border rounded-lg p-3">
							<Skeleton class="h-4 w-3/4 mb-2" />
							<Skeleton class="h-3 w-1/2 mb-1" />
							<Skeleton class="h-3 w-1/3" />
						</div>
					{/each}
				</div>
			{:else if $reportsQuery.isError}
				<div class="p-3 text-center">
					<FileTextIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-xs text-muted-foreground">Failed to load reports</p>
				</div>
			{:else if filteredReports.length === 0}
				<div class="p-3 text-center">
					<FileTextIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-xs text-muted-foreground">
						{searchTerm ? `No reports match "${searchTerm}"` : 'No reports found'}
					</p>
				</div>
			{:else}
				<div class="p-2">
					{#each filteredReports as report (report.report_id)}
						<div class="border rounded-lg p-3 mb-2 hover:bg-accent transition-colors">
							<div class="space-y-2">
								{#snippet reportIcon()}
									{@const SeverityIcon = getSeverityIcon(report.severity)}
									<SeverityIcon class="h-4 w-4 {getSeverityColor(report.severity)}" />
								{/snippet}

								<div class="flex items-start gap-2">
									{@render reportIcon()}
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium line-clamp-2">{report.message}</p>
									</div>
								</div>

								<div class="space-y-1 text-xs text-muted-foreground">
									<div class="flex items-center gap-1">
										<UserIcon class="h-3 w-3" />
										<span class="truncate">{report.user_name}</span>
									</div>
									<div class="flex items-center gap-1">
										<ClockIcon class="h-3 w-3" />
										<span>{formatRelativeTime(report.created_at)}</span>
									</div>
									<div class="text-xs truncate">
										{report.schedule_name}
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={reportsListContent} title="Reports" bind:searchTerm>
	{@render children()}
</SidebarPage>
