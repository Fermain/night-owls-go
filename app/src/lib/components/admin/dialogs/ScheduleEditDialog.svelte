<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import ScheduleForm from '$lib/components/admin/schedules/ScheduleForm.svelte';
	import type { Schedule } from '$lib/types';
	import XIcon from '@lucide/svelte/icons/x';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import EditIcon from '@lucide/svelte/icons/edit-3';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import { createSchedulesQuery } from '$lib/queries/admin/schedules/schedulesQuery';
	import { createDeleteScheduleMutation } from '$lib/queries/admin/schedules/deleteScheduleMutation';
	import DeleteConfirmDialog from './DeleteConfirmDialog.svelte';
	import { formatDistanceToNow, format } from 'date-fns';
	import { useQueryClient } from '@tanstack/svelte-query';

	let {
		open = $bindable(false),
		schedule,
		mode = 'edit',
		onSuccess,
		onCancel
	} = $props<{
		open?: boolean;
		schedule?: Schedule | null;
		mode?: 'edit' | 'create';
		onSuccess?: () => void;
		onCancel?: () => void;
	}>();

	// State for form and schedule management
	let currentSchedule = $state<Schedule | null>(schedule);
	let showDeleteConfirmDialog = $state(false);
	let scheduleToDelete = $state<Schedule | null>(null);

	// Queries
	const queryClient = useQueryClient();
	const schedulesQuery = $derived(createSchedulesQuery());
	const deleteScheduleMutation = createDeleteScheduleMutation(
		() => {
			showDeleteConfirmDialog = false;
			scheduleToDelete = null;
		},
		async () => {
			await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
		}
	);

	function handleClose() {
		open = false;
		currentSchedule = null;
		onCancel?.();
	}

	function handleFormSuccess() {
		currentSchedule = null; // Reset to create mode
		onSuccess?.();
		// Refresh the schedules list
		queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
	}

	function handleFormCancel() {
		currentSchedule = null; // Reset to create mode
	}

	function handleEditSchedule(schedule: Schedule) {
		currentSchedule = schedule;
	}

	function handleDeleteSchedule(schedule: Schedule) {
		scheduleToDelete = schedule;
		showDeleteConfirmDialog = true;
	}

	function handleCreateNew() {
		currentSchedule = null;
	}

	function getScheduleStatus(schedule: Schedule): 'active' | 'upcoming' | 'expired' {
		const now = new Date();
		const startDate = schedule.start_date ? new Date(schedule.start_date) : null;
		const endDate = schedule.end_date ? new Date(schedule.end_date) : null;

		if (endDate && endDate < now) return 'expired';
		if (startDate && startDate > now) return 'upcoming';
		return 'active';
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'active': return 'default';
			case 'upcoming': return 'secondary';
			case 'expired': return 'destructive';
			default: return 'outline';
		}
	}

	function formatDateRange(schedule: Schedule): string {
		const start = schedule.start_date ? format(new Date(schedule.start_date), 'MMM d, yyyy') : null;
		const end = schedule.end_date ? format(new Date(schedule.end_date), 'MMM d, yyyy') : null;
		
		if (start && end) return `${start} - ${end}`;
		if (start) return `From ${start}`;
		if (end) return `Until ${end}`;
		return 'No date restrictions';
	}

	// Reset current schedule when the schedule prop changes
	$effect(() => {
		currentSchedule = schedule;
	});
</script>

<Sheet.Root bind:open>
	<Sheet.Content class="w-full max-w-6xl">
		<div class="space-y-4">
			<h1 class="text-2xl font-bold">Schedule Settings</h1>
			<!-- Existing Schedules List -->
			<div class="space-y-4">

				{#if $schedulesQuery.isLoading}
					<!-- Loading Skeletons -->
					<div class="space-y-3">
						{#each Array(3) as _, i (i)}
							<div class="p-4 border rounded-lg">
								<div class="flex items-center justify-between">
									<div class="space-y-2 flex-1">
										<Skeleton class="h-5 w-48" />
										<Skeleton class="h-4 w-32" />
									</div>
									<div class="flex gap-2">
										<Skeleton class="h-8 w-16" />
										<Skeleton class="h-8 w-16" />
									</div>
								</div>
							</div>
						{/each}
					</div>
				{:else if $schedulesQuery.isError}
					<div class="p-4 border rounded-lg border-destructive/50 bg-destructive/5">
						<p class="text-destructive font-medium">Error loading schedules</p>
						<p class="text-sm text-destructive/80 mt-1">{$schedulesQuery.error.message}</p>
					</div>
				{:else if $schedulesQuery.data && $schedulesQuery.data.length > 0}
					<div class="space-y-3">
						{#each $schedulesQuery.data as schedule (schedule.schedule_id)}
							{@const status = getScheduleStatus(schedule)}
							<div class="p-4 border rounded-lg hover:bg-muted/50 transition-colors {currentSchedule?.schedule_id === schedule.schedule_id ? 'border-primary bg-primary/5' : ''}">
								<div class="flex items-center justify-between">
									<div class="flex-1 space-y-2">
										<div class="flex items-center gap-3">
											<h4 class="font-medium">{schedule.name}</h4>
											<Badge variant={getStatusColor(status)} class="text-xs">
												{status}
											</Badge>
										</div>
									</div>
									<div class="flex items-center gap-2">
										<Button 
											variant="ghost" 
											size="sm" 
											onclick={() => handleEditSchedule(schedule)}
											class={currentSchedule?.schedule_id === schedule.schedule_id ? 'bg-primary text-primary-foreground' : ''}
										>
											<EditIcon class="h-4 w-4" />
											{currentSchedule?.schedule_id === schedule.schedule_id ? 'Editing' : 'Edit'}
										</Button>
										<Button 
											variant="ghost" 
											size="sm" 
											onclick={() => handleDeleteSchedule(schedule)}
											class="text-destructive hover:text-destructive hover:bg-destructive/10"
										>
											<TrashIcon class="h-4 w-4" />
										</Button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<div class="p-8 border rounded-lg border-dashed text-center">
						<CalendarIcon class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
						<h4 class="font-medium mb-2">No schedules found</h4>
						<p class="text-sm text-muted-foreground mb-4">
							Create your first schedule to start managing shifts
						</p>
						<Button onclick={handleCreateNew}>
							<PlusIcon class="h-4 w-4 mr-2" />
							Create First Schedule
						</Button>
					</div>
				{/if}
			</div>

			<!-- Schedule Form -->
			<div class="">
				<div class="mb-4">
					<h3 class="text-lg font-medium">
						{currentSchedule ? `Edit ${currentSchedule.name} Schedule` : 'Create New Schedule'}
					</h3>
				</div>
				
				<ScheduleForm 
					schedule={currentSchedule} 
					onSuccess={handleFormSuccess} 
					onCancel={handleFormCancel} 
				/>
			</div>
		</div>
	</Sheet.Content>
</Sheet.Root>

<!-- Delete Confirmation Dialog -->
<DeleteConfirmDialog
	open={showDeleteConfirmDialog}
	name={scheduleToDelete?.name ?? ''}
	id={scheduleToDelete?.schedule_id ?? 0}
	mutation={deleteScheduleMutation}
/> 