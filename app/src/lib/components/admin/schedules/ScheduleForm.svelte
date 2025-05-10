<script context="module" lang="ts">
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

	// Type for the schedule data passed as a prop (for editing)
	// This should match the structure returned by GET /api/admin/schedules/{id}
	// and include an 'id' field.

	// Prop for existing schedule data (undefined if creating a new one)
	export let schedule: ScheduleData | undefined = undefined;

	// const isEditMode = !!schedule; // Not strictly needed due to direct schedule.schedule_id checks

	type SchedulePayload = {
		name: string;
		cron_expr: string;
		start_date?: string | null;
		end_date?: string | null;
	};

	// Add scheduleId to the mutation variables for clarity and reliable access
	type MutationVariables = {
		payload: SchedulePayload;
		scheduleId?: number; // Only present in edit mode
	};

	let formData: SchedulePayload = {
		name: '',
		cron_expr: '',
		// duration_minutes: 60, // Removed
		start_date: null,
		end_date: null
	};

	let cronError: string | null = null;
	let humanizedCron: string | null = null;

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
	$: validateAndHumanizeCron(formData.cron_expr);

	// Helper to extract string value from SQLNullString/SQLNullTime like objects or direct strings
	function getStringValue(value: string | null | undefined): string | null {
		if (typeof value === 'string') {
			// For date fields from backend, sometimes they come as YYYY-MM-DDTHH:MM:SSZ
			// We only need the YYYY-MM-DD part for the input type="date"
			if (value.includes('T')) {
				return value.split('T')[0];
			}
			return value;
		}
		return null;
	}

	onMount(() => {
		if (schedule?.schedule_id !== undefined && schedule) {
			// Check schedule_id for edit mode determination
			formData = {
				name: schedule.name,
				cron_expr: schedule.cron_expr,
				// duration_minutes: schedule.duration_minutes, // Removed
				start_date: getStringValue(schedule.start_date),
				end_date: getStringValue(schedule.end_date)
			};
		}
	});

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
			if (currentIsEditMode && mutatedScheduleId) {
				await queryClient.invalidateQueries({ queryKey: ['adminSchedule', mutatedScheduleId] });
			}
			goto('/admin/schedules');
		},
		onError: (error) => {
			toast.error(`Error: ${error.message}`);
		}
	});

	function handleSubmit() {
		const currentScheduleIdFromProp = schedule?.schedule_id;

		const payloadForSubmit: SchedulePayload = {
			...formData,
			// duration_minutes: Number(formData.duration_minutes), // Removed
			start_date: formData.start_date?.trim() === '' ? null : formData.start_date,
			end_date: formData.end_date?.trim() === '' ? null : formData.end_date
		};

		const mutationVars: MutationVariables = {
			payload: payloadForSubmit,
			scheduleId: currentScheduleIdFromProp
		};

		$mutation.mutate(mutationVars);
	}
</script>

<svelte:head>
	<title>{schedule?.schedule_id !== undefined ? 'Edit' : 'Create New'} Schedule</title>
</svelte:head>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-6">
		{schedule?.schedule_id !== undefined ? 'Edit' : 'Create New'} Schedule
	</h1>

	<form on:submit|preventDefault={handleSubmit} class="space-y-6 max-w-lg">
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

		<!-- Duration Input Removed -->
		<!-- 
		<div>
			<Label for="duration_minutes">Duration (minutes)</Label>
			<Input
				id="duration_minutes"
				type="number"
				bind:value={formData.duration_minutes}
				required
				min="1"
			/>
		</div>
		-->

		<div>
			<Label for="start_date">Start Date (Optional)</Label>
			<Input id="start_date" type="date" bind:value={formData.start_date} />
			<p class="text-sm text-muted-foreground mt-1">Format: YYYY-MM-DD</p>
		</div>

		<div>
			<Label for="end_date">End Date (Optional)</Label>
			<Input id="end_date" type="date" bind:value={formData.end_date} />
			<p class="text-sm text-muted-foreground mt-1">Format: YYYY-MM-DD</p>
		</div>

		<div>
			<Button type="submit" disabled={$mutation.isPending || !!cronError}>
				{#if $mutation.isPending}
					{schedule?.schedule_id !== undefined ? 'Updating...' : 'Creating...'}
				{:else}
					{schedule?.schedule_id !== undefined ? 'Save Changes' : 'Create Schedule'}
				{/if}
			</Button>
			{#if $mutation.isError}
				<p class="text-sm text-destructive">Error: {$mutation.error?.message}</p>
			{/if}
		</div>
	</form>
</div>
