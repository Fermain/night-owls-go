<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import {
		getMapTheme,
		getMarkerColors,
		getMapThemeCSS,
		type MapThemeKey
	} from '$lib/config/mapThemes';

	let {
		isLoading = false,
		_latitude,
		_longitude,
		_accuracy,
		_severity,
		_className = '',
		theme = 'nightOwls'
	}: {
		isLoading?: boolean;
		_latitude?: number;
		_longitude?: number;
		_accuracy?: number;
		_severity?: number;
		_className?: string;
		theme?: MapThemeKey;
	} = $props();

	// Get the map theme configuration
	const mapTheme = $derived(getMapTheme(theme));
	const markerColors = $derived(getMarkerColors(theme));
	const themeCSS = $derived(getMapThemeCSS(theme));

	// Check if we have valid GPS coordinates
	const hasValidCoordinates = $derived(
		_latitude !== undefined && _longitude !== undefined && !isNaN(_latitude) && !isNaN(_longitude)
	);

	function getSeverityIcon(severity?: number) {
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

	function getSeverityColor(severity?: number) {
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

	// Calculate appropriate zoom level based on accuracy
	const mapZoom = $derived.by(() => {
		if (!_accuracy) return 15;

		// Zoom in more for higher accuracy (lower accuracy values)
		if (_accuracy < 5) return 18;
		else if (_accuracy < 10) return 17;
		else if (_accuracy < 20) return 16;
		else if (_accuracy < 50) return 15;
		else return 14;
	});
</script>

{#if isLoading}
	<div class="map-container loading {_className}">
		<div class="flex items-center justify-center h-full">
			<div class="text-center">
				<div
					class="animate-spin h-8 w-8 border-2 border-blue-500 border-t-transparent rounded-full mx-auto mb-2"
				></div>
				<p class="text-sm text-gray-400">Loading map...</p>
			</div>
		</div>
	</div>
{:else if !hasValidCoordinates}
	<div class="map-container no-location {_className}">
		<div class="flex items-center justify-center h-full bg-gray-800/50">
			<div class="text-center p-6">
				<InfoIcon class="h-12 w-12 text-gray-400 mx-auto mb-4 opacity-50" />
				<h3 class="text-lg font-medium mb-2 text-white">No GPS Data</h3>
				<p class="text-sm text-gray-300">This report was submitted without location information</p>
			</div>
		</div>
	</div>
{:else}
	<div
		class="map-container {_className}"
		style={Object.entries(themeCSS)
			.map(([key, value]) => `${key}: ${value}`)
			.join('; ')}
	>
		<MapLibre
			style={mapTheme.style}
			center={[_longitude!, _latitude!]}
			zoom={mapZoom}
			standardControls
			class="w-full h-full night-owls-map"
		>
			<!-- Main report marker -->
			<Marker lngLat={[_longitude!, _latitude!]}>
				{@const SeverityIcon = getSeverityIcon(_severity)}
				<div class="marker-container">
					<div
						class="marker-pin"
						style="background-color: {getSeverityColor(_severity)}"
						title="Report location - GPS accuracy: ±{_accuracy || 'unknown'}m"
					>
						<SeverityIcon class="h-4 w-4 text-white" />
					</div>

					<!-- Accuracy circle -->
					{#if _accuracy && _accuracy > 0}
						<div
							class="accuracy-circle"
							style="width: {Math.max(30, Math.min(120, _accuracy / 2))}px; height: {Math.max(
								30,
								Math.min(120, _accuracy / 2)
							)}px;"
							title="GPS accuracy radius: ±{_accuracy}m"
						></div>
					{/if}
				</div>
			</Marker>
		</MapLibre>

		<!-- GPS accuracy info overlay -->
		{#if _accuracy}
			<div class="absolute bottom-4 right-4 map-overlay">
				<div class="text-xs">
					<div class="font-medium">GPS Accuracy</div>
					<div class="text-gray-300">±{_accuracy}m</div>
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
		min-height: 200px;
	}

	.map-container.loading,
	.map-container.no-location {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.marker-container {
		position: relative;
		transform: translate(-50%, -100%);
	}

	.marker-pin {
		width: 28px;
		height: 28px;
		border-radius: 50% 50% 50% 0;
		transform: rotate(-45deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
		border: 3px solid rgba(255, 255, 255, 0.9);
		transition: all 0.3s ease;
		z-index: 10;
	}

	.marker-pin:hover {
		transform: rotate(-45deg) scale(1.1);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.7);
	}

	.marker-pin :global(svg) {
		transform: rotate(45deg);
		filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.4));
	}

	.accuracy-circle {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		border: 2px solid rgba(96, 165, 250, 0.5);
		border-radius: 50%;
		background-color: rgba(96, 165, 250, 0.15);
		z-index: 1;
		animation: pulse 4s infinite;
	}

	.map-overlay {
		background: var(--map-control-bg);
		color: var(--map-control-color);
		border: var(--map-control-border);
		border-radius: var(--map-control-radius);
		backdrop-filter: var(--map-control-backdrop);
		box-shadow: var(--map-control-shadow);
		padding: 0.5rem 0.75rem;
	}

	@keyframes pulse {
		0% {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
		50% {
			opacity: 0.6;
			transform: translate(-50%, -50%) scale(1.05);
		}
		100% {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
	}
</style>
