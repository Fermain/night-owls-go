<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Command from '$lib/components/ui/command';
	import * as Popover from '$lib/components/ui/popover';
	import { Badge } from '$lib/components/ui/badge';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import {
		CheckIcon,
		ChevronsUpDownIcon,
		CalendarDaysIcon,
		UsersIcon,
		ClockIcon
	} from 'lucide-svelte';
	import { tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { authenticatedFetch } from '$lib/utils/api';
	import { formatDistanceToNow, format } from 'date-fns';
	import type { UserData } from '$lib/schemas/user';
	import type { AdminShiftSlot } from '$lib/types';

	const queryClient = useQueryClient();

	// Form state
	let selectedUserId = $state('');
	let userComboOpen = $state(false);
	let userTriggerRef = $state<HTMLButtonElement>(null!);
	let buddyName = $state('');
	let selectedShifts = $state<Set<string>>(new Set());
	let dateRangeStart = $state<string | null>(null);
	let dateRangeEnd = $state<string | null>(null);
	let showOnlyAvailable = $state(true);
	let formError = $state<string | null>(null);
	let patternMode = $state(false);
	let selectedPattern = $state<{
		scheduleName: string;
		dayOfWeek: number;
		timeSlot: string;
		scheduleId: number;
	} | null>(null);

	// Fetch users
	const usersQuery = createQuery<UserData[], Error>({
		queryKey: ['adminUsers'],
		queryFn: async () => {
			const response = await authenticatedFetch('/api/admin/users');
			if (!response.ok) throw new Error('Failed to fetch users');
			return response.json();
		}
	});

	// Fetch shift slots with date range
	const shiftsQuery = $derived.by(() => {
		let fromDate: string | undefined;
		let toDate: string | undefined;

		if (dateRangeStart && dateRangeEnd) {
			fromDate = new Date(dateRangeStart + 'T00:00:00Z').toISOString();
			toDate = new Date(dateRangeEnd + 'T23:59:59Z').toISOString();
		} else {
			// Default to next 30 days
			const now = new Date();
			const futureDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
			fromDate = now.toISOString();
			toDate = futureDate.toISOString();
		}

		return createQuery<AdminShiftSlot[], Error>({
			queryKey: ['adminShiftSlots', fromDate, toDate],
			queryFn: async () => {
				const params = new URLSearchParams();
				if (fromDate) params.append('from', fromDate);
				if (toDate) params.append('to', toDate);

				const response = await authenticatedFetch(
					`/api/admin/schedules/all-slots?${params.toString()}`
				);
				if (!response.ok) throw new Error('Failed to fetch shifts');
				return response.json();
			},
			staleTime: 1000 * 60 * 5
		});
	});

	// Bulk assignment mutation
	const bulkAssignMutation = createMutation({
		mutationFn: async (assignments: Array<{ scheduleId: number; startTime: string }>) => {
			const results = [];
			for (const assignment of assignments) {
				const response = await authenticatedFetch('/api/admin/bookings/assign', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({
						user_id: parseInt(selectedUserId),
						schedule_id: assignment.scheduleId,
						start_time: assignment.startTime
					})
				});

				if (!response.ok) {
					const errorData = await response.json();
					results.push({
						success: false,
						error: errorData.error || 'Failed to create booking',
						startTime: assignment.startTime
					});
				} else {
					const booking = await response.json();
					results.push({ success: true, booking });
				}
			}
			return results;
		},
		onSuccess: (results) => {
			queryClient.invalidateQueries({ queryKey: ['adminShiftSlots'] });
			const successCount = results.filter((r) => r.success).length;
			const errorCount = results.filter((r) => !r.success).length;

			if (errorCount === 0) {
				formError = null;
				selectedShifts.clear();
				selectedShifts = new Set(); // Trigger reactivity
			} else {
				formError = `${successCount} shifts assigned successfully, ${errorCount} failed`;
			}
		},
		onError: (error: Error) => {
			formError = error.message;
		}
	});

	// Derived values
	const users = $derived($usersQuery.data ?? []);
	const shifts = $derived($shiftsQuery.data ?? []);
	const selectedUser = $derived(users.find((u) => u.id.toString() === selectedUserId));

	// Filtered and grouped shifts
	const availableShifts = $derived.by(() => {
		return shifts
			.filter((shift) => {
				if (showOnlyAvailable && shift.is_booked) return false;
				return new Date(shift.start_time) > new Date(); // Only future shifts
			})
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
	});

	// Group shifts by date for better organization
	const groupedShifts = $derived.by(() => {
		const groups = new Map<string, AdminShiftSlot[]>();
		availableShifts.forEach((shift) => {
			const date = format(new Date(shift.start_time), 'yyyy-MM-dd');
			if (!groups.has(date)) {
				groups.set(date, []);
			}
			groups.get(date)!.push(shift);
		});
		return Array.from(groups.entries()).map(([date, shifts]) => ({
			date,
			displayDate: format(new Date(date), 'EEEE, MMMM d, yyyy'),
			shifts
		}));
	});

	function closeUserCombo() {
		userComboOpen = false;
		tick().then(() => userTriggerRef?.focus());
	}

	function toggleShift(shiftKey: string) {
		if (selectedShifts.has(shiftKey)) {
			selectedShifts.delete(shiftKey);
		} else {
			selectedShifts.add(shiftKey);
		}
		selectedShifts = new Set(selectedShifts); // Trigger reactivity
	}

	function toggleAllShiftsForDate(dateShifts: AdminShiftSlot[]) {
		const allSelected = dateShifts.every((shift) =>
			selectedShifts.has(`${shift.schedule_id}-${shift.start_time}`)
		);

		if (allSelected) {
			// Deselect all for this date
			dateShifts.forEach((shift) => {
				selectedShifts.delete(`${shift.schedule_id}-${shift.start_time}`);
			});
		} else {
			// Select all available for this date
			dateShifts
				.filter((shift) => !shift.is_booked)
				.forEach((shift) => {
					selectedShifts.add(`${shift.schedule_id}-${shift.start_time}`);
				});
		}
		selectedShifts = new Set(selectedShifts); // Trigger reactivity
	}

	function handleBulkAssign() {
		formError = null;

		if (!selectedUserId) {
			formError = 'Please select a user';
			return;
		}

		if (selectedShifts.size === 0) {
			formError = 'Please select at least one shift';
			return;
		}

		const assignments = Array.from(selectedShifts).map((shiftKey) => {
			const [scheduleId, startTime] = shiftKey.split('-', 2);
			return {
				scheduleId: parseInt(scheduleId),
				startTime: startTime
			};
		});

		$bulkAssignMutation.mutate(assignments);
	}

	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;

		// If in pattern mode, re-select matching shifts with new date range
		if (patternMode && selectedPattern) {
			selectAllMatchingShifts();
		}
	}

	function clearSelection() {
		selectedShifts.clear();
		selectedShifts = new Set();
		selectedPattern = null;
	}

	function selectPattern(shift: AdminShiftSlot) {
		const dayOfWeek = new Date(shift.start_time).getDay();
		const timeSlot = formatTimeSlot(shift.start_time, shift.end_time);

		selectedPattern = {
			scheduleName: shift.schedule_name,
			dayOfWeek,
			timeSlot,
			scheduleId: shift.schedule_id
		};

		// Auto-select all matching shifts
		selectAllMatchingShifts();
	}

	function selectAllMatchingShifts() {
		if (!selectedPattern) return;

		selectedShifts.clear();

		availableShifts
			.filter((shift) => {
				if (shift.is_booked) return false;
				const shiftDay = new Date(shift.start_time).getDay();
				const shiftTimeSlot = formatTimeSlot(shift.start_time, shift.end_time);
				return (
					shift.schedule_id === selectedPattern!.scheduleId &&
					shiftDay === selectedPattern!.dayOfWeek &&
					shiftTimeSlot === selectedPattern!.timeSlot
				);
			})
			.forEach((shift) => {
				selectedShifts.add(`${shift.schedule_id}-${shift.start_time}`);
			});

		selectedShifts = new Set(selectedShifts); // Trigger reactivity
	}

	function getPatternDescription(): string {
		if (!selectedPattern) return '';

		const dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
		const dayName = dayNames[selectedPattern.dayOfWeek];

		return `Every ${dayName} ${selectedPattern.timeSlot} - ${selectedPattern.scheduleName}`;
	}

	function getMatchingShiftsCount(): number {
		if (!selectedPattern) return 0;

		return availableShifts.filter((shift) => {
			if (shift.is_booked) return false;
			const shiftDay = new Date(shift.start_time).getDay();
			const shiftTimeSlot = formatTimeSlot(shift.start_time, shift.end_time);
			return (
				shift.schedule_id === selectedPattern!.scheduleId &&
				shiftDay === selectedPattern!.dayOfWeek &&
				shiftTimeSlot === selectedPattern!.timeSlot
			);
		}).length;
	}

	function formatTimeSlot(startTime: string, endTime: string): string {
		const start = new Date(startTime);
		const end = new Date(endTime);
		return `${format(start, 'HH:mm')} - ${format(end, 'HH:mm')}`;
	}

	// Reactive effect to re-select pattern matches when data changes
	$effect(() => {
		if (patternMode && selectedPattern && availableShifts.length > 0) {
			selectAllMatchingShifts();
		}
	});
