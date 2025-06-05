<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { themeState, themeActions, type ThemeMode } from '$lib/stores/themeStore';
	import { permissionUtils } from '$lib/stores/onboardingStore';
	import { toast } from 'svelte-sonner';
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CameraIcon from '@lucide/svelte/icons/camera';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import SettingsIcon from '@lucide/svelte/icons/settings';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	// Theme state
	let selectedTheme = $state<ThemeMode>('system');

	// Permission states
	let locationPermissionStatus = $state<string>('unknown');
	let notificationPermissionStatus = $state<string>('unknown');
	let cameraPermissionStatus = $state<string>('unknown');
	let isCheckingPermissions = $state(false);

	// Computed permission statuses for styling
	const locationGranted = $derived(locationPermissionStatus === 'granted');
	const notificationGranted = $derived(notificationPermissionStatus === 'granted');
	const cameraGranted = $derived(cameraPermissionStatus === 'granted');

	// Theme options
	const themeOptions = [
		{
			value: 'light' as ThemeMode,
			label: 'Light',
			icon: SunIcon
		},
		{
			value: 'dark' as ThemeMode,
			label: 'Dark',
			icon: MoonIcon
		},
		{
			value: 'system' as ThemeMode,
			label: 'System',
			icon: MonitorIcon
		}
	];

	// Initialize component state
	onMount(async () => {
		// Get current theme
		selectedTheme = $themeState.mode;

		// Check permissions
		isCheckingPermissions = true;
		try {
			locationPermissionStatus = await permissionUtils.checkLocationPermission();
			notificationPermissionStatus = await permissionUtils.checkNotificationPermission();
			cameraPermissionStatus = await permissionUtils.checkCameraPermission();
		} catch (_error) {
			console.warn('Failed to check permissions:', _error);
		} finally {
			isCheckingPermissions = false;
		}
	});

	// Handle theme change
	function handleThemeChange(theme: ThemeMode) {
		selectedTheme = theme;
		themeActions.setMode(theme);
		toast.success(`Theme changed to ${theme}`);
	}

	// Handle location permission request
	async function handleLocationPermission() {
		try {
			const result = await permissionUtils.requestLocationPermission();
			locationPermissionStatus = result;

			if (result === 'granted') {
				toast.success('Location access granted!');
			} else if (result === 'denied') {
				toast.warning('Location access denied. You can enable it in browser settings.');
			}
		} catch (_error) {
			toast.error('Failed to request location permission');
		}
	}

	// Handle notification permission request
	async function handleNotificationPermission() {
		try {
			const result = await permissionUtils.requestNotificationPermission();
			notificationPermissionStatus = result;

			if (result === 'granted') {
				toast.success('Notifications enabled!');
			} else if (result === 'denied') {
				toast.warning('Notifications disabled. You can enable them in browser settings.');
			}
		} catch (_error) {
			toast.error('Failed to request notification permission');
		}
	}

	// Handle camera permission (informational)
	async function handleCameraPermission() {
		try {
			await permissionUtils.requestCameraPermission();
			cameraPermissionStatus = 'prompt';
			toast.success('Camera ready for photo uploads!');
		} catch (_error) {
			toast.error('Failed to set up camera');
		}
	}

	// Get permission badge info
	function getPermissionBadge(status: string) {
		switch (status) {
			case 'granted':
				return {
					variant: 'default' as const,
					icon: CheckCircleIcon,
					text: 'Granted'
				};
			case 'denied':
				return {
					variant: 'destructive' as const,
					icon: XCircleIcon,
					text: 'Denied'
				};
			default:
				return {
					variant: 'secondary' as const,
					icon: XCircleIcon,
					text: 'Unknown'
				};
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-w-sm">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<SettingsIcon class="h-5 w-5" />
				Settings
			</Dialog.Title>
		</Dialog.Header>

		<div class="space-y-4">
			<!-- Theme Options -->
			<div>
				<h3 class="text-sm font-medium mb-2">Theme</h3>
				<div class="grid grid-cols-3 gap-2">
					{#each themeOptions as option (option.value)}
						{@const IconComponent = option.icon}
						{@const isSelected = selectedTheme === option.value}
						<button
							type="button"
							class="flex flex-col items-center gap-1 p-2 rounded-lg border-2 text-center transition-all
								{isSelected
								? 'border-primary bg-primary/5'
								: 'border-border hover:border-primary/30 opacity-75 hover:opacity-100'}"
							onclick={() => handleThemeChange(option.value)}
						>
							<IconComponent
								class="h-4 w-4 {isSelected ? 'text-primary' : 'text-muted-foreground'}"
							/>
							<span class="text-xs {isSelected ? 'font-medium' : 'text-muted-foreground'}"
								>{option.label}</span
							>
						</button>
					{/each}
				</div>
			</div>

			<!-- Permissions -->
			<div>
				<h3 class="text-sm font-medium mb-2">Permissions</h3>
				<div class="space-y-2">
					<!-- Location Permission -->
					<div
						class="flex items-center justify-between p-2 border rounded-lg
						{locationGranted
							? 'bg-green-50/50 border-green-200/50 dark:bg-green-950/20 dark:border-green-800/30 opacity-75'
							: 'border-border'}"
					>
						<div class="flex items-center gap-2 flex-1">
							<MapPinIcon
								class="h-4 w-4 {locationGranted
									? 'text-green-600 dark:text-green-400'
									: 'text-primary'}"
							/>
							<span class="text-sm {locationGranted ? 'text-green-800 dark:text-green-200' : ''}"
								>Location</span
							>
						</div>
						<div class="flex items-center gap-2">
							{#if locationPermissionStatus}
								{@const badgeInfo = getPermissionBadge(locationPermissionStatus)}
								{@const BadgeIcon = badgeInfo.icon}
								<Badge
									variant={badgeInfo.variant}
									class="text-xs {locationGranted ? 'opacity-75' : ''}"
								>
									<BadgeIcon class="h-3 w-3 mr-1" />
									{badgeInfo.text}
								</Badge>
							{/if}
							{#if !locationGranted}
								<Button
									size="sm"
									onclick={handleLocationPermission}
									disabled={isCheckingPermissions}
									class="text-xs px-2 py-1 h-6"
								>
									Enable
								</Button>
							{/if}
						</div>
					</div>

					<!-- Notification Permission -->
					<div
						class="flex items-center justify-between p-2 border rounded-lg
						{notificationGranted
							? 'bg-green-50/50 border-green-200/50 dark:bg-green-950/20 dark:border-green-800/30 opacity-75'
							: 'border-border'}"
					>
						<div class="flex items-center gap-2 flex-1">
							<BellIcon
								class="h-4 w-4 {notificationGranted
									? 'text-green-600 dark:text-green-400'
									: 'text-primary'}"
							/>
							<span
								class="text-sm {notificationGranted ? 'text-green-800 dark:text-green-200' : ''}"
								>Alerts</span
							>
						</div>
						<div class="flex items-center gap-2">
							{#if notificationPermissionStatus}
								{@const badgeInfo = getPermissionBadge(notificationPermissionStatus)}
								{@const BadgeIcon = badgeInfo.icon}
								<Badge
									variant={badgeInfo.variant}
									class="text-xs {notificationGranted ? 'opacity-75' : ''}"
								>
									<BadgeIcon class="h-3 w-3 mr-1" />
									{badgeInfo.text}
								</Badge>
							{/if}
							{#if !notificationGranted}
								<Button
									size="sm"
									onclick={handleNotificationPermission}
									disabled={isCheckingPermissions}
									class="text-xs px-2 py-1 h-6"
								>
									Enable
								</Button>
							{/if}
						</div>
					</div>

					<!-- Camera Permission -->
					<div
						class="flex items-center justify-between p-2 border rounded-lg
						{cameraGranted
							? 'bg-green-50/50 border-green-200/50 dark:bg-green-950/20 dark:border-green-800/30 opacity-75'
							: 'border-border'}"
					>
						<div class="flex items-center gap-2 flex-1">
							<CameraIcon
								class="h-4 w-4 {cameraGranted
									? 'text-green-600 dark:text-green-400'
									: 'text-primary'}"
							/>
							<span class="text-sm {cameraGranted ? 'text-green-800 dark:text-green-200' : ''}"
								>Camera</span
							>
						</div>
						<div class="flex items-center gap-2">
							{#if cameraPermissionStatus}
								{@const badgeInfo = getPermissionBadge(cameraPermissionStatus)}
								{@const BadgeIcon = badgeInfo.icon}
								<Badge
									variant={badgeInfo.variant}
									class="text-xs {cameraGranted ? 'opacity-75' : ''}"
								>
									<BadgeIcon class="h-3 w-3 mr-1" />
									{badgeInfo.text}
								</Badge>
							{/if}
							{#if !cameraGranted}
								<Button
									size="sm"
									onclick={handleCameraPermission}
									disabled={isCheckingPermissions}
									class="text-xs px-2 py-1 h-6"
								>
									Setup
								</Button>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>

		<Dialog.Footer class="pt-4">
			<Button onclick={() => (open = false)} class="w-full">Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
