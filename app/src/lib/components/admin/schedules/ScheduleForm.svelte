<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import CalendarPlusIcon from '@lucide/svelte/icons/calendar-plus';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import cronstrue from 'cronstrue';
	import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
	import type { Schedule as ScheduleData } from '$lib/types';
	import CronView from '$lib/components/cron/cron-view.svelte';
	import { scheduleZodSchema, type ZodSchemaValues } from '$lib/schemas/schedule';
	import DeleteConfirmDialog from '../dialogs/DeleteConfirmDialog.svelte';
	import { createSaveScheduleMutation } from '$lib/queries/admin/schedules/saveScheduleMutation';
	import { createDeleteScheduleMutation } from '$lib/queries/admin/schedules/deleteScheduleMutation';
	import { parseYyyyMmDdToJsDate } from '$lib/utils/date';

	let { 
		schedule,
		onSuccess,
		onCancel 
	}: { 
		schedule?: ScheduleData | null;
		onSuccess?: () => void;
		onCancel?: () => void;
	} = $props();

	type FormInputValues = {
		name: string;
		cron_expr: string;
		start_date_str: string | null;
		end_date_str: string | null;
	};

	let formData = $state<FormInputValues>({
		name: '',
		cron_expr: '',
		start_date_str: null,
		end_date_str: null
	});

	let zodErrors = $state<Partial<Record<keyof ZodSchemaValues, string>>>({});
	let showDeleteConfirm = $state(false);

	$effect(() => {
		if (schedule) {
			formData.name = schedule.name;
			formData.cron_expr = schedule.cron_expr;
			formData.start_date_str = schedule.start_date ?? null;
			formData.end_date_str = schedule.end_date ?? null;
		} else {
			formData.name = '';
			formData.cron_expr = '';
			formData.start_date_str = null;
			formData.end_date_str = null;
		}
		zodErrors = {};
	});

	const saveMutation = createSaveScheduleMutation(() => {
		if (onSuccess) {
			onSuccess();
		} else {
			selectedScheduleForForm.set(undefined);
			goto('/admin/schedules');
		}
	});

	const deleteMutation = createDeleteScheduleMutation(
		() => {
			showDeleteConfirm = false;
		},
		() => {
			if (onSuccess) {
				onSuccess();
			} else {
				selectedScheduleForForm.set(undefined);
				goto('/admin/schedules');
			}
		}
	);

	function handleDateStringsChange(dates: { start: string | null; end: string | null }) {
		formData.start_date_str = dates.start;
		formData.end_date_str = dates.end;
		if (zodErrors.start_date || zodErrors.end_date) {
			const { start_date: _start_date, end_date: _end_date, ...rest } = zodErrors;
			zodErrors = rest;
		}
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

	function handleCancel() {
		selectedScheduleForForm.set(undefined);
		if (onCancel) {
			onCancel();
		} else {
			goto('/admin/schedules');
		}
	}

	const humanizedCron = $derived.by(() => {
		const cronExpr = schedule?.cron_expr || formData.cron_expr;
		if (!cronExpr || cronExpr.trim() === '') {
			return null;
		}
		try {
			return cronstrue.toString(cronExpr);
		} catch (error) {
			return null;
		}
	});
</script>

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
			{/if}
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
	
	<div class="flex justify-between pt-4 border-t">
		<div>
			{#if schedule?.schedule_id}
				<Button
					type="button"
					variant="destructive"
					onclick={handleDeleteClick}
					disabled={$deleteMutation.isPending || $saveMutation.isPending}
				>
					{#if $deleteMutation.isPending}Deleting...{:else}Delete{/if}
				</Button>
			{/if}
		</div>
		<div class="flex gap-2">
			<Button
				variant="outline"
				onclick={handleCancel}
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

<DeleteConfirmDialog
	open={showDeleteConfirm}
	name={schedule?.name ?? ''}
	id={schedule?.schedule_id ?? 0}
	mutation={deleteMutation}
/>
