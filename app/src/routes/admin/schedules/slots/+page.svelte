<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import * as Table from '$lib/components/ui/table/index.js';
	import { toast } from 'svelte-sonner';
	import { formatDistanceToNow } from 'date-fns';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';

	// --- Types ---
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
		// aka_description?: string | null; // Future field from backend
	};

	// --- Utility Functions ---
	function formatDateForInput(date: Date): string {
		const year = date.getFullYear();
		const month = (date.getMonth() + 1).toString().padStart(2, '0');
		const day = date.getDate().toString().padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function formatTimeSlot(startTimeIso: string, endTimeIso: string): string {
		if (!startTimeIso || !endTimeIso) return 'N/A';
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);

			const options: Intl.DateTimeFormatOptions = {
				weekday: 'short',
				month: 'short',
				day: 'numeric',
				hour: 'numeric',
				minute: '2-digit',
				hour12: true
			};

			const startFormatted = startDate.toLocaleString(undefined, options);
			// Only show date part for end time if it's different from start date
			const endFormatted = endDate.toLocaleTimeString(undefined, {
				hour: 'numeric',
				minute: '2-digit',
				hour12: true
			});

			if (startDate.toDateString() === endDate.toDateString()) {
				return `${startFormatted.replace(startDate.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit', hour12: true }), '').trim()} - ${endFormatted}`;
			} else {
				const endDayFormatted = endDate.toLocaleString(undefined, {
					weekday: 'short',
					month: 'short',
					day: 'numeric'
				});
				return `${startFormatted} - ${endDayFormatted}, ${endFormatted}`;
			}
		} catch (e) {
			return 'Invalid Date Range';
		}
	}

	function formatRelativeTime(timeIso: string): string {
		if (!timeIso) return 'N/A';
		try {
			const time = new Date(timeIso);
			// Add suffix to get "ago" or "in ..."
			return formatDistanceToNow(time, { addSuffix: true });
		} catch (e) {
			return 'Invalid Date';
		}
	}

	function getAkaDescription(startTimeIso: string): string {
		if (!startTimeIso) return '';
		try {
			const startDate = new Date(startTimeIso);
			const day = startDate.getDay(); // 0 (Sun) to 6 (Sat)
			const hour = startDate.getHours(); // 0 to 23

			// Example: Saturday 00:00 - 04:59 is "Friday Night"
			if (day === 6 && hour >= 0 && hour < 5) {
				return 'Friday Night';
			}
			// Example: Sunday 00:00 - 04:59 is "Saturday Night"
			if (day === 0 && hour >= 0 && hour < 5) {
				return 'Saturday Night';
			}
			// Add more rules as needed for other AKA descriptions

			return ''; // Default if no AKA matches
		} catch (e) {
			return ''; // Return empty on error
		}
	}

	// --- State ---
	const initialFromDate = new Date();
	const initialToDate = new Date();
	initialToDate.setDate(initialToDate.getDate() + 7);

	let fromDateStr = $state(formatDateForInput(initialFromDate));
	let toDateStr = $state(formatDateForInput(initialToDate));

	// Function to handle date changes from DateRangePicker
	function handleDateFilterChange(detail: { start: string | null; end: string | null }) {
		if (detail.start) fromDateStr = detail.start;
		if (detail.end) toDateStr = detail.end;
	}

	// --- Data Fetching ---
	const fetchAdminShiftSlots = async (
		currentFromDate: string,
		currentToDate: string
	): Promise<AdminShiftSlot[]> => {
		const params = new URLSearchParams();
		// currentFromDate and currentToDate will be non-null strings here because
		// the query is only enabled when fromDateStr and toDateStr are non-null.
		const fromDT = new Date(currentFromDate);
		fromDT.setHours(0, 0, 0, 0);
		params.append('from', fromDT.toISOString());
		
		const toDT = new Date(currentToDate);
		toDT.setHours(23, 59, 59, 999);
		params.append('to', toDT.toISOString());

		const response = await fetch(`/api/admin/schedules/all-slots?${params.toString()}`);

		if (!response.ok) {
			let errorMsg = `HTTP error ${response.status}`;
			try {
				const errorData = await response.json();
				errorMsg = errorData.message || errorData.error || errorMsg;
			} catch (e) {
				/* Failed to parse JSON, use default error*/
			}
			toast.error(`Failed to fetch shift slots: ${errorMsg}`);
			throw new Error(errorMsg);
		}
		return response.json() as Promise<AdminShiftSlot[]>;
	};

	const queryOptions = $derived({
		// Ensure queryKey always has 3 string elements to match TQueryKey.
		// Use empty strings if fromDateStr or toDateStr are null; 'enabled' flag controls execution.
		queryKey: ['adminShiftSlots', fromDateStr || '', toDateStr || ''] as [string, string, string],
		queryFn: () => {
			// This function will only be called if enabled is true,
			// at which point fromDateStr and toDateStr are guaranteed to be strings.
			return fetchAdminShiftSlots(fromDateStr!, toDateStr!); 
		},
		enabled: !!fromDateStr && !!toDateStr
	});

	const slotsQuery = createQuery<
		AdminShiftSlot[],
		Error,
		AdminShiftSlot[],
		[string, string, string] // QueryKey type
	>(queryOptions);
