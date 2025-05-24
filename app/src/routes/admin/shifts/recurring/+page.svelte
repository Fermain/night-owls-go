<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Command from '$lib/components/ui/command';
	import * as Popover from '$lib/components/ui/popover';
	import {
		CheckIcon,
		ChevronsUpDownIcon,
		CalendarClockIcon,
		PlusIcon,
		TrashIcon
	} from 'lucide-svelte';
	import { tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { UserData } from '$lib/schemas/user';
	import type { AdminShiftSlot } from '$lib/types';

	const queryClient = useQueryClient();

	// Form state
	let showCreateForm = $state(false);
	let selectedUserId = $state('');
	let userComboOpen = $state(false);
	let userTriggerRef = $state<HTMLButtonElement>(null!);
	let buddyName = $state('');
	let selectedDay = $state<string>('');
	let selectedSchedule = $state<string>('');
	let selectedTimeSlot = $state<string>('');
	let description = $state('');
	let formError = $state<string | null>(null);

	// Fetch users
	const usersQuery = createQuery<UserData[], Error>({
		queryKey: ['adminUsers'],
		queryFn: async () => {
			const response = await authenticatedFetch('/api/admin/users');
			if (!response.ok) throw new Error('Failed to fetch users');
			return response.json();
		}
	});

	// Fetch schedules
	const schedulesQuery = createQuery<{ schedule_id: number; name: string }[], Error>({
		queryKey: ['adminSchedules'],
		queryFn: async () => {
			const response = await authenticatedFetch('/api/admin/schedules');
			if (!response.ok) throw new Error('Failed to fetch schedules');
			return response.json();
		}
	});

	// Fetch shift slots to analyze patterns
	const shiftsQuery = createQuery<AdminShiftSlot[], Error>({
		queryKey: ['adminShiftSlots'],
		queryFn: async () => {
			const response = await authenticatedFetch('/api/admin/schedules/all-slots');
			if (!response.ok) throw new Error('Failed to fetch shifts');
			return response.json();
		}
	});

	// Fetch recurring assignments
	const recurringAssignmentsQuery = createQuery<any[], Error>({
		queryKey: ['adminRecurringAssignments'],
		queryFn: async () => {
			const response = await authenticatedFetch('/api/admin/recurring-assignments');
			if (!response.ok) throw new Error('Failed to fetch recurring assignments');
			return response.json();
		}
	});

	// Create recurring assignment mutation
	const createRecurringAssignmentMutation = createMutation({
		mutationFn: async (data: any) => {
			const response = await authenticatedFetch('/api/admin/recurring-assignments', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(data)
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to create recurring assignment');
			}
			return response.json();
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['adminRecurringAssignments'] });
			queryClient.invalidateQueries({ queryKey: ['adminShiftSlots'] });
			resetForm();
		},
		onError: (error: Error) => {
			formError = error.message;
		}
	});

	// Delete recurring assignment mutation
	const deleteRecurringAssignmentMutation = createMutation({
		mutationFn: async (id: number) => {
			const response = await authenticatedFetch(`/api/admin/recurring-assignments/${id}`, {
				method: 'DELETE'
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to delete recurring assignment');
			}
			return response.json();
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['adminRecurringAssignments'] });
			queryClient.invalidateQueries({ queryKey: ['adminShiftSlots'] });
		}
	});

	// Derived values
	const users = $derived($usersQuery.data ?? []);
	const schedules = $derived($schedulesQuery.data ?? []);
	const shifts = $derived($shiftsQuery.data ?? []);
	const recurringAssignments = $derived($recurringAssignmentsQuery.data ?? []);
	const selectedUser = $derived(users.find((u) => u.id.toString() === selectedUserId));

	// Generate day options from actual shift data using "Previous Day Night" convention
	const dayOptions = $derived.by(() => {
		const days = new Set<string>();
		shifts.forEach((shift) => {
			const date = new Date(shift.start_time);
			const dayOfWeek = date.getUTCDay();

			// Get the previous day for the "Night" convention
			const previousDayOfWeek = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
			const dayNames = [
				'Sunday',
				'Monday',
				'Tuesday',
				'Wednesday',
				'Thursday',
				'Friday',
				'Saturday'
			];
			const previousDayName = dayNames[previousDayOfWeek];

			// Use the pattern "Friday Night" for Saturday morning shifts
			const displayLabel = `${previousDayName} Night`;
			days.add(`${dayOfWeek}:${displayLabel}`);
		});

		return Array.from(days)
			.map((dayString) => {
				const [value, label] = dayString.split(':');
				return { value, label };
			})
			.sort((a, b) => parseInt(a.value) - parseInt(b.value));
	});

	// Generate time slot options from actual shift data
	const timeSlotOptions = $derived.by(() => {
		const timeSlots = new Set<string>();
		shifts.forEach((shift) => {
			const startDate = new Date(shift.start_time);
			const endDate = new Date(shift.end_time);
			const startHour = startDate.getUTCHours().toString().padStart(2, '0');
			const startMin = startDate.getUTCMinutes().toString().padStart(2, '0');
			const endHour = endDate.getUTCHours().toString().padStart(2, '0');
			const endMin = endDate.getUTCMinutes().toString().padStart(2, '0');
			const timeSlot = `${startHour}:${startMin}-${endHour}:${endMin}`;
			timeSlots.add(timeSlot);
		});

		return Array.from(timeSlots)
			.map((slot) => ({ value: slot, label: slot }))
			.sort((a, b) => a.value.localeCompare(b.value));
	});

	function closeUserCombo() {
		userComboOpen = false;
		tick().then(() => userTriggerRef?.focus());
	}

	function handleCreateRecurring(event: SubmitEvent) {
		event.preventDefault();
		formError = null;

		if (!selectedUserId || !selectedDay || !selectedSchedule || !selectedTimeSlot) {
			formError = 'Please fill in all required fields';
			return;
		}

		const assignmentData = {
			user_id: parseInt(selectedUserId),
			buddy_name: buddyName.trim() || null,
			day_of_week: parseInt(selectedDay),
			schedule_id: parseInt(selectedSchedule),
			time_slot: selectedTimeSlot,
			description: description.trim() || null
		};

		$createRecurringAssignmentMutation.mutate(assignmentData);
	}

	function resetForm() {
		selectedUserId = '';
		buddyName = '';
		selectedDay = '';
		selectedSchedule = '';
		selectedTimeSlot = '';
		description = '';
		formError = null;
		showCreateForm = false;
	}

	function deleteRecurringAssignment(id: number) {
		if (confirm('Are you sure you want to delete this recurring assignment?')) {
			$deleteRecurringAssignmentMutation.mutate(id);
		}
	}
