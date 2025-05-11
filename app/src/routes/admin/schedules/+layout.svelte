<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/state'; // Corrected by user
	import { goto } from '$app/navigation';
	import { createQuery, type QueryKey, type CreateQueryResult } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import type { Snippet } from 'svelte';
	import { formatDistanceToNow } from 'date-fns'; // Added for upcoming shifts
	import type { Schedule } from '$lib/components/schedules_table/columns'; // Import the shared Schedule type
	import { CalendarDays, PlusCircle } from 'lucide-svelte'; // Added CalendarDays

	type AdminShiftSlot = {
		schedule_id: number;
		schedule_name: string;
		start_time: string; // ISO date string
		end_time: string; // ISO date string
		timezone?: string | null;
		is_booked: boolean;
		booking_id?: number | null;
		user_name?: string | null;
		user_phone?: string | null;
	};

	let searchTerm = $state('');
	let { children } = $props();

	const fetchSchedules = async (): Promise<Schedule[]> => {
		const response = await fetch('/api/admin/schedules');
		if (!response.ok) {
			let errorText = response.statusText;
			try {
				const errorData = await response.text(); // Try to get text for more detailed error
				console.error('Failed to fetch schedules. Status:', response.status, 'Response:', errorData);
				errorText = errorData || errorText;
			} catch (e) {
				console.error('Failed to fetch schedules and could not parse error response body. Status:', response.status);
			}
			toast.error(`Failed to fetch schedules: ${response.status} ${errorText}`);
			throw new Error(`Failed to fetch schedules: ${response.status} ${errorText}`);
		}
		// Check content type before parsing
		const contentType = response.headers.get("content-type");
		if (contentType && contentType.indexOf("application/json") !== -1) {
			return response.json();
		} else {
			const responseText = await response.text();
			console.error('Schedules response was not JSON. Received:', responseText);
			toast.error('Received non-JSON response for schedules.');
			throw new Error('Schedules response was not JSON.');
		}
	};

	const schedulesQuery = $derived(createQuery<Schedule[], Error, Schedule[], QueryKey>({
		queryKey: ['adminSchedulesForLayout'],
		queryFn: fetchSchedules
	}));

	const schedulesForTemplate = $derived.by(() => {
		// First, get the raw data from the query
		const rawData = $schedulesQuery.data;

		// If there's no raw data, return an empty array immediately
		if (!rawData) return [];

		// Filter out schedules that don't have a valid ID for the key
		const validKeyedData = rawData.filter(schedule => 
			schedule.schedule_id !== null && schedule.schedule_id !== undefined
		);

		// If no search term, return the data that has valid keys
		if (!searchTerm) return validKeyedData;

		// If there is a search term, filter the valid-keyed data further
		return validKeyedData.filter(
			// Ensure name exists before trying to use toLowerCase()
			(schedule) => schedule.name && schedule.name.toLowerCase().includes(searchTerm.toLowerCase())
		);
	});

	const fetchUpcomingShiftSlotsLayout = async (): Promise<AdminShiftSlot[]> => {
		const now = new Date();
		const toDate = new Date(now);
		toDate.setDate(now.getDate() + 30);

		const params = new URLSearchParams({
			from: now.toISOString(),
			to: toDate.toISOString()
		});

		const response = await fetch(`/api/admin/schedules/all-slots?${params.toString()}`);
		if (!response.ok) {
			let errorMsg = `HTTP error ${response.status}`;
			try { const errorData = await response.json(); errorMsg = errorData.message || errorData.error || errorMsg; } catch (e) { /* ignore */ }
			toast.error(`Failed to fetch upcoming shifts: ${errorMsg}`);
			throw new Error(errorMsg);
		}
		const allSlots = await response.json() as AdminShiftSlot[];
		return allSlots
			.filter(slot => new Date(slot.start_time) >= now)
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
	};

	const isSlotsPage = $derived(page.url.pathname.startsWith('/admin/schedules/slots'));

	const upcomingSlotsLayoutQuery = $derived(createQuery({
		queryKey: ['upcomingAdminShiftSlotsForSchedulesLayout'],
		queryFn: fetchUpcomingShiftSlotsLayout,
		enabled: isSlotsPage
	}));

	const upcomingShiftsForTemplate = $derived($upcomingSlotsLayoutQuery.data ?? []);

	function formatShiftTitleCondensed(startTimeIso: string, endTimeIso: string): string {
		if (!startTimeIso || !endTimeIso) return 'N/A';
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);
			const startDay = startDate.toLocaleDateString(undefined, { weekday: 'short' }).toUpperCase();

			const formatHourWithAmPm = (date: Date) => {
				let h = date.getHours();
				const m = date.getMinutes();
				const ampm = h >= 12 ? 'PM' : 'AM';
				h = h % 12;
				h = h ? h : 12;
				return h + (m === 0 ? '' : `:${m.toString().padStart(2, '0')}`);
			};

			const startHourStr = formatHourWithAmPm(startDate);
			const endHourStr = formatHourWithAmPm(endDate);
			const endAmPm = endDate.getHours() >= 12 ? 'PM' : 'AM';

			return `${startDay} ${startHourStr}-${endHourStr}${endAmPm}`;
		} catch (e) {
			console.error("Error formatting shift title condensed:", e);
			return 'Invalid Time';
		}
	}

	const currentListContentSnippet: Snippet = $derived(
		isSlotsPage ? upcomingShiftsLayoutListContent : scheduleListContent
	);
	const currentTitle: string = $derived(
		isSlotsPage ? 'Upcoming Shifts' : 'Schedules'
	);
	
	const isNewSchedulePage = $derived(page.url.pathname === '/admin/schedules/new');
	const editScheduleId = $derived(page.url.searchParams.get('scheduleId'));

