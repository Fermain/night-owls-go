<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as Card from '$lib/components/ui/card';
	import { themeState, themeActions, type ThemeMode } from '$lib/stores/themeStore';
	import { permissionUtils } from '$lib/stores/onboardingStore';
	import { toast } from 'svelte-sonner';
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import BellIcon from '@lucide/svelte/icons/bell';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import SettingsIcon from '@lucide/svelte/icons/settings';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	// Theme state
	let selectedTheme = $state<ThemeMode>('system');

	// Permission states
	let locationPermissionStatus = $state<string>('unknown');
	let notificationPermissionStatus = $state<string>('unknown');
	let isCheckingPermissions = $state(false);

	// Theme options
	const themeOptions = [
		{
			value: 'light' as ThemeMode,
			label: 'Light',
			description: 'Light theme always',
			icon: SunIcon
		},
		{
			value: 'dark' as ThemeMode,
			label: 'Dark',
			description: 'Dark theme always',
			icon: MoonIcon
		},
		{
			value: 'system' as ThemeMode,
			label: 'System',
			description: 'Follows system preference',
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
			notificationPermissionStatus = permissionUtils.checkNotificationPermission();
		} catch (error) {
			console.warn('Failed to check permissions:', error);
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
		} catch (error) {
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
		} catch (error) {
			toast.error('Failed to request notification permission');
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
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<SettingsIcon class="h-5 w-5" />
				Settings
			</Dialog.Title> 
		</Dialog.Header>

		<div class="space-y-6">
			<div class="grid grid-cols-1 gap-3">
        {#each themeOptions as option (option.value)}
          {@const IconComponent = option.icon}
          <button
            type="button"
            class="flex items-start gap-3 p-3 rounded-lg border-2 text-left transition-all
              {selectedTheme === option.value
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/50'}"
            onclick={() => handleThemeChange(option.value)}
          >
            <IconComponent class="h-5 w-5 mt-0.5 text-primary" />
            <div class="flex-1">
              <div class="font-medium text-sm">{option.label}</div>
              <div class="text-xs text-muted-foreground mt-0.5">
                {option.description}
              </div>
            </div>
            {#if selectedTheme === option.value}
              <CheckCircleIcon class="h-5 w-5 text-primary" />
            {/if}
          </button>
        {/each}
      </div>

			<div class="space-y-4">
        <!-- Location Permission -->
        <div class="flex items-start justify-between p-3 border rounded-lg">
          <div class="flex items-start gap-3 flex-1 min-w-0">
            <MapPinIcon class="h-4 w-4 text-primary mt-0.5 flex-shrink-0" />
            <div class="min-w-0 flex-1">
              <h3 class="font-medium text-sm">Location Access</h3>
              <p class="text-xs text-muted-foreground mt-0.5">
                Helps with incident reporting and emergency services
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            {#if locationPermissionStatus}
              {@const badgeInfo = getPermissionBadge(locationPermissionStatus)}
              {@const BadgeIcon = badgeInfo.icon}
              <Badge variant={badgeInfo.variant} class="flex items-center gap-1 text-xs">
                <BadgeIcon class="h-3 w-3" />
                {badgeInfo.text}
              </Badge>
            {/if}
            {#if locationPermissionStatus !== 'granted'}
              <Button
                size="sm"
                onclick={handleLocationPermission}
                disabled={isCheckingPermissions}
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
              <p class="text-xs text-muted-foreground mt-0.5">
                Receive important alerts and emergency broadcasts
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            {#if notificationPermissionStatus}
              {@const badgeInfo = getPermissionBadge(notificationPermissionStatus)}
              {@const BadgeIcon = badgeInfo.icon}
              <Badge variant={badgeInfo.variant} class="flex items-center gap-1 text-xs">
                <BadgeIcon class="h-3 w-3" />
                {badgeInfo.text}
              </Badge>
            {/if}
            {#if notificationPermissionStatus !== 'granted'}
              <Button
                size="sm"
                onclick={handleNotificationPermission}
                disabled={isCheckingPermissions}
                class="text-xs px-2 py-1"
              >
                Enable
              </Button>
            {/if}
          </div>
        </div>

        {#if locationPermissionStatus === 'denied' || notificationPermissionStatus === 'denied'}
          <div class="p-3 bg-muted rounded-lg">
            <p class="text-xs text-muted-foreground">
              <strong>Note:</strong> If permissions are denied, you can manually enable them in your browser's site settings.
            </p>
          </div>
        {/if}
      </div>
		</div>

		<Dialog.Footer>
			<Button onclick={() => (open = false)}>Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root> 