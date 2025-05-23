<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet';
	import { Button } from '$lib/components/ui/button';
	import ScheduleForm from '$lib/components/admin/schedules/ScheduleForm.svelte';
	import type { Schedule } from '$lib/types';
	import XIcon from '@lucide/svelte/icons/x';

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

	function handleClose() {
		open = false;
		onCancel?.();
	}

	function handleSuccess() {
		open = false;
		onSuccess?.();
	}
</script>

<Sheet.Root bind:open>
	<Sheet.Content class="w-full max-w-4xl">
		<Sheet.Header class="flex items-center justify-between">
			<Sheet.Title class="text-lg font-semibold">
				{mode === 'create' ? 'Create New Schedule' : 'Edit Schedule'}
			</Sheet.Title>
			<Button variant="ghost" size="sm" onclick={handleClose}>
				<XIcon class="h-4 w-4" />
			</Button>
		</Sheet.Header>
		
		<div class="mt-4">
			<ScheduleForm {schedule} onSuccess={handleSuccess} onCancel={handleClose} />
		</div>
	</Sheet.Content>
</Sheet.Root> 