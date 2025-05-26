<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UsersIcon from '@lucide/svelte/icons/users';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';
	import { userSession } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';

	// Type definitions
	interface Shift {
		schedule_id: number;
		schedule_name: string;
		start_time: string;
		end_time: string;
		timezone: string;
		is_booked: boolean;
		positions_available?: number;
		positions_filled?: number;
	}

	interface ProcessedShift extends Shift {
		is_tonight: boolean;
		priority: string;
		slots_available: number;
		total_slots: number;
	}

	// Simplified state management without TanStack Query
	let shifts = $state<Shift[]>([]);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	// Filter state
	let selectedFilter = $state('all');
	let showFilters = $state(false);

	const filterOptions = [
		{ value: 'all', label: 'All Shifts' },
		{ value: 'tonight', label: 'Tonight Only' },
		{ value: 'available', label: 'Available Only' },
		{ value: 'urgent', label: 'Urgent Need' }
	];

	// Load shifts on mount if authenticated
	onMount(async () => {
		if ($userSession.isAuthenticated) {
			await loadShifts();
		}
	});

	async function loadShifts() {
		isLoading = true;
		error = null;
		try {
			const response = await fetch('/shifts/available');
			if (!response.ok) {
				throw new Error(`Failed to fetch shifts: ${response.status}`);
			}
			shifts = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred while loading shifts';
			console.error('Error loading shifts:', err);
		} finally {
			isLoading = false;
		}
	}

	// Helper function to determine if shift is tonight
	function isTonight(startTime: string): boolean {
		const shiftDate = new Date(startTime);
		const today = new Date();
		return shiftDate.toDateString() === today.toDateString();
	}

	// Helper function to determine priority (based on available status)
	function getPriority(shift: Shift): string {
		if (shift.is_booked) return 'high';
		return 'normal';
	}

	// Filtered shifts
	const filteredShifts = $derived.by((): ProcessedShift[] => {
		let filteredList: ProcessedShift[] = shifts.map((shift) => ({
			...shift,
			is_tonight: isTonight(shift.start_time),
			priority: getPriority(shift),
			slots_available: shift.is_booked ? 0 : 1,
			total_slots: 1
		}));

		switch (selectedFilter) {
			case 'tonight':
				filteredList = filteredList.filter((shift) => shift.is_tonight);
				break;
			case 'available':
				filteredList = filteredList.filter((shift) => !shift.is_booked);
				break;
			case 'urgent':
				filteredList = filteredList.filter((shift) => shift.is_booked);
				break;
		}

		return filteredList.sort((a, b) => {
			if (a.priority === 'high' && b.priority !== 'high') return -1;
			if (b.priority === 'high' && a.priority !== 'high') return 1;
			return new Date(a.start_time).getTime() - new Date(b.start_time).getTime();
		});
	});

	// Helper functions
	function formatShiftTime(startTime: string, endTime: string) {
		const start = new Date(startTime);
		const end = new Date(endTime);
		return `${start.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })} - ${end.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })}`;
	}

	function formatShiftDate(startTime: string) {
		const date = new Date(startTime);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);

		if (date.toDateString() === today.toDateString()) {
			return 'Today';
		} else if (date.toDateString() === tomorrow.toDateString()) {
			return 'Tomorrow';
		} else {
			return date.toLocaleDateString('en-GB', {
				weekday: 'short',
				month: 'short',
				day: 'numeric'
			});
		}
	}

	function getShiftDuration(startTime: string, endTime: string) {
		const start = new Date(startTime);
		const end = new Date(endTime);
		const diffHours = (end.getTime() - start.getTime()) / (1000 * 60 * 60);
		return `${diffHours}h`;
	}

	function getStatusText(shift: ProcessedShift) {
		if (shift.is_tonight && !shift.is_booked) return 'Tonight';
		if (shift.is_booked) return 'Urgent';
		return 'Available';
	}

	async function handleBookShift(shift: ProcessedShift) {
		if (!$userSession.isAuthenticated) {
			toast.error('Please sign in to book shifts');
			return;
		}

		try {
			const response = await fetch('/bookings', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${$userSession.token}`
				},
				body: JSON.stringify({
					schedule_id: shift.schedule_id,
					start_time: shift.start_time
				})
			});

			if (!response.ok) {
				throw new Error('Failed to book shift');
			}

			toast.success('Shift booked successfully!');
			await loadShifts();
		} catch (err) {
			toast.error(`Failed to book shift: ${err instanceof Error ? err.message : 'Unknown error'}`);
		}
	}
</script>

<svelte:head>
	<title>Available Shifts - Night Owls</title>
</svelte:head>

<div class="min-h-screen bg-background">
	<!-- Header -->
	<header class="bg-background border-b border-border sticky top-0 z-40">
		<div class="px-4 py-3">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="border border-border p-2 rounded">
						<ShieldIcon class="h-5 w-5 text-foreground" />
					</div>
					<div>
						<h1 class="text-lg font-semibold text-foreground">Patrol Shifts</h1>
						<p class="text-sm text-muted-foreground">
							{shifts.length} total, {shifts.filter((s) => isTonight(s.start_time)).length} tonight, {shifts.filter((s) => !s.is_booked).length} available
						</p>
					</div>
				</div>
				<Button
					variant="outline"
					size="sm"
					class="h-8 px-2"
					onclick={() => (showFilters = !showFilters)}
				>
					<FilterIcon class="h-4 w-4" />
				</Button>
			</div>

			{#if showFilters}
				<div class="mt-3 pt-3 border-t border-border">
					<select
						bind:value={selectedFilter}
						class="w-full px-3 py-2 border border-input rounded bg-background text-foreground focus:ring-1 focus:ring-ring"
					>
						{#each filterOptions as option (option.value)}
							<option value={option.value}>
								{option.label}
							</option>
						{/each}
					</select>
				</div>
			{/if}
		</div>
	</header>

	<div class="px-4 py-3">
		{#if !$userSession.isAuthenticated}
			<div class="border border-border rounded p-6 text-center">
				<ShieldIcon class="h-8 w-8 mx-auto text-muted-foreground mb-3" />
				<p class="text-muted-foreground">
					Please <a href="/login" class="text-foreground hover:underline font-medium">sign in</a> to
					view available patrol shifts.
				</p>
			</div>
		{:else if isLoading}
			<div class="border border-border rounded overflow-hidden">
				{#each Array(3) as _, index (index)}
					<div class="p-4 {index > 0 ? 'border-t border-border' : ''}">
						<div class="animate-pulse space-y-3">
							<div class="h-4 bg-muted rounded w-1/4"></div>
							<div class="h-4 bg-muted rounded w-1/2"></div>
							<div class="h-4 bg-muted rounded w-1/3"></div>
						</div>
					</div>
				{/each}
			</div>
		{:else if error}
			<div class="border border-border rounded p-6 text-center">
				<CalendarIcon class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
				<p class="text-foreground font-medium">
					Error loading shifts: {error}
				</p>
			</div>
		{:else}
			<!-- Shifts List -->
			<div class="border border-border rounded overflow-hidden">
				{#if filteredShifts.length === 0}
					<div class="p-6 text-center">
						<CalendarIcon class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
						<h3 class="text-base font-medium text-foreground mb-2">No shifts found</h3>
						<p class="text-sm text-muted-foreground">
							Try adjusting your filters or check back later for new patrol opportunities.
						</p>
					</div>
				{:else}
					{#each filteredShifts as shift, index (shift.schedule_id + '-' + shift.start_time)}
						<div class="p-4 {shift.priority === 'high' ? 'bg-muted/30' : 'bg-background'} {index > 0 ? 'border-t border-border' : ''}">
							<div class="flex items-start justify-between">
								<div class="flex-1 space-y-3">
									<div class="flex items-center gap-3">
										<div class="border border-border p-2 rounded">
											<ShieldIcon class="h-4 w-4 text-foreground" />
										</div>
										<div>
											<h3 class="font-medium text-foreground">
												{shift.schedule_name}
											</h3>
											<div class="flex items-center gap-2 mt-1">
												<Badge variant={shift.priority === 'high' ? 'secondary' : 'outline'} class="text-xs h-4 px-1">
													{getStatusText(shift)}
												</Badge>
												{#if shift.priority === 'high'}
													<Badge variant="secondary" class="text-xs h-4 px-1">Urgent</Badge>
												{/if}
											</div>
										</div>
									</div>

									<div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-sm">
										<div class="flex items-center gap-2 text-muted-foreground">
											<CalendarIcon class="h-4 w-4 flex-shrink-0" />
											<span class="text-foreground">{formatShiftDate(shift.start_time)}</span>
										</div>

										<div class="flex items-center gap-2 text-muted-foreground">
											<ClockIcon class="h-4 w-4 flex-shrink-0" />
											<span class="text-foreground">{formatShiftTime(shift.start_time, shift.end_time)}</span>
											<Badge variant="outline" class="text-xs h-4 px-1 ml-1">
												{getShiftDuration(shift.start_time, shift.end_time)}
											</Badge>
										</div>

										<div class="flex items-center gap-2 text-muted-foreground">
											<MapPinIcon class="h-4 w-4 flex-shrink-0" />
											<span class="text-foreground">Main Street Area</span>
										</div>
									</div>
								</div>

								<div class="text-right space-y-2 flex-shrink-0 ml-4">
									<div class="flex items-center gap-1 text-sm text-muted-foreground justify-end">
										<UsersIcon class="h-4 w-4" />
										<span class="text-foreground">{shift.slots_available}/{shift.total_slots}</span>
									</div>

									{#if !shift.is_booked}
										<Button
											size="sm"
											class="w-full min-w-[80px] h-8"
											onclick={() => handleBookShift(shift)}
										>
											Join
										</Button>
									{:else}
										<Button size="sm" variant="outline" disabled class="w-full min-w-[80px] h-8">
											Full
										</Button>
									{/if}
								</div>
							</div>

							{#if shift.slots_available === 1 && shift.total_slots > 1}
								<div class="mt-3 p-2 border border-border rounded bg-muted/30">
									<div class="flex items-center gap-2 text-sm text-foreground">
										⚠️ Only 1 patrol slot remaining - join now!
									</div>
								</div>
							{/if}
						</div>
					{/each}
				{/if}
			</div>
		{/if}

		<!-- Bottom spacing for mobile navigation -->
		<div class="h-6"></div>
	</div>
</div>
