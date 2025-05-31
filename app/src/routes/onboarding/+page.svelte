<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import {
		onboardingState,
		onboardingActions,
		permissionUtils,
		pwaUtils,
		pwaInstallPrompt
	} from '$lib/stores/onboardingStore';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { toast } from 'svelte-sonner';
	import MapPinIcon from 'lucide-svelte/icons/map-pin';
	import BellIcon from 'lucide-svelte/icons/bell';
	import SmartphoneIcon from 'lucide-svelte/icons/smartphone';
	import CheckCircleIcon from 'lucide-svelte/icons/check-circle';
	import XCircleIcon from 'lucide-svelte/icons/x-circle';
	import SkipForwardIcon from 'lucide-svelte/icons/skip-forward';
	import DownloadIcon from 'lucide-svelte/icons/download';

	// Redirect to login if not authenticated
	$effect(() => {
		if (!$isAuthenticated) {
			goto('/login', { replaceState: true });
		}
	});

	let currentStep = $state(1);
	let isLoading = $state(false);
	let locationPermissionStatus = $state<string>('unknown');
	let notificationPermissionStatus = $state<string>('unknown');
	let canInstallPWA = $state(false);
	let isPWAInstalled = $state(false);

	const totalSteps = 2;

	// Initialize component state
	onMount(async () => {
		// Check current permission status
		locationPermissionStatus = await permissionUtils.checkLocationPermission();
		notificationPermissionStatus = permissionUtils.checkNotificationPermission();

		// Update store with current status
		onboardingActions.updateLocationPermission(locationPermissionStatus as any);
		onboardingActions.updateNotificationPermission(notificationPermissionStatus as any);

		// Check PWA capabilities
		canInstallPWA = pwaUtils.canInstallPWA();
		isPWAInstalled = pwaUtils.isPWA();

		if (isPWAInstalled) {
			onboardingActions.markPWAInstalled();
		}
	});

	// Handle location permission request
	async function handleLocationPermission() {
		isLoading = true;
		try {
			const result = await permissionUtils.requestLocationPermission();
			locationPermissionStatus = result;

			if (result === 'granted') {
				toast.success('Location access granted! This helps with incident reporting.');
			} else if (result === 'denied') {
				toast.warning('Location access denied. You can enable it later in settings.');
			}
		} catch (error) {
			toast.error('Failed to request location permission');
		} finally {
			isLoading = false;
		}
	}

	// Handle notification permission request
	async function handleNotificationPermission() {
		isLoading = true;
		try {
			const result = await permissionUtils.requestNotificationPermission();
			notificationPermissionStatus = result;

			if (result === 'granted') {
				toast.success("Notifications enabled! You'll receive important alerts.");
			} else if (result === 'denied') {
				toast.warning('Notifications disabled. You can enable them later in settings.');
			}
		} catch (error) {
			toast.error('Failed to request notification permission');
		} finally {
			isLoading = false;
		}
	}

	// Handle PWA installation
	async function handlePWAInstall() {
		isLoading = true;
		try {
			const success = await pwaUtils.showInstallPrompt($pwaInstallPrompt);
			if (success) {
				toast.success('App installed successfully!');
				isPWAInstalled = true;
			} else {
				toast.info('App installation cancelled');
			}
		} catch (error) {
			toast.error('Failed to install app');
		} finally {
			isLoading = false;
		}
	}

	// Navigate to next step
	function nextStep() {
		if (currentStep === 1) {
			onboardingActions.completePermissions();
		} else if (currentStep === 2) {
			onboardingActions.completePWAPrompt();
			completeOnboarding();
			return;
		}

		if (currentStep < totalSteps) {
			currentStep++;
		}
	}

	// Skip current step
	function skipStep() {
		if (currentStep === 1) {
			toast.info('Permissions skipped. You can enable them later in settings.');
		} else if (currentStep === 2) {
			toast.info('App installation skipped. You can install it later.');
			completeOnboarding();
			return;
		}
		nextStep();
	}

	// Complete onboarding
	function completeOnboarding() {
		onboardingActions.completeOnboarding();
		toast.success('Welcome to Night Owls! Setup complete.');
		goto('/admin', { replaceState: true });
	}

	const progress = $derived(Math.round((currentStep / totalSteps) * 100));
