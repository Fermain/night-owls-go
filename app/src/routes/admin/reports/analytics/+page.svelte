<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Badge } from '$lib/components/ui/badge';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import BarChartIcon from '@lucide/svelte/icons/bar-chart';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import TrendingDownIcon from '@lucide/svelte/icons/trending-down';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import UserIcon from '@lucide/svelte/icons/user';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';

	// Utilities with new patterns
	import { apiGet } from '$lib/utils/api';
	import { classifyError } from '$lib/utils/errors';

	// Types using our new domain types and API mappings
	import type { Report } from '$lib/types/domain';
	import type { components } from '$lib/types/api';
	import { mapAPIReportArrayToDomain } from '$lib/types/api-mappings';

	// Filter state
	let dateRangeStart = $state<string | null>(null);
	let dateRangeEnd = $state<string | null>(null);
	let timeframe = $state<string>('30d');

	const timeframeOptions = [
		{ value: '7d', label: 'Last 7 days' },
		{ value: '30d', label: 'Last 30 days' },
		{ value: '90d', label: 'Last 3 months' },
		{ value: 'custom', label: 'Custom range' }
	];

	// Fetch reports for analytics using our new API utilities
	const reportsQuery = $derived(
		createQuery<Report[], Error>({
			queryKey: ['adminReportsAnalytics', timeframe, dateRangeStart, dateRangeEnd],
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

	// Analytics calculations
	const analytics = $derived.by(() => {
		const reports = $reportsQuery.data ?? [];
		const now = new Date();

		// Filter by timeframe
		let filteredReports = reports;
		if (timeframe !== 'custom') {
			const days = parseInt(timeframe.replace('d', ''));
			const cutoff = new Date(now.getTime() - days * 24 * 60 * 60 * 1000);
			filteredReports = reports.filter((r) => new Date(r.createdAt) >= cutoff);
		} else if (dateRangeStart && dateRangeEnd) {
			const start = new Date(dateRangeStart + 'T00:00:00Z');
			const end = new Date(dateRangeEnd + 'T23:59:59Z');
			filteredReports = reports.filter((r) => {
				const date = new Date(r.createdAt);
				return date >= start && date <= end;
			});
		}

		// Basic stats
		const total = filteredReports.length;
		const critical = filteredReports.filter((r) => r.severity === 2).length;
		const warning = filteredReports.filter((r) => r.severity === 1).length;
		const info = filteredReports.filter((r) => r.severity === 0).length;

		// Trends (compare with previous period)
		const periodDays = timeframe === 'custom' ? 30 : parseInt(timeframe.replace('d', ''));
		const previousPeriodStart = new Date(now.getTime() - 2 * periodDays * 24 * 60 * 60 * 1000);
		const previousPeriodEnd = new Date(now.getTime() - periodDays * 24 * 60 * 60 * 1000);

		const previousReports = reports.filter((r) => {
			const date = new Date(r.createdAt);
			return date >= previousPeriodStart && date < previousPeriodEnd;
		});

		const previousTotal = previousReports.length;
		const totalTrend = previousTotal > 0 ? ((total - previousTotal) / previousTotal) * 100 : 0;

		// Severity distribution
		const severityDistribution = [
			{ severity: 0, label: 'Normal', count: info, color: 'bg-blue-500' },
			{ severity: 1, label: 'Suspicion', count: warning, color: 'bg-orange-500' },
			{ severity: 2, label: 'Incident', count: critical, color: 'bg-red-500' }
		];

		// Daily breakdown
		const dailyBreakdown = [];
		for (let i = periodDays - 1; i >= 0; i--) {
			const date = new Date(now.getTime() - i * 24 * 60 * 60 * 1000);
			const dayReports = filteredReports.filter((r) => {
				const reportDate = new Date(r.createdAt);
				return reportDate.toDateString() === date.toDateString();
			});

			dailyBreakdown.push({
				date: date.toISOString().split('T')[0],
				total: dayReports.length,
				critical: dayReports.filter((r) => r.severity === 2).length,
				warning: dayReports.filter((r) => r.severity === 1).length,
				info: dayReports.filter((r) => r.severity === 0).length
			});
		}

		// Top reporters
		const reporterCounts = filteredReports.reduce(
			(acc, report) => {
				const reporterName = report.userName || 'Unknown';
				acc[reporterName] = (acc[reporterName] || 0) + 1;
				return acc;
			},
			{} as Record<string, number>
		);

		const topReporters = Object.entries(reporterCounts)
			.map(([name, count]) => ({ name, count }))
			.sort((a, b) => b.count - a.count)
			.slice(0, 5);

		// Schedule breakdown
		const scheduleCounts = filteredReports.reduce(
			(acc, report) => {
				const scheduleName = report.scheduleName || 'Unknown';
				acc[scheduleName] = (acc[scheduleName] || 0) + 1;
				return acc;
			},
			{} as Record<string, number>
		);

		const scheduleBreakdown = Object.entries(scheduleCounts)
			.map(([name, count]) => ({ name, count }))
			.sort((a, b) => b.count - a.count);

		// Time of day analysis
		const hourCounts = new Array(24).fill(0);
		filteredReports.forEach((report) => {
			const hour = new Date(report.createdAt).getHours();
			hourCounts[hour]++;
		});

		const peakHour = hourCounts.indexOf(Math.max(...hourCounts));

		return {
			total,
			critical,
			warning,
			info,
			totalTrend,
			severityDistribution,
			dailyBreakdown,
			topReporters,
			scheduleBreakdown,
			peakHour,
			hourCounts
		};
	});

	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;
	}

	function formatTrend(trend: number): string {
		const sign = trend > 0 ? '+' : '';
		return `${sign}${trend.toFixed(1)}%`;
	}

	function getTrendColor(trend: number): string {
		if (trend > 0) return 'text-red-600';
		if (trend < 0) return 'text-green-600';
		return 'text-gray-600';
	}

	function getTrendIcon(trend: number) {
		if (trend > 0) return TrendingUpIcon;
		if (trend < 0) return TrendingDownIcon;
		return ClockIcon;
	}
</script>

<svelte:head>
	<title>Reports Analytics - Admin</title>
</svelte:head>

<div class="p-3 sm:p-4 lg:p-6">
	<div class="max-w-full lg:max-w-7xl mx-auto">
		<AdminPageHeader
			icon={BarChartIcon}
			heading="Reports Analytics"
			subheading="Analyze incident report patterns and trends"
		/>

		<!-- Filters -->
		<Card.Root class="p-3 sm:p-4 lg:p-6 mb-4 sm:mb-6">
			<div class="flex flex-col sm:flex-row gap-4 sm:items-end">
				<div class="space-y-2 flex-1 sm:flex-none">
					<Label>Time Period</Label>
					<Select.Root type="single" bind:value={timeframe}>
						<Select.Trigger class="w-full sm:w-40">
							{timeframeOptions.find((opt) => opt.value === timeframe)?.label ?? 'Select period'}
						</Select.Trigger>
						<Select.Content>
							{#each timeframeOptions as option (option.value)}
								<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				{#if timeframe === 'custom'}
					<div class="space-y-2 flex-1 sm:flex-none">
						<Label>Date Range</Label>
						<DateRangePicker
							initialStartDate={dateRangeStart}
							initialEndDate={dateRangeEnd}
							change={handleDateRangeChange}
							placeholderText="Select range"
						/>
					</div>
				{/if}
			</div>
		</Card.Root>

		{#if $reportsQuery.isLoading}
			<div
				class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3 sm:gap-4 lg:gap-6 mb-4 sm:mb-6"
			>
				{#each Array(4) as _, i (i)}
					<Card.Root class="p-3 sm:p-4 lg:p-6">
						<Skeleton class="h-6 w-24 mb-2" />
						<Skeleton class="h-8 w-16 mb-1" />
						<Skeleton class="h-4 w-20" />
					</Card.Root>
				{/each}
			</div>
		{:else if $reportsQuery.data}
			<!-- Key Metrics -->
			<div
				class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3 sm:gap-4 lg:gap-6 mb-4 sm:mb-6"
			>
				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium text-muted-foreground">Total Reports</p>
							<p class="text-2xl sm:text-3xl font-bold">{analytics.total}</p>
							{#if analytics.totalTrend !== undefined}
								{@const TrendIcon = getTrendIcon(analytics.totalTrend)}
								<div class="flex items-center gap-1 mt-1">
									<TrendIcon class="h-4 w-4 {getTrendColor(analytics.totalTrend)} flex-shrink-0" />
									<span class="text-xs sm:text-sm {getTrendColor(analytics.totalTrend)} truncate">
										{formatTrend(analytics.totalTrend)} vs previous period
									</span>
								</div>
							{/if}
						</div>
						<BarChartIcon class="h-6 w-6 sm:h-8 sm:w-8 text-muted-foreground flex-shrink-0" />
					</div>
				</Card.Root>

				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium text-muted-foreground">Critical Reports</p>
							<p class="text-2xl sm:text-3xl font-bold text-red-600">{analytics.critical}</p>
							<p class="text-xs sm:text-sm text-muted-foreground mt-1">
								{analytics.total > 0
									? Math.round((analytics.critical / analytics.total) * 100)
									: 0}% of total
							</p>
						</div>
						<ShieldAlertIcon class="h-6 w-6 sm:h-8 sm:w-8 text-red-600 flex-shrink-0" />
					</div>
				</Card.Root>

				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium text-muted-foreground">Warning Reports</p>
							<p class="text-2xl sm:text-3xl font-bold text-orange-600">{analytics.warning}</p>
							<p class="text-xs sm:text-sm text-muted-foreground mt-1">
								{analytics.total > 0 ? Math.round((analytics.warning / analytics.total) * 100) : 0}%
								of total
							</p>
						</div>
						<AlertTriangleIcon class="h-6 w-6 sm:h-8 sm:w-8 text-orange-600 flex-shrink-0" />
					</div>
				</Card.Root>

				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium text-muted-foreground">Peak Hour</p>
							<p class="text-2xl sm:text-3xl font-bold">
								{analytics.peakHour.toString().padStart(2, '0')}:00
							</p>
							<p class="text-xs sm:text-sm text-muted-foreground mt-1">
								{analytics.hourCounts[analytics.peakHour]} reports
							</p>
						</div>
						<ClockIcon class="h-6 w-6 sm:h-8 sm:w-8 text-muted-foreground flex-shrink-0" />
					</div>
				</Card.Root>
			</div>

			<div class="grid grid-cols-1 lg:grid-cols-2 gap-3 sm:gap-4 lg:gap-6 mb-4 sm:mb-6">
				<!-- Severity Distribution -->
				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="text-lg sm:text-xl">Severity Distribution</Card.Title>
					</Card.Header>
					<Card.Content class="px-0 pb-0">
						<div class="space-y-3 sm:space-y-4">
							{#each analytics.severityDistribution as item (item.severity)}
								<div class="space-y-2">
									<div class="flex items-center justify-between">
										<span class="text-sm font-medium">{item.label}</span>
										<span class="text-sm text-muted-foreground">{item.count}</span>
									</div>
									<div class="w-full bg-muted rounded-full h-2">
										<div
											class="{item.color} h-2 rounded-full transition-all duration-300"
											style="width: {analytics.total > 0
												? (item.count / analytics.total) * 100
												: 0}%"
										></div>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>

				<!-- Daily Trend -->
				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="text-lg sm:text-xl">Daily Trend</Card.Title>
					</Card.Header>
					<Card.Content class="px-0 pb-0">
						<div class="space-y-2">
							{#each analytics.dailyBreakdown.slice(-7) as day (day.date)}
								<div class="flex items-center justify-between gap-2">
									<span class="text-xs sm:text-sm flex-shrink-0 min-w-0">
										{new Date(day.date).toLocaleDateString('en-US', {
											weekday: 'short',
											month: 'short',
											day: 'numeric'
										})}
									</span>
									<div class="flex items-center gap-1 sm:gap-2 flex-wrap justify-end">
										{#if day.critical > 0}
											<Badge variant="destructive" class="text-xs">{day.critical}</Badge>
										{/if}
										{#if day.warning > 0}
											<Badge class="bg-orange-100 text-orange-800 text-xs">{day.warning}</Badge>
										{/if}
										{#if day.info > 0}
											<Badge variant="secondary" class="text-xs">{day.info}</Badge>
										{/if}
										<span class="text-sm font-medium w-6 sm:w-8 text-right flex-shrink-0"
											>{day.total}</span
										>
									</div>
								</div>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			</div>

			<div class="grid grid-cols-1 lg:grid-cols-2 gap-3 sm:gap-4 lg:gap-6">
				<!-- Top Reporters -->
				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="flex items-center gap-2 text-lg sm:text-xl">
							<UserIcon class="h-5 w-5" />
							Top Reporters
						</Card.Title>
					</Card.Header>
					<Card.Content class="px-0 pb-0">
						{#if analytics.topReporters.length > 0}
							<div class="space-y-3">
								{#each analytics.topReporters as reporter, index (reporter.name)}
									<div class="flex items-center justify-between gap-3">
										<div class="flex items-center gap-3 min-w-0 flex-1">
											<div
												class="w-6 h-6 rounded-full bg-muted flex items-center justify-center text-xs font-medium flex-shrink-0"
											>
												{index + 1}
											</div>
											<span class="font-medium truncate">{reporter.name}</span>
										</div>
										<Badge variant="secondary" class="text-xs flex-shrink-0"
											>{reporter.count} reports</Badge
										>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">No reports in selected period</p>
						{/if}
					</Card.Content>
				</Card.Root>

				<!-- Schedule Breakdown -->
				<Card.Root class="p-3 sm:p-4 lg:p-6">
					<Card.Header class="px-0 pt-0">
						<Card.Title class="flex items-center gap-2 text-lg sm:text-xl">
							<CalendarIcon class="h-5 w-5" />
							Reports by Schedule
						</Card.Title>
					</Card.Header>
					<Card.Content class="px-0 pb-0">
						{#if analytics.scheduleBreakdown.length > 0}
							<div class="space-y-3">
								{#each analytics.scheduleBreakdown as schedule (schedule.name)}
									<div class="space-y-2">
										<div class="flex items-center justify-between gap-3">
											<span class="font-medium truncate min-w-0 flex-1">{schedule.name}</span>
											<span class="text-sm text-muted-foreground flex-shrink-0"
												>{schedule.count}</span
											>
										</div>
										<div class="w-full bg-muted rounded-full h-2">
											<div
												class="bg-primary h-2 rounded-full transition-all duration-300"
												style="width: {analytics.total > 0
													? (schedule.count / analytics.total) * 100
													: 0}%"
											></div>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-sm text-muted-foreground">No reports in selected period</p>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>
		{/if}
	</div>
</div>
