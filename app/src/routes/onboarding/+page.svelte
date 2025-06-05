<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import {
		onboardingActions,
		permissionUtils,
		pwaUtils,
		pwaInstallPrompt
	} from '$lib/stores/onboardingStore';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { toast } from 'svelte-sonner';
	import MapPinIcon from 'lucide-svelte/icons/map-pin';
	import BellIcon from 'lucide-svelte/icons/bell';
	import CameraIcon from 'lucide-svelte/icons/camera';
	import SmartphoneIcon from 'lucide-svelte/icons/smartphone';
	import CheckCircleIcon from 'lucide-svelte/icons/check-circle';
	import XCircleIcon from 'lucide-svelte/icons/x-circle';
	import SkipForwardIcon from 'lucide-svelte/icons/skip-forward';
	import DownloadIcon from 'lucide-svelte/icons/download';
	import { Switch } from '$lib/components/ui/switch';

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
	let cameraPermissionStatus = $state<string>('unknown');
	let canInstallPWA = $state(false);
	let isPWAInstalled = $state(false);
	let notificationsEnabled = $state(false);

	const totalSteps = 2;

	// Initialize component state
	onMount(async () => {
		// Check current permission status
		locationPermissionStatus = await permissionUtils.checkLocationPermission();
		notificationPermissionStatus = permissionUtils.checkNotificationPermission();
		cameraPermissionStatus = await permissionUtils.checkCameraPermission();

		// Update store with current status
		onboardingActions.updateLocationPermission(
			locationPermissionStatus as 'granted' | 'denied' | 'prompt' | 'unknown'
		);
		onboardingActions.updateNotificationPermission(
			notificationPermissionStatus as 'granted' | 'denied' | 'default' | 'unknown'
		);
		onboardingActions.updateCameraPermission(
			cameraPermissionStatus as 'granted' | 'denied' | 'prompt' | 'unknown'
		);

		// Check PWA capabilities
		canInstallPWA = pwaUtils.canInstallPWA();
		isPWAInstalled = pwaUtils.isPWA();

		if (isPWAInstalled) {
			onboardingActions.markPWAInstalled();
		}

		notificationsEnabled = notificationPermissionStatus === 'granted';
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
		} catch (_error) {
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
		} catch (_error) {
			toast.error('Failed to request notification permission');
		} finally {
			isLoading = false;
		}
	}

	// Handle camera permission (informational)
	async function handleCameraPermission() {
		isLoading = true;
		try {
			await permissionUtils.requestCameraPermission();
			cameraPermissionStatus = 'prompt';
			toast.success('Camera ready! You can now attach photos to reports.');
		} catch (_error) {
			toast.error('Failed to set up camera');
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
		} catch (_error) {
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
			handleCompleteOnboardingStep();
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
			handleCompleteOnboardingStep();
			return;
		}
		nextStep();
	}

	// Handle completion when reaching the end of onboarding
	function handleCompleteOnboardingStep() {
		completeOnboarding();
	}

	// Complete onboarding
	function completeOnboarding() {
		onboardingActions.completeOnboarding();
		toast.success('Welcome to Night Owls! Setup complete.');
		goto('/admin', { replaceState: true });
	}

	const progress = $derived(Math.round((currentStep / totalSteps) * 100));

	function toggleNotifications() {
		notificationsEnabled = !notificationsEnabled;
		if (notificationsEnabled) {
			handleNotificationPermission();
		} else {
			notificationPermissionStatus = 'denied';
			onboardingActions.updateNotificationPermission('denied');
		}
	}
</script>

<svelte:head>
	<title>Welcome - Night Owls Setup</title>
</svelte:head>

