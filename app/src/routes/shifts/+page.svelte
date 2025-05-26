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

	function getStatusBadgeClass(shift: ProcessedShift) {
		if (shift.is_tonight && !shift.is_booked) return 'status-night';
		if (shift.is_booked) return 'status-urgent';
		return 'status-safe';
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

<div class="min-h-screen bg-patrol-gradient">
	<!-- Enhanced Header with better visual hierarchy -->
	<header
		class="bg-card/95 backdrop-blur-md border-b border-border/50 sticky top-0 z-40 animate-in"
	>
		<div class="px-4 py-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="bg-primary/10 p-2 rounded-lg">
						<ShieldIcon class="h-5 w-5 text-primary" />
					</div>
					<div>
						<h1 class="text-xl font-bold text-foreground">Patrol Shifts</h1>
						<p class="text-sm text-muted-foreground">Choose your watch time</p>
					</div>
				</div>
				<Button
					variant="outline"
					size="sm"
					class="gap-2 interactive-scale"
					onclick={() => (showFilters = !showFilters)}
				>
					<FilterIcon class="h-4 w-4" />
					Filter
				</Button>
			</div>

			{#if showFilters}
				<div class="mt-4 pt-4 border-t border-border/50 animate-in">
					<select
						bind:value={selectedFilter}
						class="w-full px-3 py-2 border border-input rounded-lg bg-background text-foreground focus:ring-2 focus:ring-ring focus:border-ring transition-all"
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

	<div class="px-4 py-6 space-y-6">
		{#if !$userSession.isAuthenticated}
			<Card.Root class="animate-in">
				<Card.Content class="pt-6">
					<div class="text-center space-y-3">
						<ShieldIcon class="h-12 w-12 mx-auto text-muted-foreground" />
						<p class="text-muted-foreground">
							Please <a href="/login" class="text-primary hover:underline font-medium">sign in</a> to
							view available patrol shifts.
						</p>
					</div>
				</Card.Content>
			</Card.Root>
		{:else if isLoading}
			<div class="space-y-4">
				{#each Array(3) as _, index (index)}
					<Card.Root class="animate-in" style="animation-delay: {index * 100}ms">
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
			<Card.Root class="animate-in">
				<Card.Content class="pt-6">
					<div class="text-center space-y-3">
						<div class="bg-destructive/10 p-3 rounded-full w-fit mx-auto">
							<CalendarIcon class="h-8 w-8 text-destructive" />
						</div>
						<p class="text-destructive font-medium">
							Error loading shifts: {error}
						</p>
					</div>
				</Card.Content>
			</Card.Root>
		{:else}
			<!-- Enhanced Quick Stats with community theming -->
			<div class="grid grid-cols-3 gap-4 mb-6">
				<Card.Root class="text-center interactive-scale animate-in">
					<Card.Content class="p-4">
						<div class="text-2xl font-bold status-night">
							{shifts.filter((s) => isTonight(s.start_time)).length}
						</div>
						<div class="text-xs text-muted-foreground font-medium">Tonight</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="text-center interactive-scale animate-in" style="animation-delay: 100ms">
					<Card.Content class="p-4">
						<div class="text-2xl font-bold status-safe">
							{shifts.filter((s) => !s.is_booked).length}
						</div>
						<div class="text-xs text-muted-foreground font-medium">Available</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="text-center interactive-scale animate-in" style="animation-delay: 200ms">
					<Card.Content class="p-4">
						<div class="text-2xl font-bold status-urgent">
							{shifts.filter((s) => s.is_booked).length}
						</div>
						<div class="text-xs text-muted-foreground font-medium">Urgent</div>
					</Card.Content>
				</Card.Root>
			</div>

			<!-- Enhanced Shifts List -->
			<div class="space-y-4">
				{#if filteredShifts.length === 0}
					<Card.Root class="text-center animate-in">
						<Card.Content class="p-8">
							<div class="bg-muted/50 p-4 rounded-full w-fit mx-auto mb-4">
								<CalendarIcon class="h-12 w-12 text-muted-foreground" />
							</div>
							<h3 class="text-lg font-semibold text-foreground mb-2">No shifts found</h3>
							<p class="text-sm text-muted-foreground">
								Try adjusting your filters or check back later for new patrol opportunities.
							</p>
						</Card.Content>
					</Card.Root>
				{:else}
					{#each filteredShifts as shift, index (shift.schedule_id + '-' + shift.start_time)}
						<Card.Root
							class="hover:shadow-lg transition-all duration-300 interactive-scale animate-in {shift.priority ===
							'high'
								? 'ring-2 ring-destructive/20 shadow-destructive/5'
								: ''}"
							style="animation-delay: {index * 50}ms"
						>
							<Card.Content class="p-5">
								<div class="flex items-start justify-between mb-4">
									<div class="flex-1 space-y-3">
										<div class="flex items-center gap-3">
											<div class="bg-primary/10 p-2 rounded-lg">
												<ShieldIcon class="h-4 w-4 text-primary" />
											</div>
											<div>
												<h3 class="font-semibold text-foreground text-base">
													{shift.schedule_name}
												</h3>
												<div class="flex items-center gap-2 mt-1">
													<Badge class={getStatusBadgeClass(shift)}>
														{getStatusText(shift)}
													</Badge>
													{#if shift.priority === 'high'}
														<Badge variant="destructive" class="text-xs">Urgent</Badge>
													{/if}
												</div>
											</div>
										</div>

										<div class="space-y-2 text-sm text-muted-foreground">
											<div class="flex items-center gap-2">
												<CalendarIcon class="h-4 w-4" />
												<span class="font-medium">{formatShiftDate(shift.start_time)}</span>
											</div>

											<div class="flex items-center gap-2">
												<ClockIcon class="h-4 w-4" />
												<span>{formatShiftTime(shift.start_time, shift.end_time)}</span>
												<span class="text-xs bg-muted px-2 py-1 rounded-full">
													{getShiftDuration(shift.start_time, shift.end_time)}
												</span>
											</div>

											<div class="flex items-center gap-2">
												<MapPinIcon class="h-4 w-4" />
												<span>Main Street Area</span>
											</div>
										</div>
									</div>

									<div class="text-right space-y-3">
										<div class="flex items-center gap-1 text-sm text-muted-foreground">
											<UsersIcon class="h-4 w-4" />
											<span class="font-medium">{shift.slots_available}/{shift.total_slots}</span>
										</div>

										{#if !shift.is_booked}
											<Button
												size="sm"
												class="w-full min-w-[90px] interactive-scale"
												onclick={() => handleBookShift(shift)}
											>
												Join Patrol
											</Button>
										{:else}
											<Button size="sm" variant="outline" disabled class="w-full min-w-[90px]">
												Full
											</Button>
										{/if}
									</div>
								</div>

								{#if shift.slots_available === 1 && shift.total_slots > 1}
									<div class="status-warning text-xs rounded-lg p-3 font-medium">
										⚠️ Only 1 patrol slot remaining - join now!
									</div>
								{/if}
							</Card.Content>
						</Card.Root>
					{/each}
				{/if}
			</div>
		{/if}

		<!-- Bottom spacing for mobile navigation -->
		<div class="h-8"></div>
	</div>
</div>
