<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';

	// Use our domain Report type
	import type { Report } from '$lib/types/domain';

	interface Props {
		reports: Report[];
		className?: string;
	}

	let { reports, className = '' }: Props = $props();

	// Calculate statistics
	const stats = $derived.by(() => {
		const total = reports.length;
		const withGPS = reports.filter((r) => r.latitude && r.longitude).length;
		const withoutGPS = total - withGPS;
		const gpsPercentage = total > 0 ? Math.round((withGPS / total) * 100) : 0;

		// Severity breakdown for reports with GPS
		const reportsWithGPS = reports.filter((r) => r.latitude && r.longitude);
		const severityBreakdown = {
			normal: reportsWithGPS.filter((r) => r.severity === 0).length,
			suspicion: reportsWithGPS.filter((r) => r.severity === 1).length,
			incident: reportsWithGPS.filter((r) => r.severity === 2).length
		};

		// Average GPS accuracy
		const accuracyValues = reportsWithGPS
			.filter((r) => r.gpsAccuracy && r.gpsAccuracy > 0)
			.map((r) => r.gpsAccuracy!);
		const avgAccuracy =
			accuracyValues.length > 0
				? Math.round(accuracyValues.reduce((sum, acc) => sum + acc, 0) / accuracyValues.length)
				: null;

		return {
			total,
			withGPS,
			withoutGPS,
			gpsPercentage,
			severityBreakdown,
			avgAccuracy
		};
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
</script>

<Card.Root class="overflow-hidden {className}">
	<Card.Header class="pb-3">
		<Card.Title class="flex items-center gap-2 text-base">
			<MapPinIcon class="h-4 w-4" />
			GPS Data Coverage
		</Card.Title>
		<Card.Description>Location data statistics for current report set</Card.Description>
	</Card.Header>
	<Card.Content class="p-4 pt-0">
		<div class="space-y-4">
			<!-- Overall GPS Statistics -->
			<div class="grid grid-cols-3 gap-3">
				<div class="text-center p-3 bg-muted/30 rounded-lg">
					<div class="text-lg font-bold">{stats.total}</div>
					<div class="text-xs text-muted-foreground">Total Reports</div>
				</div>
				<div class="text-center p-3 bg-green-50 border border-green-200 rounded-lg">
					<div class="text-lg font-bold text-green-700">{stats.withGPS}</div>
					<div class="text-xs text-green-600">With GPS</div>
				</div>
				<div class="text-center p-3 bg-gray-50 border border-gray-200 rounded-lg">
					<div class="text-lg font-bold text-gray-700">{stats.withoutGPS}</div>
					<div class="text-xs text-gray-600">No GPS</div>
				</div>
			</div>

			<!-- GPS Coverage Percentage -->
			<div class="space-y-2">
				<div class="flex items-center justify-between text-sm">
					<span class="text-muted-foreground">GPS Coverage</span>
					<span class="font-medium">{stats.gpsPercentage}%</span>
				</div>
				<div class="w-full bg-muted rounded-full h-2">
					<div
						class="bg-green-500 h-2 rounded-full transition-all duration-500"
						style="width: {stats.gpsPercentage}%"
					></div>
				</div>
			</div>

			{#if stats.withGPS > 0}
				<!-- Severity Breakdown for GPS Reports -->
				<div class="space-y-2">
					<div class="text-sm font-medium text-muted-foreground">GPS Reports by Severity</div>
					<div class="flex gap-2 flex-wrap">
						{#if stats.severityBreakdown.normal > 0}
							{@const NormalIcon = getSeverityIcon(0)}
							<Badge class="{getSeverityColor(0)} border text-xs">
								<NormalIcon class="h-3 w-3 mr-1" />
								{stats.severityBreakdown.normal} Normal
							</Badge>
						{/if}
						{#if stats.severityBreakdown.suspicion > 0}
							{@const SuspicionIcon = getSeverityIcon(1)}
							<Badge class="{getSeverityColor(1)} border text-xs">
								<SuspicionIcon class="h-3 w-3 mr-1" />
								{stats.severityBreakdown.suspicion} Suspicion
							</Badge>
						{/if}
						{#if stats.severityBreakdown.incident > 0}
							{@const IncidentIcon = getSeverityIcon(2)}
							<Badge class="{getSeverityColor(2)} border text-xs">
								<IncidentIcon class="h-3 w-3 mr-1" />
								{stats.severityBreakdown.incident} Incident
							</Badge>
						{/if}
					</div>
				</div>

				<!-- GPS Accuracy Info -->
				{#if stats.avgAccuracy}
					<div class="text-center p-2 bg-blue-50 border border-blue-200 rounded-lg">
						<div class="text-sm text-blue-700">
							<span class="font-medium">Avg GPS Accuracy:</span> ±{stats.avgAccuracy}m
						</div>
					</div>
				{/if}
			{:else if stats.total > 0}
				<!-- No GPS Data Message -->
				<div class="text-center p-4 bg-orange-50 border border-orange-200 rounded-lg">
					<FileTextIcon class="h-8 w-8 text-orange-400 mx-auto mb-2" />
					<div class="text-sm text-orange-700 font-medium">No GPS Data Available</div>
					<div class="text-xs text-orange-600 mt-1">Future reports may include location data</div>
				</div>
			{/if}
		</div>
	</Card.Content>
</Card.Root>
