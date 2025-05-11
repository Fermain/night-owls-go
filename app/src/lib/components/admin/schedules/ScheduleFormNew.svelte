<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import CalendarPlusIcon from '@lucide/svelte/icons/calendar-plus';
	import { toast } from 'svelte-sonner';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { z } from 'zod';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import cronstrue from 'cronstrue';
	import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
	import type { Schedule as ScheduleData } from '$lib/components/schedules_table/columns';
	import type { DateRange } from 'bits-ui';
	import { CalendarDate } from '@internationalized/date';
	import CronView from '$lib/components/cron/cron-view.svelte';

	let { schedule }: { schedule?: ScheduleData } = $props();

	const scheduleZodSchema = z
		.object({
			name: z.string().min(1, 'Schedule name is required'),
			cron_expr: z
				.string()
				.min(1, 'CRON expression is required')
				.refine(
					(val) => {
						try {
							cronstrue.toString(val);
							return true;
						} catch (e) {
							return false;
						}
					},
					{ message: 'Invalid CRON expression format' }
				),
			start_date: z.date().nullable().optional(),
			end_date: z.date().nullable().optional()
		})
		.refine(
			(data) => {
				if (data.start_date && data.end_date && data.start_date > data.end_date) {
					return false;
				}
				return true;
			},
			{
				message: 'End date cannot be before start date',
				path: ['end_date']
			}
		);

	type FormInputValues = {
		name: string;
		cron_expr: string;
		start_date_str: string | null;
		end_date_str: string | null;
	};
	type ZodSchemaValues = z.infer<typeof scheduleZodSchema>;

	let formData = $state<FormInputValues>({
		name: '',
		cron_expr: '',
		start_date_str: null,
		end_date_str: null
	});

	let zodErrors = $state<Partial<Record<keyof ZodSchemaValues, string>>>({});
	let humanizedCron = $state<string | null>(null);
	let showDeleteConfirm = $state(false);

	const queryClient = useQueryClient();

	$effect(() => {
		if (schedule) {
			formData.name = schedule.name;
			formData.cron_expr = schedule.cron_expr;
			formData.start_date_str = schedule.start_date ?? null;
			formData.end_date_str = schedule.end_date ?? null;
			try {
				humanizedCron = cronstrue.toString(schedule.cron_expr);
			} catch (e) {
				humanizedCron = null;
			}
		} else {
			formData.name = '';
			formData.cron_expr = '';
			formData.start_date_str = null;
			formData.end_date_str = null;
			humanizedCron = null;
		}
		zodErrors = {};
	});

	$effect(() => {
		if (formData.cron_expr) {
			try {
				humanizedCron = cronstrue.toString(formData.cron_expr);
				if (zodErrors.cron_expr) {
					const { cron_expr, ...rest } = zodErrors;
					zodErrors = rest;
				}
			} catch (e) {
				humanizedCron = null;
			}
		} else {
			humanizedCron = null;
		}
	});

	const saveMutation = createMutation<any, Error, { payload: any; scheduleId?: number }>({
		mutationFn: async ({ payload, scheduleId }) => {
			const url = scheduleId ? `/api/admin/schedules/${scheduleId}` : '/api/admin/schedules';
			const method = scheduleId ? 'PUT' : 'POST';
			const response = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: `Failed to ${scheduleId ? 'update' : 'create'} schedule` }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response.json();
		},
		onSuccess: async (_data, { scheduleId }) => {
			toast.success(`Schedule ${scheduleId ? 'updated' : 'created'} successfully!`);
			await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
			await queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] });
			if (scheduleId) {
				selectedScheduleForForm.set(undefined);
			}
			goto('/admin/schedules');
		},
		onError: (error) => {
			toast.error(`Save Error: ${error.message}`);
		}
	});

	const deleteMutation = createMutation<any, Error, number>({
		mutationFn: async (id) => {
			const response = await fetch(`/api/admin/schedules/${id}`, { method: 'DELETE' });
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: 'Failed to delete schedule' }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response.ok;
		},
		onSuccess: async () => {
			toast.success('Schedule deleted successfully!');
			await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
			await queryClient.invalidateQueries({ queryKey: ['adminSchedulesForLayout'] });
			selectedScheduleForForm.set(undefined);
			goto('/admin/schedules');
			showDeleteConfirm = false;
		},
		onError: (error) => {
			toast.error(`Delete Error: ${error.message}`);
			showDeleteConfirm = false;
		}
	});

	function handleDateStringsChange(dates: { start: string | null; end: string | null }) {
		formData.start_date_str = dates.start;
		formData.end_date_str = dates.end;
		if (zodErrors.start_date || zodErrors.end_date) {
			const { start_date, end_date, ...rest } = zodErrors;
			zodErrors = rest;
		}
	}

	function parseYyyyMmDdToJsDate(dateStr: string | null | undefined): Date | null {
		if (!dateStr) return null;
		const [year, month, day] = dateStr.split('-').map(Number);
		if (year && month && day) {
			const date = new Date(Date.UTC(year, month - 1, day));
			if (!isNaN(date.valueOf())) return date;
		}
		return null;
	}

	function validateForm(): boolean {
		const valuesToValidate: Omit<ZodSchemaValues, 'timezone'> = {
			name: formData.name,
			cron_expr: formData.cron_expr,
			start_date: parseYyyyMmDdToJsDate(formData.start_date_str),
			end_date: parseYyyyMmDdToJsDate(formData.end_date_str)
		};
		const result = scheduleZodSchema.safeParse(valuesToValidate);
		if (!result.success) {
			const newErrors: Partial<Record<keyof ZodSchemaValues, string>> = {};
			for (const issue of result.error.issues) {
				if (issue.path.length > 0) {
					newErrors[issue.path[0] as keyof ZodSchemaValues] = issue.message;
				}
			}
			zodErrors = newErrors;
			toast.error('Please correct form errors.');
			return false;
		} else {
			zodErrors = {};
		}
		return true;
	}

	function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (!validateForm()) return;
		const payload = {
			name: formData.name,
			cron_expr: formData.cron_expr,
			start_date: formData.start_date_str,
			end_date: formData.end_date_str
		};
		$saveMutation.mutate({ payload, scheduleId: schedule?.schedule_id });
	}

	function handleDeleteClick() {
		if (schedule?.schedule_id) {
			showDeleteConfirm = true;
		}
	}
	function confirmDeleteAction() {
		if (schedule?.schedule_id) {
			$deleteMutation.mutate(schedule.schedule_id);
		}
	}
	function handleCancel() {
		selectedScheduleForForm.set(undefined);
		goto('/admin/schedules');
	}
