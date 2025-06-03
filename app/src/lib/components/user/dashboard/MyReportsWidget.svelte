<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { UserApiService, type UserReport } from '$lib/services/api/user';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import UserReportsMap from './UserReportsMap.svelte';
	import UserReportDetail from './UserReportDetail.svelte';
	import { formatRelativeTime } from '$lib/utils/reports';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import ListIcon from '@lucide/svelte/icons/list';
	import FileTextIcon from '@lucide/svelte/icons/file-text';

	interface Props {
		className?: string;
	}

	let { className = '' }: Props = $props();

	// State for view mode and selected report
	let viewMode = $state<'map' | 'list'>('map');
	let selectedReportId = $state<number | null>(null);
	let showDetailModal = $state(false);

	// Fetch user's reports
	const userReportsQuery = createQuery({
		queryKey: ['user-reports'],
		queryFn: () => UserApiService.getMyReports()
	});

	// Recent reports (last 5)
	const recentReports = $derived.by(() => {
		const reports = $userReportsQuery.data ?? [];
		return reports.slice(0, 5);
	});

	// Reports with location data
	const reportsWithLocation = $derived.by(() => {
		const reports = $userReportsQuery.data ?? [];
		return reports.filter((report) => report.latitude && report.longitude);
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
				return 'bg-blue-100 text-blue-800 border-blue-200';
			case 1:
				return 'bg-orange-100 text-orange-800 border-orange-200';
			case 2:
				return 'bg-red-100 text-red-800 border-red-200';
			default:
				return 'bg-gray-100 text-gray-800 border-gray-200';
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

	function handleReportClick(report: UserReport) {
		selectedReportId = report.report_id;
		showDetailModal = true;
	}

	function handleCloseDetail() {
		showDetailModal = false;
		selectedReportId = null;
	}
</script>

<Card.Root class="overflow-hidden {className}">
	<Card.Header class="pb-3">
		<div class="flex items-center justify-between">
			<Card.Title class="flex items-center gap-2 text-lg">
				<FileTextIcon class="h-5 w-5" />
				My Reports
			</Card.Title>
			<div class="flex border rounded-lg p-1">
				<Button
					variant={viewMode === 'map' ? 'default' : 'ghost'}
					size="sm"
					onclick={() => (viewMode = 'map')}
					class="h-8 px-3"
				>
					<MapPinIcon class="h-4 w-4 mr-1" />
					Map
				</Button>
				<Button
					variant={viewMode === 'list' ? 'default' : 'ghost'}
					size="sm"
					onclick={() => (viewMode = 'list')}
					class="h-8 px-3"
				>
					<ListIcon class="h-4 w-4 mr-1" />
					List
				</Button>
			</div>
		</div>
		<Card.Description>
			{#if $userReportsQuery.isLoading}
				Loading your reports...
			{:else if ($userReportsQuery.data?.length ?? 0) === 0}
				No reports submitted yet
			{:else}
				{$userReportsQuery.data?.length ?? 0} total reports •
				{reportsWithLocation.length} with location data
			{/if}
		</Card.Description>
	</Card.Header>
	<Card.Content class="p-0">
		{#if $userReportsQuery.isLoading}
			<div class="p-4">
				<Skeleton class="h-64 w-full" />
			</div>
		{:else if $userReportsQuery.isError}
			<div class="p-4 text-center">
				<AlertTriangleIcon class="h-8 w-8 text-destructive mx-auto mb-2" />
				<p class="text-sm text-muted-foreground">Failed to load reports</p>
			</div>
		{:else if ($userReportsQuery.data?.length ?? 0) === 0}
			<div class="p-8 text-center">
				<FileTextIcon class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
				<h3 class="text-lg font-medium mb-2">No reports yet</h3>
				<p class="text-sm text-muted-foreground">
					Your submitted reports will appear here with location data on the map
				</p>
			</div>
		{:else if viewMode === 'map'}
			<UserReportsMap onReportClick={handleReportClick} />
		{:else}
			<!-- List View -->
			<div class="p-4 space-y-3">
				{#each $userReportsQuery.data || [] as report (report.report_id)}
					{@const SeverityIcon = getSeverityIcon(report.severity)}
					<button
						type="button"
						class="flex items-center justify-between p-3 bg-muted/30 rounded-lg border hover:bg-muted/50 transition-colors cursor-pointer w-full text-left"
						onclick={() => handleReportClick(report)}
					>
						<div class="flex items-center gap-3 flex-1 min-w-0">
							<div class="p-2 rounded-lg {getSeverityColor(report.severity)} flex-shrink-0">
								<SeverityIcon class="h-4 w-4" />
							</div>
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2 mb-1">
									<span class="font-medium text-sm">Report #{report.report_id}</span>
									<Badge class="{getSeverityColor(report.severity)} border text-xs">
										{getSeverityLabel(report.severity)}
									</Badge>
								</div>
								<p class="text-xs text-muted-foreground truncate">
									{report.message}
								</p>
								<div class="flex items-center gap-4 mt-1 text-xs text-muted-foreground">
									<span>{formatRelativeTime(report.created_at)}</span>
									{#if report.schedule_name}
										<span>• {report.schedule_name}</span>
									{/if}
									{#if report.latitude && report.longitude}
										<span class="flex items-center gap-1">
											<MapPinIcon class="h-3 w-3" />
											GPS
										</span>
									{/if}
								</div>
							</div>
						</div>
						<div class="flex-shrink-0 h-8 w-8 p-0 ml-2">
							<EyeIcon class="h-4 w-4" />
						</div>
					</button>
				{/each}

				{#if ($userReportsQuery.data?.length ?? 0) > 5}
					<div class="text-center pt-2">
						<p class="text-xs text-muted-foreground">
							Showing recent 5 of {$userReportsQuery.data?.length} reports
						</p>
					</div>
				{/if}
			</div>
		{/if}
	</Card.Content>
</Card.Root>

<!-- Report Detail Modal -->
<UserReportDetail
	bind:open={showDetailModal}
	reportId={selectedReportId}
	onClose={handleCloseDetail}
/>