</script>

<svelte:head>
	<title>Admin - Bulk Shift Assignment</title>
</svelte:head>

<div class="p-6">
	<div class="max-w-6xl mx-auto">
		<!-- Header -->
		<div class="grid gap-4 mb-6">
			<div>
				<h1 class="text-2xl font-bold flex items-center gap-2">
					<CalendarDaysIcon class="h-6 w-6" />
					Bulk Shift Assignment
				</h1>
				<p class="text-muted-foreground">
					Select individual shifts or commit to a pattern (e.g., "every Saturday 0-2AM until end
					date")
				</p>
			</div>
		</div>

		<!-- Controls -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
			<!-- User Selection -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">User Selection</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="space-y-2">
						<Label class="text-sm font-medium">
							Assign to User <span class="text-red-500">*</span>
						</Label>
						<Popover.Root bind:open={userComboOpen}>
							<Popover.Trigger bind:ref={userTriggerRef}>
								{#snippet child({ props })}
									<Button
										variant="outline"
										class="w-full justify-between"
										{...props}
										role="combobox"
										aria-expanded={userComboOpen}
									>
										{selectedUser ? selectedUser.name || selectedUser.phone : 'Select user...'}
										<ChevronsUpDownIcon class="ml-2 h-4 w-4 shrink-0 opacity-50" />
									</Button>
								{/snippet}
							</Popover.Trigger>
							<Popover.Content class="w-full p-0">
								<Command.Root>
									<Command.Input placeholder="Search users..." />
									<Command.List>
										<Command.Empty>No users found</Command.Empty>
										<Command.Group>
											{#each users as user (user.id)}
												<Command.Item
													value={user.id.toString()}
													onSelect={() => {
														selectedUserId = user.id.toString();
														closeUserCombo();
													}}
												>
													<CheckIcon
														class={cn(
															'mr-2 h-4 w-4',
															selectedUserId !== user.id.toString() && 'text-transparent'
														)}
													/>
													<div>
														<div class="font-medium">{user.name || 'Unnamed'}</div>
														<div class="text-sm text-muted-foreground">{user.phone}</div>
													</div>
												</Command.Item>
											{/each}
										</Command.Group>
									</Command.List>
								</Command.Root>
							</Popover.Content>
						</Popover.Root>
					</div>

					<div class="space-y-2">
						<Label for="buddy" class="text-sm font-medium">Buddy (Optional)</Label>
						<Input
							id="buddy"
							bind:value={buddyName}
							placeholder="Enter buddy name"
							class="w-full"
						/>
					</div>
				</CardContent>
			</Card>

			<!-- Filters -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">Selection Mode</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="space-y-3">
						<div class="flex items-center space-x-2">
							<Checkbox id="pattern-mode" bind:checked={patternMode} />
							<Label for="pattern-mode" class="text-sm cursor-pointer font-medium">
								Pattern Selection Mode
							</Label>
						</div>
						<p class="text-xs text-muted-foreground">
							{patternMode
								? 'Click any shift to select all matching shifts until end date'
								: 'Manually select individual shifts with checkboxes'}
						</p>
					</div>

					<div class="space-y-2">
						<Label class="text-sm font-medium">Date Range</Label>
						<DateRangePicker
							initialStartDate={dateRangeStart}
							initialEndDate={dateRangeEnd}
							change={handleDateRangeChange}
							placeholderText="Next 30 days"
						/>
					</div>

					<div class="flex items-center space-x-2">
						<Checkbox id="available-only" bind:checked={showOnlyAvailable} />
						<Label for="available-only" class="text-sm cursor-pointer">Available shifts only</Label>
					</div>
				</CardContent>
			</Card>

			<!-- Selection Summary -->
			<Card>
				<CardHeader>
					<CardTitle class="text-lg">Selection Summary</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					{#if patternMode && selectedPattern}
						<div class="border rounded-lg p-3 bg-blue-50 border-blue-200">
							<div class="text-sm font-medium text-blue-800 mb-1">Selected Pattern</div>
							<div class="text-sm text-blue-700">{getPatternDescription()}</div>
							<div class="text-xs text-blue-600 mt-1">
								{getMatchingShiftsCount()} matching shifts until {dateRangeEnd || 'end of range'}
							</div>
						</div>
					{/if}

					<div class="text-center">
						<div class="text-3xl font-bold text-primary">{selectedShifts.size}</div>
						<div class="text-sm text-muted-foreground">shifts selected</div>
					</div>

					{#if selectedShifts.size > 0}
						<Button variant="outline" size="sm" onclick={clearSelection} class="w-full">
							{patternMode ? 'Clear Pattern' : 'Clear Selection'}
						</Button>
					{/if}

					<Button
						onclick={handleBulkAssign}
						disabled={$bulkAssignMutation.isPending || selectedShifts.size === 0 || !selectedUserId}
						class="w-full"
					>
						{#if $bulkAssignMutation.isPending}
							Assigning...
						{:else if patternMode && selectedPattern}
							Assign Pattern ({selectedShifts.size} shifts)
						{:else}
							Assign {selectedShifts.size} Shifts
						{/if}
					</Button>

					{#if formError}
						<div class="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
							<p class="text-sm text-destructive">{formError}</p>
						</div>
					{/if}
				</CardContent>
			</Card>
		</div>

		<!-- Shifts List -->
		<Card>
			<CardHeader>
				<CardTitle class="text-lg">Available Shifts</CardTitle>
				<CardDescription>
					{availableShifts.length} shifts available in the selected date range
				</CardDescription>
			</CardHeader>
			<CardContent>
				{#if $shiftsQuery.isLoading}
					<div class="text-center py-8">
						<p class="text-muted-foreground">Loading shifts...</p>
					</div>
				{:else if $shiftsQuery.isError}
					<div class="text-center py-8">
						<p class="text-destructive">Error loading shifts: {$shiftsQuery.error.message}</p>
					</div>
				{:else if availableShifts.length === 0}
					<div class="text-center py-8">
						<CalendarDaysIcon class="h-12 w-12 mx-auto text-muted-foreground mb-4" />
						<p class="text-muted-foreground">No available shifts found</p>
						<p class="text-sm text-muted-foreground">Try adjusting your date range or filters</p>
					</div>
				{:else}
					<div class="space-y-6">
						{#each groupedShifts as { date, displayDate, shifts } (date)}
							<div class="border rounded-lg p-4">
								<div class="flex items-center justify-between mb-4">
									<h3 class="font-semibold text-lg">{displayDate}</h3>
									<Button
										variant="outline"
										size="sm"
										onclick={() => toggleAllShiftsForDate(shifts)}
									>
										{shifts
											.filter((s) => !s.is_booked)
											.every((s) => selectedShifts.has(`${s.schedule_id}-${s.start_time}`))
											? 'Deselect All'
											: 'Select Available'}
									</Button>
								</div>

								<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
									{#each shifts as shift (shift.schedule_id + '-' + shift.start_time)}
										{@const shiftKey = `${shift.schedule_id}-${shift.start_time}`}
										{@const isSelected = selectedShifts.has(shiftKey)}
										{@const isPatternMatch =
											patternMode &&
											selectedPattern &&
											shift.schedule_id === selectedPattern.scheduleId &&
											new Date(shift.start_time).getDay() === selectedPattern.dayOfWeek &&
											formatTimeSlot(shift.start_time, shift.end_time) === selectedPattern.timeSlot}
										<div
											class="border rounded-lg p-3 transition-colors cursor-pointer {isSelected
												? 'border-primary bg-primary/5'
												: isPatternMatch
													? 'border-blue-300 bg-blue-50'
													: shift.is_booked
														? 'border-muted bg-muted/50'
														: 'border-border hover:bg-accent'}"
											role="button"
											tabindex={shift.is_booked ? -1 : 0}
											onclick={() => {
												if (shift.is_booked) return;
												if (patternMode) {
													selectPattern(shift);
												} else {
													toggleShift(shiftKey);
												}
											}}
											onkeydown={(event) => {
												if (shift.is_booked) return;
												if (event.key === 'Enter' || event.key === ' ') {
													event.preventDefault();
													if (patternMode) {
														selectPattern(shift);
													} else {
														toggleShift(shiftKey);
													}
												}
											}}
											aria-label={`${patternMode ? 'Select pattern for' : 'Toggle selection of'} ${shift.schedule_name} shift on ${formatTimeSlot(shift.start_time, shift.end_time)}`}
											aria-pressed={isSelected}
										>
											<div class="flex items-start gap-3">
												{#if !shift.is_booked}
													{#if patternMode}
														<div class="w-4 h-4 mt-1 flex items-center justify-center">
															{#if isPatternMatch}
																<div class="w-3 h-3 bg-blue-500 rounded-full"></div>
															{:else}
																<div class="w-3 h-3 border-2 border-gray-300 rounded-full"></div>
															{/if}
														</div>
													{:else}
														<Checkbox
															checked={isSelected}
															onCheckedChange={() => toggleShift(shiftKey)}
															class="mt-1"
														/>
													{/if}
												{:else}
													<div class="w-4 h-4 mt-1 rounded border bg-muted"></div>
												{/if}

												<div class="flex-1 min-w-0">
													<div class="font-medium text-sm truncate">{shift.schedule_name}</div>
													<div class="text-sm text-muted-foreground flex items-center gap-1">
														<ClockIcon class="h-3 w-3" />
														{formatTimeSlot(shift.start_time, shift.end_time)}
													</div>

													{#if patternMode && !shift.is_booked}
														<div class="text-xs text-blue-600 mt-1">
															{new Date(shift.start_time).toLocaleDateString('en-US', {
																weekday: 'short'
															})}
															{#if isPatternMatch}
																<Badge variant="secondary" class="text-xs ml-1">Pattern Match</Badge
																>
															{/if}
														</div>
													{/if}

													{#if shift.is_booked}
														<div class="text-xs text-green-700 flex items-center gap-1 mt-1">
															<UsersIcon class="h-3 w-3" />
															{shift.user_name || 'Assigned'}
															{#if shift.is_recurring_reservation}
																<Badge variant="secondary" class="text-xs">Recurring</Badge>
															{/if}
														</div>
													{:else if !patternMode}
														<div class="text-xs text-muted-foreground mt-1">
															{formatDistanceToNow(new Date(shift.start_time), { addSuffix: true })}
														</div>
													{/if}
												</div>
											</div>
										</div>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</CardContent>
		</Card>
	</div>
</div>
