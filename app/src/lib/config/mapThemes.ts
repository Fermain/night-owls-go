/**
 * Centralized map themes configuration for Night Owls
 * Provides consistent styling across all map components
 */

// MapLibre GL JS style types
interface RasterSource {
	type: 'raster';
	tiles: string[];
	tileSize: number;
	attribution: string;
}

interface RasterLayer {
	id: string;
	type: 'raster';
	source: string;
	minzoom: number;
	maxzoom: number;
}

export interface MapTheme {
	name: string;
	description: string;
	style: {
		version: 8;
		sources: Record<string, RasterSource>;
		layers: RasterLayer[];
		glyphs?: string;
		sprite?: string;
	};
	attribution: string;
}

// Night Owls Dark Theme - Primary theme for the application
export const nightOwlsTheme: MapTheme = {
	name: 'Night Owls Dark',
	description: 'Dark theme optimized for night security operations with high contrast roads',
	style: {
		version: 8,
		sources: {
			'carto-dark-voyager': {
				type: 'raster',
				tiles: [
					'https://a.basemaps.cartocdn.com/rastertiles/voyager_nolabels/{z}/{x}/{y}.png',
					'https://b.basemaps.cartocdn.com/rastertiles/voyager_nolabels/{z}/{x}/{y}.png',
					'https://c.basemaps.cartocdn.com/rastertiles/voyager_nolabels/{z}/{x}/{y}.png',
					'https://d.basemaps.cartocdn.com/rastertiles/voyager_nolabels/{z}/{x}/{y}.png'
				],
				tileSize: 256,
				attribution:
					'© <a href="https://carto.com/">CARTO</a> © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
			},
			'carto-dark-labels': {
				type: 'raster',
				tiles: [
					'https://a.basemaps.cartocdn.com/rastertiles/voyager_only_labels/{z}/{x}/{y}.png',
					'https://b.basemaps.cartocdn.com/rastertiles/voyager_only_labels/{z}/{x}/{y}.png',
					'https://c.basemaps.cartocdn.com/rastertiles/voyager_only_labels/{z}/{x}/{y}.png',
					'https://d.basemaps.cartocdn.com/rastertiles/voyager_only_labels/{z}/{x}/{y}.png'
				],
				tileSize: 256,
				attribution: ''
			}
		},
		layers: [
			{
				id: 'carto-voyager-base',
				type: 'raster',
				source: 'carto-dark-voyager',
				minzoom: 0,
				maxzoom: 20
			},
			{
				id: 'carto-voyager-labels',
				type: 'raster',
				source: 'carto-dark-labels',
				minzoom: 0,
				maxzoom: 20
			}
		]
	},
	attribution:
		'© <a href="https://carto.com/">CARTO</a> © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
};

// Light theme for daytime operations or user preference
export const dayOwlsTheme: MapTheme = {
	name: 'Day Owls Light',
	description: 'Light theme for daytime security operations',
	style: {
		version: 8,
		sources: {
			'carto-light': {
				type: 'raster',
				tiles: [
					'https://a.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png',
					'https://b.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png',
					'https://c.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png',
					'https://d.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png'
				],
				tileSize: 256,
				attribution:
					'© <a href="https://carto.com/">CARTO</a> © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
			}
		},
		layers: [
			{
				id: 'carto-light-layer',
				type: 'raster',
				source: 'carto-light',
				minzoom: 0,
				maxzoom: 20
			}
		]
	},
	attribution:
		'© <a href="https://carto.com/">CARTO</a> © <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
};

// Satellite theme for detailed geographical context
export const satelliteTheme: MapTheme = {
	name: 'Satellite View',
	description: 'High-resolution satellite imagery for detailed location analysis',
	style: {
		version: 8,
		sources: {
			'esri-satellite': {
				type: 'raster',
				tiles: [
					'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}'
				],
				tileSize: 256,
				attribution: 'Esri, Maxar, Earthstar Geographics, and the GIS User Community'
			}
		},
		layers: [
			{
				id: 'esri-satellite-layer',
				type: 'raster',
				source: 'esri-satellite',
				minzoom: 0,
				maxzoom: 19
			}
		]
	},
	attribution: 'Esri, Maxar, Earthstar Geographics, and the GIS User Community'
};

// Available map themes
export const mapThemes = {
	nightOwls: nightOwlsTheme,
	dayOwls: dayOwlsTheme,
	satellite: satelliteTheme
} as const;

export type MapThemeKey = keyof typeof mapThemes;

// Default theme for the application
export const DEFAULT_THEME: MapThemeKey = 'nightOwls';

// Get a specific theme
export function getMapTheme(themeKey: MapThemeKey = DEFAULT_THEME): MapTheme {
	return mapThemes[themeKey];
}

// Map control styling for consistent appearance
export const mapControlStyles = {
	dark: {
		background: 'rgba(255, 255, 255, 0.95)', // Light background for contrast
		color: '#1e293b', // Dark text for readability
		border: '1px solid rgba(203, 213, 225, 0.8)', // Light border
		borderRadius: '8px',
		backdropFilter: 'blur(8px)',
		boxShadow: '0 4px 12px rgba(0, 0, 0, 0.3)'
	},
	light: {
		background: 'rgba(255, 255, 255, 0.95)',
		color: '#0f172a', // slate-900
		border: '1px solid rgba(203, 213, 225, 0.5)', // slate-300 border
		borderRadius: '8px',
		backdropFilter: 'blur(8px)',
		boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)'
	}
};

// Marker colors optimized for each theme
export const markerStyles = {
	nightOwls: {
		normal: '#60a5fa', // blue-400 - more visible on dark background
		suspicion: '#fbbf24', // amber-400 - warmer, more visible
		incident: '#f87171', // red-400 - softer red for dark theme
		default: '#9ca3af' // gray-400
	},
	dayOwls: {
		normal: '#3b82f6', // blue-500 - standard blue
		suspicion: '#f59e0b', // amber-500 - standard orange
		incident: '#ef4444', // red-500 - standard red
		default: '#6b7280' // gray-500
	},
	satellite: {
		normal: '#60a5fa', // blue-400 - visible on satellite imagery
		suspicion: '#fbbf24', // amber-400 - high contrast
		incident: '#f87171', // red-400 - visible but not harsh
		default: '#e5e7eb' // gray-200 - light for contrast
	}
};

// Get marker colors for a specific theme
export function getMarkerColors(themeKey: MapThemeKey = DEFAULT_THEME) {
	return markerStyles[themeKey];
}

// Utility function to get CSS custom properties for map styling
export function getMapThemeCSS(themeKey: MapThemeKey = DEFAULT_THEME) {
	const colors = getMarkerColors(themeKey);
	const controls = themeKey === 'nightOwls' ? mapControlStyles.dark : mapControlStyles.light;

	return {
		'--map-marker-normal': colors.normal,
		'--map-marker-suspicion': colors.suspicion,
		'--map-marker-incident': colors.incident,
		'--map-marker-default': colors.default,
		'--map-control-bg': controls.background,
		'--map-control-color': controls.color,
		'--map-control-border': controls.border,
		'--map-control-radius': controls.borderRadius,
		'--map-control-backdrop': controls.backdropFilter,
		'--map-control-shadow': controls.boxShadow
	};
}
