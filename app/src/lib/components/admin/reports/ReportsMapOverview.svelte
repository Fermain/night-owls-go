<script lang="ts">
	import { MapLibre, Marker } from 'svelte-maplibre';
	import 'maplibre-gl/dist/maplibre-gl.css';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';

	interface Report {
		report_id: number;
		severity: number;
		latitude?: number;
		longitude?: number;
		message: string;
		user_name: string;
		created_at: string;
	}

	interface Props {
		reports: Report[];
		className?: string;
		onReportClick?: (reportId: number) => void;
	}

	let { reports, className = '', onReportClick }: Props = $props();

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

	// Filter reports that have GPS coordinates (for demo, we'll add mock coordinates)
	const reportsWithLocation = $derived.by(() => {
		return reports.map((report, _index) => ({
			...report,
			// Mock GPS coordinates for demo - in real app these would come from the database
			latitude: -33.9249 + (Math.random() - 0.5) * 0.1, // Cape Town area with some spread
			longitude: 18.4241 + (Math.random() - 0.5) * 0.1
		}));
	});

	// Calculate map bounds to fit all reports
	const mapBounds = $derived.by(() => {
		if (reportsWithLocation.length === 0) {
			return { center: [18.4241, -33.9249] as [number, number], zoom: 12 }; // Default to Cape Town
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

	function handleMarkerClick(reportId: number) {
		if (onReportClick) {
			onReportClick(reportId);
		}
	}
</script>

<div class="map-container {className}">
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
						title="Report #{report.report_id} - {report.message.slice(0, 50)}..."
					>
						<SeverityIcon class="h-3 w-3 text-white" />
					</button>
				</div>
			</Marker>
		{/each}
	</MapLibre>

	{#if reportsWithLocation.length === 0}
		<div class="absolute inset-0 flex items-center justify-center bg-muted/50">
			<div class="text-center">
				<InfoIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
				<p class="text-sm text-muted-foreground">No reports with location data</p>
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
		width: 24px;
		height: 24px;
		border-radius: 50% 50% 50% 0;
		transform: rotate(-45deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
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
