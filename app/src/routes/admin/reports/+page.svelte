<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import ArchiveIcon from '@lucide/svelte/icons/archive';

	// Utilities with new patterns
	import { apiGet } from '$lib/utils/api';
	import { classifyError } from '$lib/utils/errors';

	// Components
	import ReportDetail from '$lib/components/admin/reports/ReportDetail.svelte';
	import ReportsMapOverview from '$lib/components/admin/reports/ReportsMapOverview.svelte';
	import ReportsFilters from '$lib/components/admin/reports/ReportsFilters.svelte';
	import ReportsStats from '$lib/components/admin/reports/ReportsStats.svelte';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';

	// Types using our new domain types and API mappings
	import type { Report } from '$lib/types/domain';
	import type { components } from '$lib/types/api';
	import { mapAPIReportArrayToDomain } from '$lib/types/api-mappings';

	// Utilities
	import {
		getSeverityIcon,
		getSeverityColor,
		getSeverityLabel,
		formatRelativeTime
	} from '$lib/utils/reports';

	// Get current selected report ID from URL
	const currentSelectedReportId = $derived.by(() => {
		const reportIdFromUrl = page.url.searchParams.get('reportId');
		return reportIdFromUrl ? parseInt(reportIdFromUrl, 10) : undefined;
	});

	// Enhanced filter state
	let searchQuery = $state<string>('');
	let severityFilter = $state<string>('all');
	let scheduleFilter = $state<string>('all');
	let dateRangeStart = $state<string | null>(null);
	let dateRangeEnd = $state<string | null>(null);
	let sortBy = $state<string>('newest');
	let viewMode = $state<'list' | 'map'>('list');
	let showArchived = $state<boolean>(false);

	// Fetch shift reports using our new API utilities
	const reportsQuery = $derived(
		createQuery<Report[], Error>({
			queryKey: [
				'adminReports',
				showArchived,
				searchQuery,
				severityFilter,
				scheduleFilter,
				dateRangeStart,
				dateRangeEnd,
				sortBy
			],
			queryFn: async () => {
				try {
					const endpoint = showArchived ? '/api/admin/reports/archived' : '/api/admin/reports';
					const apiReports =
						await apiGet<components['schemas']['api.AdminReportResponse'][]>(endpoint);
					const domainReports = mapAPIReportArrayToDomain(apiReports);

					// Apply client-side filters and search
					let filteredReports = domainReports.filter((report) => {
						// Search filter
						if (searchQuery.trim()) {
							const query = searchQuery.toLowerCase();
							const searchableText = [
								report.message,
								report.userName || '',
								report.scheduleName || '',
								getSeverityLabel(report.severity)
							]
								.join(' ')
								.toLowerCase();

							if (!searchableText.includes(query)) {
								return false;
							}
						}

						// Severity filter
						if (severityFilter !== 'all' && report.severity.toString() !== severityFilter) {
							return false;
						}

						// Schedule filter
						if (scheduleFilter !== 'all' && report.scheduleName !== scheduleFilter) {
							return false;
						}

						// Date range filter
						if (dateRangeStart && dateRangeEnd) {
							const reportDate = new Date(report.createdAt);
							const startDate = new Date(dateRangeStart + 'T00:00:00Z');
							const endDate = new Date(dateRangeEnd + 'T23:59:59Z');

							if (reportDate < startDate || reportDate > endDate) {
								return false;
							}
						}

						return true;
					});

					// Apply sorting
					filteredReports.sort((a, b) => {
						switch (sortBy) {
							case 'oldest':
								return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
							case 'severity':
								return b.severity - a.severity;
							case 'schedule':
								return (a.scheduleName || '').localeCompare(b.scheduleName || '');
							case 'newest':
							default:
								return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
						}
					});

					return filteredReports;
				} catch (error) {
					throw classifyError(error);
				}
			}
		})
	);

	function formatShiftTime(dateString: string) {
		return new Date(dateString).toLocaleString('en-ZA', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'UTC'
		});
	}

	function formatFullDateTime(dateString: string) {
		return new Date(dateString).toLocaleString('en-ZA', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'UTC'
		});
	}

	function viewReportDetail(reportId: number) {
		goto(`/admin/reports?reportId=${reportId}`);
	}

	// Check if any filters are active
	const hasActiveFilters = $derived.by(() => {
		return (
			searchQuery.trim() !== '' ||
			severityFilter !== 'all' ||
			scheduleFilter !== 'all' ||
			dateRangeStart !== null ||
			dateRangeEnd !== null ||
			sortBy !== 'newest'
		);
	});
</script>

<svelte:head>
	<title>Admin - Reports</title>
</svelte:head>

