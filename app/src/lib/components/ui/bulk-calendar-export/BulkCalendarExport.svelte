<!--
Bulk Calendar Export Component
Allows users to export all their shifts to calendar applications
-->
<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import { toast } from 'svelte-sonner';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import type { UserBooking } from '$lib/services/api/user';
	import { downloadAllShiftsICS, getCalendarSummary } from '$lib/utils/calendar';

	let {
		bookings,
		variant = 'default',
		size = 'default',
		class: className = ''
	}: {
		bookings: UserBooking[];
		variant?: 'default' | 'outline' | 'secondary' | 'ghost';
		size?: 'sm' | 'default' | 'lg';
		class?: string;
	} = $props();

	// Get calendar summary statistics
	const summary = $derived(getCalendarSummary(bookings));

	// Get upcoming bookings for calendar integration
	const upcomingBookings = $derived(
		bookings.filter((booking) => {
			const shiftStart = new Date(booking.shift_start);
			const now = new Date();
			return shiftStart > now;
		})
	);

	// Check if there are any shifts to export
	const hasShifts = $derived(upcomingBookings.length > 0);

	// Handle bulk download
	function handleBulkDownload() {
		if (!hasShifts) {
			toast.error('No upcoming shifts to export');
			return;
		}

		try {
			downloadAllShiftsICS(upcomingBookings);
			toast.success(`Downloaded ${upcomingBookings.length} shifts to calendar`);
		} catch (error) {
			console.error('Error downloading calendar file:', error);
			toast.error('Failed to download calendar file');
		}
	}

	// Handle provider selection for multiple events
	function handleProviderSelect(providerName: string) {
		if (!hasShifts) {
			toast.error('No upcoming shifts to export');
			return;
		}

		try {
			// For bulk export, we'll download an ICS file that can be imported
			// into any calendar application
			downloadAllShiftsICS(upcomingBookings);
			toast.success(`Downloaded ${upcomingBookings.length} shifts for ${providerName}`);
			toast.info(`Import the downloaded .ics file into ${providerName}`);
		} catch (error) {
			console.error('Error downloading calendar file:', error);
			toast.error(`Failed to export shifts for ${providerName}`);
		}
	}

	// Get next shift for preview
	const nextShift = $derived(upcomingBookings.length > 0 ? upcomingBookings[0] : null);
</script>

