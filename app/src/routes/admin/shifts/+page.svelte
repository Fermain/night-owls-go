<script lang="ts">
	import { createQuery, useQueryClient, createMutation } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { formatDistanceToNow } from 'date-fns';
	import { page } from '$app/stores';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Calendar } from '$lib/components/ui/calendar';
	import * as Chart from '$lib/components/ui/chart';
	import { BarChart, PieChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import { CalendarDate, today, getLocalTimeZone } from '@internationalized/date';
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import * as Select from '$lib/components/ui/select';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { AdminShiftSlot, Schedule } from '$lib/types';
	import type { UserData } from '$lib/schemas/user';
	import { Label } from '$lib/components/ui/label';
	import CheckIcon from '@lucide/svelte/icons/check';
	import XIcon from '@lucide/svelte/icons/x';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import ShiftsDashboard from '$lib/components/dashboard/ShiftsDashboard.svelte';

	// State
	let selectedShift = $state<AdminShiftSlot | null>(null);
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));
	let selectedUserIdForBooking = $state<string | undefined>(undefined);
	let bookingFormError = $state<string | null>(null);
	
	// Schedule dialog state
	let showScheduleDialog = $state(false);
	let selectedScheduleForEdit = $state<Schedule | null>(null);
	let scheduleDialogMode = $state<'create' | 'edit'>('create');

	const queryClient = useQueryClient();
	let currentDisplayMonth = $state(today(getLocalTimeZone()));
	let isBookingFormEnabled = $derived(!!selectedShift && !selectedShift.is_booked);

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
				hour12: false,
				timeZone: 'UTC'
			});

			const endFormatted = endDate.toLocaleTimeString('en-ZA', {
				hour: '2-digit',
				minute: '2-digit',
				hour12: false,
				timeZone: 'UTC'
			});

			return `${startFormatted} - ${endFormatted}`;
		} catch (e) {
			return 'Invalid Date Range';
		}
	}

	function formatRelativeTime(timeIso: string): string {
		if (!timeIso) return 'N/A';
		try {
			return formatDistanceToNow(new Date(timeIso), { addSuffix: true });
		} catch (e) {
			return 'Invalid Date';
		}
	}

	function getAkaDescription(startTimeIso: string): string {
		if (!startTimeIso) return '';
		try {
			const startDate = new Date(startTimeIso);
			const day = startDate.getDay();
			const hour = startDate.getHours();
			if (day === 6 && hour >= 0 && hour < 5) {
				return 'Friday Night';
			}
			if (day === 0 && hour >= 0 && hour < 5) {
				return 'Saturday Night';
			}
			return '';
		} catch (e) {
			return '';
		}
	}

	// Data Fetching
	async function fetchShiftDetails(startTime: string) {
		try {
			const response = await authenticatedFetch('/api/admin/schedules/all-slots');
			if (!response.ok) {
				throw new Error('Failed to fetch shift details');
			}
			const allShifts = (await response.json()) as AdminShiftSlot[];
			return allShifts.find((shift) => shift.start_time === startTime) || null;
		} catch (error) {
			console.error('Error fetching shift details:', error);
			throw error;
		}
	}

	async function fetchDashboardShifts() {
		try {
			const now = new Date();
			const futureDate = new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000);

			const params = new URLSearchParams();
			params.append('from', now.toISOString());
			params.append('to', futureDate.toISOString());

			const response = await authenticatedFetch(
				`/api/admin/schedules/all-slots?${params.toString()}`
			);
			if (!response.ok) {
				throw new Error('Failed to fetch calendar shifts');
			}
			return response.json() as Promise<AdminShiftSlot[]>;
		} catch (error) {
			console.error('Error fetching calendar shifts:', error);
			throw error;
		}
	}

	async function fetchUsers() {
		try {
			const response = await authenticatedFetch('/api/admin/users');
			if (!response.ok) {
				throw new Error('Failed to fetch users');
			}
			return response.json() as Promise<UserData[]>;
		} catch (error) {
			toast.error('Failed to load users');
			console.error('Fetch users error:', error);
			throw error;
		}
	}

	// Data Processing Functions
	function processScheduleData(shifts: AdminShiftSlot[]) {
		const scheduleStats = new Map<string, { total: number; filled: number; name: string }>();

		shifts.forEach((shift) => {
			const key = shift.schedule_id.toString();
			if (!scheduleStats.has(key)) {
				scheduleStats.set(key, { total: 0, filled: 0, name: shift.schedule_name });
			}
			const stats = scheduleStats.get(key)!;
			stats.total += 1;
			if (shift.is_booked) stats.filled += 1;
		});

		return Array.from(scheduleStats.entries())
			.map(([id, stats]) => ({
				schedule: stats.name,
				total: stats.total,
				filled: stats.filled,
				fillRate: stats.total > 0 ? Math.round((stats.filled / stats.total) * 100) : 0
			}))
			.sort((a, b) => b.total - a.total);
	}

	function processTimeSlotData(shifts: AdminShiftSlot[]) {
		const timeSlots = new Map<string, { total: number; filled: number }>();

		shifts.forEach((shift) => {
			const start = new Date(shift.start_time);
			const hour = start.getUTCHours();
			const timeSlotLabels = [
				'00:00-02:00',
				'02:00-04:00',
				'04:00-06:00',
				'06:00-08:00',
				'08:00-10:00',
				'10:00-12:00',
				'12:00-14:00',
				'14:00-16:00',
				'16:00-18:00',
				'18:00-20:00',
				'20:00-22:00',
				'22:00-24:00'
			];
			const slotIndex = Math.floor(hour / 2);
			const slot = timeSlotLabels[slotIndex] || timeSlotLabels[timeSlotLabels.length - 1];

			if (!timeSlots.has(slot)) {
				timeSlots.set(slot, { total: 0, filled: 0 });
			}
			const stats = timeSlots.get(slot)!;
			stats.total += 1;
			if (shift.is_booked) stats.filled += 1;
		});

		return Array.from(timeSlots.entries())
			.map(([timeSlot, stats]) => ({
				timeSlot,
				total: stats.total,
				fillRate: stats.total > 0 ? Math.round((stats.filled / stats.total) * 100) : 0
			}))
			.filter((item) => item.total > 0)
			.sort((a, b) => a.timeSlot.localeCompare(b.timeSlot));
	}

	// Queries
	const shiftDetailsQuery = $derived(
		createQuery<AdminShiftSlot | null, Error>({
			queryKey: ['shiftDetails', shiftStartTimeFromUrl || ''],
			queryFn: () => fetchShiftDetails(shiftStartTimeFromUrl!),
			enabled: !!shiftStartTimeFromUrl
		})
	);

	const dashboardShiftsQuery = $derived(
		createQuery<AdminShiftSlot[], Error>({
			queryKey: ['dashboardShifts'],
			queryFn: fetchDashboardShifts,
			enabled: !shiftStartTimeFromUrl
		})
	);

	const usersQuery = $derived.by(() => {
		return createQuery<UserData[], Error>({
			queryKey: ['allAdminUsersForBooking'],
			queryFn: fetchUsers,
			enabled: isBookingFormEnabled
		});
	});

	// Fetch schedules for management section
	const schedulesQuery = $derived(
		createQuery<Schedule[], Error>({
			queryKey: ['adminSchedulesForShifts'],
			queryFn: async () => {
				const response = await authenticatedFetch('/api/admin/schedules');
				if (!response.ok) {
					throw new Error('Failed to fetch schedules');
				}
				return response.json();
			}
		})
	);

	// Chart configurations
	const chartConfig = {
		filled: { label: 'Filled', color: 'var(--color-chart-1)' },
		available: { label: 'Available', color: 'var(--color-chart-2)' },
		total: { label: 'Total Shifts', color: 'var(--color-chart-1)' },
		fillRate: { label: 'Fill Rate', color: 'var(--color-chart-3)' }
	} satisfies Chart.ChartConfig;

	// Dashboard metrics
	const dashboardMetrics = $derived.by(() => {
		const shifts = $dashboardShiftsQuery.data ?? [];
		if (shifts.length === 0) return null;

		const totalShifts = shifts.length;
		const filledShifts = shifts.filter((s) => s.is_booked).length;
		const availableShifts = totalShifts - filledShifts;
		const fillRate = totalShifts > 0 ? Math.round((filledShifts / totalShifts) * 100) : 0;

		return {
			totalShifts,
			filledShifts,
			availableShifts,
			fillRate,
			scheduleData: processScheduleData(shifts),
			timeSlotData: processTimeSlotData(shifts),
			fillRateData: [
				{ label: 'Filled', value: filledShifts },
				{ label: 'Available', value: availableShifts }
			]
		};
	});

	// Calendar functions
	function prevMonth() {
		currentDisplayMonth = currentDisplayMonth.add({ months: -1 });
	}

	function nextMonth() {
		currentDisplayMonth = currentDisplayMonth.add({ months: 1 });
	}

	const shiftDatesForCalendar = $derived.by(() => {
		const shifts = $dashboardShiftsQuery.data ?? [];
		const uniqueDates = new Set<string>();
		shifts.forEach((shift: AdminShiftSlot) => {
			const dateStr = shift.start_time.split('T')[0];
			uniqueDates.add(dateStr);
		});
		return Array.from(uniqueDates).map((dateStr) => {
			const [year, month, day] = dateStr.split('-').map(Number);
			return new CalendarDate(year, month, day);
		});
	});

	// Booking Mutation
	type AdminBookingPayload = {
		schedule_id: number;
		start_time: string;
		user_id: number;
	};

	type MutationResponse = {
		success: boolean;
		message?: string;
	};

	const bookShiftMutation = createMutation<MutationResponse, Error, AdminBookingPayload>({
		mutationFn: async (bookingData: AdminBookingPayload) => {
			const response = await authenticatedFetch('/api/admin/bookings/assign', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(bookingData)
			});
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: 'Failed to assign shift' }));
				throw new Error(errorData.message || `HTTP error ${response.status}`);
			}
			return response.json();
		},
		onSuccess: () => {
			toast.success('Shift assigned to user successfully!');
			queryClient.invalidateQueries({ queryKey: ['shiftDetails'] });
			selectedUserIdForBooking = undefined;
		},
		onError: (error: Error) => {
			toast.error(`Assignment failed: ${error.message}`);
			bookingFormError = error.message;
		}
	});

	function handleBookShift(event: SubmitEvent) {
		event.preventDefault();
		bookingFormError = null;
		const userIdToBook = selectedUserIdForBooking ? parseInt(selectedUserIdForBooking) : undefined;

		if (!selectedShift || !userIdToBook) {
			bookingFormError = 'Please select a user.';
			return;
		}
		$bookShiftMutation.mutate({
			schedule_id: selectedShift.schedule_id,
			start_time: selectedShift.start_time,
			user_id: userIdToBook
		});
	}

	// Schedule management functions
	function openCreateScheduleDialog() {
		selectedScheduleForEdit = null;
		scheduleDialogMode = 'create';
		showScheduleDialog = true;
	}

	function openEditScheduleDialog(schedule: Schedule) {
		selectedScheduleForEdit = schedule;
		scheduleDialogMode = 'edit';
		showScheduleDialog = true;
	}

	function closeScheduleDialog() {
		showScheduleDialog = false;
		selectedScheduleForEdit = null;
	}

	// Effects
	$effect(() => {
		if ($shiftDetailsQuery.data) {
			selectedShift = $shiftDetailsQuery.data;
		} else if (!shiftStartTimeFromUrl) {
			selectedShift = null;
		}
	});

	$effect(() => {
		if ($usersQuery.data) {
			// Add any additional logic here if needed when usersQuery.data is available
		}
	});
