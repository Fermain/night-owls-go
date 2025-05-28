<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import * as Command from '$lib/components/ui/command';
	import * as Popover from '$lib/components/ui/popover';
	import { CheckIcon, ChevronsUpDownIcon, UserIcon, Users2Icon, ClockIcon } from 'lucide-svelte';
	import { tick } from 'svelte';
	import { cn } from '$lib/utils';
	import { formatTimeSlot, formatRelativeTime } from '$lib/utils/dateFormatting';
	import type { AdminShiftSlot } from '$lib/types';
	import type { UserData } from '$lib/schemas/user';
	import { authenticatedFetch } from '$lib/utils/api';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';

	interface Props {
		selectedShift: AdminShiftSlot;
		onBookingSuccess: () => void;
	}

	let { selectedShift, onBookingSuccess }: Props = $props();

	const queryClient = useQueryClient();

	// State for user selection
	let primaryUserOpen = $state(false);
	let primaryUserValue = $state('');
	let primaryUserTriggerRef = $state<HTMLButtonElement>(null!);
	let buddyName = $state('');
	let assignmentError = $state<string | null>(null);
	let showReassignForm = $state(false);

	// Fetch users query
	const usersQuery = createQuery<UserData[], Error>({
		queryKey: ['allAdminUsersForBooking'],
		queryFn: async () => {
			const response = await authenticatedFetch('/api/admin/users');
			if (!response.ok) {
				throw new Error('Failed to fetch users');
			}
			return response.json();
		}
	});

	// Derived values
	const users = $derived($usersQuery.data ?? []);
	const selectedUser = $derived(users.find((u) => u.id.toString() === primaryUserValue));

	// Assignment mutation
	const assignShiftMutation = createMutation({
		mutationFn: async (assignmentData: {
			schedule_id: number;
			start_time: string;
			user_id: number;
			buddy_name?: string;
		}) => {
			// Note: Admin assignment may need to be extended to support buddy
			// For now, we'll send the basic assignment and buddy separately if needed
			const response = await authenticatedFetch('/api/admin/bookings/assign', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					schedule_id: assignmentData.schedule_id,
					start_time: assignmentData.start_time,
					user_id: assignmentData.user_id
					// TODO: Add buddy support to admin endpoint
					// buddy_name: assignmentData.buddy_name
				})
			});

			if (!response.ok) {
				const error = await response.json().catch(() => ({ message: 'Assignment failed' }));
				throw new Error(error.message || 'Failed to assign shift');
			}

			return response.json();
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['adminShiftSlots'] });
			queryClient.invalidateQueries({ queryKey: ['shiftDetails'] });
			queryClient.invalidateQueries({ queryKey: ['dashboardShifts'] });
			primaryUserValue = '';
			buddyName = '';
			assignmentError = null;
			showReassignForm = false;
			onBookingSuccess();
		},
		onError: (error: Error) => {
			assignmentError = error.message;
		}
	});

	// Clear assignment mutation
	const clearAssignmentMutation = createMutation({
		mutationFn: async () => {
			// Note: This endpoint may need to be implemented in the backend
			const response = await authenticatedFetch(`/api/admin/bookings/unassign`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					schedule_id: selectedShift.schedule_id,
					start_time: selectedShift.start_time
				})
			});

			if (!response.ok) {
				const error = await response
					.json()
					.catch(() => ({ message: 'Failed to clear assignment' }));
				throw new Error(error.message || 'Failed to clear assignment');
			}

			return response.json();
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['adminShiftSlots'] });
			queryClient.invalidateQueries({ queryKey: ['shiftDetails'] });
			queryClient.invalidateQueries({ queryKey: ['dashboardShifts'] });
			showReassignForm = false;
			onBookingSuccess();
		},
		onError: (error: Error) => {
			assignmentError = error.message;
		}
	});

	function closeAndFocusTrigger() {
		primaryUserOpen = false;
		tick().then(() => {
			primaryUserTriggerRef?.focus();
		});
	}

	function handleAssignShift(event: SubmitEvent) {
		event.preventDefault();
		assignmentError = null;

		if (!primaryUserValue) {
			assignmentError = 'Please select a primary user';
			return;
		}

		$assignShiftMutation.mutate({
			schedule_id: selectedShift.schedule_id,
			start_time: selectedShift.start_time,
			user_id: parseInt(primaryUserValue),
			buddy_name: buddyName.trim() || undefined
		});
	}

	function handleReassignShift(event: SubmitEvent) {
		event.preventDefault();
		assignmentError = null;

		if (!primaryUserValue) {
			assignmentError = 'Please select a primary user';
			return;
		}

		$assignShiftMutation.mutate({
			schedule_id: selectedShift.schedule_id,
			start_time: selectedShift.start_time,
			user_id: parseInt(primaryUserValue),
			buddy_name: buddyName.trim() || undefined
		});
	}

	function handleClearAssignment() {
		if (confirm('Are you sure you want to clear this shift assignment?')) {
			$clearAssignmentMutation.mutate();
		}
	}

	function startReassignment() {
		showReassignForm = true;
		// Pre-populate with current assignment if available
		if (selectedShift.user_name) {
			const currentUser = users.find(
				(u) => u.name === selectedShift.user_name || u.phone === selectedShift.user_phone
			);
			if (currentUser) {
				primaryUserValue = currentUser.id.toString();
			}
		}
		assignmentError = null;
	}

	function cancelReassignment() {
		showReassignForm = false;
		primaryUserValue = '';
		buddyName = '';
		assignmentError = null;
	}

	function handleBackToCalendar() {
		goto('/admin/shifts');
	}

	// Generate proper shift title following "XYZ Night 0-2AM" convention
	function getShiftTitle(startTimeIso: string, endTimeIso: string): string {
		try {
			const startDate = new Date(startTimeIso);
			const endDate = new Date(endTimeIso);

			// Get the previous day name for the shift title
			const previousDay = new Date(startDate);
			previousDay.setDate(previousDay.getDate() - 1);
			const dayName = previousDay.toLocaleDateString('en-US', { weekday: 'long' });

			// Format time range using UTC hours
			const startHour = startDate.getUTCHours();
			const endHour = endDate.getUTCHours();
			const timeRange = `${startHour}-${endHour}AM`;

			return `${dayName} Night ${timeRange}`;
		} catch (e) {
			return 'Shift Assignment';
		}
	}

	// Check if shift is assigned
	const isAssigned = $derived(selectedShift.is_booked);
