<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';

	// Filter state
	let severityFilter = $state<string>('all');
	let timeFilter = $state<string>('all');

	// Filter options
	const severityOptions = [
		{ value: 'all', label: 'All Severities' },
		{ value: '0', label: 'Info (0)' },
		{ value: '1', label: 'Warning (1)' },
		{ value: '2', label: 'Critical (2)' }
	];

	const timeOptions = [
		{ value: 'all', label: 'All Time' },
		{ value: 'today', label: 'Today' },
		{ value: 'week', label: 'This Week' },
		{ value: 'month', label: 'This Month' }
	];

	// Fetch shift reports (simulated for now since API exists)
	const reportsQuery = $derived(
		createQuery({
			queryKey: ['shiftReports', severityFilter, timeFilter],
			queryFn: async () => {
				// This would use the real API: GET /api/admin/reports
				// For now, simulate the data structure
				await new Promise((resolve) => setTimeout(resolve, 800));

				const mockReports = [
					{
						report_id: 1,
						booking_id: 123,
						message: 'Visitor seemed intoxicated and was asked to leave. No incidents.',
						severity: 1,
						created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
						user_name: 'John Doe',
						user_phone: '+27123456789',
						shift_start: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
						schedule_name: 'Friday Night Security'
					},
					{
						report_id: 2,
						booking_id: 124,
						message: 'All quiet during shift. Routine patrol completed.',
						severity: 0,
						created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
						user_name: 'Jane Smith',
						user_phone: '+27987654321',
						shift_start: new Date(Date.now() - 26 * 60 * 60 * 1000).toISOString(),
						schedule_name: 'Thursday Night Security'
					},
					{
						report_id: 3,
						booking_id: 125,
						message:
							'Attempted break-in at rear entrance. Police called and responded. Suspect fled.',
						severity: 2,
						created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
						user_name: 'Mike Johnson',
						user_phone: '+27555666777',
						shift_start: new Date(
							Date.now() - 3 * 24 * 60 * 60 * 1000 - 2 * 60 * 60 * 1000
						).toISOString(),
						schedule_name: 'Tuesday Night Security'
					}
				];

				// Apply filters
				return mockReports.filter((report) => {
					if (severityFilter !== 'all' && report.severity.toString() !== severityFilter) {
						return false;
					}
					if (timeFilter !== 'all') {
						const reportDate = new Date(report.created_at);
						const now = new Date();
						switch (timeFilter) {
							case 'today':
								return reportDate.toDateString() === now.toDateString();
							case 'week':
								const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
								return reportDate >= weekAgo;
							case 'month':
								const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
								return reportDate >= monthAgo;
						}
					}
					return true;
				});
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

	// Summary stats
	const reportStats = $derived.by(() => {
		const reports = $reportsQuery.data ?? [];
		return {
			total: reports.length,
			critical: reports.filter((r) => r.severity === 2).length,
			warning: reports.filter((r) => r.severity === 1).length,
			info: reports.filter((r) => r.severity === 0).length
		};
	});
</script>

<svelte:head>
	<title>Admin - Reports</title>
</svelte:head>

<div class="p-6">
	<div class="max-w-6xl mx-auto">
		<div class="mb-6">
			<h1 class="text-2xl font-semibold mb-2">Shift Reports</h1>
			<p class="text-muted-foreground">
				Incident reports and updates submitted by volunteers during shifts
			</p>
		</div>

		<!-- Summary Stats -->
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
			<Card.Root class="p-4">
				<div class="flex items-center gap-3">
					<FileTextIcon class="h-8 w-8 text-muted-foreground" />
					<div>
						<p class="text-sm font-medium text-muted-foreground">Total Reports</p>
						<p class="text-2xl font-bold">{reportStats.total}</p>
					</div>
				</div>
			</Card.Root>

			<Card.Root class="p-4">
				<div class="flex items-center gap-3">
					<ShieldAlertIcon class="h-8 w-8 text-red-600" />
					<div>
						<p class="text-sm font-medium text-muted-foreground">Critical</p>
						<p class="text-2xl font-bold text-red-600">{reportStats.critical}</p>
					</div>
				</div>
			</Card.Root>

			<Card.Root class="p-4">
				<div class="flex items-center gap-3">
					<AlertTriangleIcon class="h-8 w-8 text-orange-600" />
					<div>
						<p class="text-sm font-medium text-muted-foreground">Warning</p>
						<p class="text-2xl font-bold text-orange-600">{reportStats.warning}</p>
					</div>
				</div>
			</Card.Root>

			<Card.Root class="p-4">
				<div class="flex items-center gap-3">
					<InfoIcon class="h-8 w-8 text-blue-600" />
					<div>
						<p class="text-sm font-medium text-muted-foreground">Info</p>
						<p class="text-2xl font-bold text-blue-600">{reportStats.info}</p>
					</div>
				</div>
			</Card.Root>
		</div>

		<!-- Filters -->
		<Card.Root class="p-4 mb-6">
			<div class="flex gap-4 items-end">
				<div class="space-y-2">
					<Label>Severity</Label>
					<Select.Root type="single" bind:value={severityFilter}>
						<Select.Trigger class="w-40">
							{severityOptions.find((opt) => opt.value === severityFilter)?.label ??
								'Select severity'}
						</Select.Trigger>
						<Select.Content>
							{#each severityOptions as option (option.value)}
								<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>

				<div class="space-y-2">
					<Label>Time Period</Label>
					<Select.Root type="single" bind:value={timeFilter}>
						<Select.Trigger class="w-40">
							{timeOptions.find((opt) => opt.value === timeFilter)?.label ?? 'Select time'}
						</Select.Trigger>
						<Select.Content>
							{#each timeOptions as option (option.value)}
								<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
			</div>
		</Card.Root>

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
						<h3 class="text-lg font-semibold mb-2">No Reports Found</h3>
						<p class="text-muted-foreground">No shift reports match your current filters.</p>
					</div>
				</Card.Root>
			{:else}
				{#each $reportsQuery.data ?? [] as report (report.report_id)}
					<Card.Root class="p-6">
						<div class="space-y-4">
							<!-- Header -->
							<div class="flex items-start justify-between">
								{#snippet reportHeader()}
									{@const SeverityIcon = getSeverityIcon(report.severity)}
									<div class="flex items-center gap-3">
										<SeverityIcon
											class="h-5 w-5 {getSeverityColor(report.severity).split(' ')[0]}"
										/>
										<div>
											<h3 class="font-semibold text-sm">Report #{report.report_id}</h3>
											<p class="text-xs text-muted-foreground">
												{formatRelativeTime(report.created_at)}
											</p>
										</div>
									</div>
								{/snippet}
								{@render reportHeader()}
								<div
									class="px-2 py-1 text-xs font-medium rounded-full border {getSeverityColor(
										report.severity
									)}"
								>
									{getSeverityLabel(report.severity)}
								</div>
							</div>

							<!-- Message -->
							<div class="bg-muted/30 rounded-lg p-4">
								<p class="text-sm">{report.message}</p>
							</div>

							<!-- Details -->
							<div class="flex flex-wrap gap-4 text-xs text-muted-foreground">
								<div class="flex items-center gap-1">
									<UserIcon class="h-3 w-3" />
									<span>{report.user_name}</span>
									<span class="text-muted-foreground">({report.user_phone})</span>
								</div>
								<div class="flex items-center gap-1">
									<CalendarIcon class="h-3 w-3" />
									<span>{report.schedule_name}</span>
								</div>
								<div class="flex items-center gap-1">
									<ClockIcon class="h-3 w-3" />
									<span>{formatShiftTime(report.shift_start)}</span>
								</div>
							</div>
						</div>
					</Card.Root>
				{/each}
			{/if}
		</div>
	</div>
</div>
