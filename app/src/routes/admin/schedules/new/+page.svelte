<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';

	type SchedulePayload = {
		name: string;
		cron_expr: string;
		start_date?: string | null;
		end_date?: string | null;
		duration_minutes: number;
		timezone?: string | null;
	};

	let formData: SchedulePayload = {
		name: '',
		cron_expr: '',
		duration_minutes: 60,
		start_date: null,
		end_date: null,
		timezone: null
	};

	const queryClient = useQueryClient();

	const mutation = createMutation<
		Response, // Response type from fetch
		Error, // Error type
		SchedulePayload // Variables type
	>({
		mutationFn: async (payload) => {
			const response = await fetch('/api/admin/schedules', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});
			if (!response.ok) {
				const errorData = await response.json().catch(() => ({ message: 'Failed to create schedule' }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response;
		},
		onSuccess: async () => {
			toast.success('Schedule created successfully!');
			await queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
			goto('/admin/schedules');
		},
		onError: (error) => {
			toast.error(`Error: ${error.message}`);
		}
	});

	function handleSubmit() {
		const payload: SchedulePayload = {
			...formData,
			duration_minutes: Number(formData.duration_minutes), // Ensure it's a number
			// Ensure optional fields are null if empty, otherwise backend might get empty strings
			start_date: formData.start_date?.trim() === '' ? null : formData.start_date,
			end_date: formData.end_date?.trim() === '' ? null : formData.end_date,
			timezone: formData.timezone?.trim() === '' ? null : formData.timezone
		};
		$mutation.mutate(payload);
	}
</script>

<svelte:head>
	<title>Create New Schedule</title>
</svelte:head>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-6">Create New Schedule</h1>

	<form on:submit|preventDefault={handleSubmit} class="space-y-6 max-w-lg">
		<div>
			<Label for="name">Name</Label>
			<Input id="name" type="text" bind:value={formData.name} required />
		</div>

		<div>
			<Label for="cron_expr">CRON Expression</Label>
			<Input id="cron_expr" type="text" bind:value={formData.cron_expr} required />
			<p class="text-sm text-muted-foreground mt-1">
				E.g., "0 0 * * *" for daily at midnight.
			</p>
		</div>

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
			<Label for="timezone">Timezone (Optional)</Label>
			<Input id="timezone" type="text" bind:value={formData.timezone} />
			<p class="text-sm text-muted-foreground mt-1">E.g., "America/New_York", "UTC".</p>
		</div>

		<Button type="submit" disabled={$mutation.isPending}>
			{#if $mutation.isPending}
				Creating...
			{:else}
				Create Schedule
			{/if}
		</Button>
		{#if $mutation.isError}
			<p class="text-sm text-destructive">Error: {$mutation.error?.message}</p>
		{/if}
	</form>
</div> 