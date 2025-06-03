<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';

	let {
		isLoading = false,
		_latitude,
		_longitude,
		_accuracy,
		_severity,
		_className = ''
	}: {
		isLoading?: boolean;
		_latitude?: number;
		_longitude?: number;
		_accuracy?: number;
		_severity?: number;
		_className?: string;
	} = $props();

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

	// Determine zoom level based on accuracy
	const zoomLevel = $derived.by(() => {
		if (!_accuracy) return 15;
		if (_accuracy < 10) return 17;
		if (_accuracy < 50) return 16;
		if (_accuracy < 100) return 15;
		return 14;
	});
</script>

<div class="map-container {_className}">
	{#if isLoading}
		<div class="h-full w-full flex items-center justify-center bg-muted">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
		</div>
	{:else if _latitude && _longitude}
		<MapLibre
			style={mapStyle}
			center={[_longitude, _latitude]}
			zoom={zoomLevel}
			standardControls
			class="w-full h-full"
		>
			{@const SeverityIcon = getSeverityIcon(_severity ?? 0)}
			<Marker lngLat={[_longitude, _latitude]}>
				<div class="marker-container">
					<div
						class="marker-pin"
						style="background-color: {getSeverityColor(_severity ?? 0)}"
						title="Report location - accuracy: ±{_accuracy ?? 0}m"
					>
						<SeverityIcon class="h-3 w-3 text-white" />
					</div>
					{#if _accuracy && _accuracy > 0}
						<div
							class="accuracy-circle"
							style="width: {Math.max(20, Math.min(100, _accuracy / 5))}px; height: {Math.max(
								20,
								Math.min(100, _accuracy / 5)
							)}px;"
						></div>
					{/if}
				</div>
			</Marker>
		</MapLibre>
	{:else}
		<div class="h-full w-full flex items-center justify-center bg-muted text-muted-foreground">
			<div class="text-center">
				<InfoIcon class="h-8 w-8 mx-auto mb-2 opacity-50" />
				<p class="text-sm">No location data available</p>
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
	}

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
		position: relative;
		z-index: 2;
	}

	.marker-pin :global(svg) {
		transform: rotate(45deg);
	}

	.accuracy-circle {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		border: 2px solid rgba(59, 130, 246, 0.5);
		border-radius: 50%;
		background-color: rgba(59, 130, 246, 0.1);
		z-index: 1;
	}
</style>
