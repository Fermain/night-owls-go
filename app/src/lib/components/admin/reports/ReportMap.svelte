<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';

	interface Props {
		latitude: number;
		longitude: number;
		accuracy?: number;
		severity?: number;
		className?: string;
		interactive?: boolean;
		showAccuracyCircle?: boolean;
		zoom?: number;
	}

	let {
		latitude,
		longitude,
		accuracy = 10,
		severity = 0,
		className = '',
		interactive = true,
		showAccuracyCircle = true,
		zoom = 16
	}: Props = $props();

	// OpenStreetMap style using free tile servers
	const mapStyle = {
		version: 8 as const,
		sources: {
			'osm-tiles': {
				type: 'raster' as const,
				tiles: [
					'https://tile.openstreetmap.org/{z}/{x}/{y}.png'
				],
				tileSize: 256,
				attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
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
				return MapPinIcon;
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

	// Create accuracy circle data if needed
	const accuracyCircle = $derived.by(() => {
		if (!showAccuracyCircle) return null;
		
		// Simple circle approximation using GeoJSON
		const points = 64;
		const coords = [];
		const earthRadius = 6371000; // meters
		
		for (let i = 0; i < points; i++) {
			const angle = (i * 360) / points;
			const angleRad = (angle * Math.PI) / 180;
			
			const latRad = (latitude * Math.PI) / 180;
			const deltaLat = (accuracy * Math.cos(angleRad)) / earthRadius;
			const deltaLng = (accuracy * Math.sin(angleRad)) / (earthRadius * Math.cos(latRad));
			
			const newLat = latitude + (deltaLat * 180) / Math.PI;
			const newLng = longitude + (deltaLng * 180) / Math.PI;
			
			coords.push([newLng, newLat]);
		}
		
		// Close the circle
		coords.push(coords[0]);
		
		return {
			type: 'Feature' as const,
			properties: {},
			geometry: {
				type: 'Polygon' as const,
				coordinates: [coords]
			}
		};
	});

	// Map style with accuracy circle
	const mapStyleWithCircle = $derived.by(() => {
		if (!accuracyCircle) return mapStyle;
		
		return {
			...mapStyle,
			sources: {
				...mapStyle.sources,
				'accuracy-circle': {
					type: 'geojson' as const,
					data: accuracyCircle
				}
			},
			layers: [
				...mapStyle.layers,
				{
					id: 'accuracy-circle-fill',
					type: 'fill' as const,
					source: 'accuracy-circle',
					paint: {
						'fill-color': getSeverityColor(severity),
						'fill-opacity': 0.1
					}
				},
				{
					id: 'accuracy-circle-stroke',
					type: 'line' as const,
					source: 'accuracy-circle',
					paint: {
						'line-color': getSeverityColor(severity),
						'line-width': 2,
						'line-opacity': 0.5
					}
				}
			]
		};
	});
</script>

<div class="map-container {className}">
	<MapLibre
		style={mapStyleWithCircle}
		center={[longitude, latitude]}
		{zoom}
		standardControls={interactive}
		class="w-full h-full"
	>
		<Marker lngLat={[longitude, latitude]}>
			<div class="marker-container">
				{#snippet children()}
					{@const SeverityIcon = getSeverityIcon(severity)}
					<div 
						class="marker-pin"
						style="background-color: {getSeverityColor(severity)}"
					>
						<SeverityIcon class="h-4 w-4 text-white" />
					</div>
				{/snippet}
			</div>
		</Marker>
	</MapLibre>
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
		width: 32px;
		height: 32px;
		border-radius: 50% 50% 50% 0;
		transform: rotate(-45deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
		border: 2px solid white;
	}

	.marker-pin :global(svg) {
		transform: rotate(45deg);
	}
</style> 