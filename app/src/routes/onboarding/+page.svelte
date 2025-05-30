<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
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

	const totalSteps = 3;

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
				toast.success('Notifications enabled! You\'ll receive important alerts.');
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
		}
		
		if (currentStep < totalSteps) {
			currentStep++;
		} else {
			completeOnboarding();
		}
	}

	// Skip current step
	function skipStep() {
		if (currentStep === 1) {
			toast.info('Permissions skipped. You can enable them later in settings.');
		} else if (currentStep === 2) {
			toast.info('App installation skipped. You can install it later.');
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

<div class="min-h-screen bg-gradient-to-br from-primary/5 via-background to-secondary/5">
	<div class="container mx-auto px-4 py-8">
		<!-- Header -->
		<div class="text-center mb-8">
			<div class="flex items-center justify-center mb-4">
				<div class="h-12 w-12 bg-primary rounded-lg flex items-center justify-center mr-3">
					<img src="/logo.png" alt="Night Owls" class="h-8 w-8" />
				</div>
				<h1 class="text-3xl font-bold">Welcome to Night Owls</h1>
			</div>
			<p class="text-muted-foreground max-w-2xl mx-auto">
				Hi {$currentUser?.name || 'there'}! Let's set up your account to get the best experience 
				keeping our community safe.
			</p>
		</div>

		<!-- Progress indicator -->
		<div class="max-w-2xl mx-auto mb-8">
			<div class="flex items-center justify-between text-sm text-muted-foreground mb-2">
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
		<div class="max-w-2xl mx-auto">
			{#if currentStep === 1}
				<!-- Step 1: Permissions -->
				<Card class="mb-6">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<MapPinIcon class="h-5 w-5" />
							Location & Notification Permissions
						</CardTitle>
						<CardDescription>
							Enable these permissions to get the full Night Owls experience
						</CardDescription>
					</CardHeader>
					<CardContent class="space-y-6">
						<!-- Location Permission -->
						<div class="flex items-center justify-between p-4 border rounded-lg">
							<div class="flex items-center gap-3">
								<MapPinIcon class="h-5 w-5 text-primary" />
								<div>
									<h3 class="font-medium">Location Access</h3>
									<p class="text-sm text-muted-foreground">
										Helps with accurate incident reporting
									</p>
								</div>
							</div>
							<div class="flex items-center gap-2">
								{#if locationPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-1">
										<CheckCircleIcon class="h-3 w-3" />
										Granted
									</Badge>
								{:else if locationPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-1">
										<XCircleIcon class="h-3 w-3" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-1">
										<XCircleIcon class="h-3 w-3" />
										Unknown
									</Badge>
								{/if}
								{#if locationPermissionStatus !== 'granted'}
									<Button 
										size="sm" 
										onclick={handleLocationPermission}
										disabled={isLoading}
									>
										Enable
									</Button>
								{/if}
							</div>
						</div>

						<!-- Notification Permission -->
						<div class="flex items-center justify-between p-4 border rounded-lg">
							<div class="flex items-center gap-3">
								<BellIcon class="h-5 w-5 text-primary" />
								<div>
									<h3 class="font-medium">Notifications</h3>
									<p class="text-sm text-muted-foreground">
										Receive important safety alerts and updates
									</p>
								</div>
							</div>
							<div class="flex items-center gap-2">
								{#if notificationPermissionStatus === 'granted'}
									<Badge variant="default" class="flex items-center gap-1">
										<CheckCircleIcon class="h-3 w-3" />
										Granted
									</Badge>
								{:else if notificationPermissionStatus === 'denied'}
									<Badge variant="destructive" class="flex items-center gap-1">
										<XCircleIcon class="h-3 w-3" />
										Denied
									</Badge>
								{:else}
									<Badge variant="secondary" class="flex items-center gap-1">
										<XCircleIcon class="h-3 w-3" />
										Unknown
									</Badge>
								{/if}
								{#if notificationPermissionStatus !== 'granted'}
									<Button 
										size="sm" 
										onclick={handleNotificationPermission}
										disabled={isLoading}
									>
										Enable
									</Button>
								{/if}
							</div>
						</div>

						<!-- Actions -->
						<div class="flex justify-between pt-4">
							<Button variant="outline" onclick={skipStep}>
								<SkipForwardIcon class="h-4 w-4 mr-2" />
								Skip for now
							</Button>
							<Button onclick={nextStep}>
								Continue
							</Button>
						</div>
					</CardContent>
				</Card>
			{:else if currentStep === 2}
				<!-- Step 2: PWA Installation -->
				<Card class="mb-6">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<SmartphoneIcon class="h-5 w-5" />
							Install App for Offline Use
						</CardTitle>
						<CardDescription>
							Install Night Owls as an app for better performance and offline access
						</CardDescription>
					</CardHeader>
					<CardContent class="space-y-6">
						<div class="text-center py-8">
							{#if isPWAInstalled}
								<CheckCircleIcon class="h-16 w-16 text-green-500 mx-auto mb-4" />
								<h3 class="text-lg font-medium mb-2">App Already Installed!</h3>
								<p class="text-muted-foreground">
									Night Owls is running as an installed app. You'll have offline access 
									and better performance.
								</p>
							{:else if canInstallPWA && $pwaInstallPrompt}
								<DownloadIcon class="h-16 w-16 text-primary mx-auto mb-4" />
								<h3 class="text-lg font-medium mb-2">Install Night Owls App</h3>
								<p class="text-muted-foreground mb-6">
									Installing the app gives you:
								</p>
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
								<Button 
									onclick={handlePWAInstall}
									disabled={isLoading}
									size="lg"
									class="mb-4"
								>
									<DownloadIcon class="h-4 w-4 mr-2" />
									Install App
								</Button>
							{:else}
								<SmartphoneIcon class="h-16 w-16 text-muted-foreground mx-auto mb-4" />
								<h3 class="text-lg font-medium mb-2">App Installation Not Available</h3>
								<p class="text-muted-foreground">
									Your browser doesn't support app installation, but you can still 
									use Night Owls normally through your browser.
								</p>
							{/if}
						</div>

						<!-- Actions -->
						<div class="flex justify-between pt-4">
							<Button variant="outline" onclick={skipStep}>
								<SkipForwardIcon class="h-4 w-4 mr-2" />
								Skip for now
							</Button>
							<Button onclick={nextStep}>
								Continue
							</Button>
						</div>
					</CardContent>
				</Card>
			{:else if currentStep === 3}
				<!-- Step 3: Completion -->
				<Card class="mb-6">
					<CardHeader>
						<CardTitle class="flex items-center gap-2">
							<CheckCircleIcon class="h-5 w-5 text-green-500" />
							You're All Set!
						</CardTitle>
						<CardDescription>
							Night Owls is ready to help keep our community safe
						</CardDescription>
					</CardHeader>
					<CardContent class="space-y-6">
						<div class="text-center py-8">
							<CheckCircleIcon class="h-16 w-16 text-green-500 mx-auto mb-4" />
							<h3 class="text-lg font-medium mb-2">Welcome to the Team!</h3>
							<p class="text-muted-foreground mb-6">
								You're now ready to coordinate with fellow Night Owls, report incidents, 
								and help keep Mount Moreland safe.
							</p>
							
							<div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
								<div class="p-4 bg-secondary/50 rounded-lg">
									<h4 class="font-medium mb-1">Report Incidents</h4>
									<p class="text-muted-foreground">
										Quickly report security issues or suspicious activity
									</p>
								</div>
								<div class="p-4 bg-secondary/50 rounded-lg">
									<h4 class="font-medium mb-1">Join Shifts</h4>
									<p class="text-muted-foreground">
										Sign up for patrol shifts and coordinate with your team
									</p>
								</div>
								<div class="p-4 bg-secondary/50 rounded-lg">
									<h4 class="font-medium mb-1">Stay Connected</h4>
									<p class="text-muted-foreground">
										Receive broadcasts and stay updated on community safety
									</p>
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex justify-center pt-4">
							<Button onclick={completeOnboarding} size="lg">
								Get Started
							</Button>
						</div>
					</CardContent>
				</Card>
			{/if}
		</div>
	</div>
</div> 