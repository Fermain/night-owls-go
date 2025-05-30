<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Badge } from '$lib/components/ui/badge';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import InfoIcon from '@lucide/svelte/icons/info';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import SendIcon from '@lucide/svelte/icons/send';
	import WifiOffIcon from '@lucide/svelte/icons/wifi-off';
	import CloudIcon from '@lucide/svelte/icons/cloud';
	import { toast } from 'svelte-sonner';
	import EmergencyContacts from '$lib/components/emergency/EmergencyContacts.svelte';
	import GPSCapture from '$lib/components/ui/gps-capture/GPSCapture.svelte';
	import { UserApiService } from '$lib/services/api/user';
	import { goto } from '$app/navigation';
	import { offlineService } from '$lib/services/offlineService';

	// Form state
	let selectedSeverity = $state('0'); // Default to Normal (severity 0)
	let reportMessage = $state('');
	let isSubmitting = $state(false);
	let isOnline = $state(true);
	let gpsLocation = $state<{
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: string;
	} | null>(null);

	// Current shift interface
	interface CurrentShift {
		id: number;
		schedule_name: string;
		start_time: string;
		end_time: string;
		location: string;
	}

	// Mock current shift data (would come from user session)
	// Set to null to simulate no active shift - in real implementation this would be determined by checking user's active bookings
	const mockCurrentShift = $state<CurrentShift | null>(null);

	const severityOptions = [
		{
			value: '0',
			label: 'Normal',
			description: 'Routine patrol notes, minor observations',
			color:
				'text-green-700 bg-green-50 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800',
			icon: CheckCircleIcon
		},
		{
			value: '1',
			label: 'Suspicion',
			description: 'General incidents, noise complaints, suspicious activity',
			color:
				'text-orange-700 bg-orange-50 border-orange-200 dark:bg-orange-950 dark:text-orange-300 dark:border-orange-800',
			icon: InfoIcon
		},
		{
			value: '2',
			label: 'Incident',
			description: 'Security threats, property damage, immediate attention needed',
			color:
				'text-red-700 bg-red-50 border-red-200 dark:bg-red-950 dark:text-red-300 dark:border-red-800',
			icon: AlertTriangleIcon
		}
	];

	// Initialize offline service and monitor status
	onMount(() => {
		let unsubscribe: (() => void) | undefined;

		const initializeService = async () => {
			try {
				await offlineService.init();

				// Subscribe to online/offline status
				unsubscribe = offlineService.state.subscribe((state) => {
					isOnline = state.isOnline;
				});
			} catch (error) {
				console.error('Failed to initialize offline service:', error);
			}
		};

		// Start the async initialization
		initializeService();

		// Return cleanup function
		return () => {
			if (unsubscribe) {
				unsubscribe();
			}
		};
	});

	// Get current time formatted
	function getCurrentTime() {
		return new Date().toLocaleString('en-GB', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getSeverityOption(value: string) {
		return severityOptions.find((opt) => opt.value === value);
	}

	function handleLocationCaptured(location: {
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: string;
	}) {
		gpsLocation = location;
		toast.success('Location captured successfully');
	}

	function handleLocationError(error: string) {
		console.log('Location capture failed:', error);
		// Don't show error toast since location is optional
		// Users can still submit reports without location data
	}

	async function handleSubmit() {
		if (!selectedSeverity || !reportMessage.trim()) {
			toast.error('Please select severity and provide a message');
			return;
		}

		isSubmitting = true;

		try {
			// Determine if user is on shift
			let activeBooking = null;
			let isOffShift = true;

			if (isOnline) {
				try {
					const userBookings = await UserApiService.getMyBookings();
					const now = new Date();
					activeBooking = userBookings.find((booking) => {
						const shiftStart = new Date(booking.shift_start);
						const shiftEnd = new Date(booking.shift_end);
						return now >= shiftStart && now <= shiftEnd && booking.checked_in_at;
					});
					isOffShift = !activeBooking;
				} catch (error) {
					console.warn('Could not check shift status, assuming off-shift:', error);
				}
			}

			const reportData = {
				severity: parseInt(selectedSeverity),
				message: reportMessage.trim(),
				latitude: gpsLocation?.latitude,
				longitude: gpsLocation?.longitude,
				accuracy: gpsLocation?.accuracy,
				locationTimestamp: gpsLocation?.timestamp,
				bookingId: activeBooking?.booking_id,
				isOffShift
			};

			// Use offline service (works both online and offline)
			const reportId = await offlineService.createIncidentReport(reportData);

			if (isOnline) {
				toast.success('Report submitted successfully');
			} else {
				toast.success('Report saved offline - will sync when connection returns', {
					description:
						'Your report is safely stored and will be submitted automatically when you reconnect.'
				});
			}

			// Reset form
			selectedSeverity = '0'; // Reset to default Normal severity
			reportMessage = '';
			gpsLocation = null;

			// Navigate back to home
			goto('/');
		} catch (error) {
			console.error('Failed to submit report:', error);

			if (isOnline) {
				toast.error(error instanceof Error ? error.message : 'Failed to submit report');
			} else {
				toast.error('Failed to save report offline', {
					description: 'Please try again or check your device storage.'
				});
			}
		} finally {
			isSubmitting = false;
		}
	}

	function handleEmergency() {
		// Handle emergency call
		if (confirm('This will call emergency services immediately. Continue?')) {
			window.location.href = 'tel:999';
		}
	}
</script>

<svelte:head>
	<title>Report Incident - Night Owls</title>
</svelte:head>

<div class="px-4 py-4 space-y-4">
	<!-- Offline Status Banner -->
	{#if !isOnline}
		<Card.Root class="bg-orange-50 border-orange-200 dark:bg-orange-950/50 dark:border-orange-800">
			<Card.Content class="p-4">
				<div class="flex items-center gap-3">
					<WifiOffIcon class="h-5 w-5 text-orange-600 dark:text-orange-400" />
					<div class="flex-1">
						<h3 class="font-medium text-orange-900 dark:text-orange-100 text-sm">Offline Mode</h3>
						<p class="text-xs text-orange-700 dark:text-orange-300">
							You can still create incident reports. They'll be saved locally and submitted when
							connection returns.
						</p>
					</div>
					<CloudIcon class="h-5 w-5 text-orange-400" />
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Current Shift Context -->
	{#if mockCurrentShift && mockCurrentShift.schedule_name && mockCurrentShift.location}
		<Card.Root class="bg-blue-50 dark:bg-blue-950/50 border-blue-200 dark:border-blue-800">
			<Card.Content class="p-4">
				<div class="flex items-center gap-3 mb-2">
					<ClockIcon class="h-5 w-5 text-blue-600 dark:text-blue-400" />
					<div>
						<h3 class="font-medium text-slate-900 dark:text-slate-100">Current Shift</h3>
						<p class="text-sm text-slate-600 dark:text-slate-400">
							{mockCurrentShift.schedule_name} • {mockCurrentShift.location}
						</p>
					</div>
				</div>
				<div class="text-xs text-slate-500 dark:text-slate-400">
					Report time: {getCurrentTime()}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Report Form -->
	<Card.Root>
		<Card.Header>
			<Card.Title class="flex items-center gap-2 text-base">
				Incident Details
				{#if !isOnline}
					<Badge variant="outline" class="text-xs bg-orange-50 border-orange-200 text-orange-700">
						<WifiOffIcon class="h-3 w-3 mr-1" />
						Offline
					</Badge>
				{/if}
			</Card.Title>
		</Card.Header>
		<Card.Content class="space-y-6">
			<!-- Severity Selection -->
			<div class="space-y-3">
				<Label class="text-base font-medium">Incident Severity *</Label>
				<div class="space-y-2">
					{#each severityOptions as severity (severity.value)}
						{@const IconComponent = severity.icon}
						<button
							type="button"
							class="w-full p-4 rounded-lg border-2 text-left transition-all
								{selectedSeverity === severity.value
								? severity.color
								: severity.value === '0'
									? 'border-slate-200 dark:border-slate-700 hover:border-green-300 dark:hover:border-green-600'
									: severity.value === '1'
										? 'border-slate-200 dark:border-slate-700 hover:border-orange-300 dark:hover:border-orange-600'
										: 'border-slate-200 dark:border-slate-700 hover:border-red-300 dark:hover:border-red-600'}"
							onclick={() => (selectedSeverity = severity.value)}
						>
							<div class="flex items-start gap-3">
								<IconComponent
									class="h-5 w-5 mt-0.5 {selectedSeverity === severity.value
										? ''
										: 'text-slate-400'}"
								/>
								<div class="flex-1">
									<div class="font-medium text-sm">{severity.label}</div>
									<div class="text-xs text-slate-600 dark:text-slate-400 mt-1">
										{severity.description}
									</div>
								</div>
								{#if selectedSeverity === severity.value}
									<CheckCircleIcon class="h-5 w-5 text-current" />
								{/if}
							</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- Message Input -->
			<div class="space-y-2">
				<Label for="message" class="text-base font-medium">Incident Description *</Label>
				<Textarea
					id="message"
					bind:value={reportMessage}
					placeholder="Describe what happened, where, and any relevant details..."
					rows={6}
					class="resize-none"
				/>
				<div class="flex justify-between text-xs text-slate-500 dark:text-slate-400">
					<span>Be specific about location, time, and circumstances</span>
					<span>{reportMessage.length}/1000</span>
				</div>
			</div>

			<!-- GPS Location Capture -->
			<div class="space-y-2">
				<Label class="text-base font-medium">Location Information (Optional)</Label>
				<GPSCapture
					autoCapture={false}
					onLocationCaptured={handleLocationCaptured}
					onError={handleLocationError}
					className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-3"
				/>
				<p class="text-xs text-slate-500 dark:text-slate-400">
					Location data helps emergency services and improves incident response. If GPS fails, you
					can enter coordinates manually or submit without location data.
					{#if !isOnline}
						<span class="text-orange-600"> GPS works offline.</span>
					{/if}
				</p>
			</div>

			<!-- Submit Button -->
			<Button
				onclick={handleSubmit}
				disabled={isSubmitting || !selectedSeverity || !reportMessage.trim()}
				class="w-full"
				size="lg"
			>
				{#if isSubmitting}
					{#if isOnline}
						Submitting...
					{:else}
						Saving offline...
					{/if}
				{:else}
					<SendIcon class="h-4 w-4 mr-2" />
					{#if isOnline}
						Submit Report
					{:else}
						Save Report (Offline)
					{/if}
				{/if}
			</Button>

			{#if !isOnline}
				<p class="text-xs text-center text-orange-600">
					Reports created offline will be automatically submitted when you reconnect to the
					internet.
				</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<!-- Emergency Contacts -->
	<EmergencyContacts />

	<!-- Emergency Notice -->
	<Card.Root class="border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950/50">
		<Card.Content class="p-4">
			<div class="flex items-start gap-3">
				<AlertTriangleIcon class="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
				<div class="flex-1">
					<h3 class="font-medium text-red-900 dark:text-red-100 text-sm">Emergency Situations</h3>
					<p class="text-xs text-red-700 dark:text-red-300 mt-1">
						For immediate threats or emergencies requiring police, medical, or fire response, use
						the Emergency button in the header to call 999 directly rather than submitting a report.
						{#if !isOnline}
							<strong>Emergency calling works offline.</strong>
						{/if}
					</p>
				</div>
			</div>
		</Card.Content>
	</Card.Root>
</div>
