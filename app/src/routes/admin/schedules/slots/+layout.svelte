<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { createQuery, type QueryKey, type CreateQueryResult } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { formatDistanceToNow } from 'date-fns';
	import type { Snippet } from 'svelte';

	// Type for AdminShiftSlot (copied from the parent layout)
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

	const fetchUpcomingShiftSlotsLayout = async (): Promise<AdminShiftSlot[]> => {
		const now = new Date();
		const toDate = new Date(now);
		toDate.setDate(now.getDate() + 30); // Look ahead 30 days

		const params = new URLSearchParams({
			from: now.toISOString(), // Only fetch from now onwards
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
		// Backend should ideally handle filtering by start_time >= now, but double check here.
		return allSlots
			.filter(slot => new Date(slot.start_time) >= now) 
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
	};

	const upcomingSlotsLayoutQuery: CreateQueryResult<AdminShiftSlot[], Error> = createQuery({
		queryKey: ['upcomingAdminShiftSlotsLayoutForSlotsPage'] as QueryKey, // Unique queryKey
		queryFn: fetchUpcomingShiftSlotsLayout
	});

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

	const title = 'Upcoming Shifts';

	// Reverted: Derived state for the active shift based on URL param
	// let currentShiftStartTime = $derived(page.url.searchParams.get('shiftStartTime'));
</script>

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

<SidebarPage listContent={upcomingShiftsLayoutListContent} {title} bind:searchTerm>
	{@render children()}
</SidebarPage> 