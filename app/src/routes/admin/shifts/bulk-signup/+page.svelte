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
	import { CheckIcon, ChevronsUpDownIcon, CalendarDaysIcon, UsersIcon } from 'lucide-svelte';
	import { tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';

	// Utilities with new patterns
	import { apiGet, apiPost } from '$lib/utils/api';
	import { classifyError } from '$lib/utils/errors';

	import { formatDistanceToNow, format } from 'date-fns';

	// Types - using domain User type but keeping AdminShiftSlot for now
	import type { User } from '$lib/types/domain';
	import type { components } from '$lib/types/api';
	import { mapAPIUserArrayToDomain } from '$lib/types/api-mappings';
	import type { AdminShiftSlot } from '$lib/types';
	import { formatShiftTitle, formatTimeSlot } from '$lib/utils/shiftFormatting';

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

	// Fetch users using our new API utilities
	const usersQuery = createQuery<User[], Error>({
		queryKey: ['adminUsers'],
		queryFn: async () => {
			try {
				const apiUsers =
					await apiGet<components['schemas']['api.UserAPIResponse'][]>('/api/admin/users');
				return mapAPIUserArrayToDomain(apiUsers);
			} catch (error) {
				throw classifyError(error);
			}
		}
	});

	// Fetch shift slots with date range using our new API utilities
	const shiftsQuery = createQuery<AdminShiftSlot[], Error>({
		queryKey: ['adminShiftSlots', 'bulk-assignment'],
		queryFn: async () => {
			try {
				// Calculate date range on each query
				let fromDate: string;
				let toDate: string;

				if (dateRangeStart) {
					fromDate = new Date(dateRangeStart + 'T00:00:00Z').toISOString();
					console.log('Using selected start date:', dateRangeStart, '→', fromDate);
				} else {
					fromDate = new Date().toISOString();
					console.log('Using default start date (now):', fromDate);
				}

				if (dateRangeEnd) {
					toDate = new Date(dateRangeEnd + 'T23:59:59Z').toISOString();
					console.log('Using selected end date:', dateRangeEnd, '→', toDate);
				} else {
					const futureDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000);
					toDate = futureDate.toISOString();
					console.log('Using default end date (30 days):', toDate);
				}

				const params = { from: fromDate, to: toDate };

				console.log('Fetching shifts with URL:', `/api/admin/schedules/all-slots`);

				const data = await apiGet<AdminShiftSlot[]>('/api/admin/schedules/all-slots', { params });
				console.log('Received shifts data:', data.length, 'shifts');
				return data;
			} catch (error) {
				throw classifyError(error);
			}
		},
		staleTime: 1000 * 60 * 5
	});

	// Bulk assignment mutation using our new API utilities
	const bulkAssignMutation = createMutation({
		mutationFn: async (assignments: Array<{ scheduleId: number; startTime: string }>) => {
			const results = [];
			for (const assignment of assignments) {
				try {
					const booking = await apiPost('admin/bookings/assign', {
						user_id: parseInt(selectedUserId),
						schedule_id: assignment.scheduleId,
						start_time: assignment.startTime
					});
					results.push({ success: true, booking });
				} catch (error) {
					const appError = classifyError(error);
					results.push({
						success: false,
						error: appError.message || 'Failed to create booking',
						startTime: assignment.startTime
					});
				}
			}
			return results;
		},
		onSuccess: (results) => {
			// Invalidate all adminShiftSlots queries regardless of date parameters
			queryClient.invalidateQueries({
				queryKey: ['adminShiftSlots'],
				exact: false // This allows matching queries with additional parameters
			});
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
			const appError = classifyError(error);
			formError = appError.message;
		}
	});

	// Derived values
	const users = $derived($usersQuery.data ?? []);
	const shifts = $derived(($shiftsQuery.data ?? []) as AdminShiftSlot[]);
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
			// Split only on the first hyphen to preserve the full timestamp
			const firstHyphenIndex = shiftKey.indexOf('-');
			const scheduleId = shiftKey.substring(0, firstHyphenIndex);
			const startTime = shiftKey.substring(firstHyphenIndex + 1);

			return {
				scheduleId: parseInt(scheduleId),
				startTime: startTime
			};
		});

		$bulkAssignMutation.mutate(assignments);
	}

	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		console.log('Date range changed:', range);
		dateRangeStart = range.start;
		dateRangeEnd = range.end;

		// If in pattern mode, re-select matching shifts with new date range
		if (patternMode && selectedPattern) {
			selectAllMatchingShifts();
		}

		// Force query refetch with new date range
		$shiftsQuery.refetch();
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
						<div data-testid="date-range-picker">
							<DateRangePicker
								initialStartDate={dateRangeStart}
								initialEndDate={dateRangeEnd}
								change={handleDateRangeChange}
								placeholderText="Next 30 days"
							/>
						</div>
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
						<div class="border rounded-lg p-3 border-primary/30 bg-primary/10">
							<div class="text-sm font-medium text-card-foreground mb-1">Selected Pattern</div>
							<div class="text-sm text-card-foreground/80">{getPatternDescription()}</div>
							<div class="text-xs text-muted-foreground mt-1">
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
					<div class="space-y-6" data-testid="shifts-list">
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
											class="group border rounded-lg p-3 transition-all duration-200 cursor-pointer {isSelected
												? 'border-primary bg-primary/10 shadow-primary/5'
												: isPatternMatch
													? 'border-primary/40 bg-primary/5'
													: shift.is_booked
														? 'border-border/50 bg-muted/50 opacity-75'
														: 'border-border/50 hover:bg-accent/50 hover:border-border'}"
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
																<div class="w-3 h-3 bg-primary rounded-full"></div>
															{:else}
																<div class="w-3 h-3 border-2 border-border rounded-full"></div>
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
													<div
														class="font-medium text-sm text-card-foreground truncate group-hover:text-accent-foreground transition-colors"
													>
														{formatShiftTitle(shift.start_time, shift.end_time)}
													</div>
													<div class="flex items-center gap-2 mt-1">
														<Badge variant="secondary" class="text-xs">{shift.schedule_name}</Badge>
													</div>

													{#if patternMode && !shift.is_booked}
														<div class="text-xs text-primary mt-1">
															{new Date(shift.start_time).toLocaleDateString('en-US', {
																weekday: 'short'
															})}
															{#if isPatternMatch}
																<Badge
																	variant="secondary"
																	class="text-xs ml-1 bg-primary/20 text-primary border-primary/30"
																	>Pattern Match</Badge
																>
															{/if}
														</div>
													{/if}

													{#if shift.is_booked}
														<div
															class="text-xs flex items-center gap-1 mt-1"
															style="color: hsl(var(--safety-green))"
														>
															<UsersIcon class="h-3 w-3" />
															{shift.user_name || 'Assigned'}
															{#if shift.is_recurring_reservation}
																<Badge
																	variant="secondary"
																	class="text-xs bg-primary/20 text-primary border-primary/30"
																	>Recurring</Badge
																>
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
