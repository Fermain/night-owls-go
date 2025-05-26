<script lang="ts">
	import { userSession } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';
	import ShiftsHeader from '$lib/components/shifts/ShiftsHeader.svelte';
	import ShiftsFilter from '$lib/components/shifts/ShiftsFilter.svelte';
	import ShiftsList from '$lib/components/shifts/ShiftsList.svelte';
	import type { Shift, ProcessedShift, FilterOption } from '$lib/types/shifts';

	// State management
	let shifts = $state<Shift[]>([]);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	// Filter state
	let selectedFilter = $state<FilterOption>('all');
	let showFilters = $state(false);

	// Load shifts on mount - works for both authenticated and unauthenticated users
	onMount(async () => {
		await loadShifts();
	});

	async function loadShifts() {
		isLoading = true;
		error = null;
		try {
			// Add query parameters for better data
			const params = new URLSearchParams();
			const now = new Date();
			const twoWeeksFromNow = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000);
			
			params.append('from', now.toISOString());
			params.append('to', twoWeeksFromNow.toISOString());
			params.append('limit', '50');

			const response = await fetch(`/shifts/available?${params.toString()}`);
			if (!response.ok) {
				throw new Error(`Failed to fetch shifts: ${response.status}`);
			}
			const data = await response.json();
			shifts = Array.isArray(data) ? data : [];
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

	// Helper function to determine priority (based on available status and timing)
	function getPriority(shift: Shift): string {
		if (shift.is_booked) return 'high';
		if (isTonight(shift.start_time)) return 'medium';
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
			if (a.priority === 'medium' && b.priority === 'normal') return -1;
			if (b.priority === 'medium' && a.priority === 'normal') return 1;
			return new Date(a.start_time).getTime() - new Date(b.start_time).getTime();
		});
	});

	// Event handlers
	function handleToggleFilters() {
		showFilters = !showFilters;
	}

	function handleFilterChange(filter: FilterOption) {
		selectedFilter = filter;
	}

	function handleSignIn() {
		window.location.href = '/login';
	}

	function handleShowAllShifts() {
		selectedFilter = 'all';
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
	<ShiftsHeader 
		{shifts} 
		onToggleFilters={handleToggleFilters}
	/>

	{#if showFilters}
		<div class="px-4">
			<ShiftsFilter 
				{selectedFilter} 
				onFilterChange={handleFilterChange}
			/>
		</div>
	{/if}

	<div class="px-4 py-3">
		<ShiftsList 
			shifts={filteredShifts}
			{isLoading}
			{error}
			{selectedFilter}
			isAuthenticated={$userSession.isAuthenticated}
			onBookShift={handleBookShift}
			onSignIn={handleSignIn}
			onRetry={loadShifts}
			onShowAllShifts={handleShowAllShifts}
		/>

		<!-- Bottom spacing for mobile navigation -->
		<div class="h-6"></div>
	</div>
</div>
