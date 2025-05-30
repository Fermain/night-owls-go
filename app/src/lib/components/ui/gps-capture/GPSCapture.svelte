<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import LoaderIcon from '@lucide/svelte/icons/loader-2';
	import EditIcon from '@lucide/svelte/icons/edit';
	import XIcon from '@lucide/svelte/icons/x';
	import LockIcon from '@lucide/svelte/icons/lock';

	interface GeolocationData {
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: string;
	}

	interface Props {
		onLocationCaptured?: (location: GeolocationData) => void;
		onError?: (error: string) => void;
		autoCapture?: boolean;
		className?: string;
	}

	let { onLocationCaptured, onError, autoCapture = false, className = '' }: Props = $props();

	// State
	let permissionStatus = $state<'granted' | 'denied' | 'prompt' | 'checking' | 'unsupported'>(
		'checking'
	);
	let isCapturing = $state(false);
	let capturedLocation = $state<GeolocationData | null>(null);
	let error = $state<string | null>(null);
	let showManualInput = $state(false);
	let manualLat = $state('');
	let manualLng = $state('');

	// Check permissions on mount
	onMount(async () => {
		await checkLocationPermission();
		// Auto-capture if enabled and permissions granted
		if (autoCapture && permissionStatus === 'granted' && !capturedLocation) {
			captureLocation();
		}
	});

	// Check location permission status
	async function checkLocationPermission() {
		// Check if geolocation is supported
		if (!navigator.geolocation) {
			permissionStatus = 'unsupported';
			return;
		}

		// Check permissions API if available
		if ('permissions' in navigator) {
			try {
				const result = await navigator.permissions.query({ name: 'geolocation' });
				permissionStatus = result.state as any;

				// Listen for permission changes
				result.addEventListener('change', () => {
					permissionStatus = result.state as any;
				});

				return;
			} catch (error) {
				// Fallback to 'prompt' if permissions API fails
				permissionStatus = 'prompt';
			}
		}

		// Fallback: assume prompt state
		permissionStatus = 'prompt';
	}

	// Capture location
	async function captureLocation() {
		if (permissionStatus === 'denied') {
			error =
				'Location access denied. Please enable location permissions in your browser settings.';
			onError?.(error);
			return;
		}

		if (!navigator.geolocation) {
			error = 'Geolocation is not supported by this browser.';
			onError?.(error);
			return;
		}

		isCapturing = true;
		error = null;

		try {
			const position = await new Promise<GeolocationPosition>((resolve, reject) => {
				navigator.geolocation.getCurrentPosition(resolve, reject, {
					enableHighAccuracy: true,
					timeout: 15000,
					maximumAge: 300000 // 5 minutes
				});
			});

			const locationData: GeolocationData = {
				latitude: position.coords.latitude,
				longitude: position.coords.longitude,
				accuracy: position.coords.accuracy,
				timestamp: new Date().toISOString()
			};

			capturedLocation = locationData;
			onLocationCaptured?.(locationData);

			// Update permission status if successful
			if (permissionStatus === 'prompt') {
				permissionStatus = 'granted';
			}
		} catch (err) {
			let errorMessage = 'Location capture failed';

			if (err instanceof GeolocationPositionError) {
				switch (err.code) {
					case err.PERMISSION_DENIED:
						permissionStatus = 'denied';
						errorMessage = 'Location access denied';
						break;
					case err.POSITION_UNAVAILABLE:
						errorMessage = 'Location unavailable. Please try again or use manual entry.';
						break;
					case err.TIMEOUT:
						errorMessage = 'Location request timed out. Please try again.';
						break;
				}
			}

			error = errorMessage;
			onError?.(errorMessage);
		} finally {
			isCapturing = false;
		}
	}

	// Manual location input
	function handleManualLocationSubmit() {
		const lat = parseFloat(manualLat);
		const lng = parseFloat(manualLng);

		if (isNaN(lat) || isNaN(lng)) {
			onError?.('Please enter valid latitude and longitude coordinates');
			return;
		}

		if (lat < -90 || lat > 90) {
			onError?.('Latitude must be between -90 and 90 degrees');
			return;
		}

		if (lng < -180 || lng > 180) {
			onError?.('Longitude must be between -180 and 180 degrees');
			return;
		}

		const locationData: GeolocationData = {
			latitude: lat,
			longitude: lng,
			accuracy: 1000, // Manual input has lower accuracy
			timestamp: new Date().toISOString()
		};

		capturedLocation = locationData;
		showManualInput = false;
		manualLat = '';
		manualLng = '';
		error = null;
		onLocationCaptured?.(locationData);
	}

	function clearLocation() {
		capturedLocation = null;
		error = null;
		showManualInput = false;
		manualLat = '';
		manualLng = '';
	}

	function formatCoordinates(lat: number, lng: number): string {
		return `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
	}

	function formatAccuracy(accuracy: number): string {
		return accuracy < 1000 ? `±${Math.round(accuracy)}m` : `±${(accuracy / 1000).toFixed(1)}km`;
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

		{#if capturedLocation}
			<Badge class="{getAccuracyColor(capturedLocation.accuracy)} border text-xs">
				{formatAccuracy(capturedLocation.accuracy)} accuracy
			</Badge>
		{/if}
	</div>

	<!-- Permission Denied State -->
	{#if permissionStatus === 'denied'}
		<div class="p-4 bg-orange-50 rounded-lg border border-orange-200">
			<div class="flex items-start gap-3">
				<LockIcon class="h-5 w-5 text-orange-600 mt-0.5" />
				<div class="flex-1">
					<p class="text-sm font-medium text-orange-900">Location access denied</p>
					<p class="text-xs text-orange-700 mt-1">
						Location permissions have been denied for this site. Reports can be submitted without
						location data.
					</p>
					<details class="mt-2">
						<summary class="text-xs text-orange-600 cursor-pointer hover:text-orange-700">
							How to enable location access
						</summary>
						<div class="mt-1 text-xs text-orange-600 space-y-1">
							<p><strong>Chrome/Safari:</strong> Click the location icon in the address bar</p>
							<p>
								<strong>Firefox:</strong> Click the shield icon and select "Allow Location Access"
							</p>
							<p>After enabling, refresh the page and try again.</p>
						</div>
					</details>
				</div>
			</div>
		</div>

		<!-- Unsupported State -->
	{:else if permissionStatus === 'unsupported'}
		<div class="p-4 bg-gray-50 rounded-lg border border-gray-200">
			<div class="flex items-center gap-3">
				<AlertCircleIcon class="h-5 w-5 text-gray-600" />
				<div>
					<p class="text-sm font-medium text-gray-900">Location not supported</p>
					<p class="text-xs text-gray-700">This browser doesn't support location services.</p>
				</div>
			</div>
		</div>

		<!-- Checking Permissions State -->
	{:else if permissionStatus === 'checking'}
		<div class="p-4 bg-blue-50 rounded-lg border border-blue-200">
			<div class="flex items-center gap-3">
				<LoaderIcon class="h-5 w-5 text-blue-600 animate-spin" />
				<p class="text-sm text-blue-900">Checking location permissions...</p>
			</div>
		</div>

		<!-- Active Location Interface (when permissions granted or prompt) -->
	{:else}
		<!-- Capturing State -->
		{#if isCapturing}
			<div class="flex items-center gap-3 p-4 bg-blue-50 rounded-lg border border-blue-200">
				<LoaderIcon class="h-5 w-5 text-blue-600 animate-spin" />
				<div>
					<p class="text-sm font-medium text-blue-900">Capturing location...</p>
					<p class="text-xs text-blue-700">This may take a moment</p>
				</div>
			</div>

			<!-- Error State -->
		{:else if error}
			<div class="flex items-start gap-3 p-4 bg-orange-50 rounded-lg border border-orange-200">
				<AlertCircleIcon class="h-5 w-5 text-orange-600 mt-0.5" />
				<div class="flex-1">
					<p class="text-sm font-medium text-orange-900">Location capture failed</p>
					<p class="text-xs text-orange-700">{error}</p>
				</div>
			</div>

			<!-- Success State -->
		{:else if capturedLocation}
			<div class="p-3 bg-green-50 rounded-lg border border-green-200">
				<div class="flex items-start gap-2">
					<CheckCircleIcon class="h-4 w-4 text-green-600 mt-0.5" />
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium text-green-900">Location captured</p>
						<p class="text-xs text-green-700 font-mono">
							{formatCoordinates(capturedLocation.latitude, capturedLocation.longitude)}
						</p>
						<p class="text-xs text-green-600">
							Accuracy: {formatAccuracy(capturedLocation.accuracy)}
						</p>
					</div>
					<Button
						variant="ghost"
						size="sm"
						onclick={captureLocation}
						disabled={isCapturing}
						class="text-green-600 hover:text-green-700"
					>
						<RefreshCwIcon class="h-3 w-3" />
					</Button>
				</div>
			</div>

			<!-- Default State -->
		{:else}
			<div class="p-4 bg-muted/20 rounded-lg border border-dashed">
				<div class="text-center">
					<MapPinIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-sm text-muted-foreground">No location captured</p>
					<p class="text-xs text-muted-foreground mt-1">Location data is optional for reports</p>
				</div>
			</div>
		{/if}

		<!-- Manual Location Input (only show if GPS fails and permissions are OK) -->
		{#if showManualInput}
			<div class="p-4 bg-blue-50 rounded-lg border border-blue-200">
				<div class="flex items-center justify-between mb-3">
					<h4 class="text-sm font-medium text-blue-900">Manual Location Entry</h4>
					<Button variant="ghost" size="sm" onclick={() => (showManualInput = false)}>
						<XIcon class="h-4 w-4" />
					</Button>
				</div>
				<div class="space-y-3">
					<div class="grid grid-cols-2 gap-3">
						<div>
							<Label for="manual-lat" class="text-xs">Latitude</Label>
							<Input
								id="manual-lat"
								bind:value={manualLat}
								placeholder="-33.9249"
								class="text-sm"
							/>
						</div>
						<div>
							<Label for="manual-lng" class="text-xs">Longitude</Label>
							<Input id="manual-lng" bind:value={manualLng} placeholder="18.4241" class="text-sm" />
						</div>
					</div>
					<p class="text-xs text-blue-700">
						Enter coordinates in decimal degrees format (e.g., -33.9249, 18.4241 for Cape Town)
					</p>
					<Button onclick={handleManualLocationSubmit} size="sm" class="w-full">
						<CheckCircleIcon class="h-4 w-4 mr-2" />
						Use Manual Location
					</Button>
				</div>
			</div>
		{/if}

		<!-- Action Buttons -->
		<div class="flex gap-2">
			{#if !capturedLocation && !isCapturing}
				<Button variant="outline" onclick={captureLocation} disabled={isCapturing} class="flex-1">
					{#if isCapturing}
						<LoaderIcon class="h-4 w-4 mr-2 animate-spin" />
						Capturing location...
					{:else}
						<MapPinIcon class="h-4 w-4 mr-2" />
						Capture GPS Location
					{/if}
				</Button>
				{#if error}
					<Button variant="ghost" onclick={() => (showManualInput = true)}>
						<EditIcon class="h-4 w-4 mr-2" />
						Manual Entry
					</Button>
				{/if}
			{:else if capturedLocation}
				<Button variant="outline" onclick={captureLocation} disabled={isCapturing}>
					<RefreshCwIcon class="h-4 w-4 mr-2" />
					Refresh
				</Button>
				<Button variant="ghost" onclick={clearLocation}>Clear</Button>
			{:else if error}
				<Button variant="outline" onclick={captureLocation} disabled={isCapturing} class="flex-1">
					{#if isCapturing}
						<LoaderIcon class="h-4 w-4 mr-2 animate-spin" />
						Retrying...
					{:else}
						<MapPinIcon class="h-4 w-4 mr-2" />
						Try Again
					{/if}
				</Button>
				<Button variant="ghost" onclick={() => (showManualInput = true)}>
					<EditIcon class="h-4 w-4 mr-2" />
					Manual Entry
				</Button>
			{/if}
		</div>
	{/if}
</div>