</script>

<svelte:head>
	<title>Admin - Shift Slots</title>
</svelte:head>

<div class="container mx-auto p-4 space-y-6">
	<h1 class="text-2xl font-bold mb-4">Shift Slots Dashboard</h1>
	<div class="flex flex-col sm:flex-row gap-4 items-end p-4 border rounded-lg bg-card">
		<DateRangePicker
			initialStartDate={fromDateStr}
			initialEndDate={toDateStr}
			change={handleDateFilterChange}
			placeholderText="Select date range for slots"
		/>
	</div>

	{#if $slotsQuery.isLoading}
		<div class="border rounded-lg mt-4">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head class="w-[20%]">Schedule</Table.Head>
						<Table.Head class="w-[30%]">Time Slot</Table.Head>
						<Table.Head class="w-[15%]">Starts / Started</Table.Head>
						<Table.Head class="w-[25%]">Booking Status</Table.Head>
						<Table.Head class="w-[10%]">AKA</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each Array(5) as _}
						<Table.Row>
							{#each Array(5) as __}
								<Table.Cell><div class="h-4 bg-gray-200 rounded animate-pulse"></div></Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	{:else if $slotsQuery.isError}
		<p class="text-destructive">Error loading shift slots: {$slotsQuery.error.message}</p>
	{:else if $slotsQuery.data}
		{#if $slotsQuery.data.length === 0}
			<p>No shift slots found for the selected period.</p>
		{:else}
			<div class="border rounded-lg">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-[20%]">Schedule</Table.Head>
							<Table.Head class="w-[30%]">Time Slot</Table.Head>
							<Table.Head class="w-[15%]">Starts / Started</Table.Head>
							<Table.Head class="w-[25%]">Booking Status</Table.Head>
							<Table.Head class="w-[10%]">AKA</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each $slotsQuery.data as slot (slot.schedule_id + slot.start_time)}
							<Table.Row class="hover:bg-muted/50">
								<Table.Cell class="font-medium">{slot.schedule_name}</Table.Cell>
								<Table.Cell>{formatTimeSlot(slot.start_time, slot.end_time)}</Table.Cell>
								<Table.Cell>{formatRelativeTime(slot.start_time)}</Table.Cell>
								<Table.Cell>
									{#if slot.is_booked}
										<span class="text-orange-600 font-semibold">Taken</span>
										{#if slot.user_name || slot.user_phone}
											<span class="text-xs text-muted-foreground ml-1">
												by: {slot.user_name ?? 'N/A'}{#if slot.user_phone}
													({slot.user_phone}){/if}
											</span>
										{/if}
									{:else}
										<span class="text-green-600 font-semibold">Available</span>
									{/if}
								</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">
									{getAkaDescription(slot.start_time) || '-'}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		{/if}
	{:else}
		<p>Select dates to view shift slots.</p>
	{/if}
</div>