</script>

<div class="p-6">
	<div class="max-w-2xl mx-auto">
		<!-- Header with shift title -->
		<div class="mb-6">
			<div class="flex items-center justify-between">
				<AdminPageHeader
					icon={ClockIcon}
					heading={getShiftTitle(selectedShift.start_time, selectedShift.end_time)}
					subheading="Manage shift assignment and team details"
				/>
				<Button variant="outline" onclick={handleBackToCalendar}>← Back</Button>
			</div>
			<div class="flex items-center gap-4 mt-4">
				<div class="flex items-center gap-2 text-sm text-muted-foreground">
					<ClockIcon class="h-4 w-4" />
					{formatTimeSlot(selectedShift.start_time, selectedShift.end_time)}
				</div>
				<Badge
					variant={isAssigned ? 'default' : 'secondary'}
					class={isAssigned
						? 'bg-green-100 text-green-700 border-green-200'
						: 'bg-orange-100 text-orange-700 border-orange-200'}
				>
					{isAssigned ? 'Assigned' : 'Available'}
				</Badge>
			</div>
		</div>

		<!-- Current Assignment Display -->
		{#if isAssigned && selectedShift.user_name}
			<Card class="mb-6">
				<CardHeader>
					<CardTitle class="text-lg flex items-center gap-2">
						<UserIcon class="h-5 w-5" />
						Current Assignment
					</CardTitle>
				</CardHeader>
				<CardContent>
					{#if !showReassignForm}
						<div class="space-y-4">
							<div>
								<p class="font-medium">{selectedShift.user_name}</p>
								{#if selectedShift.user_phone}
									<p class="text-sm text-muted-foreground">{selectedShift.user_phone}</p>
								{/if}
							</div>

							<!-- Action buttons -->
							<div class="flex gap-2">
								<Button variant="outline" size="sm" onclick={startReassignment}>Reassign</Button>
								<Button
									variant="outline"
									size="sm"
									onclick={handleClearAssignment}
									disabled={$clearAssignmentMutation.isPending}
									class="text-destructive hover:text-destructive"
								>
									{#if $clearAssignmentMutation.isPending}
										Clearing...
									{:else}
										Clear Assignment
									{/if}
								</Button>
							</div>
						</div>
					{:else}
						<!-- Reassignment Form -->
						<form onsubmit={handleReassignShift} class="space-y-4">
							<div class="space-y-2">
								<Label for="reassign-user" class="text-sm font-medium">
									Reassign to User <span class="text-red-500">*</span>
								</Label>
								<Popover.Root bind:open={primaryUserOpen}>
									<Popover.Trigger bind:ref={primaryUserTriggerRef}>
										{#snippet child({ props })}
											<Button
												variant="outline"
												class="w-full justify-between"
												{...props}
												role="combobox"
												aria-expanded={primaryUserOpen}
											>
												{selectedUser
													? selectedUser.name || selectedUser.phone
													: 'Select new user...'}
												<ChevronsUpDownIcon class="ml-2 h-4 w-4 shrink-0 opacity-50" />
											</Button>
										{/snippet}
									</Popover.Trigger>
									<Popover.Content class="w-full p-0">
										<Command.Root>
											<Command.Input placeholder="Search users..." />
											<Command.List>
												<Command.Empty>
													{#if $usersQuery.isLoading}
														Loading users...
													{:else if $usersQuery.isError}
														Error loading users
													{:else}
														No users found
													{/if}
												</Command.Empty>
												<Command.Group>
													{#each users as user (user.id)}
														<Command.Item
															value={user.id.toString()}
															onSelect={() => {
																primaryUserValue = user.id.toString();
																closeAndFocusTrigger();
															}}
														>
															<CheckIcon
																class={cn(
																	'mr-2 h-4 w-4',
																	primaryUserValue !== user.id.toString() && 'text-transparent'
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

							<!-- Buddy Field for Reassignment -->
							<div class="space-y-2">
								<Label for="reassign-buddy" class="text-sm font-medium">Buddy (Optional)</Label>
								<Input
									id="reassign-buddy"
									bind:value={buddyName}
									placeholder="Enter buddy name"
									class="w-full"
								/>
							</div>

							<!-- Error Display -->
							{#if assignmentError}
								<div class="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
									<p class="text-sm text-destructive">{assignmentError}</p>
								</div>
							{/if}

							<!-- Action buttons -->
							<div class="flex gap-2">
								<Button
									type="submit"
									disabled={$assignShiftMutation.isPending || !primaryUserValue}
									size="sm"
								>
									{#if $assignShiftMutation.isPending}
										Reassigning...
									{:else}
										Confirm Reassignment
									{/if}
								</Button>
								<Button
									type="button"
									variant="outline"
									size="sm"
									onclick={cancelReassignment}
									disabled={$assignShiftMutation.isPending}
								>
									Cancel
								</Button>
							</div>
						</form>
					{/if}
				</CardContent>
			</Card>
		{/if}

		<!-- Assignment Form -->
		{#if !isAssigned}
			<Card>
				<CardHeader>
					<CardTitle class="text-lg flex items-center gap-2">
						<Users2Icon class="h-5 w-5" />
						Assign Team
					</CardTitle>
					<CardDescription>
						Select the primary user and optional buddy for this shift
					</CardDescription>
				</CardHeader>
				<CardContent>
					<form onsubmit={handleAssignShift} class="space-y-6">
						<!-- Primary User Selection (Combobox) -->
						<div class="space-y-2">
							<Label for="primary-user" class="text-sm font-medium">
								Primary User <span class="text-red-500">*</span>
							</Label>
							<Popover.Root bind:open={primaryUserOpen}>
								<Popover.Trigger bind:ref={primaryUserTriggerRef}>
									{#snippet child({ props })}
										<Button
											variant="outline"
											class="w-full justify-between"
											{...props}
											role="combobox"
											aria-expanded={primaryUserOpen}
										>
											{selectedUser
												? selectedUser.name || selectedUser.phone
												: 'Select primary user...'}
											<ChevronsUpDownIcon class="ml-2 h-4 w-4 shrink-0 opacity-50" />
										</Button>
									{/snippet}
								</Popover.Trigger>
								<Popover.Content class="w-full p-0">
									<Command.Root>
										<Command.Input placeholder="Search users..." />
										<Command.List>
											<Command.Empty>
												{#if $usersQuery.isLoading}
													Loading users...
												{:else if $usersQuery.isError}
													Error loading users
												{:else}
													No users found
												{/if}
											</Command.Empty>
											<Command.Group>
												{#each users as user (user.id)}
													<Command.Item
														value={user.id.toString()}
														onSelect={() => {
															primaryUserValue = user.id.toString();
															closeAndFocusTrigger();
														}}
													>
														<CheckIcon
															class={cn(
																'mr-2 h-4 w-4',
																primaryUserValue !== user.id.toString() && 'text-transparent'
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

						<!-- Buddy Field -->
						<div class="space-y-2">
							<Label for="buddy-name" class="text-sm font-medium">Buddy (Optional)</Label>
							<Input
								id="buddy-name"
								bind:value={buddyName}
								placeholder="Enter buddy name (spouse, family member, etc.)"
								class="w-full"
							/>
							<p class="text-xs text-muted-foreground">
								Buddy can be anyone - they don't need to be a registered user
							</p>
						</div>

						<!-- Error Display -->
						{#if assignmentError}
							<div class="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
								<p class="text-sm text-destructive">{assignmentError}</p>
							</div>
						{/if}

						<!-- Submit Button -->
						<Button
							type="submit"
							disabled={$assignShiftMutation.isPending || !primaryUserValue}
							class="w-full"
						>
							{#if $assignShiftMutation.isPending}
								Assigning Shift...
							{:else}
								Assign Shift
							{/if}
						</Button>
					</form>
				</CardContent>
			</Card>
		{/if}

		<!-- Shift Details (Compact) -->
		<Card class="mt-6">
			<CardContent class="pt-6">
				<div class="grid grid-cols-2 gap-4 text-sm">
					<div>
						<p class="text-muted-foreground">Schedule</p>
						<p class="font-medium">{selectedShift.schedule_name}</p>
					</div>
					<div>
						<p class="text-muted-foreground">When</p>
						<p class="font-medium">{formatRelativeTime(selectedShift.start_time)}</p>
					</div>
				</div>
			</CardContent>
		</Card>
	</div>
</div>