</script>

<svelte:head>
	<title>Admin - Recurring Shift Assignments</title>
</svelte:head>

<div class="p-6">
	<div class="max-w-4xl mx-auto">
		<!-- Header -->
		<div class="grid gap-4 mb-6">
			<div>
				<h1 class="text-2xl font-bold flex items-center gap-2">
					<CalendarClockIcon class="h-6 w-6" />
					Recurring Assignments
				</h1>
				<p class="text-muted-foreground">
					Set up automatic shift assignments for regular volunteers
				</p>
			</div>
			<div class="flex gap-2">
				<Button onclick={() => (showCreateForm = true)} disabled={showCreateForm}>
					<PlusIcon class="h-4 w-4 mr-2" />
					New Recurring Assignment
				</Button>
			</div>
		</div>

		<!-- Create Form -->
		{#if showCreateForm}
			<Card class="mb-6">
				<CardHeader>
					<CardTitle>Create Recurring Assignment</CardTitle>
					<CardDescription>
						Set up automatic assignment for shifts matching a pattern
					</CardDescription>
				</CardHeader>
				<CardContent>
					<form onsubmit={handleCreateRecurring} class="space-y-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<!-- Primary User -->
							<div class="space-y-2">
								<Label class="text-sm font-medium">
									Primary User <span class="text-red-500">*</span>
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

							<!-- Buddy -->
							<div class="space-y-2">
								<Label for="buddy" class="text-sm font-medium">Buddy (Optional)</Label>
								<Input
									id="buddy"
									bind:value={buddyName}
									placeholder="Enter buddy name"
									class="w-full"
								/>
							</div>

							<!-- Schedule -->
							<div class="space-y-2">
								<Label class="text-sm font-medium">
									Schedule <span class="text-red-500">*</span>
								</Label>
								<Select.Root type="single" bind:value={selectedSchedule}>
									<Select.Trigger class="w-full">
										{selectedSchedule
											? schedules.find((s) => s.schedule_id.toString() === selectedSchedule)
													?.name || 'Select schedule'
											: 'Select schedule'}
									</Select.Trigger>
									<Select.Content>
										{#each schedules as schedule (schedule.schedule_id)}
											<Select.Item value={schedule.schedule_id.toString()}>
												{schedule.name}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>

							<!-- Day of Week -->
							<div class="space-y-2">
								<Label class="text-sm font-medium">
									Day of Week <span class="text-red-500">*</span>
								</Label>
								<Select.Root type="single" bind:value={selectedDay}>
									<Select.Trigger class="w-full">
										{selectedDay
											? dayOptions.find((d) => d.value === selectedDay)?.label || 'Select day'
											: 'Select day'}
									</Select.Trigger>
									<Select.Content>
										{#each dayOptions as option (option.value)}
											<Select.Item value={option.value}>
												{option.label}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>

							<!-- Time Slot -->
							<div class="space-y-2">
								<Label class="text-sm font-medium">
									Time Slot <span class="text-red-500">*</span>
								</Label>
								<Select.Root type="single" bind:value={selectedTimeSlot}>
									<Select.Trigger class="w-full">
										{selectedTimeSlot || 'Select time slot'}
									</Select.Trigger>
									<Select.Content>
										{#each timeSlotOptions as option (option.value)}
											<Select.Item value={option.value}>
												{option.label}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>

							<!-- Description -->
							<div class="space-y-2">
								<Label for="description" class="text-sm font-medium">Description</Label>
								<Input
									id="description"
									bind:value={description}
									placeholder="Optional description"
									class="w-full"
								/>
							</div>
						</div>

						{#if formError}
							<div class="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
								<p class="text-sm text-destructive">{formError}</p>
							</div>
						{/if}

						<div class="flex gap-2">
							<Button type="submit" disabled={$createRecurringAssignmentMutation.isPending}>
								{#if $createRecurringAssignmentMutation.isPending}
									Creating...
								{:else}
									Create Assignment
								{/if}
							</Button>
							<Button type="button" variant="outline" onclick={resetForm}>Cancel</Button>
						</div>
					</form>
				</CardContent>
			</Card>
		{/if}

		<!-- Status -->
		<Card>
			<CardContent class="pt-6">
				{#if $recurringAssignmentsQuery.isLoading}
					<div class="text-center py-8">
						<p class="text-muted-foreground">Loading recurring assignments...</p>
					</div>
				{:else if $recurringAssignmentsQuery.isError}
					<div class="text-center py-8">
						<p class="text-destructive">
							Error loading recurring assignments: {$recurringAssignmentsQuery.error.message}
						</p>
					</div>
				{:else if recurringAssignments.length === 0}
					<div class="text-center py-8">
						<CalendarClockIcon class="h-12 w-12 mx-auto text-muted-foreground mb-4" />
						<p class="text-muted-foreground">No recurring assignments configured</p>
						<p class="text-sm text-muted-foreground">
							Found {dayOptions.length} unique days and {timeSlotOptions.length} time slots available
							for recurring assignments
						</p>
					</div>
				{:else}
					<div class="space-y-4">
						<h3 class="text-lg font-semibold">Active Recurring Assignments</h3>
						{#each recurringAssignments as assignment (assignment.recurring_assignment_id)}
							<div class="border rounded-lg p-4 flex items-center justify-between">
								<div>
									<div class="font-medium">
										{users.find((u) => u.id === assignment.user_id)?.name ||
											users.find((u) => u.id === assignment.user_id)?.phone ||
											'Unknown User'}
									</div>
									<div class="text-sm text-muted-foreground">
										{dayOptions.find((d) => d.value === assignment.day_of_week.toString())?.label ||
											'Unknown Day'} •
										{assignment.time_slot} •
										{schedules.find((s) => s.schedule_id === assignment.schedule_id)?.name ||
											'Unknown Schedule'}
									</div>
									{#if assignment.buddy_name && assignment.buddy_name.Valid && assignment.buddy_name.String}
										<div class="text-sm text-muted-foreground">
											With buddy: {assignment.buddy_name.String}
										</div>
									{/if}
									{#if assignment.description && assignment.description.Valid && assignment.description.String}
										<div class="text-sm text-muted-foreground">
											{assignment.description.String}
										</div>
									{/if}
								</div>
								<Button
									variant="destructive"
									size="sm"
									onclick={() => deleteRecurringAssignment(assignment.recurring_assignment_id)}
									disabled={$deleteRecurringAssignmentMutation.isPending}
								>
									<TrashIcon class="h-4 w-4" />
								</Button>
							</div>
						{/each}
					</div>
				{/if}
			</CardContent>
		</Card>
	</div>
</div>
