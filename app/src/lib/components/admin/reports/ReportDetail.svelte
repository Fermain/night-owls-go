<script lang="ts">
	import { goto } from '$app/navigation';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Separator } from '$lib/components/ui/separator';
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
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import InfoIcon from '@lucide/svelte/icons/info';

	// Utilities with new patterns
	import { apiGet, apiPut, apiDelete } from '$lib/utils/api';
	import { classifyError, getErrorMessage } from '$lib/utils/errors';
	import { toast } from 'svelte-sonner';

	// Components
	import ReportMap from './ReportMap.svelte';

	// Types using our new domain types and API mappings
	import type { Report } from '$lib/types/domain';
	import type { components } from '$lib/types/api';
	import { mapAPIReportToDomain } from '$lib/types/api-mappings';

	// Utilities
	import { formatShiftTime, formatRelativeTime } from '$lib/utils/dateFormatting';
	import { getSeverityIcon, getSeverityColor, getSeverityLabel } from '$lib/utils/reports';
	import { differenceInMinutes, format } from 'date-fns';

	interface Props {
		reportId: number;
	}

	let { reportId }: Props = $props();
	const queryClient = useQueryClient();

	// Fetch report details using our new API utilities
	const reportQuery = $derived(
		createQuery<Report, Error>({
			queryKey: ['adminReport', reportId],
			queryFn: async () => {
				try {
					const apiReport = await apiGet<components['schemas']['api.AdminReportResponse']>(
						`/api/admin/reports/${reportId}`
					);
					return mapAPIReportToDomain(apiReport);
				} catch (error) {
					throw classifyError(error);
				}
			}
		})
	);

	// Archive mutation using our new API utilities
	const archiveMutation = createMutation({
		mutationFn: async () => {
			await apiPut(`/api/admin/reports/${reportId}/archive`);
		},
		onSuccess: () => {
			toast.success('Report archived successfully');
			queryClient.invalidateQueries({ queryKey: ['adminReport', reportId] });
			queryClient.invalidateQueries({ queryKey: ['adminReports'] });
			queryClient.invalidateQueries({ queryKey: ['adminReportsForLayout'] });
		},
		onError: (error: Error) => {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
		}
	});

	// Unarchive mutation using our new API utilities
	const unarchiveMutation = createMutation({
		mutationFn: async () => {
			await apiPut(`/api/admin/reports/${reportId}/unarchive`);
		},
		onSuccess: () => {
			toast.success('Report unarchived successfully');
			queryClient.invalidateQueries({ queryKey: ['adminReport', reportId] });
			queryClient.invalidateQueries({ queryKey: ['adminReports'] });
			queryClient.invalidateQueries({ queryKey: ['adminReportsForLayout'] });
		},
		onError: (error: Error) => {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
		}
	});

	// Delete mutation using our new API utilities
	const deleteMutation = createMutation({
		mutationFn: async () => {
			await apiDelete(`/api/admin/reports/${reportId}`);
		},
		onSuccess: () => {
			toast.success('Report deleted successfully');
			queryClient.invalidateQueries({ queryKey: ['adminReports'] });
			queryClient.invalidateQueries({ queryKey: ['adminReportsForLayout'] });
			goto('/admin/reports');
		},
		onError: (error: Error) => {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
		}
	});

	// Calculate duration between two dates
	function calculateDuration(startDate: Date, endDate: Date): string {
		const totalMinutes = differenceInMinutes(endDate, startDate);
		const hours = Math.floor(totalMinutes / 60);
		const minutes = totalMinutes % 60;
		return `${hours}h ${minutes}m`;
	}

	// Get full color classes for severity badges
	function getSeverityFullColor(severity: number) {
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

	function goBack() {
		goto('/admin/reports');
	}

	function handleArchive() {
		if (
			confirm('Are you sure you want to archive this report? You can unarchive it later if needed.')
		) {
			$archiveMutation.mutate();
		}
	}

	function handleUnarchive() {
		if (confirm('Are you sure you want to unarchive this report?')) {
			$unarchiveMutation.mutate();
		}
	}

	function handleDelete() {
		if (
			confirm(
				'Are you sure you want to permanently delete this report? This action cannot be undone.'
			)
		) {
			$deleteMutation.mutate();
		}
	}

	// Timeline events for the shift
	const timelineEvents = $derived.by(() => {
		const report = $reportQuery.data;
		if (!report) return [];

		// Check if this is an off-shift report (no booking_id or schedule_name is "Off-Shift Report")
		const isOffShiftReport = !report.bookingId || report.scheduleName === 'Off-Shift Report';
		if (isOffShiftReport) return [];

		if (!report.shiftStart || !report.shiftEnd) return [];

		const shiftStart = new Date(report.shiftStart);
		const reportTime = new Date(report.createdAt);
		const shiftEnd = new Date(report.shiftEnd);

		return [
			{
				time: shiftStart,
				title: 'Shift Started',
				description: `${report.userName || 'Unknown'} began their shift`,
				icon: CheckCircleIcon,
				color: 'text-green-600'
			},
			{
				time: reportTime,
				title: 'Incident Reported',
				description: `${getSeverityLabel(report.severity)} report submitted`,
				icon: getSeverityIcon(report.severity),
				color: getSeverityColor(report.severity)
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
		return !report.bookingId || report.scheduleName === 'Off-Shift Report';
	});

	// GPS data from the report
	const gpsData = $derived.by(() => {
		const report = $reportQuery.data;
		if (!report) return null;

		// Check if GPS data is available
		if (report.latitude && report.longitude) {
			return {
				latitude: report.latitude,
				longitude: report.longitude,
				accuracy: report.gpsAccuracy || 0,
				timestamp: report.gpsTimestamp || report.createdAt
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
					<div class="p-3 rounded-lg {getSeverityFullColor(report.severity)}">
						<SeverityIcon class="h-6 w-6" />
					</div>
					<div>
						<h1 class="text-3xl font-bold">Report #{report.id}</h1>
						<p class="text-muted-foreground">
							Submitted {formatRelativeTime(report.createdAt)} • {formatShiftTime(report.createdAt)}
						</p>
					</div>
					<div class="ml-auto flex gap-2">
						<Badge class="{getSeverityFullColor(report.severity)} border text-sm px-3 py-1">
							{getSeverityLabel(report.severity)}
						</Badge>
						{#if report.archivedAt}
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
							<div
								class="bg-muted/30 rounded-lg p-4 border-l-4 {report.severity === 2
									? 'border-l-red-500'
									: report.severity === 1
										? 'border-l-orange-500'
										: 'border-l-blue-500'}"
							>
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
												<div
													class="p-2 rounded-full bg-background border-2 {event.color ===
													'text-green-600'
														? 'border-green-200'
														: event.color === 'text-red-600'
															? 'border-red-200'
															: event.color === 'text-orange-600'
																? 'border-orange-200'
																: event.color === 'text-blue-600'
																	? 'border-blue-200'
																	: 'border-gray-200'}"
												>
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
														{format(event.time, 'HH:mm')}
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
											<p class="text-sm">{formatRelativeTime(gpsData.timestamp)}</p>
										</div>
									</div>
									<div class="h-48">
										<ReportMap
											_latitude={gpsData.latitude}
											_longitude={gpsData.longitude}
											_accuracy={gpsData.accuracy}
											_severity={report.severity}
											_className="h-full"
										/>
									</div>
								</div>
							</Card.Content>
						</Card.Root>
					{/if}
				</div>

				<!-- Sidebar -->
				<div class="space-y-6">
					<!-- Incident Date Card (simplified) -->
					<Card.Root class="p-6">
						<Card.Header class="px-0 pt-0">
							<Card.Title class="flex items-center gap-2">
								<CalendarIcon class="h-5 w-5" />
								Incident Date
							</Card.Title>
						</Card.Header>
						<Card.Content class="px-0 pb-0">
							<div class="space-y-2">
								<p class="text-2xl font-semibold">
									{format(new Date(report.createdAt), 'dd MMM yyyy')}
								</p>
								<p class="text-sm text-muted-foreground">
									{format(new Date(report.createdAt), 'EEEE, HH:mm:ss')} UTC
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
										<p class="text-sm font-medium">{report.scheduleName}</p>
									</div>
									<Separator />
									<div>
										<span class="text-sm font-medium text-muted-foreground">Shift Time</span>
										<p class="text-sm">
											{report.shiftStart ? formatShiftTime(report.shiftStart) : 'N/A'} - {report.shiftEnd
												? format(new Date(report.shiftEnd), 'HH:mm')
												: 'N/A'}
										</p>
									</div>
									<div>
										<span class="text-sm font-medium text-muted-foreground">Duration</span>
										<p class="text-sm">
											{report.shiftStart && report.shiftEnd
												? calculateDuration(new Date(report.shiftStart), new Date(report.shiftEnd))
												: 'N/A'}
										</p>
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
								<div
									class="p-3 bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg"
								>
									<div class="flex items-center gap-2">
										<InfoIcon class="h-4 w-4 text-blue-600 dark:text-blue-400" />
										<div>
											<p class="text-sm font-medium text-blue-900 dark:text-blue-100">
												Off-Shift Report
											</p>
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
										<p class="font-medium">{report.userName || 'Unknown'}</p>
										<p class="text-sm text-muted-foreground">Volunteer</p>
									</div>
								</div>
								<Separator />
								<div class="flex items-center gap-2">
									<PhoneIcon class="h-4 w-4 text-muted-foreground" />
									<span class="text-sm">{report.userPhone}</span>
								</div>
								<div class="flex gap-2">
									<Button variant="outline" size="sm" class="flex-1">
										<PhoneIcon class="h-4 w-4 mr-2" />
										Call
									</Button>
									<Button variant="outline" size="sm" class="flex-1">View Profile</Button>
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
								{#if gpsData}
									<Button
										variant="outline"
										class="w-full justify-start"
										onclick={() => {
											const mapUrl = `https://www.openstreetmap.org/?mlat=${gpsData.latitude}&mlon=${gpsData.longitude}&zoom=18`;
											window.open(mapUrl, '_blank');
										}}
									>
										<MapPinIcon class="h-4 w-4 mr-2" />
										View on Map
									</Button>
								{/if}
								<Button variant="outline" class="w-full justify-start">
									<UserIcon class="h-4 w-4 mr-2" />
									Contact Reporter
								</Button>

								{#if report.archivedAt}
									<Button
										variant="outline"
										class="w-full justify-start text-green-600 hover:text-green-700"
										onclick={handleUnarchive}
									>
										<ArchiveRestoreIcon class="h-4 w-4 mr-2" />
										Unarchive Report
									</Button>
									<Button
										variant="outline"
										class="w-full justify-start text-destructive hover:text-destructive"
										onclick={handleDelete}
									>
										<TrashIcon class="h-4 w-4 mr-2" />
										Delete Report
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