</script>

<svelte:head>
	<title>Welcome - Night Owls Setup</title>
</svelte:head>

<div
	class="min-h-screen bg-gradient-to-br from-primary/5 via-background to-secondary/5 flex flex-col"
>
	<div class="container mx-auto px-4 py-4 sm:py-6 flex-1 flex flex-col">
		<!-- Header -->
		<div class="text-center mb-4 sm:mb-6">
			<div class="flex items-center justify-center mb-3">
				<div
					class="h-8 w-8 sm:h-10 sm:w-10 bg-primary rounded-lg flex items-center justify-center mr-2"
				>
					<img src="/logo.png" alt="Night Owls" class="h-6 w-6 sm:h-8 sm:w-8" />
				</div>
				<h1 class="text-xl sm:text-2xl md:text-3xl font-bold">Welcome to Night Owls</h1>
			</div>
			<p class="text-sm sm:text-base text-muted-foreground max-w-xl mx-auto">
				Hi {$currentUser?.name || 'there'}! Let's set up your account to get the best experience
				keeping our community safe.
			</p>
		</div>

		<!-- Progress indicator -->
		<div class="max-w-xl mx-auto mb-4 sm:mb-6">
			<div class="flex items-center justify-between text-xs sm:text-sm text-muted-foreground mb-2">
				<span>Step {currentStep} of {totalSteps}</span>
				<span>{progress}% complete</span>
			</div>
			<div class="w-full bg-secondary rounded-full h-2">
				<div
					class="bg-primary h-2 rounded-full transition-all duration-300"
					style="width: {progress}%"
				></div>
			</div>
		</div>

		<!-- Onboarding Steps -->
		<div class="max-w-xl mx-auto flex-1 flex flex-col justify-center">
			{#if currentStep === 1}
				<!-- Step 1: Permissions -->
				<Card.Root class="mb-4 sm:mb-6">
					<Card.Header class="pb-3">
						<Card.Title class="flex items-center gap-2 text-lg">
							<MapPinIcon class="h-5 w-5" />
							Permissions
						</Card.Title>
						<Card.Description class="text-sm">
							Enable these to get the full experience
						</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-4">
						<!-- Location Permission -->
						<div class="flex items-start justify-between p-3 border rounded-lg">
							<div class="flex items-start gap-3 flex-1 min-w-0">
								<MapPinIcon class="h-4 w-4 text-primary mt-0.5 flex-shrink-0" />
								<div class="min-w-0 flex-1">
									<h3 class="font-medium text-sm">Location Access</h3>
									<p class="text-xs text-muted-foreground">Helps with incident reporting</p>
								</div>
							</div>
							<div class="flex items-center gap-2 flex-shrink-0">
								{#if locationPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-1 text-xs">
										<CheckCircleIcon class="h-3 w-3" />
										Granted
									</Badge>
								{:else if locationPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-1 text-xs">
										<XCircleIcon class="h-3 w-3" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-1 text-xs">
										<XCircleIcon class="h-3 w-3" />
										Unknown
									</Badge>
								{/if}
								{#if locationPermissionStatus !== 'granted'}
									<Button
										size="sm"
										onclick={handleLocationPermission}
										disabled={isLoading}
										class="text-xs px-2 py-1"
									>
										Enable
									</Button>
								{/if}
							</div>
						</div>

						<!-- Notification Permission -->
						<div class="flex items-start justify-between p-3 border rounded-lg">
							<div class="flex items-start gap-3 flex-1 min-w-0">
								<BellIcon class="h-4 w-4 text-primary mt-0.5 flex-shrink-0" />
								<div class="min-w-0 flex-1">
									<h3 class="font-medium text-sm">Notifications</h3>
									<p class="text-xs text-muted-foreground">Receive important alerts</p>
								</div>
							</div>
							<div class="flex items-center gap-2 flex-shrink-0">
								{#if notificationPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-1 text-xs">
										<CheckCircleIcon class="h-3 w-3" />
										Granted
									</Badge>
								{:else if notificationPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-1 text-xs">
										<XCircleIcon class="h-3 w-3" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-1 text-xs">
										<XCircleIcon class="h-3 w-3" />
										Unknown
									</Badge>
								{/if}
								{#if notificationPermissionStatus !== 'granted'}
									<Button
										size="sm"
										onclick={handleNotificationPermission}
										disabled={isLoading}
										class="text-xs px-2 py-1"
									>
										Enable
									</Button>
								{/if}
							</div>
						</div>

						<!-- Actions -->
						<div class="flex justify-between pt-3">
							<Button variant="outline" onclick={skipStep} size="sm">
								<SkipForwardIcon class="h-4 w-4 mr-2" />
								Skip
							</Button>
							<Button onclick={nextStep} size="sm">Continue</Button>
						</div>
					</Card.Content>
				</Card.Root>
			{:else if currentStep === 2}
				<!-- Step 2: PWA Installation -->
				<Card.Root class="mb-6">
					<Card.Header>
						<Card.Title class="flex items-center gap-2">
							<SmartphoneIcon class="h-5 w-5" />
							Install App for Offline Use
						</Card.Title>
						<Card.Description>
							Install Night Owls as an app for better performance and offline access
						</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-6">
						<div class="text-center py-8">
							{#if isPWAInstalled}
								<CheckCircleIcon class="h-16 w-16 text-green-500 mx-auto mb-4" />
								<h3 class="text-lg font-medium mb-2">App Already Installed!</h3>
								<p class="text-muted-foreground">
									Night Owls is running as an installed app. You'll have offline access and better
									performance.
								</p>
							{:else if canInstallPWA && $pwaInstallPrompt}
								<DownloadIcon class="h-16 w-16 text-primary mx-auto mb-4" />
								<h3 class="text-lg font-medium mb-2">Install Night Owls App</h3>
								<p class="text-muted-foreground mb-6">Installing the app gives you:</p>
								<ul class="text-left space-y-2 mb-6 max-w-sm mx-auto">
									<li class="flex items-center gap-2 text-sm">
										<CheckCircleIcon class="h-4 w-4 text-green-500" />
										Offline access to key features
									</li>
									<li class="flex items-center gap-2 text-sm">
										<CheckCircleIcon class="h-4 w-4 text-green-500" />
										Faster loading and performance
									</li>
									<li class="flex items-center gap-2 text-sm">
										<CheckCircleIcon class="h-4 w-4 text-green-500" />
										Home screen icon
									</li>
									<li class="flex items-center gap-2 text-sm">
										<CheckCircleIcon class="h-4 w-4 text-green-500" />
										Push notifications
									</li>
								</ul>
								<Button onclick={handlePWAInstall} disabled={isLoading} size="lg" class="mb-4">
									<DownloadIcon class="h-4 w-4 mr-2" />
									Install App
								</Button>
							{:else}
								<SmartphoneIcon class="h-16 w-16 text-muted-foreground mx-auto mb-4" />
								<h3 class="text-lg font-medium mb-2">App Installation Not Available</h3>
								<p class="text-muted-foreground">
									Your browser doesn't support app installation, but you can still use Night Owls
									normally through your browser.
								</p>
							{/if}
						</div>

						<!-- Actions -->
						<div class="flex justify-between pt-4">
							<Button variant="outline" onclick={skipStep}>
								<SkipForwardIcon class="h-4 w-4 mr-2" />
								Skip for now
							</Button>
							<Button onclick={nextStep}>Continue</Button>
						</div>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>
	</div>
</div>
