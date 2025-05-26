<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';

	interface GPSLocation {
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: number;
	}

	interface Props {
		onLocationCapture?: (location: GPSLocation) => void;
		autoCapture?: boolean;
		showMap?: boolean;
		className?: string;
	}

	let { 
		onLocationCapture = () => {}, 
		autoCapture = false, 
		showMap = false,
		className = ''
	}: Props = $props();

	let isCapturing = $state(false);
	let location = $state<GPSLocation | null>(null);
	let error = $state<string | null>(null);
	let watchId = $state<number | null>(null);

	// Auto-capture on mount if enabled
	$effect(() => {
		if (autoCapture && !location && !isCapturing) {
			captureLocation();
		}
	});

	// Cleanup watch on unmount
	$effect(() => {
		return () => {
			if (watchId !== null) {
				navigator.geolocation.clearWatch(watchId);
			}
		};
	});

	async function captureLocation() {
		if (!navigator.geolocation) {
			error = 'Geolocation is not supported by this browser';
			return;
		}

		isCapturing = true;
		error = null;

		const options: PositionOptions = {
			enableHighAccuracy: true,
			timeout: 10000,
			maximumAge: 60000 // 1 minute
		};

		try {
			const position = await new Promise<GeolocationPosition>((resolve, reject) => {
				navigator.geolocation.getCurrentPosition(resolve, reject, options);
			});

			const newLocation: GPSLocation = {
				latitude: position.coords.latitude,
				longitude: position.coords.longitude,
				accuracy: position.coords.accuracy,
				timestamp: position.timestamp
			};

			location = newLocation;
			onLocationCapture(newLocation);
		} catch (err) {
			if (err instanceof GeolocationPositionError) {
				switch (err.code) {
					case err.PERMISSION_DENIED:
						error = 'Location access denied. Please enable location permissions.';
						break;
					case err.POSITION_UNAVAILABLE:
						error = 'Location information unavailable.';
						break;
					case err.TIMEOUT:
						error = 'Location request timed out. Please try again.';
						break;
					default:
						error = 'An unknown error occurred while retrieving location.';
						break;
				}
			} else {
				error = 'Failed to capture location';
			}
		} finally {
			isCapturing = false;
		}
	}

	function startWatching() {
		if (!navigator.geolocation || watchId !== null) return;

		const options: PositionOptions = {
			enableHighAccuracy: true,
			timeout: 5000,
			maximumAge: 30000 // 30 seconds
		};

		watchId = navigator.geolocation.watchPosition(
			(position) => {
				const newLocation: GPSLocation = {
					latitude: position.coords.latitude,
					longitude: position.coords.longitude,
					accuracy: position.coords.accuracy,
					timestamp: position.timestamp
				};

				location = newLocation;
				onLocationCapture(newLocation);
			},
			(err) => {
				console.error('GPS watch error:', err);
			},
			options
		);
	}

	function stopWatching() {
		if (watchId !== null) {
			navigator.geolocation.clearWatch(watchId);
			watchId = null;
		}
	}

	function clearLocation() {
		location = null;
		error = null;
		stopWatching();
	}

	function formatCoordinates(lat: number, lng: number): string {
		return `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
	}

	function formatAccuracy(accuracy: number): string {
		if (accuracy < 10) return 'High';
		if (accuracy < 50) return 'Medium';
		return 'Low';
	}

	function getAccuracyColor(accuracy: number): string {
		if (accuracy < 10) return 'text-green-600 bg-green-50 border-green-200';
		if (accuracy < 50) return 'text-orange-600 bg-orange-50 border-orange-200';
		return 'text-red-600 bg-red-50 border-red-200';
	}
</script>

<div class="space-y-4 {className}">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-2">
			<MapPinIcon class="h-5 w-5" />
			<h3 class="font-medium">Location</h3>
		</div>
		
		{#if location}
			<Badge class="{getAccuracyColor(location.accuracy)} border text-xs">
				{formatAccuracy(location.accuracy)} accuracy
			</Badge>
		{/if}
	</div>

	<!-- Status Display -->
	{#if isCapturing}
		<div class="flex items-center gap-3 p-4 bg-blue-50 rounded-lg border border-blue-200">
			<RefreshCwIcon class="h-5 w-5 text-blue-600 animate-spin" />
			<div>
				<p class="text-sm font-medium text-blue-900">Capturing location...</p>
				<p class="text-xs text-blue-700">Please ensure location services are enabled</p>
			</div>
		</div>
	{:else if error}
		<div class="flex items-start gap-3 p-4 bg-red-50 rounded-lg border border-red-200">
			<AlertCircleIcon class="h-5 w-5 text-red-600 mt-0.5" />
			<div class="flex-1">
				<p class="text-sm font-medium text-red-900">Location Error</p>
				<p class="text-xs text-red-700">{error}</p>
			</div>
		</div>
	{:else if location}
		<div class="flex items-start gap-3 p-4 bg-green-50 rounded-lg border border-green-200">
			<CheckCircleIcon class="h-5 w-5 text-green-600 mt-0.5" />
			<div class="flex-1">
				<p class="text-sm font-medium text-green-900">Location captured</p>
				<p class="text-xs text-green-700 font-mono">
					{formatCoordinates(location.latitude, location.longitude)}
				</p>
				<p class="text-xs text-green-700 mt-1">
					Accuracy: ±{location.accuracy.toFixed(0)}m • 
					{new Date(location.timestamp).toLocaleTimeString()}
				</p>
			</div>
		</div>
	{:else}
		<div class="p-4 bg-muted/20 rounded-lg border border-dashed">
			<div class="text-center">
				<MapPinIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
				<p class="text-sm text-muted-foreground">No location captured</p>
			</div>
		</div>
	{/if}

	<!-- Map Placeholder (if enabled) -->
	{#if showMap && location}
		<div class="bg-muted/20 rounded-lg p-4 border">
			<div class="aspect-video flex items-center justify-center">
				<div class="text-center">
					<MapPinIcon class="h-12 w-12 text-muted-foreground mx-auto mb-2" />
					<p class="text-sm text-muted-foreground">Map integration</p>
					<p class="text-xs text-muted-foreground">
						{formatCoordinates(location.latitude, location.longitude)}
					</p>
				</div>
			</div>
		</div>
	{/if}

	<!-- Action Buttons -->
	<div class="flex gap-2">
		{#if !location && !isCapturing}
			<Button onclick={captureLocation} class="flex-1">
				<MapPinIcon class="h-4 w-4 mr-2" />
				Capture Location
			</Button>
		{:else if location}
			<Button variant="outline" onclick={captureLocation} disabled={isCapturing}>
				<RefreshCwIcon class="h-4 w-4 mr-2" />
				Refresh
			</Button>
			
			{#if watchId === null}
				<Button variant="outline" onclick={startWatching}>
					Track Live
				</Button>
			{:else}
				<Button variant="outline" onclick={stopWatching}>
					Stop Tracking
				</Button>
			{/if}
			
			<Button variant="ghost" onclick={clearLocation}>
				Clear
			</Button>
		{:else if error}
			<Button onclick={captureLocation} variant="outline" class="flex-1">
				<RefreshCwIcon class="h-4 w-4 mr-2" />
				Try Again
			</Button>
		{/if}
	</div>

	<!-- Technical Details (for debugging) -->
	{#if location}
		<details class="text-xs">
			<summary class="cursor-pointer text-muted-foreground hover:text-foreground">
				Technical Details
			</summary>
			<div class="mt-2 p-3 bg-muted/20 rounded font-mono text-xs space-y-1">
				<div>Latitude: {location.latitude}</div>
				<div>Longitude: {location.longitude}</div>
				<div>Accuracy: {location.accuracy}m</div>
				<div>Timestamp: {new Date(location.timestamp).toISOString()}</div>
			</div>
		</details>
	{/if}
</div> 