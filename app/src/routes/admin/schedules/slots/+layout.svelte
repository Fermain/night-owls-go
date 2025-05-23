<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CheckIcon from '@lucide/svelte/icons/check';
	import XIcon from '@lucide/svelte/icons/x';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { formatDistanceToNow } from 'date-fns';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { AdminShiftSlot } from '$lib/types';

	let searchTerm = $state('');

	// Filters state
	let dateRangeStart = $state<string | null>(null);
	let dateRangeEnd = $state<string | null>(null);
	let showOnlyAvailable = $state(false);
	let showOnlyFilled = $state(false);

	// Define navigation items for the shifts section
	const shiftsNavItems = [
		{
			title: 'Calendar Dashboard',
			url: '/admin/schedules/slots',
			icon: LayoutDashboardIcon
		}
	];

	// Get selected shift from URL
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));

	// Utility Functions
	function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
		if (!startTimeIso || !endTimeIso) return 'N/A';
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);
			
			const startFormatted = startDate.toLocaleString('en-ZA', {
				weekday: 'short',
				day: 'numeric',
				month: 'short',
				hour: '2-digit',
				minute: '2-digit',
				hour12: false
			});
			
			const endFormatted = endDate.toLocaleTimeString('en-ZA', {
				hour: '2-digit',
				minute: '2-digit',
				hour12: false
			});
			
			return `${startFormatted} - ${endFormatted}`;
		} catch {
			return 'Invalid Date Range';
		}
	}

	function formatRelativeTime(timeIso: string): string {
		if (!timeIso) return 'N/A';
		try {
			return formatDistanceToNow(new Date(timeIso), { addSuffix: true });
		} catch {
			return 'Invalid Date';
		}
	}

	// Data Fetching
	async function fetchShiftSlots(from?: string, to?: string) {
		try {
			const params = new URLSearchParams();
			if (from) params.append('from', from);
			if (to) params.append('to', to);

			const response = await authenticatedFetch(
				`/api/admin/schedules/all-slots?${params.toString()}`
			);
			if (!response.ok) {
				let errorMsg = `HTTP error ${response.status}`;
				try {
					const errorData = await response.json();
					errorMsg = errorData.message || errorData.error || errorMsg;
				} catch {
					/* ignore */
				}
				throw new Error(errorMsg);
			}
			return response.json() as Promise<AdminShiftSlot[]>;
		} catch (error) {
			console.error('Error fetching shift slots:', error);
			throw error;
		}
	}

	// Main query for shift slots with filtering
	const shiftsQuery = $derived.by(() => {
		let fromDate: string | undefined;
		let toDate: string | undefined;

		if (dateRangeStart && dateRangeEnd) {
			fromDate = new Date(dateRangeStart + 'T00:00:00Z').toISOString();
			toDate = new Date(dateRangeEnd + 'T23:59:59Z').toISOString();
			
			// Safety check to prevent invalid ranges
			if (new Date(fromDate) > new Date(toDate)) {
				const now = new Date();
				const futureDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
				fromDate = now.toISOString();
				toDate = futureDate.toISOString();
			}
		} else {
			// Default to next 30 days if no range selected
			const now = new Date();
			const futureDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
			fromDate = now.toISOString();
			toDate = futureDate.toISOString();
		}

		return createQuery<AdminShiftSlot[], Error>({
			queryKey: ['adminShiftSlots', fromDate, toDate, showOnlyAvailable, showOnlyFilled],
			queryFn: () => fetchShiftSlots(fromDate, toDate),
			staleTime: 1000 * 60 * 5
		});
	});

	// Filtered shifts for display
	const filteredShifts = $derived.by(() => {
		const shifts = $shiftsQuery.data ?? [];
		return shifts.filter((shift) => {
			if (showOnlyAvailable && shift.is_booked) return false;
			if (showOnlyFilled && !shift.is_booked) return false;
			return true;
		});
	});

	// Upcoming shifts (next 7 days) for sidebar
	const upcomingShifts = $derived.by(() => {
		const now = new Date();
		const nextWeek = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
		const upcoming = filteredShifts
			.filter((shift) => {
				const shiftDate = new Date(shift.start_time);
				return shiftDate >= now && shiftDate <= nextWeek;
			})
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
			.slice(0, 20); // Limit to 20 items for sidebar
		
		return upcoming;
	});

	// Event Handlers
	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;
	}

	function selectShift(shift: AdminShiftSlot) {
		goto(`/admin/schedules/slots?shiftStartTime=${encodeURIComponent(shift.start_time)}`);
	}

	function clearFilters() {
		dateRangeStart = null;
		dateRangeEnd = null;
		showOnlyAvailable = false;
		showOnlyFilled = false;
	}

	function goToDashboard() {
		goto('/admin/schedules/slots');
	}

	let { children } = $props();
