<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import SearchIcon from '@lucide/svelte/icons/search';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import { authenticatedFetch } from '$lib/utils/api';
	import ReportDetail from '$lib/components/admin/reports/ReportDetail.svelte';
	import ReportsMapOverview from '$lib/components/admin/reports/ReportsMapOverview.svelte';

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

	// Filter options
	const severityOptions = [
		{ value: 'all', label: 'All Severities', icon: FileTextIcon, color: 'text-gray-600' },
		{ value: '0', label: 'Info', icon: InfoIcon, color: 'text-blue-600' },
		{ value: '1', label: 'Warning', icon: AlertTriangleIcon, color: 'text-orange-600' },
		{ value: '2', label: 'Critical', icon: ShieldAlertIcon, color: 'text-red-600' }
	];

	const scheduleOptions = [
		{ value: 'all', label: 'All Schedules' },
		{ value: 'Old schedule', label: 'Old schedule' },
		{ value: 'New schedule', label: 'New schedule' }
	];

	const sortOptions = [
		{ value: 'newest', label: 'Newest First' },
		{ value: 'oldest', label: 'Oldest First' },
		{ value: 'severity', label: 'Severity (High to Low)' },
		{ value: 'schedule', label: 'Schedule Name' }
	];

	// Fetch shift reports from the real API
	const reportsQuery = $derived(
		createQuery({
			queryKey: ['adminReports', showArchived, searchQuery, severityFilter, scheduleFilter, dateRangeStart, dateRangeEnd, sortBy],
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
					[key: string]: any;
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
						].join(' ').toLowerCase();
						
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
				return 'text-blue-600 bg-blue-50 border-blue-200';
			case 1:
				return 'text-orange-600 bg-orange-50 border-orange-200';
			case 2:
				return 'text-red-600 bg-red-50 border-red-200';
			default:
				return 'text-gray-600 bg-gray-50 border-gray-200';
		}
	}

	function getSeverityLabel(severity: number) {
		switch (severity) {
			case 0:
				return 'Info';
			case 1:
				return 'Warning';
			case 2:
				return 'Critical';
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

	// Summary stats with enhanced calculations
	const reportStats = $derived.by(() => {
		const reports = $reportsQuery.data ?? [];
		const now = new Date();
		const last24h = new Date(now.getTime() - 24 * 60 * 60 * 1000);
		const last7d = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

		return {
			total: reports.length,
			critical: reports.filter((r) => r.severity === 2).length,
			warning: reports.filter((r) => r.severity === 1).length,
			info: reports.filter((r) => r.severity === 0).length,
			last24h: reports.filter((r) => new Date(r.created_at) >= last24h).length,
			last7d: reports.filter((r) => new Date(r.created_at) >= last7d).length,
			criticalLast24h: reports.filter((r) => r.severity === 2 && new Date(r.created_at) >= last24h).length
		};
	});

	// Event handlers
	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;
	}

	function clearFilters() {
		searchQuery = '';
		severityFilter = 'all';
		scheduleFilter = 'all';
		dateRangeStart = null;
		dateRangeEnd = null;
		sortBy = 'newest';
	}

	function viewReportDetail(reportId: number) {
		goto(`/admin/reports?reportId=${reportId}`);
	}

	// Check if any filters are active
	const hasActiveFilters = $derived.by(() => {
		return searchQuery.trim() !== '' || 
			   severityFilter !== 'all' || 
			   scheduleFilter !== 'all' || 
			   dateRangeStart !== null || 
			   dateRangeEnd !== null ||
			   sortBy !== 'newest';
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
				<div class="flex items-center justify-between">
					<div>
						<h1 class="text-3xl font-bold mb-2">
							{showArchived ? 'Archived' : 'Active'} Incident Reports
						</h1>
						<p class="text-muted-foreground">
							{showArchived 
								? 'View and manage archived incident reports' 
								: 'Monitor and analyze incident reports submitted by volunteers during shifts'
							}
						</p>
					</div>
					<div class="flex gap-2">
						<div class="flex border rounded-lg p-1">
							<Button 
								variant={!showArchived ? 'default' : 'ghost'} 
								size="sm"
								onclick={() => showArchived = false}
							>
								<FileTextIcon class="h-4 w-4 mr-2" />
								Active
							</Button>
							<Button 
								variant={showArchived ? 'default' : 'ghost'} 
								size="sm"
								onclick={() => showArchived = true}
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
									onclick={() => viewMode = 'list'}
								>
									<FileTextIcon class="h-4 w-4 mr-2" />
									List
								</Button>
								<Button 
									variant={viewMode === 'map' ? 'default' : 'ghost'} 
									size="sm"
									onclick={() => viewMode = 'map'}
								>
									<MapPinIcon class="h-4 w-4 mr-2" />
									Map
								</Button>
							</div>
						{/if}
						<Button onclick={() => goto('/admin/reports/analytics')} variant="outline">
							<TrendingUpIcon class="h-4 w-4 mr-2" />
							View Analytics
						</Button>
					</div>
				</div>
			</div>

			<!-- Enhanced Summary Stats -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
				<Card.Root class="p-4 hover:shadow-md transition-shadow">
					<div class="flex items-center gap-3">
						<div class="p-2 rounded-lg bg-blue-50">
							<FileTextIcon class="h-6 w-6 text-blue-600" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Total Reports</p>
							<p class="text-2xl font-bold">{reportStats.total}</p>
							<p class="text-xs text-muted-foreground">
								{reportStats.last24h} in last 24h
							</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-4 hover:shadow-md transition-shadow">
					<div class="flex items-center gap-3">
						<div class="p-2 rounded-lg bg-red-50">
							<ShieldAlertIcon class="h-6 w-6 text-red-600" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Critical</p>
							<p class="text-2xl font-bold text-red-600">{reportStats.critical}</p>
							<p class="text-xs text-muted-foreground">
								{reportStats.criticalLast24h} in last 24h
							</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-4 hover:shadow-md transition-shadow">
					<div class="flex items-center gap-3">
						<div class="p-2 rounded-lg bg-orange-50">
							<AlertTriangleIcon class="h-6 w-6 text-orange-600" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Warning</p>
							<p class="text-2xl font-bold text-orange-600">{reportStats.warning}</p>
							<p class="text-xs text-muted-foreground">
								{Math.round((reportStats.warning / Math.max(reportStats.total, 1)) * 100)}% of total
							</p>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="p-4 hover:shadow-md transition-shadow">
					<div class="flex items-center gap-3">
						<div class="p-2 rounded-lg bg-green-50">
							<TrendingUpIcon class="h-6 w-6 text-green-600" />
						</div>
						<div>
							<p class="text-sm font-medium text-muted-foreground">Weekly Activity</p>
							<p class="text-2xl font-bold text-green-600">{reportStats.last7d}</p>
							<p class="text-xs text-muted-foreground">
								Reports this week
							</p>
						</div>
					</div>
				</Card.Root>
			</div>

			<!-- Enhanced Filters -->
			<Card.Root class="p-6 mb-6">
				<div class="space-y-4">
					<div class="flex items-center gap-2 mb-4">
						<FilterIcon class="h-5 w-5" />
						<h3 class="text-lg font-semibold">Filters & Search</h3>
						{#if hasActiveFilters}
							<Badge variant="secondary" class="ml-2">
								{($reportsQuery.data?.length ?? 0)} results
							</Badge>
						{/if}
					</div>

					<!-- Search Bar -->
					<div class="relative">
						<SearchIcon class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
						<Input
							bind:value={searchQuery}
							placeholder="Search reports by message, user, schedule, or severity..."
							class="pl-10"
						/>
					</div>

					<!-- Filter Controls -->
					<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
						<div class="space-y-2">
							<Label>Severity</Label>
							<Select.Root type="single" bind:value={severityFilter}>
								<Select.Trigger>
									{#if severityFilter === 'all'}
										All Severities
									{:else}
										{@const option = severityOptions.find(opt => opt.value === severityFilter)}
										{#if option}
											<div class="flex items-center gap-2">
												<option.icon class="h-4 w-4 {option.color}" />
												{option.label}
											</div>
										{/if}
									{/if}
								</Select.Trigger>
								<Select.Content>
									{#each severityOptions as option (option.value)}
										<Select.Item value={option.value} label={option.label}>
											<div class="flex items-center gap-2">
												<option.icon class="h-4 w-4 {option.color}" />
												{option.label}
											</div>
										</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>

						<div class="space-y-2">
							<Label>Schedule</Label>
							<Select.Root type="single" bind:value={scheduleFilter}>
								<Select.Trigger>
									{scheduleOptions.find((opt) => opt.value === scheduleFilter)?.label ?? 'Select schedule'}
								</Select.Trigger>
								<Select.Content>
									{#each scheduleOptions as option (option.value)}
										<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>

						<div class="space-y-2">
							<Label>Sort By</Label>
							<Select.Root type="single" bind:value={sortBy}>
								<Select.Trigger>
									{sortOptions.find((opt) => opt.value === sortBy)?.label ?? 'Select sort'}
								</Select.Trigger>
								<Select.Content>
									{#each sortOptions as option (option.value)}
										<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>

						<div class="space-y-2">
							<Label>Date Range</Label>
							<DateRangePicker
								initialStartDate={dateRangeStart}
								initialEndDate={dateRangeEnd}
								change={handleDateRangeChange}
								placeholderText="Select range"
							/>
						</div>

						<div class="space-y-2">
							<Label class="invisible">Actions</Label>
							{#if hasActiveFilters}
								<Button variant="outline" onclick={clearFilters} class="w-full">
									Clear Filters
								</Button>
							{:else}
								<div class="h-10"></div>
							{/if}
						</div>
					</div>
				</div>
			</Card.Root>

			<!-- Reports Content -->
			{#if viewMode === 'map'}
				<Card.Root class="p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="flex items-center gap-2">
							<MapPinIcon class="h-5 w-5" />
							Reports Map Overview
						</Card.Title>
						<Card.Description>
							Click on markers to view report details
						</Card.Description>
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
							{#if hasActiveFilters}
								<Button variant="outline" onclick={clearFilters} class="mt-4">
									Clear All Filters
								</Button>
							{/if}
						</div>
					</Card.Root>
				{:else}
					{#each $reportsQuery.data ?? [] as report (report.report_id)}
						<Card.Root class="p-6 hover:shadow-md transition-all duration-200 cursor-pointer" onclick={() => viewReportDetail(report.report_id)}>
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
								<div class="bg-muted/30 rounded-lg p-4 border-l-4 {
									report.severity === 2 ? 'border-l-red-500' :
									report.severity === 1 ? 'border-l-orange-500' : 'border-l-blue-500'
								}">
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
										<Button variant="outline" size="sm" onclick={(e) => { e.stopPropagation(); viewReportDetail(report.report_id); }}>
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