{#if hasShifts}
	<Card.Root class={className}>
		<Card.Header>
			<Card.Title class="flex items-center gap-2">
				<CalendarIcon class="h-5 w-5" />
				Calendar Export
			</Card.Title>
			<Card.Description>Add all your upcoming shifts to your calendar app</Card.Description>
		</Card.Header>

		<Card.Content class="space-y-4">
			<!-- Summary Statistics -->
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="text-center">
					<div class="text-2xl font-bold text-primary">{summary.upcoming}</div>
					<div class="text-xs text-muted-foreground">Upcoming</div>
				</div>
				<div class="text-center">
					<div class="text-2xl font-bold text-orange-600">{summary.thisWeek}</div>
					<div class="text-xs text-muted-foreground">This Week</div>
				</div>
				<div class="text-center">
					<div class="text-2xl font-bold text-blue-600">{summary.thisMonth}</div>
					<div class="text-xs text-muted-foreground">This Month</div>
				</div>
				<div class="text-center">
					<div class="text-2xl font-bold text-green-600">{summary.total}</div>
					<div class="text-xs text-muted-foreground">Total</div>
				</div>
			</div>

			<Separator />

			<!-- Next Shift Preview -->
			{#if nextShift}
				<div class="border rounded-lg p-3 bg-accent/50">
					<div class="flex items-center gap-2 mb-2">
						<ClockIcon class="h-4 w-4 text-muted-foreground" />
						<span class="text-sm font-medium">Next Shift</span>
					</div>
					<div class="space-y-1">
						<div class="font-medium">{nextShift.schedule_name}</div>
						<div class="text-sm text-muted-foreground">
							{new Date(nextShift.shift_start).toLocaleDateString('en-GB', {
								weekday: 'long',
								month: 'long',
								day: 'numeric'
							})}
						</div>
						<div class="text-sm text-muted-foreground">
							{new Date(nextShift.shift_start).toLocaleTimeString('en-GB', {
								hour: '2-digit',
								minute: '2-digit'
							})} - {new Date(nextShift.shift_end).toLocaleTimeString('en-GB', {
								hour: '2-digit',
								minute: '2-digit'
							})}
						</div>
						{#if nextShift.buddy_name}
							<div class="text-sm text-muted-foreground">
								👥 Buddy: {nextShift.buddy_name}
							</div>
						{/if}
					</div>
				</div>
			{/if}

			<!-- Export Options -->
			<div class="flex flex-col sm:flex-row gap-2">
				<!-- Direct Download Button -->
				<Button
					{variant}
					{size}
					class="flex-1"
					onclick={handleBulkDownload}
					aria-label="Download all shifts as calendar file"
				>
					<DownloadIcon class="h-4 w-4 mr-2" />
					Download All Shifts
				</Button>

				<!-- Calendar Provider Dropdown -->
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						<Button variant="outline" {size} aria-label="Choose calendar app">
							<CalendarIcon class="h-4 w-4 mr-2" />
							Choose App
						</Button>
					</DropdownMenu.Trigger>

					<DropdownMenu.Content align="end" class="w-48">
						<DropdownMenu.Group>
							<DropdownMenu.Label>Calendar Apps</DropdownMenu.Label>
							<DropdownMenu.Separator />

							<DropdownMenu.Item
								class="flex items-center gap-2 cursor-pointer"
								onclick={() => handleProviderSelect('Google Calendar')}
							>
								<span class="text-base">📅</span>
								<span>Google Calendar</span>
							</DropdownMenu.Item>

							<DropdownMenu.Item
								class="flex items-center gap-2 cursor-pointer"
								onclick={() => handleProviderSelect('Outlook')}
							>
								<span class="text-base">📘</span>
								<span>Outlook</span>
							</DropdownMenu.Item>

							<DropdownMenu.Item
								class="flex items-center gap-2 cursor-pointer"
								onclick={() => handleProviderSelect('Apple Calendar')}
							>
								<span class="text-base">🍎</span>
								<span>Apple Calendar</span>
							</DropdownMenu.Item>

							<DropdownMenu.Item
								class="flex items-center gap-2 cursor-pointer"
								onclick={() => handleProviderSelect('Yahoo Calendar')}
							>
								<span class="text-base">🟣</span>
								<span>Yahoo Calendar</span>
							</DropdownMenu.Item>

							<DropdownMenu.Separator />

							<DropdownMenu.Item
								class="flex items-center gap-2 cursor-pointer"
								onclick={handleBulkDownload}
							>
								<DownloadIcon class="h-4 w-4" />
								<span>Download .ics file</span>
							</DropdownMenu.Item>
						</DropdownMenu.Group>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</div>

			<!-- Help Text -->
			<div class="text-xs text-muted-foreground bg-accent/30 p-3 rounded-lg">
				<div class="font-medium mb-1">📋 How to use:</div>
				<ul class="space-y-1">
					<li>
						• <strong>Download All Shifts:</strong> Gets a .ics file with all your upcoming shifts
					</li>
					<li>
						• <strong>Choose App:</strong> Download and import into your preferred calendar app
					</li>
					<li>• <strong>Calendar Reminders:</strong> Get notified 1 hour before each shift</li>
					<li>
						• <strong>Buddy Info:</strong> Your buddy's name is included in the calendar event
					</li>
				</ul>
			</div>
		</Card.Content>
	</Card.Root>
{:else}
	<Card.Root class={className}>
		<Card.Content class="text-center py-8">
			<CalendarIcon class="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
			<h3 class="text-lg font-semibold mb-2">No Upcoming Shifts</h3>
			<p class="text-muted-foreground">
				{#if summary.total > 0}
					You have {summary.total} total shifts, but no upcoming ones to export.
				{:else}
					You haven't booked any shifts yet. Check the available shifts to get started!
				{/if}
			</p>
		</Card.Content>
	</Card.Root>
{/if}
