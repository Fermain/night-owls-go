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
	import { authenticatedFetch } from '$lib/utils/api';
	import ReportDetail from '$lib/components/admin/reports/ReportDetail.svelte';
	import ReportsMapOverview from '$lib/components/admin/reports/ReportsMapOverview.svelte';
	import ReportsFilters from '$lib/components/admin/reports/ReportsFilters.svelte';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';
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

	// Fetch shift reports from the real API
	const reportsQuery = $derived(
		createQuery({
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
				const endpoint = showArchived ? '/api/admin/reports/archived' : '/api/admin/reports';
				const response = await authenticatedFetch(endpoint);
				if (!response.ok) {
					throw new Error(`Failed to fetch reports: ${response.status}`);
				}
				const reports = (await response.json()) as Array<{
					report_id: number;
					severity: number;
					message: string;
					created_at: string;
					archived_at?: string;
					schedule_name: string;
					user_name: string;
					user_phone: string;
					shift_start: string;
					shift_end: string;
					latitude?: number;
					longitude?: number;
					gps_accuracy?: number;
					gps_timestamp?: string;
					user_id?: number;
					booking_id?: number;
				}>;

				// Apply client-side filters and search
				let filteredReports = reports.filter((report) => {
					// Search filter
					if (searchQuery.trim()) {
						const query = searchQuery.toLowerCase();
						const searchableText = [
							report.message,
							report.user_name,
							report.schedule_name,
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
					if (scheduleFilter !== 'all' && report.schedule_name !== scheduleFilter) {
						return false;
					}

					// Date range filter
					if (dateRangeStart && dateRangeEnd) {
						const reportDate = new Date(report.created_at);
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
							return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
						case 'severity':
							return b.severity - a.severity;
						case 'schedule':
							return a.schedule_name.localeCompare(b.schedule_name);
						case 'newest':
						default:
							return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
					}
				});

				return filteredReports;
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
	<div class="p-6">
		<div class="max-w-7xl mx-auto">
			<div class="mb-6">
				<AdminPageHeader
					icon={FileTextIcon}
					heading="{showArchived ? 'Archived' : 'Active'} Incident Reports"
					subheading={showArchived
						? 'View and manage archived incident reports'
						: 'Monitor and analyze incident reports submitted by volunteers during shifts'}
				/>
				<div class="flex items-center justify-between">
					<div class="flex gap-2">
						<div class="flex border rounded-lg p-1">
							<Button
								variant={!showArchived ? 'default' : 'ghost'}
								size="sm"
								onclick={() => (showArchived = false)}
							>
								<FileTextIcon class="h-4 w-4 mr-2" />
								Active
							</Button>
							<Button
								variant={showArchived ? 'default' : 'ghost'}
								size="sm"
								onclick={() => (showArchived = true)}
							>
								<ArchiveIcon class="h-4 w-4 mr-2" />
								Archived
							</Button>
						</div>
						{#if !showArchived}
							<div class="flex border rounded-lg p-1">
								<Button
									variant={viewMode === 'list' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => (viewMode = 'list')}
								>
									<FileTextIcon class="h-4 w-4 mr-2" />
									List
								</Button>
								<Button
									variant={viewMode === 'map' ? 'default' : 'ghost'}
									size="sm"
									onclick={() => (viewMode = 'map')}
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
			<div class="mb-6">
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

			<!-- Reports Content -->
			{#if viewMode === 'map'}
				<Card.Root class="p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="flex items-center gap-2">
							<MapPinIcon class="h-5 w-5" />
							Reports Map Overview
						</Card.Title>
						<Card.Description>Click on markers to view report details</Card.Description>
					</Card.Header>
					<Card.Content class="px-0 pb-0">
						<div class="h-96">
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
				<div class="space-y-4">
					{#if $reportsQuery.isLoading}
						{#each Array(3) as _, i (i)}
							<Card.Root class="p-6">
								<div class="space-y-3">
									<div class="flex items-center justify-between">
										<Skeleton class="h-6 w-32" />
										<Skeleton class="h-6 w-16" />
									</div>
									<Skeleton class="h-4 w-full" />
									<Skeleton class="h-4 w-3/4" />
									<div class="flex gap-4">
										<Skeleton class="h-4 w-24" />
										<Skeleton class="h-4 w-24" />
										<Skeleton class="h-4 w-24" />
									</div>
								</div>
							</Card.Root>
						{/each}
					{:else if $reportsQuery.isError}
						<Card.Root class="p-8">
							<div class="text-center">
								<FileTextIcon class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
								<h3 class="text-lg font-semibold mb-2">Error Loading Reports</h3>
								<p class="text-muted-foreground">
									{$reportsQuery.error?.message || 'Failed to load reports'}
								</p>
							</div>
						</Card.Root>
					{:else if ($reportsQuery.data?.length ?? 0) === 0}
						<Card.Root class="p-8">
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
						{#each $reportsQuery.data ?? [] as report (report.report_id)}
							<Card.Root
								class="p-6 hover:shadow-md transition-all duration-200 cursor-pointer"
								onclick={() => viewReportDetail(report.report_id)}
							>
								<div class="space-y-4">
									<!-- Header -->
									<div class="flex items-start justify-between">
										{#snippet reportHeader()}
											{@const SeverityIcon = getSeverityIcon(report.severity)}
											<div class="flex items-center gap-3">
												<div class="p-2 rounded-lg {getSeverityColor(report.severity)}">
													<SeverityIcon class="h-5 w-5" />
												</div>
												<div>
													<h3 class="font-semibold text-lg">Report #{report.report_id}</h3>
													<p class="text-sm text-muted-foreground">
														{formatFullDateTime(report.created_at)}
													</p>
												</div>
											</div>
										{/snippet}
										{@render reportHeader()}

										<div class="flex items-center gap-2">
											<Badge class="{getSeverityColor(report.severity)} border">
												{getSeverityLabel(report.severity)}
											</Badge>
											<Button variant="ghost" size="sm" class="h-8 w-8 p-0">
												<EyeIcon class="h-4 w-4" />
											</Button>
										</div>
									</div>

									<!-- Message -->
									<div
										class="bg-muted/30 rounded-lg p-4 border-l-4 {report.severity === 2
											? 'border-l-red-500'
											: report.severity === 1
												? 'border-l-orange-500'
												: 'border-l-blue-500'}"
									>
										<p class="text-sm leading-relaxed">{report.message}</p>
									</div>

									<!-- Enhanced Details -->
									<div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
										<div class="flex items-center gap-2 p-3 bg-muted/20 rounded-lg">
											<UserIcon class="h-4 w-4 text-muted-foreground" />
											<div>
												<p class="font-medium">{report.user_name}</p>
												<p class="text-xs text-muted-foreground">{report.user_phone}</p>
											</div>
										</div>

										<div class="flex items-center gap-2 p-3 bg-muted/20 rounded-lg">
											<CalendarIcon class="h-4 w-4 text-muted-foreground" />
											<div>
												<p class="font-medium">{report.schedule_name}</p>
												<p class="text-xs text-muted-foreground">Schedule</p>
											</div>
										</div>

										<div class="flex items-center gap-2 p-3 bg-muted/20 rounded-lg">
											<ClockIcon class="h-4 w-4 text-muted-foreground" />
											<div>
												<p class="font-medium">{formatShiftTime(report.shift_start)}</p>
												<p class="text-xs text-muted-foreground">Shift Time</p>
											</div>
										</div>
									</div>

									<!-- Quick Actions -->
									<div class="flex items-center justify-between pt-2 border-t">
										<div class="flex items-center gap-2 text-xs text-muted-foreground">
											<span>Reported {formatRelativeTime(report.created_at)}</span>
										</div>
										<div class="flex items-center gap-2">
											<Button
												variant="outline"
												size="sm"
												onclick={(e) => {
													e.stopPropagation();
													viewReportDetail(report.report_id);
												}}
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
