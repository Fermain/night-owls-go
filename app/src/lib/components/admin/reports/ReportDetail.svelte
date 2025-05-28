<script lang="ts">
	import { goto } from '$app/navigation';
	import { createQuery } from '@tanstack/svelte-query';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Separator } from '$lib/components/ui/separator';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import ArchiveRestoreIcon from '@lucide/svelte/icons/archive-restore';
	import { authenticatedFetch } from '$lib/utils/api';
	import ReportMap from './ReportMap.svelte';

	interface Props {
		reportId: number;
	}

	let { reportId }: Props = $props();

	// Fetch report details
	const reportQuery = $derived(
		createQuery({
			queryKey: ['adminReport', reportId],
			queryFn: async () => {
				const response = await authenticatedFetch(`/api/admin/reports/${reportId}`);
				if (!response.ok) {
					throw new Error(`Failed to fetch report: ${response.status}`);
				}
				return await response.json() as {
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
					booking_id: number;
					user_id: number;
					schedule_id: number;
					latitude?: number;
					longitude?: number;
					gps_accuracy?: number;
					gps_timestamp?: string;
				};
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
				return 'Normal';
			case 1:
				return 'Suspicion';
			case 2:
				return 'Incident';
			default:
				return 'Unknown';
		}
	}

	function formatFullDateTime(dateString: string) {
		return new Date(dateString).toLocaleString('en-ZA', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
			timeZone: 'UTC'
		});
	}

	function formatShiftDuration(startString: string, endString: string) {
		const start = new Date(startString);
		const end = new Date(endString);
		const durationMs = end.getTime() - start.getTime();
		const hours = Math.floor(durationMs / (1000 * 60 * 60));
		const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));
		return `${hours}h ${minutes}m`;
	}

	function goBack() {
		goto('/admin/reports');
	}

	async function handleArchive() {
		try {
			const response = await authenticatedFetch(`/api/admin/reports/${reportId}/archive`, {
				method: 'PUT'
			});
			
			if (response.ok) {
				// Refresh the report data
				$reportQuery.refetch();
			} else {
				console.error('Failed to archive report');
			}
		} catch (error) {
			console.error('Error archiving report:', error);
		}
	}

	async function handleUnarchive() {
		try {
			const response = await authenticatedFetch(`/api/admin/reports/${reportId}/unarchive`, {
				method: 'PUT'
			});
			
			if (response.ok) {
				// Refresh the report data
				$reportQuery.refetch();
			} else {
				console.error('Failed to unarchive report');
			}
		} catch (error) {
			console.error('Error unarchiving report:', error);
		}
	}

	// Timeline events for the shift
	const timelineEvents = $derived.by(() => {
		const report = $reportQuery.data;
		if (!report) return [];

		// Check if this is an off-shift report (no booking_id or schedule_name is "Off-Shift Report")
		const isOffShiftReport = !report.booking_id || report.schedule_name === 'Off-Shift Report';
		if (isOffShiftReport) return [];

		const shiftStart = new Date(report.shift_start);
		const reportTime = new Date(report.created_at);
		const shiftEnd = new Date(report.shift_end);

		return [
			{
				time: shiftStart,
				title: 'Shift Started',
				description: `${report.user_name} began their shift`,
				icon: CheckCircleIcon,
				color: 'text-green-600'
			},
			{
				time: reportTime,
				title: 'Incident Reported',
				description: `${getSeverityLabel(report.severity)} report submitted`,
				icon: getSeverityIcon(report.severity),
				color: getSeverityColor(report.severity).split(' ')[0]
			},
			{
				time: shiftEnd,
				title: 'Shift Ended',
				description: 'Scheduled shift completion',
				icon: CheckCircleIcon,
				color: 'text-gray-600'
			}
		].sort((a, b) => a.time.getTime() - b.time.getTime());
	});

	// Check if this is an off-shift report
	const isOffShiftReport = $derived.by(() => {
		const report = $reportQuery.data;
		if (!report) return false;
		return !report.booking_id || report.schedule_name === 'Off-Shift Report';
	});

	// GPS data from the report (replace mock data)
	const gpsData = $derived.by(() => {
		const report = $reportQuery.data;
		if (!report) return null;
		
		// Check if GPS data is available
		if (report.latitude && report.longitude) {
			return {
				latitude: report.latitude,
				longitude: report.longitude,
				accuracy: report.gps_accuracy || 0,
				timestamp: report.gps_timestamp || report.created_at
			};
		}
		
		return null;
	});