</script>

<div class="container mx-auto p-4">
	<div class="w-full p-6 shadow-md rounded-lg border bg-card text-card-foreground">
		<div class="mb-4">
			<h2 class="text-2xl font-semibold tracking-tight">
				{schedule?.schedule_id ? 'Edit Schedule' : 'Create New Schedule'}
			</h2>
			{#if schedule?.schedule_id}
				<p class="text-sm text-muted-foreground">ID: {schedule.schedule_id}</p>
			{/if}
		</div>
		<div class="space-y-6">
			<form onsubmit={handleSubmit} id="scheduleFormInternal" class="space-y-6">
				<div>
					<Label for="name">Name</Label>
					<Input
						id="name"
						type="text"
						bind:value={formData.name}
						disabled={$saveMutation.isPending}
					/>
					{#if zodErrors.name}<p class="text-sm text-destructive mt-1">
							{zodErrors.name}
						</p>{/if}
				</div>
				<div>
					<Label for="cron_expr">CRON Expression</Label>
					<Input
						id="cron_expr"
						type="text"
						bind:value={formData.cron_expr}
						disabled={$saveMutation.isPending}
					/>
					{#if zodErrors.cron_expr}<p class="text-sm text-destructive mt-1">
							{zodErrors.cron_expr}
						</p>
					{:else if humanizedCron}<p class="text-sm text-muted-foreground mt-1">
							Interprets as: {humanizedCron}
						</p>{/if}
					{#if !zodErrors.cron_expr && formData.cron_expr && humanizedCron}
						<div class="mt-2">
							<CronView cronExpr={formData.cron_expr} />
						</div>
					{/if}
				</div>
				<div>
					<Label for="date_range_picker_trigger_id">Date Range (Optional)</Label>
					<DateRangePicker
						initialStartDate={formData.start_date_str}
						initialEndDate={formData.end_date_str}
						change={handleDateStringsChange}
						placeholderText="Pick start and end dates"
					/>
					{#if zodErrors.start_date}<p class="text-sm text-destructive mt-1">
							Start Date: {zodErrors.start_date}
						</p>{/if}
					{#if zodErrors.end_date}<p class="text-sm text-destructive mt-1">
							End Date: {zodErrors.end_date}
						</p>{/if}
				</div>
			</form>
		</div>
		<div class="flex justify-between mt-6 pt-4 border-t">
			<div>
				{#if schedule?.schedule_id}
					<!-- @ts-ignore -->
					<Button
						variant="destructive"
						on:click={handleDeleteClick}
						disabled={$deleteMutation.isPending || $saveMutation.isPending}
					>
						{#if $deleteMutation.isPending}Deleting...{:else}Delete{/if}
					</Button>
				{/if}
			</div>
			<div class="flex gap-2">
				<!-- @ts-ignore -->
				<Button
					variant="outline"
					on:click={handleCancel}
					disabled={$saveMutation.isPending || $deleteMutation.isPending}>Cancel</Button
				>
				<Button
					type="submit"
					form="scheduleFormInternal"
					disabled={$saveMutation.isPending || $deleteMutation.isPending}
				>
					{#if $saveMutation.isPending}<Loader2Icon class="animate-spin mr-2 h-4 w-4" />Saving...
					{:else}<CalendarPlusIcon class="mr-2 h-4 w-4" />{schedule?.schedule_id
							? 'Update'
							: 'Create'} Schedule{/if}
				</Button>
			</div>
		</div>
	</div>
</div>

{#if showDeleteConfirm}
	<AlertDialog.Root bind:open={showDeleteConfirm}>
		<AlertDialog.Content>
			<AlertDialog.Header
				><AlertDialog.Title>Are you sure?</AlertDialog.Title>
				<AlertDialog.Description
					>This will permanently delete "{schedule?.name}".</AlertDialog.Description
				>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<!-- @ts-ignore -->
				<AlertDialog.Cancel
					on:click={() => (showDeleteConfirm = false)}
					disabled={$deleteMutation.isPending}>Cancel</AlertDialog.Cancel
				>
				<!-- @ts-ignore -->
				<AlertDialog.Action
					on:click={confirmDeleteAction}
					disabled={$deleteMutation.isPending}
					class={$deleteMutation.isPending ? 'bg-destructive/50' : 'bg-destructive'}
				>
					{#if $deleteMutation.isPending}Deleting...{:else}Yes, delete{/if}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