<div class="bg-gradient-to-br from-primary/5 via-background to-secondary/5 flex-1 flex flex-col">
	<div class="container mx-auto px-4 py-6 lg:py-12 flex-1 flex flex-col max-w-4xl">
		<!-- Header -->
		<header class="text-center mb-8 lg:mb-12">
			<div class="flex items-center justify-center mb-4 lg:mb-6">
				<div
					class="h-12 w-12 lg:h-16 lg:w-16 bg-primary rounded-xl flex items-center justify-center mr-3"
				>
					<img src="/logo.png" alt="Night Owls" class="h-8 w-8 lg:h-12 lg:w-12" />
				</div>
				<h1 class="text-2xl lg:text-4xl xl:text-5xl font-bold">Welcome to Night Owls</h1>
			</div>
			<p class="text-base lg:text-lg text-muted-foreground max-w-2xl mx-auto leading-relaxed">
				Hi {$currentUser?.name || 'there'}! Let's set up your account to get the best experience
				keeping our community safe.
			</p>
		</header>

		<!-- Progress indicator -->
		<div class="max-w-2xl mx-auto mb-8 lg:mb-12 w-full">
			<div
				class="flex items-center justify-between text-sm lg:text-base text-muted-foreground mb-3"
			>
				<span>Step {currentStep} of {totalSteps}</span>
				<span>{progress}% complete</span>
			</div>
			<div class="w-full bg-secondary rounded-full h-2.5 lg:h-3">
				<div
					class="bg-primary h-2.5 lg:h-3 rounded-full transition-all duration-500"
					style="width: {progress}%"
				></div>
			</div>
		</div>

		<!-- Onboarding Steps -->
		<div class="flex-1 flex flex-col justify-center max-w-3xl mx-auto w-full">
			{#if currentStep === 1}
				<!-- Step 1: Permissions -->
				<Card.Root class="shadow-lg">
					<Card.Header class="pb-4 lg:pb-6">
						<Card.Title class="flex items-center gap-3 text-xl lg:text-2xl">
							<MapPinIcon class="h-6 w-6 lg:h-7 lg:w-7" />
							Permissions
						</Card.Title>
						<Card.Description class="text-base lg:text-lg">
							Enable these to get the full Night Owls experience
						</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-6">
						<!-- Location Permission -->
						<div class="flex items-start justify-between p-4 lg:p-6 border rounded-xl bg-card/50">
							<div class="flex items-start gap-4 flex-1 min-w-0">
								<div class="p-2 bg-primary/10 rounded-lg">
									<MapPinIcon class="h-5 w-5 lg:h-6 lg:w-6 text-primary" />
								</div>
								<div class="min-w-0 flex-1">
									<h3 class="font-semibold text-base lg:text-lg">Location Access</h3>
									<p class="text-sm lg:text-base text-muted-foreground mt-1">
										Helps with accurate incident reporting and location-based features
									</p>
								</div>
							</div>
							<div class="flex items-center gap-3 flex-shrink-0 ml-4">
								{#if locationPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-2 px-3 py-1">
										<CheckCircleIcon class="h-4 w-4" />
										Granted
									</Badge>
								{:else if locationPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-2 px-3 py-1">
										<XCircleIcon class="h-4 w-4" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-2 px-3 py-1">
										<XCircleIcon class="h-4 w-4" />
										Not Set
									</Badge>
								{/if}
								{#if locationPermissionStatus !== 'granted'}
									<Button
										onclick={handleLocationPermission}
										disabled={isLoading}
										size="sm"
										class="px-4"
									>
										Enable
									</Button>
								{/if}
							</div>
						</div>

						<!-- Notification Permission -->
						<div class="flex items-start justify-between p-4 lg:p-6 border rounded-xl bg-card/50">
							<div class="flex items-start gap-4 flex-1 min-w-0">
								<div class="p-2 bg-primary/10 rounded-lg">
									<BellIcon class="h-5 w-5 lg:h-6 lg:w-6 text-primary" />
								</div>
								<div class="min-w-0 flex-1">
									<h3 class="font-semibold text-base lg:text-lg">Alerts</h3>
									<p class="text-sm lg:text-base text-muted-foreground mb-4">
										Stay updated with important shift information and community updates.
									</p>

									<div
										class="flex items-center space-x-3 p-3 rounded-lg border border-border bg-card"
									>
										<div class="flex-shrink-0">
											<BellIcon class="h-5 w-5 text-primary" />
										</div>
										<div class="flex-1 min-w-0">
											<div class="flex items-center justify-between">
												<span class="text-sm lg:text-base">Push alerts</span>
												<Switch
													id="notifications"
													bind:checked={notificationsEnabled}
													onCheckedChange={toggleNotifications}
												/>
											</div>
											<p class="text-xs lg:text-sm text-muted-foreground mt-1">
												Get notified about shift reminders and important announcements
											</p>
										</div>
									</div>
								</div>
							</div>
							<div class="flex items-center gap-3 flex-shrink-0 ml-4">
								{#if notificationPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-2 px-3 py-1">
										<CheckCircleIcon class="h-4 w-4" />
										Granted
									</Badge>
								{:else if notificationPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-2 px-3 py-1">
										<XCircleIcon class="h-4 w-4" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-2 px-3 py-1">
										<XCircleIcon class="h-4 w-4" />
										Not Set
									</Badge>
								{/if}
								{#if notificationPermissionStatus !== 'granted'}
									<Button
										onclick={handleNotificationPermission}
										disabled={isLoading}
										size="sm"
										class="px-4"
									>
										Enable
									</Button>
								{/if}
							</div>
						</div>

						<!-- Camera Permission -->
						<div class="flex items-start justify-between p-4 lg:p-6 border rounded-xl bg-card/50">
							<div class="flex items-start gap-4 flex-1 min-w-0">
								<div class="p-2 bg-primary/10 rounded-lg">
									<CameraIcon class="h-5 w-5 lg:h-6 lg:w-6 text-primary" />
								</div>
								<div class="min-w-0 flex-1">
									<h3 class="font-semibold text-base lg:text-lg">Camera Access</h3>
									<p class="text-sm lg:text-base text-muted-foreground mt-1">
										Attach photos to incident reports for visual evidence
									</p>
								</div>
							</div>
							<div class="flex items-center gap-3 flex-shrink-0 ml-4">
								{#if cameraPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-2 px-3 py-1">
										<CheckCircleIcon class="h-4 w-4" />
										Available
									</Badge>
								{:else if cameraPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-2 px-3 py-1">
										<XCircleIcon class="h-4 w-4" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-2 px-3 py-1">
										<CameraIcon class="h-4 w-4" />
										Ready
									</Badge>
								{/if}
								{#if cameraPermissionStatus !== 'granted'}
									<Button
										onclick={handleCameraPermission}
										disabled={isLoading}
										size="sm"
										class="px-4"
									>
										Setup
									</Button>
								{/if}
							</div>
						</div>

						<!-- Actions -->
						<div class="flex flex-col sm:flex-row justify-between gap-4 pt-6">
							<Button variant="outline" onclick={skipStep} size="lg" class="px-8">
								<SkipForwardIcon class="h-4 w-4 mr-2" />
								Skip for now
							</Button>
							<Button onclick={nextStep} size="lg" class="px-8">Continue</Button>
						</div>
					</Card.Content>
				</Card.Root>
			{:else if currentStep === 2}
				<!-- Step 2: PWA Installation -->
				<Card.Root class="shadow-lg">
					<Card.Header class="pb-4 lg:pb-6">
						<Card.Title class="flex items-center gap-3 text-xl lg:text-2xl">
							<SmartphoneIcon class="h-6 w-6 lg:h-7 lg:w-7" />
							Install App for Offline Use
						</Card.Title>
						<Card.Description class="text-base lg:text-lg">
							Install Night Owls as an app for better performance and offline access
						</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-8">
						<div class="text-center py-8 lg:py-12">
							{#if isPWAInstalled}
								<div class="p-4 bg-green-50 rounded-2xl w-fit mx-auto mb-6">
									<CheckCircleIcon class="h-16 w-16 lg:h-20 lg:w-20 text-green-500 mx-auto" />
								</div>
								<h3 class="text-2xl lg:text-3xl font-semibold mb-4">App Already Installed!</h3>
								<p class="text-base lg:text-lg text-muted-foreground max-w-lg mx-auto">
									Night Owls is running as an installed app. You'll have offline access and better
									performance.
								</p>
							{:else if canInstallPWA && $pwaInstallPrompt}
								<div class="p-4 bg-primary/10 rounded-2xl w-fit mx-auto mb-6">
									<DownloadIcon class="h-16 w-16 lg:h-20 lg:w-20 text-primary mx-auto" />
								</div>
								<h3 class="text-2xl lg:text-3xl font-semibold mb-4">Install Night Owls App</h3>
								<p class="text-base lg:text-lg text-muted-foreground mb-8">
									Installing the app gives you:
								</p>

								<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 max-w-2xl mx-auto mb-8">
									<div class="flex items-center gap-3 text-left p-4 rounded-lg bg-muted/30">
										<CheckCircleIcon class="h-5 w-5 text-green-500 flex-shrink-0" />
										<span class="text-sm lg:text-base">Offline access to key features</span>
									</div>
									<div class="flex items-center gap-3 text-left p-4 rounded-lg bg-muted/30">
										<CheckCircleIcon class="h-5 w-5 text-green-500 flex-shrink-0" />
										<span class="text-sm lg:text-base">Faster loading and performance</span>
									</div>
									<div class="flex items-center gap-3 text-left p-4 rounded-lg bg-muted/30">
										<CheckCircleIcon class="h-5 w-5 text-green-500 flex-shrink-0" />
										<span class="text-sm lg:text-base">Home screen icon</span>
									</div>
									<div class="flex items-center gap-3 text-left p-4 rounded-lg bg-muted/30">
										<CheckCircleIcon class="h-5 w-5 text-green-500 flex-shrink-0" />
										<span class="text-sm lg:text-base">Push alerts</span>
									</div>
								</div>

								<Button
									onclick={handlePWAInstall}
									disabled={isLoading}
									size="lg"
									class="px-8 py-3 text-lg"
								>
									<DownloadIcon class="h-5 w-5 mr-2" />
									Install App
								</Button>
							{:else}
								<div class="p-4 bg-muted/20 rounded-2xl w-fit mx-auto mb-6">
									<SmartphoneIcon class="h-16 w-16 lg:h-20 lg:w-20 text-muted-foreground mx-auto" />
								</div>
								<h3 class="text-2xl lg:text-3xl font-semibold mb-4">
									App Installation Not Available
								</h3>
								<p class="text-base lg:text-lg text-muted-foreground max-w-lg mx-auto">
									Your browser doesn't support app installation, but you can still use Night Owls
									normally through your browser.
								</p>
							{/if}
						</div>

						<!-- Actions -->
						<div class="flex flex-col sm:flex-row justify-between gap-4 pt-6">
							<Button variant="outline" onclick={skipStep} size="lg" class="px-8">
								<SkipForwardIcon class="h-4 w-4 mr-2" />
								Skip for now
							</Button>
							<Button onclick={nextStep} size="lg" class="px-8">Complete Setup</Button>
						</div>
					</Card.Content>
				</Card.Root>
			{/if}
		</div>
	</div>
</div>
