<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { createQuery } from '@tanstack/svelte-query';
	import { UserApiService } from '$lib/services/api/user';
	import { formatRelativeTime } from '$lib/utils/reports';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';

	interface Props {
		open: boolean;
		reportId: number | null;
		onClose: () => void;
	}

	let { open = $bindable(), reportId, onClose }: Props = $props();

	// Fetch user's reports to find the selected one
	const userReportsQuery = createQuery({
		queryKey: ['user-reports'],
		queryFn: () => UserApiService.getMyReports()
	});

	// Find the selected report
	const selectedReport = $derived.by(() => {
		if (!reportId || !$userReportsQuery.data) return null;
		return $userReportsQuery.data.find((report) => report.report_id === reportId) ?? null;
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

	function formatDateTime(dateString: string) {
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

	function openInMap() {
		if (selectedReport && selectedReport.latitude && selectedReport.longitude) {
			const mapUrl = `https://www.openstreetmap.org/?mlat=${selectedReport.latitude}&mlon=${selectedReport.longitude}&zoom=18`;
			window.open(mapUrl, '_blank');
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>
				{#if selectedReport}
					Report #{selectedReport.report_id}
				{:else}
					Report Details
				{/if}
			</Dialog.Title>
			<Dialog.Description>
				{#if selectedReport}
					Submitted {formatRelativeTime(selectedReport.created_at)}
				{:else}
					Loading report details...
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4">
			{#if $userReportsQuery.isLoading}
				<div class="space-y-4">
					<Skeleton class="h-6 w-32" />
					<Skeleton class="h-20 w-full" />
					<div class="grid grid-cols-2 gap-4">
						<Skeleton class="h-16 w-full" />
						<Skeleton class="h-16 w-full" />
					</div>
				</div>
			{:else if selectedReport}
				{@const SeverityIcon = getSeverityIcon(selectedReport.severity)}

				<!-- Severity Badge -->
				<div class="flex items-center gap-3">
					<div class="p-2 rounded-lg {getSeverityColor(selectedReport.severity)}">
						<SeverityIcon class="h-5 w-5" />
					</div>
					<div>
						<Badge class="{getSeverityColor(selectedReport.severity)} border">
							{getSeverityLabel(selectedReport.severity)}
						</Badge>
					</div>
				</div>

				<!-- Report Message -->
				<Card.Root>
					<Card.Header class="pb-3">
						<Card.Title class="text-base">Report Details</Card.Title>
					</Card.Header>
					<Card.Content>
						<div
							class="bg-muted/30 rounded-lg p-4 border-l-4 {selectedReport.severity === 2
								? 'border-l-red-500'
								: selectedReport.severity === 1
									? 'border-l-orange-500'
									: 'border-l-blue-500'}"
						>
							<p class="text-sm leading-relaxed">{selectedReport.message}</p>
						</div>
					</Card.Content>
				</Card.Root>

				<!-- Report Info Grid -->
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<!-- Date & Time -->
					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Title class="flex items-center gap-2 text-sm">
								<CalendarIcon class="h-4 w-4" />
								Date & Time
							</Card.Title>
						</Card.Header>
						<Card.Content class="pt-0">
							<p class="text-sm font-medium">{formatDateTime(selectedReport.created_at)}</p>
						</Card.Content>
					</Card.Root>

					<!-- Schedule Info -->
					{#if selectedReport.schedule_name}
						<Card.Root>
							<Card.Header class="pb-2">
								<Card.Title class="flex items-center gap-2 text-sm">
									<ClockIcon class="h-4 w-4" />
									Schedule
								</Card.Title>
							</Card.Header>
							<Card.Content class="pt-0">
								<p class="text-sm font-medium">{selectedReport.schedule_name}</p>
								{#if selectedReport.shift_start}
									<p class="text-xs text-muted-foreground mt-1">
										{new Date(selectedReport.shift_start).toLocaleTimeString('en-GB', {
											hour: '2-digit',
											minute: '2-digit'
										})} - {selectedReport.shift_end
											? new Date(selectedReport.shift_end).toLocaleTimeString('en-GB', {
													hour: '2-digit',
													minute: '2-digit'
												})
											: ''}
									</p>
								{/if}
							</Card.Content>
						</Card.Root>
					{/if}
				</div>

				<!-- Location Info -->
				{#if selectedReport.latitude && selectedReport.longitude}
					<Card.Root>
						<Card.Header class="pb-2">
							<Card.Title class="flex items-center gap-2 text-sm">
								<MapPinIcon class="h-4 w-4" />
								Location Information
							</Card.Title>
						</Card.Header>
						<Card.Content class="pt-0 space-y-3">
							<div class="grid grid-cols-2 gap-4 text-sm">
								<div>
									<span class="text-muted-foreground">Coordinates</span>
									<p class="font-mono text-xs">
										{selectedReport.latitude.toFixed(6)}, {selectedReport.longitude.toFixed(6)}
									</p>
								</div>
								{#if selectedReport.gps_accuracy}
									<div>
										<span class="text-muted-foreground">Accuracy</span>
										<p class="text-xs">±{selectedReport.gps_accuracy}m</p>
									</div>
								{/if}
							</div>
							<Button variant="outline" size="sm" onclick={openInMap} class="w-full">
								<MapPinIcon class="h-4 w-4 mr-2" />
								View on Map
							</Button>
						</Card.Content>
					</Card.Root>
				{/if}
			{:else}
				<div class="text-center py-8">
					<AlertTriangleIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-sm text-muted-foreground">Report not found</p>
				</div>
			{/if}
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={onClose}>Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
