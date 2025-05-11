<script module lang="ts">
	export type ScheduleData = {
		schedule_id: number;
		name: string;
		cron_expr: string;
		start_date?: string | null;
		end_date?: string | null;
		timezone?: string | null;
	};
</script>

<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import cronstrue from 'cronstrue';
	import CronView from '$lib/components/cron/cron-view.svelte';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import { parseIsoToYyyyMmDd } from '$lib/utils/date';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Trash2 from 'lucide-svelte/icons/trash-2';

	// Type for the schedule data passed as a prop (for editing)
	// This should match the structure returned by GET /api/admin/schedules/{id}
	// and include an 'id' field.

	type Props = {
		schedule?: ScheduleData;
	};
	const { schedule = undefined }: Props = $props();

	// const isEditMode = !!schedule; // Not strictly needed due to direct schedule.schedule_id checks

	type SchedulePayload = {
		name: string;
		cron_expr: string;
		start_date?: string | null;
		end_date?: string | null;
		// timezone?: string | null; // Removed timezone
	};

	// Add scheduleId to the mutation variables for clarity and reliable access
	type MutationVariables = {
		payload: SchedulePayload;
		scheduleId?: number; // Only present in edit mode
	};

	let formData = $state<SchedulePayload>({
		name: '',
		cron_expr: '',
		start_date: null,
		end_date: null
		// timezone: null // Removed timezone
	});

	let cronError: string | null = $state(null);
	let humanizedCron: string | null = $state(null);
	let isDeleteDialogOpen = $state(false);

	function validateAndHumanizeCron(cronValue: string) {
		if (!cronValue || cronValue.trim() === '') {
			cronError = 'CRON expression is required.'; // Or handle as per form's `required` attribute
			humanizedCron = null;
			return;
		}
		try {
			humanizedCron = cronstrue.toString(cronValue, { verbose: true });
			cronError = null; // Clear error if cronstrue parsing succeeds
		} catch (e) {
			cronError = e instanceof Error ? e.message : 'Invalid CRON expression';
			humanizedCron = null;
		}
	}

	// Validate and humanize when cron_expr changes
	$effect(() => {
		validateAndHumanizeCron(formData.cron_expr);
	});

	// Helper to extract string value from SQLNullString/SQLNullTime like objects or direct strings
	// This function is now effectively replaced by parseIsoToYyyyMmDd for dates,
	// but we might keep it if other string-like props need similar null/undefined handling.
	// For now, let's assume it's mainly for dates and can be superseded by the new util.
	function getStringValue(value: string | null | undefined): string | null {
		return parseIsoToYyyyMmDd(value);
	}

	// $effect to reactively update formData when the schedule prop changes.
	$effect(() => {
		console.log('[ScheduleForm $effect] schedule prop changed:', schedule);
		if (schedule?.schedule_id !== undefined && schedule) {
			formData.name = schedule.name;
			formData.cron_expr = schedule.cron_expr;
			formData.start_date = parseIsoToYyyyMmDd(schedule.start_date); 
			formData.end_date = parseIsoToYyyyMmDd(schedule.end_date);
			console.log('[ScheduleForm $effect] formData updated:', JSON.stringify(formData));
		} else {
			formData.name = '';
			formData.cron_expr = '';
			formData.start_date = null;
			formData.end_date = null;
			console.log('[ScheduleForm $effect] schedule prop undefined, formData reset.');
		}
	});

	// Function to handle date changes from DateRangePicker
	function handleDateRangeChange(detail: { start: string | null; end: string | null }) {
		formData.start_date = detail.start;
		formData.end_date = detail.end;
	}

	const queryClient = useQueryClient();

	const mutation = createMutation<
		Response, // Response type from fetch
		Error, // Error type
		MutationVariables // Variables type updated here
	>({
		mutationFn: async (vars) => {
			// vars is now MutationVariables
			const { payload, scheduleId: currentScheduleIdToUse } = vars;
			const currentIsEditMode = currentScheduleIdToUse !== undefined;

			const url = currentIsEditMode
				? `/api/admin/schedules/${currentScheduleIdToUse}`
				: '/api/admin/schedules';
			const method = currentIsEditMode ? 'PUT' : 'POST';

			const response = await fetch(url, {
				method: method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({
					message: `Failed to ${currentIsEditMode ? 'update' : 'create'} schedule`
				}));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response;
		},
		onSuccess: async (_data, vars) => {
			// vars is MutationVariables here too
			const { scheduleId: mutatedScheduleId } = vars;
			const currentIsEditMode = mutatedScheduleId !== undefined;
			toast.success(`Schedule ${currentIsEditMode ? 'updated' : 'created'} successfully!`);
			await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] }); // Invalidate list
			await queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] }); // Also invalidate layout query
			if (currentIsEditMode && mutatedScheduleId) {
				await queryClient.invalidateQueries({ queryKey: ['adminScheduleDetail', String(mutatedScheduleId)] });
			}
			goto('/admin/schedules');
		},
		onError: (error) => {
			toast.error(`Error: ${error.message}`);
		}
	});

	// Delete Mutation
	const deleteMutation = createMutation<
		Response, 
		Error, 
		number // Expects schedule_id as variable
	>({
		mutationFn: async (scheduleId: number) => {
			const response = await fetch(`/api/admin/schedules/${scheduleId}`, {
				method: 'DELETE'
			});
			if (!response.ok) {
				const errorData = await response.json().catch(() => ({ message: 'Failed to delete schedule' }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response;
		},
		onSuccess: (_data, scheduleId) => {
			toast.success('Schedule deleted successfully!');
			queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
			queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] });
			queryClient.invalidateQueries({ queryKey: ['adminScheduleDetail', String(scheduleId)] }); // Invalidate specific detail
			goto('/admin/schedules');
		},
		onError: (error) => {
			toast.error(`Error deleting schedule: ${error.message}`);
		}
	});

	function handleSubmit() {
		console.log('[ScheduleForm handleSubmit] formData before parentSubmit:', formData);
		const currentScheduleIdFromProp = schedule?.schedule_id;

		const payloadForSubmit: SchedulePayload = {
			...formData,
			start_date:
				typeof formData.start_date === 'string' && formData.start_date.trim() === ''
					? null
					: formData.start_date,
			end_date:
				typeof formData.end_date === 'string' && formData.end_date.trim() === ''
					? null
					: formData.end_date
			// timezone: typeof formData.timezone === 'string' && formData.timezone.trim() === '' ? null : formData.timezone // Removed timezone
		};

		const mutationVars: MutationVariables = {
			payload: payloadForSubmit,
			scheduleId: currentScheduleIdFromProp
		};

		$mutation.mutate(mutationVars);
	}

	function openDeleteDialog() {
		if (schedule?.schedule_id !== undefined) {
			isDeleteDialogOpen = true;
		}
	}

	function confirmDelete() {
		if (schedule?.schedule_id !== undefined) {
			$deleteMutation.mutate(schedule.schedule_id);
		}
		isDeleteDialogOpen = false; // Close dialog regardless of mutation outcome initially
	}