{#if currentSelectedReportId}
	<!-- Show report detail view -->
	<ReportDetail reportId={currentSelectedReportId} />
{:else}
	<!-- Show reports list view -->
	<div class="p-3 sm:p-4 lg:p-6">
		<div class="max-w-full lg:max-w-7xl mx-auto">
			<div class="mb-4 sm:mb-6">
				<AdminPageHeader
					icon={FileTextIcon}
					heading="{showArchived ? 'Archived' : 'Active'} Incident Reports"
					subheading={showArchived
						? 'View and manage archived incident reports'
						: 'Monitor and analyze incident reports submitted by volunteers during shifts'}
				/>
				<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-0">
					<div class="flex flex-col sm:flex-row gap-2">
						<div class="flex border rounded-lg p-1 w-full sm:w-auto">
							<Button
								variant={!showArchived ? 'default' : 'ghost'}
								size="sm"
								onclick={() => (showArchived = false)}
								class="flex-1 sm:flex-none"
							>
								<FileTextIcon class="h-4 w-4 mr-2" />
								Active
							</Button>
							<Button
								variant={showArchived ? 'default' : 'ghost'}
								size="sm"
								onclick={() => (showArchived = true)}
								class="flex-1 sm:flex-none"
							>
								<ArchiveIcon class="h-4 w-4 mr-2" />
								Archived
							</Button>
						</div>
						{#if !showArchived}
							<div class="flex border rounded-lg p-1 w-full sm:w-auto">
								<Button
									variant={viewMode === 'list' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => (viewMode = 'list')}
									class="flex-1 sm:flex-none"
								>
									<FileTextIcon class="h-4 w-4 mr-2" />
									List
								</Button>
								<Button
									variant={viewMode === 'map' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => (viewMode = 'map')}
									class="flex-1 sm:flex-none"
								>
									<MapPinIcon class="h-4 w-4 mr-2" />
									Map
								</Button>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Compact Filters -->
			<div class="mb-4 sm:mb-6">
				<ReportsFilters
					bind:searchQuery
					bind:severityFilter
					bind:scheduleFilter
					bind:dateRangeStart
					bind:dateRangeEnd
					bind:sortBy
					resultCount={$reportsQuery.data?.length ?? 0}
				/>
			</div>

			<!-- GPS Statistics (only show in map view or when there are reports) -->
			{#if viewMode === 'map' && $reportsQuery.data && $reportsQuery.data.length > 0}
				<div class="mb-4 sm:mb-6">
					<ReportsStats reports={$reportsQuery.data} />
				</div>
			{/if}

			<!-- Reports Content -->
			{#if viewMode === 'map'}
				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="flex items-center gap-2 text-lg sm:text-xl">
							<MapPinIcon class="h-5 w-5" />
							Reports Map Overview
						</Card.Title>
						<Card.Description>
							{#if $reportsQuery.isLoading}
								Loading reports map data...
							{:else}
								{@const reportsWithGPS =
									$reportsQuery.data?.filter((r) => r.latitude && r.longitude) ?? []}
								{#if reportsWithGPS.length === 0}
									No reports contain GPS location data to display on map
								{:else}
									Showing {reportsWithGPS.length} of {$reportsQuery.data?.length ?? 0} reports with GPS
									location data • Click markers to view details
								{/if}
							{/if}
						</Card.Description>
					</Card.Header>
					<Card.Content class="px-0 pb-0">
						<div class="h-64 sm:h-80 lg:h-96">
							<ReportsMapOverview
								reports={$reportsQuery.data ?? []}
								className="h-full"
								onReportClick={(reportId) => viewReportDetail(reportId)}
							/>
						</div>
					</Card.Content>
				</Card.Root>
			{:else}
				<!-- Reports List -->
				<div class="space-y-3 sm:space-y-4">
					{#if $reportsQuery.isLoading}
						{#each Array(3) as _, i (i)}
							<Card.Root class="p-3 sm:p-4 lg:p-6">
								<div class="space-y-3">
									<div class="flex items-center justify-between">
										<Skeleton class="h-6 w-32" />
										<Skeleton class="h-6 w-16" />
									</div>
									<Skeleton class="h-4 w-full" />
									<Skeleton class="h-4 w-3/4" />
									<div class="flex flex-col sm:flex-row gap-2 sm:gap-4">
										<Skeleton class="h-4 w-full sm:w-24" />
										<Skeleton class="h-4 w-full sm:w-24" />
										<Skeleton class="h-4 w-full sm:w-24" />
									</div>
								</div>
							</Card.Root>
						{/each}
					{:else if $reportsQuery.isError}
						<Card.Root class="p-4 sm:p-6 lg:p-8">
							<div class="text-center">
								<FileTextIcon class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
								<h3 class="text-lg font-semibold mb-2">Error Loading Reports</h3>
								<p class="text-muted-foreground">
									{$reportsQuery.error?.message || 'Failed to load reports'}
								</p>
							</div>
						</Card.Root>
					{:else if ($reportsQuery.data?.length ?? 0) === 0}
						<Card.Root class="p-4 sm:p-6 lg:p-8">
							<div class="text-center">
								<FileTextIcon class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
								<h3 class="text-lg font-semibold mb-2">
									{hasActiveFilters ? 'No Reports Match Your Filters' : 'No Reports Found'}
								</h3>
								<p class="text-muted-foreground">
									{hasActiveFilters
										? 'Try adjusting your search criteria or clearing filters.'
										: 'No incident reports have been submitted yet.'}
								</p>
							</div>
						</Card.Root>
					{:else}
						{#each $reportsQuery.data ?? [] as report (report.id)}
							<Card.Root
								class="p-3 sm:p-4 lg:p-6 hover:shadow-md transition-all duration-200 cursor-pointer"
								onclick={() => viewReportDetail(report.id)}
							>
								<div class="space-y-3 sm:space-y-4">
									<!-- Header -->
									<div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
										{#snippet reportHeader()}
											{@const SeverityIcon = getSeverityIcon(report.severity)}
											<div class="flex items-center gap-3">
												<div
													class="p-2 rounded-lg {getSeverityColor(report.severity)} flex-shrink-0"
												>
													<SeverityIcon class="h-4 w-4 sm:h-5 sm:w-5" />
												</div>
												<div class="min-w-0 flex-1">
													<h3 class="font-semibold text-base sm:text-lg truncate">
														Report #{report.id}
													</h3>
													<p class="text-xs sm:text-sm text-muted-foreground">
														{formatFullDateTime(report.createdAt)}
													</p>
												</div>
											</div>
										{/snippet}
										{@render reportHeader()}

										<div class="flex items-center gap-2 flex-shrink-0">
											<Badge class="{getSeverityColor(report.severity)} border text-xs sm:text-sm">
												{getSeverityLabel(report.severity)}
											</Badge>
											<Button variant="ghost" size="sm" class="h-8 w-8 p-0 hidden sm:flex">
												<EyeIcon class="h-4 w-4" />
											</Button>
										</div>
									</div>

									<!-- Message -->
									<div
										class="bg-muted/30 rounded-lg p-3 sm:p-4 border-l-4 {report.severity === 2
											? 'border-l-red-500'
											: report.severity === 1
												? 'border-l-orange-500'
												: 'border-l-blue-500'}"
									>
										<p class="text-sm leading-relaxed">{report.message}</p>
									</div>

									<!-- Enhanced Details - Mobile-first grid -->
									<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
										<div class="flex items-center gap-2 p-3 bg-muted/20 rounded-lg">
											<UserIcon class="h-4 w-4 text-muted-foreground flex-shrink-0" />
											<div class="min-w-0 flex-1">
												<p class="font-medium truncate">{report.userName || 'Unknown'}</p>
												<p class="text-xs text-muted-foreground truncate">
													{report.userPhone || ''}
												</p>
											</div>
										</div>

										<div class="flex items-center gap-2 p-3 bg-muted/20 rounded-lg">
											<CalendarIcon class="h-4 w-4 text-muted-foreground flex-shrink-0" />
											<div class="min-w-0 flex-1">
												<p class="font-medium truncate">{report.scheduleName || 'Unknown'}</p>
												<p class="text-xs text-muted-foreground">Schedule</p>
											</div>
										</div>

										<div
											class="flex items-center gap-2 p-3 bg-muted/20 rounded-lg sm:col-span-2 lg:col-span-1"
										>
											<ClockIcon class="h-4 w-4 text-muted-foreground flex-shrink-0" />
											<div class="min-w-0 flex-1">
												<p class="font-medium truncate">
													{report.shiftStart ? formatShiftTime(report.shiftStart) : 'N/A'}
												</p>
												<p class="text-xs text-muted-foreground">Shift Time</p>
											</div>
										</div>
									</div>

									<!-- Quick Actions -->
									<div
										class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 pt-2 border-t"
									>
										<div class="flex items-center gap-2 text-xs text-muted-foreground">
											<span>Reported {formatRelativeTime(report.createdAt)}</span>
										</div>
										<div class="flex items-center gap-2">
											<Button
												variant="outline"
												size="sm"
												onclick={(e) => {
													e.stopPropagation();
													viewReportDetail(report.id);
												}}
												class="w-full sm:w-auto"
											>
												<EyeIcon class="h-4 w-4 mr-2" />
												View Details
											</Button>
										</div>
									</div>
								</div>
							</Card.Root>
						{/each}
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}
