<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import { formatRelativeTime } from '$lib/utils/reports';
	import {
		getMapTheme,
		getMarkerColors,
		getMapThemeCSS,
		type MapThemeKey
	} from '$lib/config/mapThemes';

	// Use our domain Report type
	import type { Report } from '$lib/types/domain';

	interface Props {
		reports: Report[];
		className?: string;
		onReportClick?: (reportId: number) => void;
		theme?: MapThemeKey;
	}

	let { reports, className = '', onReportClick, theme = 'nightOwls' }: Props = $props();

	// Get the map theme configuration
	const mapTheme = $derived(getMapTheme(theme));
	const markerColors = $derived(getMarkerColors(theme));
	const themeCSS = $derived(getMapThemeCSS(theme));

	// Filter reports that have REAL GPS coordinates from the database
	const reportsWithLocation = $derived.by(() => {
		return reports.filter(
			(report) =>
				report.latitude !== undefined &&
				report.longitude !== undefined &&
				report.latitude !== null &&
				report.longitude !== null &&
				!isNaN(report.latitude) &&
				!isNaN(report.longitude)
		);
	});

	// Calculate map bounds to fit all real report locations
	const mapBounds = $derived.by(() => {
		if (reportsWithLocation.length === 0) {
			// Default to Cape Town area when no GPS data is available
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
				return markerColors.normal;
			case 1:
				return markerColors.suspicion;
			case 2:
				return markerColors.incident;
			default:
				return markerColors.default;
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

<div
	class="map-container {className}"
	style={Object.entries(themeCSS)
		.map(([key, value]) => `${key}: ${value}`)
		.join('; ')}
>
	<MapLibre
		style={mapTheme.style}
		center={mapBounds.center}
		zoom={mapBounds.zoom}
		standardControls
		class="w-full h-full night-owls-map"
	>
		{#each reportsWithLocation as report (report.id)}
			{@const SeverityIcon = getSeverityIcon(report.severity)}
			<Marker lngLat={[report.longitude!, report.latitude!]}>
				<div class="marker-container">
					<button
						class="marker-pin"
						style="background-color: {getSeverityColor(report.severity)}"
						onclick={() => handleMarkerClick(report.id)}
						title="Report #{report.id} - {getSeverityLabel(report.severity)} - {report.userName ||
							'Unknown'} - {formatRelativeTime(report.createdAt)}"
					>
						<SeverityIcon class="h-3 w-3 text-white" />
					</button>
					{#if report.gpsAccuracy && report.gpsAccuracy > 0}
						<div
							class="accuracy-circle"
							style="width: {Math.max(
								20,
								Math.min(100, report.gpsAccuracy / 5)
							)}px; height: {Math.max(20, Math.min(100, report.gpsAccuracy / 5))}px;"
						></div>
					{/if}
				</div>
			</Marker>
		{/each}
	</MapLibre>

	{#if reportsWithLocation.length === 0}
		<div class="absolute inset-0 flex items-center justify-center bg-black/50 backdrop-blur-sm">
			<div class="text-center p-8">
				<MapPinIcon class="h-12 w-12 text-gray-400 mx-auto mb-4 opacity-50" />
				<h3 class="text-lg font-medium mb-2 text-white">No reports with location data</h3>
				<p class="text-sm text-gray-300 max-w-md">
					{reports.length === 0
						? 'No reports found matching current filters'
						: `${reports.length} reports found, but none contain GPS location data`}
				</p>
				<p class="text-xs text-gray-400 mt-2">
					Reports with GPS coordinates will appear here on the map
				</p>
			</div>
		</div>
	{/if}
</div>

<style>
	.map-container {
		position: relative;
		border-radius: 0.5rem;
		overflow: hidden;
		border: 1px solid hsl(var(--border));
		background: #1a1a1a;
	}

	.marker-container {
		position: relative;
		transform: translate(-50%, -100%);
	}

	.marker-pin {
		width: 24px;
		height: 24px;
		border-radius: 50% 50% 50% 0;
		transform: rotate(-45deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
		border: 2px solid rgba(255, 255, 255, 0.9);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.marker-pin:hover {
		transform: rotate(-45deg) scale(1.15);
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.6);
	}

	.marker-pin :global(svg) {
		transform: rotate(45deg);
		filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.3));
	}

	.accuracy-circle {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		border: 2px solid rgba(96, 165, 250, 0.4);
		border-radius: 50%;
		background-color: rgba(96, 165, 250, 0.1);
		z-index: 1;
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
		100% {
			opacity: 1;
		}
	}
</style>