</script>

<svelte:head>
	<title
		>Admin - {selectedShift ? `Shift: ${selectedShift.schedule_name}` : 'Shifts Dashboard'}</title
	>
</svelte:head>

{#if shiftStartTimeFromUrl}
	<!-- Shift Detail View -->
	{#if $shiftDetailsQuery.isLoading}
		<div class="p-6">
			<Skeleton class="h-8 w-64 mb-4" />
			<Skeleton class="h-4 w-48 mb-6" />
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div>
					<Skeleton class="h-6 w-32 mb-4" />
					<div class="space-y-2">
						<Skeleton class="h-4 w-full" />
						<Skeleton class="h-4 w-3/4" />
						<Skeleton class="h-4 w-1/2" />
					</div>
				</div>
				<div>
					<Skeleton class="h-6 w-32 mb-4" />
					<Skeleton class="h-10 w-full mb-2" />
					<Skeleton class="h-10 w-full" />
				</div>
			</div>
		</div>
	{:else if $shiftDetailsQuery.isError}
		<div class="p-6">
			<div class="text-center">
				<h1 class="text-xl font-semibold text-destructive mb-2">Error Loading Shift</h1>
				<p class="text-muted-foreground">
					{$shiftDetailsQuery.error.message}
				</p>
			</div>
		</div>
	{:else if !selectedShift}
		<div class="p-6">
			<div class="text-center">
				<h1 class="text-xl font-semibold mb-2">Shift Not Found</h1>
				<p class="text-muted-foreground">The requested shift could not be found.</p>
			</div>
		</div>
	{:else}
		<!-- Shift Details -->
		<div class="p-6">
			<div class="max-w-7xl mx-auto">
				<h1 class="text-2xl font-semibold mb-2">{selectedShift.schedule_name}</h1>
				<p class="text-muted-foreground mb-6">
					{formatTimeSlot(selectedShift.start_time, selectedShift.end_time)}
					({formatRelativeTime(selectedShift.start_time)})
				</p>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<!-- Shift Information -->
					<div class="space-y-4">
						<h2 class="text-lg font-medium">Shift Details</h2>
						<div class="space-y-3">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium">Status:</span>
								{#if selectedShift.is_booked}
									<span class="inline-flex items-center gap-1 text-green-600 font-semibold">
										<CheckIcon class="h-4 w-4" />
										Filled
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 text-orange-600 font-semibold">
										<XIcon class="h-4 w-4" />
										Available
									</span>
								{/if}
							</div>

							{#if selectedShift.is_booked && selectedShift.user_name}
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium">Assigned to:</span>
									<span class="text-sm">{selectedShift.user_name}</span>
									{#if selectedShift.user_phone}
										<span class="text-xs text-muted-foreground">({selectedShift.user_phone})</span>
									{/if}
								</div>
							{/if}

							{#if getAkaDescription(selectedShift.start_time)}
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium">AKA:</span>
									<span class="text-sm">{getAkaDescription(selectedShift.start_time)}</span>
								</div>
							{/if}
						</div>
					</div>

					<!-- Booking Form -->
					{#if !selectedShift.is_booked}
						<div class="space-y-4">
							<h2 class="text-lg font-medium">Assign Shift</h2>
							{#if $usersQuery.isLoading}
								<p class="text-sm text-muted-foreground">Loading users...</p>
							{:else if $usersQuery.isError}
								<p class="text-sm text-destructive">
									Error loading users: {$usersQuery.error.message}
								</p>
							{:else if $usersQuery.data}
								<form onsubmit={handleBookShift} class="space-y-4">
									<div>
										<Label for="user-select" class="text-sm font-medium">Select User</Label>
										<Select.Root
											type="single"
											value={selectedUserIdForBooking}
											onValueChange={(val?: string) => {
												selectedUserIdForBooking = val;
												bookingFormError = null;
											}}
										>
											<Select.Trigger class="w-full mt-1" placeholder="Choose a user"
											></Select.Trigger>
											<Select.Content>
												{#each $usersQuery.data as user (user.id)}
													<Select.Item value={user.id.toString()} label={user.name || user.phone}>
														{user.name || user.phone}
													</Select.Item>
												{/each}
											</Select.Content>
										</Select.Root>
									</div>

									{#if bookingFormError}
										<p class="text-sm text-destructive">{bookingFormError}</p>
									{/if}

									<Button type="submit" disabled={$bookShiftMutation.isPending} class="w-full">
										{#if $bookShiftMutation.isPending}
											Assigning...
										{:else}
											Assign User to Shift
										{/if}
									</Button>
								</form>
							{/if}
						</div>
					{:else}
						<div class="space-y-4">
							<h2 class="text-lg font-medium">Shift Management</h2>
							<p class="text-sm text-muted-foreground">
								This shift is currently assigned. Future options for reassignment or cancellation
								will appear here.
							</p>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
{:else}
	<!-- Dashboard View -->
	<ShiftsDashboard 
		isLoading={$dashboardShiftsQuery.isLoading}
		isError={$dashboardShiftsQuery.isError}
		error={$dashboardShiftsQuery.error || undefined}
		metrics={dashboardMetrics}
	/>
{/if}
