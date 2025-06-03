<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import { createQuery } from '@tanstack/svelte-query';
	import { UserApiService, type UserReport } from '$lib/services/api/user';
	import { Skeleton } from '$lib/components/ui/skeleton';
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

	interface Props {
		className?: string;
		onReportClick?: (report: UserReport) => void;
		theme?: MapThemeKey;
	}

	let { className = '', onReportClick, theme = 'nightOwls' }: Props = $props();

	// Get the map theme configuration
	const mapTheme = $derived(getMapTheme(theme));
	const markerColors = $derived(getMarkerColors(theme));
	const themeCSS = $derived(getMapThemeCSS(theme));

	// Fetch user's reports
	const reportsQuery = $derived(
		createQuery({
			queryKey: ['userReports'],
			queryFn: UserApiService.getMyReports
		})
	);

	// Filter reports with GPS data
	const reportsWithLocation = $derived.by(() => {
		if (!$reportsQuery.data) return [];
		return $reportsQuery.data.filter(
			(report) =>
				report.latitude !== undefined &&
				report.longitude !== undefined &&
				report.latitude !== null &&
				report.longitude !== null &&
				!isNaN(report.latitude) &&
				!isNaN(report.longitude)
		);
	});

	// Calculate map bounds
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

		let zoom = 15; // Start closer for user reports
		if (maxDiff < 0.005) zoom = 17;
		else if (maxDiff < 0.01) zoom = 16;
		else if (maxDiff < 0.05) zoom = 14;
		else if (maxDiff < 0.1) zoom = 13;
		else zoom = 12;

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

	function handleMarkerClick(report: UserReport) {
		if (onReportClick) {
			onReportClick(report);
		}
	}
</script>

{#if $reportsQuery.isLoading}
	<div class="map-container {className}">
		<div class="flex items-center justify-center h-64">
			<div class="text-center">
				<Skeleton class="h-8 w-8 rounded-full mx-auto mb-2" />
				<Skeleton class="h-4 w-32" />
			</div>
		</div>
	</div>
{:else if $reportsQuery.isError}
	<div class="map-container {className}">
		<div class="flex items-center justify-center h-64 bg-red-50 dark:bg-red-950/20">
			<div class="text-center">
				<AlertTriangleIcon class="h-8 w-8 text-red-500 mx-auto mb-2" />
				<p class="text-sm text-red-600 dark:text-red-400">Failed to load your reports</p>
			</div>
		</div>
	</div>
{:else}
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
			{#each reportsWithLocation as report (report.report_id)}
				{@const SeverityIcon = getSeverityIcon(report.severity)}
				<Marker lngLat={[report.longitude!, report.latitude!]}>
					<div class="marker-container">
						<button
							class="marker-pin"
							style="background-color: {getSeverityColor(report.severity)}"
							onclick={() => handleMarkerClick(report)}
							title="Report #{report.report_id} - {getSeverityLabel(
								report.severity
							)} - {formatRelativeTime(report.created_at)}"
						>
							<SeverityIcon class="h-3 w-3 text-white" />
						</button>
						{#if report.gps_accuracy && report.gps_accuracy > 0}
							<div
								class="accuracy-circle"
								style="width: {Math.max(
									20,
									Math.min(80, report.gps_accuracy / 3)
								)}px; height: {Math.max(20, Math.min(80, report.gps_accuracy / 3))}px;"
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
					<h3 class="text-lg font-medium mb-2 text-white">No location data available</h3>
					<p class="text-sm text-gray-300 max-w-md">
						{$reportsQuery.data && $reportsQuery.data.length > 0
							? `You have ${$reportsQuery.data.length} reports, but none contain GPS location data`
							: "You haven't submitted any reports yet"}
					</p>
					<p class="text-xs text-gray-400 mt-2">
						Submit reports from your mobile device to see them on the map
					</p>
				</div>
			</div>
		{/if}

		<!-- Report count overlay -->
		{#if reportsWithLocation.length > 0}
			<div class="absolute top-4 left-4 map-overlay">
				<div class="flex items-center gap-2 text-sm">
					<MapPinIcon class="h-4 w-4" />
					<span
						>{reportsWithLocation.length} location{reportsWithLocation.length !== 1
							? 's'
							: ''}</span
					>
				</div>
			</div>
		{/if}
	</div>
{/if}

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
		width: 20px;
		height: 20px;
		border-radius: 50% 50% 50% 0;
		transform: rotate(-45deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 3px 10px rgba(0, 0, 0, 0.4);
		border: 2px solid rgba(255, 255, 255, 0.9);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.marker-pin:hover {
		transform: rotate(-45deg) scale(1.2);
		box-shadow: 0 5px 15px rgba(0, 0, 0, 0.6);
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
		animation: pulse 3s infinite;
	}

	.map-overlay {
		background: var(--map-control-bg);
		color: var(--map-control-color);
		border: var(--map-control-border);
		border-radius: var(--map-control-radius);
		backdrop-filter: var(--map-control-backdrop);
		box-shadow: var(--map-control-shadow);
		padding: 0.5rem 0.75rem;
		font-weight: 500;
	}

	@keyframes pulse {
		0% {
			opacity: 1;
		}
		50% {
			opacity: 0.4;
		}
		100% {
			opacity: 1;
		}
	}
</style>