</script>

<svelte:head>
	<title>Report #{reportId} - Admin</title>
</svelte:head>

<div class="p-6">
	<div class="max-w-6xl mx-auto">
		<!-- Header -->
		<div class="mb-6">
			<Button variant="ghost" onclick={goBack} class="mb-4">
				<ArrowLeftIcon class="h-4 w-4 mr-2" />
				Back to Reports
			</Button>
			
			{#if $reportQuery.isLoading}
				<Skeleton class="h-8 w-48 mb-2" />
				<Skeleton class="h-4 w-96" />
			{:else if $reportQuery.data}
				{@const report = $reportQuery.data}
				{@const SeverityIcon = getSeverityIcon(report.severity)}
				<div class="flex items-center gap-4 mb-2">
					<div class="p-3 rounded-lg {getSeverityColor(report.severity)}">
						<SeverityIcon class="h-6 w-6" />
					</div>
					<div>
						<h1 class="text-3xl font-bold">Report #{report.report_id}</h1>
						<p class="text-muted-foreground">
							Submitted on {formatFullDateTime(report.created_at)}
						</p>
					</div>
					<div class="ml-auto flex gap-2">
						<Badge class="{getSeverityColor(report.severity)} border text-sm px-3 py-1">
							{getSeverityLabel(report.severity)}
						</Badge>
						{#if report.archived_at}
							<Badge variant="secondary" class="text-sm px-3 py-1">
								<ArchiveIcon class="h-3 w-3 mr-1" />
								Archived
							</Badge>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		{#if $reportQuery.isLoading}
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<div class="lg:col-span-2 space-y-6">
					<Card.Root class="p-6">
						<Skeleton class="h-6 w-32 mb-4" />
						<Skeleton class="h-20 w-full" />
					</Card.Root>
				</div>
				<div class="space-y-6">
					<Card.Root class="p-6">
						<Skeleton class="h-6 w-24 mb-4" />
						<Skeleton class="h-32 w-full" />
					</Card.Root>
				</div>
			</div>
		{:else if $reportQuery.isError}
			<Card.Root class="p-8">
				<div class="text-center">
					<AlertCircleIcon class="h-12 w-12 text-red-500 mx-auto mb-4" />
					<h3 class="text-lg font-semibold mb-2">Error Loading Report</h3>
					<p class="text-muted-foreground mb-4">
						{$reportQuery.error?.message || 'Failed to load report details'}
					</p>
					<Button onclick={goBack}>Return to Reports</Button>
				</div>
			</Card.Root>
		{:else if $reportQuery.data}
			{@const report = $reportQuery.data}
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Main Content -->
				<div class="lg:col-span-2 space-y-6">
					<!-- Report Message -->
					<Card.Root class="p-6">
						<Card.Header class="px-0 pt-0">
							<Card.Title class="flex items-center gap-2">
								<FileTextIcon class="h-5 w-5" />
								Incident Details
							</Card.Title>
						</Card.Header>
						<Card.Content class="px-0 pb-0">
							<div class="bg-muted/30 rounded-lg p-4 border-l-4 {
								report.severity === 2 ? 'border-l-red-500' :
								report.severity === 1 ? 'border-l-orange-500' : 'border-l-blue-500'
							}">
								<p class="text-sm leading-relaxed">{report.message}</p>
							</div>
						</Card.Content>
					</Card.Root>

					<!-- Timeline -->
					{#if !isOffShiftReport && timelineEvents.length > 0}
						<Card.Root class="p-6">
							<Card.Header class="px-0 pt-0">
								<Card.Title class="flex items-center gap-2">
									<ClockIcon class="h-5 w-5" />
									Shift Timeline
								</Card.Title>
							</Card.Header>
							<Card.Content class="px-0 pb-0">
								<div class="space-y-4">
									{#each timelineEvents as event, index (index)}
										<div class="flex items-start gap-4">
											<div class="flex flex-col items-center">
												<div class="p-2 rounded-full bg-background border-2 {
													event.color === 'text-green-600' ? 'border-green-200' :
													event.color === 'text-red-600' ? 'border-red-200' :
													event.color === 'text-orange-600' ? 'border-orange-200' :
													event.color === 'text-blue-600' ? 'border-blue-200' : 'border-gray-200'
												}">
													<event.icon class="h-4 w-4 {event.color}" />
												</div>
												{#if index < timelineEvents.length - 1}
													<div class="w-px h-8 bg-border mt-2"></div>
												{/if}
											</div>
											<div class="flex-1 min-w-0">
												<div class="flex items-center justify-between">
													<h4 class="font-medium">{event.title}</h4>
													<span class="text-sm text-muted-foreground">
														{event.time.toLocaleTimeString('en-ZA', { 
															hour: '2-digit', 
															minute: '2-digit',
															timeZone: 'UTC'
														})}
													</span>
												</div>
												<p class="text-sm text-muted-foreground">{event.description}</p>
											</div>
										</div>
									{/each}
								</div>
							</Card.Content>
						</Card.Root>
					{/if}

					<!-- Location Information (GPS) -->
					{#if gpsData}
						<Card.Root class="p-6">
							<Card.Header class="px-0 pt-0">
								<Card.Title class="flex items-center gap-2">
									<MapPinIcon class="h-5 w-5" />
									Location Information
								</Card.Title>
							</Card.Header>
							<Card.Content class="px-0 pb-0">
								<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
									<div class="space-y-3">
										<div>
											<span class="text-sm font-medium text-muted-foreground">Coordinates</span>
											<p class="text-sm font-mono">
												{gpsData.latitude.toFixed(6)}, {gpsData.longitude.toFixed(6)}
											</p>
										</div>
										<div>
											<span class="text-sm font-medium text-muted-foreground">Accuracy</span>
											<p class="text-sm">±{gpsData.accuracy}m</p>
										</div>
										<div>
											<span class="text-sm font-medium text-muted-foreground">Captured</span>
											<p class="text-sm">{formatFullDateTime(gpsData.timestamp)}</p>
										</div>
									</div>
									<div class="h-48">
										<ReportMap
											latitude={gpsData.latitude}
											longitude={gpsData.longitude}
											accuracy={gpsData.accuracy}
											severity={report.severity}
											className="h-full"
										/>
									</div>
								</div>
							</Card.Content>
						</Card.Root>
					{/if}
				</div>

				<!-- Sidebar -->
				<div class="space-y-6">
					<!-- Calendar -->
					<Card.Root class="p-6">
						<Card.Header class="px-0 pt-0">
							<Card.Title class="flex items-center gap-2">
								<CalendarIcon class="h-5 w-5" />
								Incident Date
							</Card.Title>
						</Card.Header>
						<Card.Content class="px-0 pb-0">
							<div class="w-full p-4 bg-muted/20 rounded-lg border">
								<div class="text-center">
									<CalendarIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
									<p class="text-sm font-medium">
										{new Date(report.created_at).toLocaleDateString('en-ZA', {
											weekday: 'long',
											year: 'numeric',
											month: 'long',
											day: 'numeric'
										})}
									</p>
								</div>
							</div>
							<div class="mt-4 p-3 bg-muted/20 rounded-lg">
								<p class="text-sm font-medium">
									{formatFullDateTime(report.created_at)}
								</p>
							</div>
						</Card.Content>
					</Card.Root>

					<!-- Shift Information -->
					{#if !isOffShiftReport}
						<Card.Root class="p-6">
							<Card.Header class="px-0 pt-0">
								<Card.Title class="flex items-center gap-2">
									<ClockIcon class="h-5 w-5" />
									Shift Details
								</Card.Title>
							</Card.Header>
							<Card.Content class="px-0 pb-0">
								<div class="space-y-4">
									<div>
										<span class="text-sm font-medium text-muted-foreground">Schedule</span>
										<p class="text-sm font-medium">{report.schedule_name}</p>
									</div>
									<Separator />
									<div>
										<span class="text-sm font-medium text-muted-foreground">Start Time</span>
										<p class="text-sm">{formatFullDateTime(report.shift_start)}</p>
									</div>
									<div>
										<span class="text-sm font-medium text-muted-foreground">End Time</span>
										<p class="text-sm">{formatFullDateTime(report.shift_end)}</p>
									</div>
									<div>
										<span class="text-sm font-medium text-muted-foreground">Duration</span>
										<p class="text-sm">{formatShiftDuration(report.shift_start, report.shift_end)}</p>
									</div>
								</div>
							</Card.Content>
						</Card.Root>
					{:else}
						<!-- Off-Shift Report Info -->
						<Card.Root class="p-6">
							<Card.Header class="px-0 pt-0">
								<Card.Title class="flex items-center gap-2">
									<ClockIcon class="h-5 w-5" />
									Report Type
								</Card.Title>
							</Card.Header>
							<Card.Content class="px-0 pb-0">
								<div class="p-3 bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg">
									<div class="flex items-center gap-2">
										<InfoIcon class="h-4 w-4 text-blue-600 dark:text-blue-400" />
										<div>
											<p class="text-sm font-medium text-blue-900 dark:text-blue-100">Off-Shift Report</p>
											<p class="text-xs text-blue-700 dark:text-blue-300">
												This report was submitted outside of a scheduled shift
											</p>
										</div>
									</div>
								</div>
							</Card.Content>
						</Card.Root>
					{/if}

					<!-- Reporter Information -->
					<Card.Root class="p-6">
						<Card.Header class="px-0 pt-0">
							<Card.Title class="flex items-center gap-2">
								<UserIcon class="h-5 w-5" />
								Reporter
							</Card.Title>
						</Card.Header>
						<Card.Content class="px-0 pb-0">
							<div class="space-y-4">
								<div class="flex items-center gap-3">
									<div class="p-2 rounded-full bg-muted">
										<UserIcon class="h-4 w-4" />
									</div>
									<div>
										<p class="font-medium">{report.user_name}</p>
										<p class="text-sm text-muted-foreground">Volunteer</p>
									</div>
								</div>
								<Separator />
								<div class="flex items-center gap-2">
									<PhoneIcon class="h-4 w-4 text-muted-foreground" />
									<span class="text-sm">{report.user_phone}</span>
								</div>
								<div class="flex gap-2">
									<Button variant="outline" size="sm" class="flex-1">
										<PhoneIcon class="h-4 w-4 mr-2" />
										Call
									</Button>
									<Button variant="outline" size="sm" class="flex-1">
										View Profile
									</Button>
								</div>
							</div>
						</Card.Content>
					</Card.Root>

					<!-- Quick Actions -->
					<Card.Root class="p-6">
						<Card.Header class="px-0 pt-0">
							<Card.Title>Actions</Card.Title>
						</Card.Header>
						<Card.Content class="px-0 pb-0">
							<div class="space-y-2">
								<Button variant="outline" class="w-full justify-start">
									<FileTextIcon class="h-4 w-4 mr-2" />
									Export Report
								</Button>
								<Button 
									variant="outline" 
									class="w-full justify-start"
									onclick={() => {
										// Open in a new window with a larger map
										if (gpsData) {
											const mapUrl = `https://www.openstreetmap.org/?mlat=${gpsData.latitude}&mlon=${gpsData.longitude}&zoom=18`;
											window.open(mapUrl, '_blank');
										}
									}}
								>
									<MapPinIcon class="h-4 w-4 mr-2" />
									View on Map
								</Button>
								<Button variant="outline" class="w-full justify-start">
									<UserIcon class="h-4 w-4 mr-2" />
									Contact Reporter
								</Button>
								
								{#if report.archived_at}
									<Button 
										variant="outline" 
										class="w-full justify-start text-green-600 hover:text-green-700"
										onclick={handleUnarchive}
									>
										<ArchiveRestoreIcon class="h-4 w-4 mr-2" />
										Unarchive Report
									</Button>
								{:else}
									<Button 
										variant="outline" 
										class="w-full justify-start text-orange-600 hover:text-orange-700"
										onclick={handleArchive}
									>
										<ArchiveIcon class="h-4 w-4 mr-2" />
										Archive Report
									</Button>
								{/if}
							</div>
						</Card.Content>
					</Card.Root>
				</div>
			</div>
		{/if}
	</div>
</div> 