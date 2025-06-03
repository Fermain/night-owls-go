<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import * as Card from '$lib/components/ui/card';
	import { createQuery } from '@tanstack/svelte-query';
	import { UserApiService } from '$lib/services/api/user';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import { formatRelativeTime } from '$lib/utils/reports';

	interface Props {
		className?: string;
		onReportClick?: (reportId: number) => void;
	}

	let { className = '', onReportClick }: Props = $props();

	// Fetch user's reports
	const userReportsQuery = createQuery({
		queryKey: ['user-reports'],
		queryFn: () => UserApiService.getMyReports()
	});

	// OpenStreetMap style
	const mapStyle = {
		version: 8 as const,
		sources: {
			'osm-tiles': {
				type: 'raster' as const,
				tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
				tileSize: 256,
				attribution:
					'© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
			}
		},
		layers: [
			{
				id: 'osm-tiles',
				type: 'raster' as const,
				source: 'osm-tiles',
				minzoom: 0,
				maxzoom: 19
			}
		]
	};

	// Filter reports that have GPS coordinates
	const reportsWithLocation = $derived.by(() => {
		const reports = $userReportsQuery.data ?? [];
		return reports.filter((report) => report.latitude && report.longitude);
	});

	// Calculate map bounds to fit all reports
	const mapBounds = $derived.by(() => {
		if (reportsWithLocation.length === 0) {
			// Default to Cape Town area
			return { center: [18.4241, -33.9249] as [number, number], zoom: 12 };
		}

		const lats = reportsWithLocation.map((r) => r.latitude!);
		const lngs = reportsWithLocation.map((r) => r.longitude!);

		const minLat = Math.min(...lats);
		const maxLat = Math.max(...lats);
		const minLng = Math.min(...lngs);
		const maxLng = Math.max(...lngs);

		const centerLat = (minLat + maxLat) / 2;
		const centerLng = (minLng + maxLng) / 2;

		// Calculate zoom level based on bounds
		const latDiff = maxLat - minLat;
		const lngDiff = maxLng - minLng;
		const maxDiff = Math.max(latDiff, lngDiff);

		let zoom = 12;
		if (maxDiff < 0.01) zoom = 15;
		else if (maxDiff < 0.05) zoom = 13;
		else if (maxDiff < 0.1) zoom = 11;
		else zoom = 10;

		return { center: [centerLng, centerLat] as [number, number], zoom };
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
				return '#3b82f6'; // blue
			case 1:
				return '#f59e0b'; // orange
			case 2:
				return '#ef4444'; // red
			default:
				return '#6b7280'; // gray
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

	function handleMarkerClick(reportId: number) {
		if (onReportClick) {
			onReportClick(reportId);
		}
	}
</script>

<Card.Root class="overflow-hidden {className}">
	<Card.Header class="pb-3">
		<Card.Title class="flex items-center gap-2 text-lg">
			<MapPinIcon class="h-5 w-5" />
			My Report Locations
		</Card.Title>
		<Card.Description>
			{#if $userReportsQuery.isLoading}
				Loading your reports...
			{:else if reportsWithLocation.length === 0}
				No reports with location data found
			{:else}
				{reportsWithLocation.length} of {$userReportsQuery.data?.length ?? 0} reports have location data
			{/if}
		</Card.Description>
	</Card.Header>
	<Card.Content class="p-0">
		{#if $userReportsQuery.isLoading}
			<div class="h-64 p-4">
				<Skeleton class="h-full w-full" />
			</div>
		{:else if $userReportsQuery.isError}
			<div class="h-64 flex items-center justify-center p-4">
				<div class="text-center">
					<AlertTriangleIcon class="h-8 w-8 text-destructive mx-auto mb-2" />
					<p class="text-sm text-muted-foreground">Failed to load reports</p>
				</div>
			</div>
		{:else}
			<div class="h-64 relative">
				<MapLibre
					style={mapStyle}
					center={mapBounds.center}
					zoom={mapBounds.zoom}
					standardControls
					class="w-full h-full"
				>
					{#each reportsWithLocation as report (report.report_id)}
						{@const SeverityIcon = getSeverityIcon(report.severity)}
						<Marker lngLat={[report.longitude!, report.latitude!]}>
							<div class="marker-container">
								<button
									class="marker-pin"
									style="background-color: {getSeverityColor(report.severity)}"
									onclick={() => handleMarkerClick(report.report_id)}
									title="Report #{report.report_id} - {getSeverityLabel(
										report.severity
									)} - {formatRelativeTime(report.created_at)}"
								>
									<SeverityIcon class="h-3 w-3 text-white" />
								</button>
							</div>
						</Marker>
					{/each}
				</MapLibre>

				{#if reportsWithLocation.length === 0 && !$userReportsQuery.isLoading}
					<div class="absolute inset-0 flex items-center justify-center bg-muted/50">
						<div class="text-center">
							<MapPinIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
							<p class="text-sm text-muted-foreground">No reports with location data</p>
							<p class="text-xs text-muted-foreground mt-1">
								Future reports with GPS will appear here
							</p>
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</Card.Content>

	{#if reportsWithLocation.length > 0}
		<Card.Footer class="pt-3">
			<div class="flex items-center justify-between w-full text-sm text-muted-foreground">
				<div class="flex items-center gap-2">
					<ClockIcon class="h-4 w-4" />
					<span>Click markers to view details</span>
				</div>
				<div class="flex items-center gap-4">
					<div class="flex items-center gap-1">
						<div class="w-2 h-2 rounded-full bg-blue-500"></div>
						<span class="text-xs">Normal</span>
					</div>
					<div class="flex items-center gap-1">
						<div class="w-2 h-2 rounded-full bg-orange-500"></div>
						<span class="text-xs">Suspicion</span>
					</div>
					<div class="flex items-center gap-1">
						<div class="w-2 h-2 rounded-full bg-red-500"></div>
						<span class="text-xs">Incident</span>
					</div>
				</div>
			</div>
		</Card.Footer>
	{/if}
</Card.Root>

<style>
	:global(.maplibregl-map) {
		font-family: inherit;
	}

	:global(.maplibregl-ctrl-attrib) {
		font-size: 10px;
		background-color: rgba(255, 255, 255, 0.8);
	}

	.marker-container {
		position: relative;
		transform: translate(-50%, -100%);
	}

	.marker-pin {
		width: 20px;
		height: 20px;
		border-radius: 50% 50% 50% 0;
		transform: rotate(-45deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
		border: 2px solid white;
		cursor: pointer;
		transition: transform 0.2s ease;
	}

	.marker-pin:hover {
		transform: rotate(-45deg) scale(1.1);
	}

	.marker-pin :global(svg) {
		transform: rotate(45deg);
	}
</style>
