<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UsersIcon from '@lucide/svelte/icons/users';
	import FilterIcon from '@lucide/svelte/icons/filter';
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

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'high':
				return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
			case 'normal':
				return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
			case 'low':
				return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300';
			default:
				return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300';
		}
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

<div
	class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800"
>
	<!-- Header -->
	<header
		class="bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border-b border-slate-200 dark:border-slate-700 sticky top-0 z-40"
	>
		<div class="px-4 py-3">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-lg font-semibold text-slate-900 dark:text-slate-100">Available Shifts</h1>
					<p class="text-sm text-slate-600 dark:text-slate-400">Choose your patrol time</p>
				</div>
				<Button
					variant="outline"
					size="sm"
					class="gap-2"
					onclick={() => (showFilters = !showFilters)}
				>
					<FilterIcon class="h-4 w-4" />
					Filter
				</Button>
			</div>

			{#if showFilters}
				<div class="mt-3 pt-3 border-t border-slate-200 dark:border-slate-700">
					<select
						bind:value={selectedFilter}
						class="w-full px-3 py-2 border border-slate-200 dark:border-slate-700 rounded-md bg-background text-foreground"
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

	<div class="px-4 py-6">
		{#if !$userSession.isAuthenticated}
			<Card.Root>
				<Card.Content class="pt-6">
					<p class="text-center text-muted-foreground">
						Please <a href="/login" class="text-primary hover:underline">sign in</a> to view available
						shifts.
					</p>
				</Card.Content>
			</Card.Root>
		{:else if isLoading}
			<div class="space-y-4">
				{#each Array(3) as _, index (index)}
					<Card.Root>
						<Card.Content class="pt-6">
							<div class="animate-pulse space-y-3">
								<div class="h-4 bg-muted rounded w-1/4"></div>
								<div class="h-4 bg-muted rounded w-1/2"></div>
								<div class="h-4 bg-muted rounded w-1/3"></div>
							</div>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		{:else if error}
			<Card.Root>
				<Card.Content class="pt-6">
					<p class="text-destructive text-center">
						Error loading shifts: {error}
					</p>
				</Card.Content>
			</Card.Root>
		{:else}
			<!-- Quick Stats -->
			<div class="grid grid-cols-3 gap-3 mb-6">
				<Card.Root class="text-center">
					<Card.Content class="p-3">
						<div class="text-lg font-bold text-blue-600 dark:text-blue-400">
							{shifts.filter((s) => isTonight(s.start_time)).length}
						</div>
						<div class="text-xs text-slate-600 dark:text-slate-400">Tonight</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="text-center">
					<Card.Content class="p-3">
						<div class="text-lg font-bold text-green-600 dark:text-green-400">
							{shifts.filter((s) => !s.is_booked).length}
						</div>
						<div class="text-xs text-slate-600 dark:text-slate-400">Available</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="text-center">
					<Card.Content class="p-3">
						<div class="text-lg font-bold text-orange-600 dark:text-orange-400">
							{shifts.filter((s) => s.is_booked).length}
						</div>
						<div class="text-xs text-slate-600 dark:text-slate-400">Urgent</div>
					</Card.Content>
				</Card.Root>
			</div>

			<!-- Shifts List -->
			<div class="space-y-4">
				{#if filteredShifts.length === 0}
					<Card.Root class="text-center">
						<Card.Content class="p-8">
							<CalendarIcon class="h-12 w-12 text-slate-400 mx-auto mb-3" />
							<h3 class="text-lg font-medium text-slate-900 dark:text-slate-100 mb-2">
								No shifts found
							</h3>
							<p class="text-sm text-slate-600 dark:text-slate-400">
								Try adjusting your filters or check back later.
							</p>
						</Card.Content>
					</Card.Root>
				{:else}
					{#each filteredShifts as shift (shift.schedule_id + '-' + shift.start_time)}
						<Card.Root
							class="hover:shadow-md transition-shadow {shift.priority === 'high'
								? 'ring-2 ring-red-200 dark:ring-red-800'
								: ''}"
						>
							<Card.Content class="p-4">
								<div class="flex items-start justify-between mb-3">
									<div class="flex-1">
										<div class="flex items-center gap-2 mb-1">
											<h3 class="font-semibold text-slate-900 dark:text-slate-100">
												{shift.schedule_name}
											</h3>
											{#if shift.priority === 'high'}
												<Badge variant="destructive" class="text-xs">Urgent</Badge>
											{/if}
											{#if shift.is_tonight}
												<Badge class={getPriorityColor('normal')}>Tonight</Badge>
											{/if}
										</div>

										<div class="space-y-1 text-sm text-slate-600 dark:text-slate-400">
											<div class="flex items-center gap-2">
												<CalendarIcon class="h-4 w-4" />
												<span>{formatShiftDate(shift.start_time)}</span>
											</div>

											<div class="flex items-center gap-2">
												<ClockIcon class="h-4 w-4" />
												<span>{formatShiftTime(shift.start_time, shift.end_time)}</span>
												<span class="text-xs"
													>({getShiftDuration(shift.start_time, shift.end_time)})</span
												>
											</div>
										</div>
									</div>

									<div class="text-right">
										<div
											class="flex items-center gap-1 text-sm text-slate-600 dark:text-slate-400 mb-2"
										>
											<UsersIcon class="h-4 w-4" />
											<span>{shift.slots_available}/{shift.total_slots}</span>
										</div>

										{#if !shift.is_booked}
											<Button
												size="sm"
												class="w-full min-w-[80px]"
												onclick={() => handleBookShift(shift)}
											>
												Book Slot
											</Button>
										{:else}
											<Button size="sm" variant="outline" disabled class="w-full min-w-[80px]">
												Full
											</Button>
										{/if}
									</div>
								</div>

								{#if shift.slots_available === 1 && shift.total_slots > 1}
									<div
										class="text-xs text-orange-600 dark:text-orange-400 bg-orange-50 dark:bg-orange-950/50 rounded p-2"
									>
										⚠️ Only 1 slot remaining
									</div>
								{/if}
							</Card.Content>
						</Card.Root>
					{/each}
				{/if}
			</div>
		{/if}

		<!-- Bottom spacing for mobile navigation -->
		<div class="h-6"></div>
	</div>
</div>
