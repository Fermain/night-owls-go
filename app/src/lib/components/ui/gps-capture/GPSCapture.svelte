<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import LoaderIcon from '@lucide/svelte/icons/loader-2';
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

	function formatAccuracy(accuracy: number): string {
		return accuracy < 1000 ? `±${Math.round(accuracy)}m` : `±${(accuracy / 1000).toFixed(1)}km`;
	}
</script>

<div class="space-y-3 {className}">
	{#if permissionStatus === 'denied' || permissionStatus === 'unsupported'}
		<!-- Permission issues - show minimal error -->
		<div class="flex items-center gap-2 text-orange-600 text-sm">
			<LockIcon class="h-4 w-4" />
			<span>Location unavailable</span>
		</div>
	{:else if permissionStatus === 'checking'}
		<!-- Checking permissions -->
		<div class="flex items-center gap-2 text-blue-600 text-sm">
			<LoaderIcon class="h-4 w-4 animate-spin" />
			<span>Checking location...</span>
		</div>
	{:else if isCapturing}
		<!-- Capturing location -->
		<div class="flex items-center gap-2 text-blue-600 text-sm">
			<LoaderIcon class="h-4 w-4 animate-spin" />
			<span>Getting location...</span>
		</div>
	{:else if capturedLocation}
		<!-- Location captured successfully -->
		<div class="flex items-center justify-between p-3 bg-green-50 rounded-lg border border-green-200">
			<div class="flex items-center gap-2">
				<CheckCircleIcon class="h-4 w-4 text-green-600" />
				<div>
					<p class="text-sm font-medium text-green-900">Location captured</p>
					<p class="text-xs text-green-700">
						{formatAccuracy(capturedLocation.accuracy)} accuracy
					</p>
				</div>
			</div>
			<Button
				variant="ghost"
				size="sm"
				onclick={captureLocation}
				disabled={isCapturing}
				class="text-green-600 hover:text-green-700 h-8 w-8 p-0"
			>
				<RefreshCwIcon class="h-3 w-3" />
			</Button>
		</div>
	{:else if error}
		<!-- Error with retry option -->
		<div class="flex items-center justify-between p-3 bg-orange-50 rounded-lg border border-orange-200">
			<div class="flex items-center gap-2">
				<AlertCircleIcon class="h-4 w-4 text-orange-600" />
				<div>
					<p class="text-sm font-medium text-orange-900">Location failed</p>
					<p class="text-xs text-orange-700">Optional for reports</p>
				</div>
			</div>
			<Button
				variant="ghost"
				size="sm"
				onclick={captureLocation}
				disabled={isCapturing}
				class="text-orange-600 hover:text-orange-700 h-8 w-8 p-0"
			>
				<RefreshCwIcon class="h-3 w-3" />
			</Button>
		</div>
	{:else}
		<!-- Default state - offer to capture location -->
		<Button 
			variant="outline" 
			onclick={captureLocation} 
			disabled={isCapturing}
			size="sm"
			class="w-full"
		>
			<MapPinIcon class="h-4 w-4 mr-2" />
			Capture Location (Optional)
		</Button>
	{/if}
</div>