</script>

{#snippet scheduleListContent()}
	<Sidebar.Menu class="p-2">
		<Sidebar.MenuItem class="mb-2">
			<Button
				variant={isNewSchedulePage ? 'default' : 'outline'}
				class="w-full justify-start"
				onclick={() => goto('/admin/schedules/new')}
			>
				Create New Schedule 
			</Button>
		</Sidebar.MenuItem>

		{#if $schedulesQuery.isLoading}
			<p class="p-2 text-sm text-muted-foreground">Loading schedules...</p>
		{:else if $schedulesQuery.isError}
			<p class="p-2 text-sm text-destructive">
				Error loading schedules: {$schedulesQuery.error?.message ?? 'Unknown error'}
			</p>
		{:else if !$schedulesQuery.data}
			 <p class="p-2 text-sm text-muted-foreground">No data received for schedules.</p>
		{:else if schedulesForTemplate.length > 0}
			{#each schedulesForTemplate as schedule (schedule.schedule_id)}
				<Sidebar.MenuItem>
					<Sidebar.MenuButton
						onclick={() => goto(`/admin/schedules?scheduleId=${schedule.schedule_id}`)}
						isActive={editScheduleId === String(schedule.schedule_id) && !isSlotsPage && !isNewSchedulePage }
					>
						<CalendarDays class="mr-2 h-4 w-4 text-muted-foreground" />
						{schedule.name}
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			{/each}
		{:else if $schedulesQuery.data && schedulesForTemplate.length === 0 && searchTerm}
			<p class="p-2 text-sm text-muted-foreground">No schedules match "{searchTerm}".</p>
		{:else}
			<p class="p-2 text-sm text-muted-foreground">No schedules found.</p>
		{/if}
	</Sidebar.Menu>
{/snippet}

{#snippet upcomingShiftsLayoutListContent()}
	<div class="flex flex-col h-full overflow-y-auto">
		{#if $upcomingSlotsLayoutQuery.isLoading}
			<p class="p-4 text-sm text-muted-foreground">Loading upcoming shifts...</p>
		{:else if $upcomingSlotsLayoutQuery.isError}
			<p class="p-4 text-sm text-destructive">Error: {$upcomingSlotsLayoutQuery.error?.message ?? 'Unknown error'}</p>
		{:else if upcomingShiftsForTemplate.length > 0}
			<Sidebar.Menu class="p-2">
				{#each upcomingShiftsForTemplate as shift (shift.schedule_id + shift.start_time)}
					<Sidebar.MenuItem>
						<Sidebar.MenuButton
							onclick={() => goto(`/admin/schedules/slots?shiftStartTime=${encodeURIComponent(shift.start_time)}`)}
							isActive={page.url.searchParams.get('shiftStartTime') === shift.start_time}
							class="flex flex-col items-start h-auto py-2 w-full text-left"
						>
							<span class="font-semibold text-sm">{formatShiftTitleCondensed(shift.start_time, shift.end_time)}</span>
							<span class="text-xs text-muted-foreground">{shift.schedule_name}</span>
							<span class="text-xs text-muted-foreground">{formatDistanceToNow(new Date(shift.start_time), { addSuffix: true })}</span>
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				{/each}
			</Sidebar.Menu>
		{:else}
			<p class="p-4 text-sm text-muted-foreground">No upcoming shifts found.</p>
		{/if}
	</div>
{/snippet}

<SidebarPage listContent={currentListContentSnippet} title={currentTitle} bind:searchTerm>
	{@render children()}
</SidebarPage>