</script>

{#snippet shiftsListContent()}
	<div class="flex flex-col h-full">
		<!-- Top navigation (Dashboard) -->
		{#each shiftsNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={$page.url.pathname === '/admin/schedules/slots' && !shiftStartTimeFromUrl}
				onclick={(event) => {
					event.preventDefault();
					goToDashboard();
				}}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Filters Section -->
		<div class="p-3 border-b space-y-3">
			<div class="flex items-center gap-2 mb-2">
				<FilterIcon class="h-4 w-4" />
				<Label class="text-sm font-medium">Filters</Label>
			</div>

			<div class="space-y-3">
				<div class="flex items-center space-x-2">
					<Switch id="available-only" bind:checked={showOnlyAvailable} />
					<Label for="available-only" class="text-sm cursor-pointer">Available only</Label>
				</div>
				<div class="flex items-center space-x-2">
					<Switch id="filled-only" bind:checked={showOnlyFilled} />
					<Label for="filled-only" class="text-sm cursor-pointer">Filled only</Label>
				</div>
			</div>

			<div class="space-y-2">
				<Label class="text-sm font-medium">Date Range</Label>
				<DateRangePicker
					initialStartDate={dateRangeStart}
					initialEndDate={dateRangeEnd}
					change={handleDateRangeChange}
					placeholderText="Select range"
				/>
			</div>

			{#if dateRangeStart || dateRangeEnd || showOnlyAvailable || showOnlyFilled}
				<Button variant="outline" size="sm" onclick={clearFilters} class="w-full">
					<FilterIcon class="h-4 w-4 mr-2" />
					Clear Filters
				</Button>
			{/if}
		</div>

		<!-- Upcoming Shifts Header -->
		<div class="p-3 border-b bg-muted/50">
			<div class="flex items-center gap-2">
				<ClockIcon class="h-4 w-4" />
				<span class="text-sm font-medium">Upcoming Shifts</span>
			</div>
			<p class="text-xs text-muted-foreground">Next 7 days</p>
		</div>

		<!-- Upcoming Shifts List -->
		<div class="flex-grow overflow-y-auto">
			{#if $shiftsQuery.isLoading}
				<div class="p-3 space-y-3">
					{#each Array(5) as _, index (index)}
						<Skeleton class="h-16 w-full" />
					{/each}
				</div>
			{:else if $shiftsQuery.isError}
				<div class="p-3 text-sm text-destructive">
					Error loading shifts: {$shiftsQuery.error.message}
				</div>
			{:else if upcomingShifts.length === 0}
				<div class="p-3 text-sm text-muted-foreground">
					No upcoming shifts found matching your filters.
				</div>
			{:else}
				<div class="p-2">
					{#each upcomingShifts as shift (shift.schedule_id + '-' + shift.start_time)}
						<button
							class="w-full p-3 mb-2 text-left rounded-lg border transition-colors hover:bg-accent
								{shiftStartTimeFromUrl === shift.start_time ? 'border-primary bg-primary/5' : 'border-border'}"
							onclick={() => selectShift(shift)}
						>
							<div class="space-y-1">
								<div class="flex items-center justify-between">
									<h3 class="font-medium text-sm truncate">{shift.schedule_name}</h3>
									{#if shift.is_booked}
										<CheckIcon class="h-4 w-4 text-green-600" />
									{:else}
										<XIcon class="h-4 w-4 text-orange-600" />
									{/if}
								</div>
								<p class="text-xs text-muted-foreground">
									{formatTimeSlot(shift.start_time, shift.end_time)}
								</p>
								<p class="text-xs text-muted-foreground">
									{formatRelativeTime(shift.start_time)}
								</p>
								{#if shift.is_booked && shift.user_name}
									<p class="text-xs text-green-700">
										Assigned to: {shift.user_name}
									</p>
								{/if}
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={shiftsListContent} bind:searchTerm>
	{@render children()}
</SidebarPage>
