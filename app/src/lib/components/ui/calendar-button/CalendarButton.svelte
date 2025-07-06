<!--
Calendar Button Component
Provides calendar integration options for shift bookings
-->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from 'svelte-sonner';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
	import type { UserBooking } from '$lib/services/api/user';
	import {
		bookingToCalendarEvent,
		downloadShiftICS,
		getCalendarProviders,
		openCalendarProvider,
		canAddToCalendar
	} from '$lib/utils/calendar';

	let {
		booking,
		variant = 'default',
		size = 'sm',
		showDropdown = true,
		disabled = false,
		class: className = ''
	}: {
		booking: UserBooking;
		variant?: 'default' | 'outline' | 'secondary' | 'ghost';
		size?: 'sm' | 'default' | 'lg';
		showDropdown?: boolean;
		disabled?: boolean;
		class?: string;
	} = $props();

	// Convert booking to calendar event
	const calendarEvent = $derived(bookingToCalendarEvent(booking));

	// Get calendar providers
	const providers = $derived(getCalendarProviders(calendarEvent));

	// Check if this booking can be added to calendar
	const canAdd = $derived(canAddToCalendar(booking));

	// Handle direct download without dropdown
	function handleDirectDownload() {
		if (!canAdd) {
			toast.error('This shift has already passed');
			return;
		}

		try {
			downloadShiftICS(booking);
			toast.success('Calendar file downloaded');
		} catch (error) {
			console.error('Error downloading calendar file:', error);
			toast.error('Failed to download calendar file');
		}
	}

	// Handle provider selection
	function handleProviderSelect(providerUrl: string, providerName: string) {
		if (!canAdd) {
			toast.error('This shift has already passed');
			return;
		}

		try {
			openCalendarProvider(providerUrl);
			toast.success(`Opening ${providerName}`);
		} catch (error) {
			console.error('Error opening calendar provider:', error);
			toast.error(`Failed to open ${providerName}`);
		}
	}

	// Handle direct ICS download
	function handleDownloadICS() {
		if (!canAdd) {
			toast.error('This shift has already passed');
			return;
		}

		try {
			downloadShiftICS(booking);
			toast.success('Calendar file downloaded');
		} catch (error) {
			console.error('Error downloading calendar file:', error);
			toast.error('Failed to download calendar file');
		}
	}
</script>

{#if showDropdown}
	<DropdownMenu.Root>
		<DropdownMenu.Trigger>
			<Button {variant} {size} {disabled} class={className} aria-label="Add shift to calendar">
				<CalendarIcon class="h-4 w-4" />
				{#if size !== 'sm'}
					<span class="ml-2">Add to Calendar</span>
				{/if}
			</Button>
		</DropdownMenu.Trigger>

		<DropdownMenu.Content align="end" class="w-56">
			<DropdownMenu.Group>
				<DropdownMenu.Label class="flex items-center gap-2">
					<CalendarIcon class="h-4 w-4" />
					Add to Calendar
				</DropdownMenu.Label>
				<DropdownMenu.Separator />

				{#if canAdd}
					<!-- Calendar Providers -->
					{#each providers as provider (provider.name)}
						<DropdownMenu.Item
							class="flex items-center gap-2 cursor-pointer"
							onclick={() => handleProviderSelect(provider.url, provider.name)}
						>
							<span class="text-base">{provider.icon}</span>
							<span>{provider.name}</span>
							<ExternalLinkIcon class="h-3 w-3 ml-auto text-muted-foreground" />
						</DropdownMenu.Item>
					{/each}

					<DropdownMenu.Separator />

					<!-- Download Options -->
					<DropdownMenu.Item
						class="flex items-center gap-2 cursor-pointer"
						onclick={handleDownloadICS}
					>
						<DownloadIcon class="h-4 w-4" />
						<span>Download .ics file</span>
					</DropdownMenu.Item>
				{:else}
					<DropdownMenu.Item disabled class="text-muted-foreground">
						<span>Shift has already passed</span>
					</DropdownMenu.Item>
				{/if}
			</DropdownMenu.Group>

			<DropdownMenu.Separator />

			<!-- Shift Info -->
			<div class="px-2 py-1.5 text-xs text-muted-foreground">
				<div class="font-medium">{booking.schedule_name}</div>
				<div class="flex items-center gap-1 mt-1">
					<CalendarIcon class="h-3 w-3" />
					{new Date(booking.shift_start).toLocaleDateString()}
				</div>
				<div class="flex items-center gap-1">
					<span>🕐</span>
					{new Date(booking.shift_start).toLocaleTimeString('en-GB', {
						hour: '2-digit',
						minute: '2-digit'
					})} - {new Date(booking.shift_end).toLocaleTimeString('en-GB', {
						hour: '2-digit',
						minute: '2-digit'
					})}
				</div>
				{#if booking.buddy_name}
					<div class="flex items-center gap-1 mt-1">
						<span>👥</span>
						<span class="text-xs">Buddy: {booking.buddy_name}</span>
					</div>
				{/if}
			</div>
		</DropdownMenu.Content>
	</DropdownMenu.Root>
{:else}
	<!-- Simple button without dropdown -->
	<Button
		{variant}
		{size}
		{disabled}
		class={className}
		onclick={handleDirectDownload}
		aria-label="Download calendar file"
	>
		<CalendarIcon class="h-4 w-4" />
		{#if size !== 'sm'}
			<span class="ml-2">Add to Calendar</span>
		{/if}
	</Button>
{/if}

<!-- Status indicator for past shifts -->
{#if !canAdd}
	<Badge variant="outline" class="text-xs ml-2">Past</Badge>
{/if}