</script>

<svelte:head>
	<title>{schedule?.schedule_id !== undefined ? 'Edit' : 'Create New'} Schedule</title>
</svelte:head>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-6">
		{schedule?.schedule_id !== undefined ? 'Edit' : 'Create New'} Schedule
	</h1>

	<form onsubmit={handleSubmit} class="space-y-4">
		<div>
			<Label for="name">Name</Label>
			<Input id="name" type="text" bind:value={formData.name} required />
		</div>

		<div>
			<Label for="cron_expr">CRON Expression</Label>
			<Input id="cron_expr" type="text" bind:value={formData.cron_expr} required />
			{#if cronError}
				<p class="text-sm text-destructive mt-1">{cronError}</p>
			{:else if humanizedCron}
				<p class="text-sm text-muted-foreground mt-1">Interprets as: {humanizedCron}</p>
				{#if formData.cron_expr.trim() !== ''}
					<div class="mt-2">
						<CronView cronExpr={formData.cron_expr} />
					</div>
				{/if}
			{:else}
				<p class="text-sm text-muted-foreground mt-1">E.g., "0 0 * * *" for daily at midnight.</p>
			{/if}
		</div>

		<div>
			<DateRangePicker
				initialStartDate={formData.start_date}
				initialEndDate={formData.end_date}
				change={handleDateRangeChange}
				placeholderText="Pick start and end dates"
			/>
		</div>

		<div class="flex items-center space-x-2">
			<Button type="submit" disabled={$mutation.isPending || !!cronError}>
				{#if $mutation.isPending}
					{schedule?.schedule_id !== undefined ? 'Updating...' : 'Creating...'}
				{:else}
					{schedule?.schedule_id !== undefined ? 'Save Changes' : 'Create Schedule'}
				{/if}
			</Button>

			{#if schedule?.schedule_id !== undefined}
				<Button variant="destructive" type="button" onclick={openDeleteDialog} disabled={$deleteMutation.isPending}>
					<Trash2 class="mr-2 h-4 w-4" />
					Delete Schedule
				</Button>
			{/if}

			{#if $mutation.isError}
				<p class="text-sm text-destructive">Error: {$mutation.error?.message}</p>
			{/if}
		</div>
	</form>

	{#if schedule?.schedule_id !== undefined}
		<AlertDialog.Root bind:open={isDeleteDialogOpen}>
			<AlertDialog.Content>
				<AlertDialog.Header>
					<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
					<AlertDialog.Description>
						This action cannot be undone. This will permanently delete the schedule
						"{schedule.name}" and all of its associated data.
					</AlertDialog.Description>
				</AlertDialog.Header>
				<AlertDialog.Footer>
					<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
					<AlertDialog.Action onclick={confirmDelete} disabled={$deleteMutation.isPending}>
						{#if $deleteMutation.isPending}Deleting...{:else}Yes, delete schedule{/if}
					</AlertDialog.Action>
				</AlertDialog.Footer>
			</AlertDialog.Content>
		</AlertDialog.Root>
	{/if}
</div>
