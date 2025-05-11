<script lang="ts">
	import { page } from '$app/stores';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Card from '$lib/components/ui/card';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import { DateRangePicker } from '$lib/components/ui/date-range-picker';
	import type { Schedule } from '$lib/components/schedules_table/columns';
	import {পাট, type SubmitFunction, type FormPathLeavesWithErrors } from 'sveltekit-forms';
	import { scheduleFormSchema, type ScheduleFormSchema } from './schema';
	import { 실패했을때 } from '$lib/utils/forms';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import type { DateRange } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import { AlertCircle } from 'lucide-svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';

	export let schedule: Schedule | undefined = undefined;
	export let close: () => void;

	let formElement: HTMLFormElement;
	let showDeleteConfirm = $state(false);

	const { form, errors, state, validate, reset, submit } = পাট(scheduleFormSchema, {
		logging: true,
		onSubmit: async ($form, event) => {
			await mutation.mutateAsync($form);
			close();
		},
		onValidate: 실패했을때
	});

	let initialFormData: ScheduleFormSchema | undefined;

	$effect(() => {
		console.log('ScheduleForm: schedule prop changed:', schedule);
		if (schedule) {
			const formData: ScheduleFormSchema = {
				name: schedule.name,
				location: schedule.location,
				date_range: {
					start: schedule.start_date ? new Date(schedule.start_date) : undefined,
					end: schedule.end_date ? new Date(schedule.end_date) : undefined
				}
			};
			console.log('ScheduleForm: Resetting form with data:', formData);
			reset(formData);
			initialFormData = JSON.parse(JSON.stringify(formData)); // Deep copy for later comparison
		} else {
			console.log('ScheduleForm: Resetting form (no schedule)');
			reset();
			initialFormData = undefined;
		}
	});

	const client = $page.data.queryClient;

	const mutation = createMutation({
		mutationFn: async (data: ScheduleFormSchema) => {
			const url = schedule?.schedule_id
				? `/api/admin/schedules/${schedule.schedule_id}`
				: '/api/admin/schedules';
			const method = schedule?.schedule_id ? 'PUT' : 'POST';

			const body = {
				name: data.name,
				location: data.location,
				start_date: data.date_range?.start
					? data.date_range.start.toISOString().split('T')[0]
					: null,
				end_date: data.date_range?.end ? data.date_range.end.toISOString().split('T')[0] : null
			};

			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(body)
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.message || 'Failed to save schedule');
			}
			return response.json();
		},
		onSuccess: () => {
			toast.success(schedule?.schedule_id ? 'Schedule updated' : 'Schedule created');
			client.invalidateQueries({ queryKey: ['schedules'] });
		},
		onError: (error) => {
			toast.error(error.message);
		}
	});

	const deleteMutation = createMutation({
		mutationFn: async () => {
			if (!schedule?.schedule_id) {
				throw new Error('No schedule selected for deletion');
			}
			const response = await fetch(`/api/admin/schedules/${schedule.schedule_id}`, {
				method: 'DELETE'
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.message || 'Failed to delete schedule');
			}
			// Assuming the API returns nothing or a success message on DELETE
			return;
		},
		onSuccess: () => {
			toast.success('Schedule deleted');
			client.invalidateQueries({ queryKey: ['schedules'] });
			close(); // Close the form/drawer
			reset(); // Reset the form
		},
		onError: (error) => {
			toast.error(error.message);
		}
	});

	function handleDateRangeChange(newRange: DateRange | undefined) {
		console.log('DateRangePicker changed in form: ', newRange);
		$form.date_range = newRange;
		// Manually trigger validation for the date_range field if needed, or rely on form-level validation
		validate('date_range');
	}

	function handleDelete() {
		if (schedule?.schedule_id) {
			showDeleteConfirm = true;
		}
	}

	function confirmDelete() {
		deleteMutation.mutate();
		showDeleteConfirm = false;
	}

	let title = $derived(schedule?.schedule_id ? 'Edit Schedule' : 'Create New Schedule');
	let actionButtonText = $derived(schedule?.schedule_id ? 'Update Schedule' : 'Create Schedule');

	// Expose form state for debugging if needed
	$: console.log('ScheduleForm $form state:', $form);
	$: console.log('ScheduleForm $errors:', $errors);
	$: console.log('ScheduleForm $state:', $state);
</script>

// ... existing code ...
					<DateRangePicker
						range={$form.date_range}
						onRangeChange={handleDateRangeChange}
						disabled={$state.submitting}
						initialMonth={new Date(new Date().setMonth(new Date().getMonth()))}
					/>
					{#if $errors.date_range?.length}
						<div class="text-destructive text-sm mt-1" transition:slide|local>
							{$errors.date_range.join(', ')}
						</div>
					{/if}
				</div>
			</Card.Content>
			<Card.Footer class="flex justify-between">
				{#if schedule?.schedule_id}
					<Button variant="destructive" on:click={handleDelete} disabled={$state.submitting || deleteMutation.isPending}>
						Delete
					</Button>
				{:else}
					<div></div> <!-- Placeholder to keep "Create" button to the right -->
				{/if}
				<div class="flex gap-2">
					<Button variant="outline" on:click={close} disabled={$state.submitting || deleteMutation.isPending}>Cancel</Button>
					<Button type="submit" disabled={$state.submitting || !$state.dirty || deleteMutation.isPending}
						>{#if $state.submitting}Saving...{:else}{actionButtonText}{/if}</Button
					>
				</div>
			</Card.Footer>
		</Card.Root>
	</form>
{/if}

<AlertDialog.Root bind:open={showDeleteConfirm}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Are you sure?</AlertDialog.Title>
			<AlertDialog.Description>
				This action cannot be undone. This will permanently delete the schedule
				"{schedule?.name}".
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={deleteMutation.isPending}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				on:click={confirmDelete}
				disabled={deleteMutation.isPending}
				class={deleteMutation.isPending ? 'cursor-not-allowed' : ''}
			>
				{#if deleteMutation.isPending}Deleting...{:else}Delete{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<style>
	.icon-container {
		display: flex;
		align-items: center;
		gap: 0.5rem; /* Adjust gap as needed */
	}
</style>
