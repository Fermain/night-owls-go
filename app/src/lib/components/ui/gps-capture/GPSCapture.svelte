<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import LoaderIcon from '@lucide/svelte/icons/loader-2';
	import EditIcon from '@lucide/svelte/icons/edit';
	import XIcon from '@lucide/svelte/icons/x';

	interface GPSLocation {
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: number;
	}

	interface Props {
		onLocationCaptured?: (location: GeolocationData) => void;
		onError?: (error: string) => void;
		autoCapture?: boolean;
		showMap?: boolean;
		className?: string;
	}

	interface GeolocationData {
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: string;
	}

	let { 
		onLocationCaptured, 
		onError, 
		autoCapture = false, 
		showMap = false,
		className = ''
	}: Props = $props();

	let isCapturing = $state(false);
	let capturedLocation = $state<GeolocationData | null>(null);
	let error = $state<string | null>(null);
	let watchId = $state<number | null>(null);
	let showManualInput = $state(false);
	let manualLat = $state('');
	let manualLng = $state('');

	// Auto-capture on mount if enabled
	$effect(() => {
		if (autoCapture && !capturedLocation && !error) {
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
		console.log('🔍 GPS Capture: Starting location capture process...');
		
		if (!navigator.geolocation) {
			const errorMsg = 'Geolocation is not supported by this browser';
			console.error('❌ GPS Capture: Geolocation not supported');
			error = errorMsg;
			onError?.(errorMsg);
			return;
		}

		console.log('✅ GPS Capture: Geolocation API available');
		console.log('🌐 GPS Capture: User Agent:', navigator.userAgent);
		console.log('🖥️ GPS Capture: Platform:', navigator.platform);
		console.log('🔒 GPS Capture: Secure context (HTTPS):', window.isSecureContext);
		console.log('🌍 GPS Capture: Online status:', navigator.onLine);
		console.log('🕐 GPS Capture: Current time:', new Date().toISOString());
		
		// Check if we're in localhost (which should work for geolocation)
		console.log('🏠 GPS Capture: Hostname:', window.location.hostname);
		console.log('🔗 GPS Capture: Protocol:', window.location.protocol);
		
		// Check for any blocking extensions or privacy settings
		console.log('🍪 GPS Capture: Cookies enabled:', navigator.cookieEnabled);
		console.log('🔍 GPS Capture: Do Not Track:', navigator.doNotTrack);
		
		// Check if there are any other location-related APIs
		console.log('🧭 GPS Capture: DeviceOrientationEvent:', typeof DeviceOrientationEvent !== 'undefined');
		console.log('📱 GPS Capture: DeviceMotionEvent:', typeof DeviceMotionEvent !== 'undefined');
		
		// Check permissions API if available
		if ('permissions' in navigator) {
			try {
				const permission = await navigator.permissions.query({ name: 'geolocation' });
				console.log('🔐 GPS Capture: Permission state:', permission.state);
				console.log('🔐 GPS Capture: Permission object:', permission);
			} catch (permErr) {
				console.warn('⚠️ GPS Capture: Could not check permissions:', permErr);
			}
		} else {
			console.log('⚠️ GPS Capture: Permissions API not available');
		}
		
		// Try to get cached location first (very permissive)
		console.log('🗂️ GPS Capture: Attempting to get cached location first...');
		try {
			const cachedPosition = await new Promise<GeolocationPosition>((resolve, reject) => {
				navigator.geolocation.getCurrentPosition(
					(pos) => {
						console.log('✅ GPS Capture: Got cached position:', pos);
						resolve(pos);
					},
					(err) => {
						console.log('❌ GPS Capture: No cached position available:', err);
						reject(err);
					},
					{
						enableHighAccuracy: true,
						timeout: 5000, // Longer timeout for cache check
						maximumAge: Infinity // Accept any cached data
					}
				);
			});
			
			// If we got cached data, use it
			const locationData: GeolocationData = {
				latitude: cachedPosition.coords.latitude,
				longitude: cachedPosition.coords.longitude,
				accuracy: cachedPosition.coords.accuracy,
				timestamp: new Date().toISOString()
			};

			console.log('🎉 GPS Capture: Using cached location data:', locationData);
			capturedLocation = locationData;
			onLocationCaptured?.(locationData);
			isCapturing = false;
			return;
		} catch (cacheErr) {
			console.log('📝 GPS Capture: No cached location, proceeding with fresh request...');
		}

		isCapturing = true;
		error = null;
		console.log('🚀 GPS Capture: Starting getCurrentPosition...');

		try {
			const position = await new Promise<GeolocationPosition>((resolve, reject) => {
				console.log('📍 GPS Capture: Calling navigator.geolocation.getCurrentPosition...');
				
				navigator.geolocation.getCurrentPosition(
					(pos) => {
						console.log('✅ GPS Capture: Success callback triggered');
						console.log('📍 GPS Capture: Position:', pos);
						resolve(pos);
					},
					(err) => {
						console.error('❌ GPS Capture: Error callback triggered');
						console.error('❌ GPS Capture: Error object:', err);
						console.error('❌ GPS Capture: Error code:', err.code);
						console.error('❌ GPS Capture: Error message:', err.message);
						
						if (err.code === err.POSITION_UNAVAILABLE) {
							console.error('📍 GPS Capture: POSITION_UNAVAILABLE - CoreLocation framework issue');
						}
						
						reject(err);
					},
					{
						enableHighAccuracy: false,
						timeout: 30000, // Increased to 30 seconds
						maximumAge: 300000
					}
				);
			});

			const locationData: GeolocationData = {
				latitude: position.coords.latitude,
				longitude: position.coords.longitude,
				accuracy: position.coords.accuracy,
				timestamp: new Date().toISOString()
			};

			console.log('📊 GPS Capture: Processed location data:', locationData);
			capturedLocation = locationData;
			onLocationCaptured?.(locationData);
			
		} catch (err) {
			console.error('💥 GPS Capture: Location capture failed:', err);
			
			let errorMsg = 'Location capture failed';
			let fallbackSuggestion = '';
			
			if (err instanceof GeolocationPositionError) {
				switch (err.code) {
					case err.PERMISSION_DENIED:
						errorMsg = 'Location access denied';
						fallbackSuggestion = 'Please enable location permissions in your browser settings.';
						break;
					case err.POSITION_UNAVAILABLE:
						errorMsg = 'Location unavailable';
						fallbackSuggestion = 'This appears to be a macOS CoreLocation framework issue. You can submit the report without location data or use manual entry.';
						break;
					case err.TIMEOUT:
						errorMsg = 'Location request timed out';
						fallbackSuggestion = 'Please try again or use manual entry.';
						break;
				}
			}

			error = errorMsg;
			onError?.(`${errorMsg}${fallbackSuggestion ? ` ${fallbackSuggestion}` : ''}`);
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

				capturedLocation = {
					latitude: newLocation.latitude,
					longitude: newLocation.longitude,
					accuracy: newLocation.accuracy,
					timestamp: new Date(newLocation.timestamp).toISOString()
				};
				onLocationCaptured?.(capturedLocation);
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
		capturedLocation = null;
		error = null;
		showManualInput = false;
		manualLat = '';
		manualLng = '';
		stopWatching();
	}

	function toggleManualInput() {
		showManualInput = !showManualInput;
		if (!showManualInput) {
			manualLat = '';
			manualLng = '';
		}
	}

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

	<!-- Status Display -->
	{#if isCapturing}
		<div class="flex items-center gap-3 p-4 bg-blue-50 rounded-lg border border-blue-200">
			<LoaderIcon class="h-5 w-5 text-blue-600 animate-spin" />
			<div>
				<p class="text-sm font-medium text-blue-900">Capturing location...</p>
				<p class="text-xs text-blue-700">This may take a moment</p>
			</div>
		</div>
	{:else if error}
		<div class="flex items-start gap-3 p-4 bg-orange-50 rounded-lg border border-orange-200">
			<AlertCircleIcon class="h-5 w-5 text-orange-600 mt-0.5" />
			<div class="flex-1">
				<p class="text-sm font-medium text-orange-900">Location capture failed</p>
				<p class="text-xs text-orange-700">{error}</p>
				<p class="text-xs text-orange-600 mt-1">💡 You can submit the report without location data or enter coordinates manually.</p>
			</div>
		</div>
	{:else if capturedLocation}
		<div class="p-3 bg-green-50 dark:bg-green-950/50 border border-green-200 dark:border-green-800 rounded-lg">
			<div class="flex items-start gap-2">
				<CheckCircleIcon class="h-4 w-4 text-green-600 dark:text-green-400 mt-0.5" />
				<div class="flex-1 min-w-0">
					<p class="text-sm font-medium text-green-900 dark:text-green-100">Location captured</p>
					<p class="text-xs text-green-700 dark:text-green-300 font-mono">
						{formatCoordinates(capturedLocation.latitude, capturedLocation.longitude)}
					</p>
					<p class="text-xs text-green-600 dark:text-green-400">
						Accuracy: {formatAccuracy(capturedLocation.accuracy)}
					</p>
				</div>
				<Button
					variant="ghost"
					size="sm"
					onclick={captureLocation}
					disabled={isCapturing}
					class="text-green-600 hover:text-green-700 dark:text-green-400 dark:hover:text-green-300"
				>
					{#if isCapturing}
						<LoaderIcon class="h-3 w-3 animate-spin" />
					{:else}
						<MapPinIcon class="h-3 w-3" />
					{/if}
				</Button>
			</div>
		</div>
	{:else}
		<div class="p-4 bg-muted/20 rounded-lg border border-dashed">
			<div class="text-center">
				<MapPinIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
				<p class="text-sm text-muted-foreground">No location captured</p>
				<p class="text-xs text-muted-foreground mt-1">Location data is optional for reports</p>
			</div>
		</div>
	{/if}

	<!-- Manual Location Input -->
	{#if showManualInput}
		<div class="p-4 bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800 rounded-lg">
			<div class="flex items-center justify-between mb-3">
				<h4 class="text-sm font-medium text-blue-900 dark:text-blue-100">Manual Location Entry</h4>
				<Button variant="ghost" size="sm" onclick={toggleManualInput}>
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
						<Input
							id="manual-lng"
							bind:value={manualLng}
							placeholder="18.4241"
							class="text-sm"
						/>
					</div>
				</div>
				<p class="text-xs text-blue-700 dark:text-blue-300">
					Enter coordinates in decimal degrees format (e.g., -33.9249, 18.4241 for Cape Town)
				</p>
				<Button onclick={handleManualLocationSubmit} size="sm" class="w-full">
					<CheckCircleIcon class="h-4 w-4 mr-2" />
					Use Manual Location
				</Button>
			</div>
		</div>
	{/if}

	<!-- Map Placeholder (if enabled) -->
	{#if showMap && capturedLocation}
		<div class="bg-muted/20 rounded-lg p-4 border">
			<div class="aspect-video flex items-center justify-center">
				<div class="text-center">
					<MapPinIcon class="h-12 w-12 text-muted-foreground mx-auto mb-2" />
					<p class="text-sm text-muted-foreground">Map integration</p>
					<p class="text-xs text-muted-foreground">
						{formatCoordinates(capturedLocation.latitude, capturedLocation.longitude)}
					</p>
				</div>
			</div>
		</div>
	{/if}

	<!-- Action Buttons -->
	<div class="flex gap-2">
		{#if !capturedLocation && !isCapturing}
			<Button
				variant="outline"
				onclick={captureLocation}
				disabled={isCapturing}
				class="flex-1 justify-start"
			>
				{#if isCapturing}
					<LoaderIcon class="h-4 w-4 mr-2 animate-spin" />
					Capturing location...
				{:else}
					<MapPinIcon class="h-4 w-4 mr-2" />
					Capture GPS Location
				{/if}
			</Button>
			{#if error}
				<Button
					variant="ghost"
					onclick={toggleManualInput}
					class="flex-shrink-0"
				>
					<EditIcon class="h-4 w-4 mr-2" />
					Manual Entry
				</Button>
			{/if}
		{:else if capturedLocation}
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
			<Button
				variant="outline"
				onclick={captureLocation}
				disabled={isCapturing}
				class="flex-1 justify-start"
			>
				{#if isCapturing}
					<LoaderIcon class="h-4 w-4 mr-2 animate-spin" />
					Retrying...
				{:else}
					<MapPinIcon class="h-4 w-4 mr-2" />
					Try Again
				{/if}
			</Button>
			<Button
				variant="ghost"
				onclick={toggleManualInput}
				class="flex-shrink-0"
			>
				<EditIcon class="h-4 w-4 mr-2" />
				Manual Entry
			</Button>
		{/if}
	</div>

	<!-- Technical Details (for debugging) -->
	{#if capturedLocation}
		<details class="text-xs">
			<summary class="cursor-pointer text-muted-foreground hover:text-foreground">
				Technical Details
			</summary>
			<div class="mt-2 p-3 bg-muted/20 rounded font-mono text-xs space-y-1">
				<div>Latitude: {capturedLocation.latitude}</div>
				<div>Longitude: {capturedLocation.longitude}</div>
				<div>Accuracy: {capturedLocation.accuracy}m</div>
				<div>Timestamp: {new Date(capturedLocation.timestamp).toISOString()}</div>
			</div>
		</details>
	{/if}

	<!-- Troubleshooting Information -->
	{#if error}
		<details class="text-xs">
			<summary class="cursor-pointer text-muted-foreground hover:text-foreground">
				Troubleshooting & Alternatives
			</summary>
			<div class="mt-2 p-3 bg-muted/20 rounded text-xs space-y-2">
				<div class="space-y-1">
					<p class="font-medium">Quick Solutions:</p>
					<ul class="list-disc list-inside space-y-1 text-muted-foreground">
						<li><strong>Use Manual Entry:</strong> Click "Manual Entry" to input coordinates directly</li>
						<li><strong>Submit Without Location:</strong> Reports can be submitted without location data</li>
						<li><strong>Find Coordinates:</strong> Use Google Maps or Apple Maps to get your coordinates</li>
					</ul>
				</div>
				<div class="space-y-1">
					<p class="font-medium">macOS Location Services (if needed):</p>
					<ul class="list-disc list-inside space-y-1 text-muted-foreground">
						<li>System Preferences → Security & Privacy → Privacy → Location Services</li>
						<li>Enable location for your browser (Chrome/Safari/Firefox)</li>
						<li>Restart browser after making changes</li>
					</ul>
				</div>
			</div>
		</details>
	{/if}
</div> 