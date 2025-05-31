<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import type { AvailableShiftSlot, CreateBookingRequest } from '$lib/services/api/user';

	let {
		open = $bindable(false),
		shift = $bindable<AvailableShiftSlot | null>(null),
		isLoading = false,
		onConfirm,
		onCancel
	}: {
		open: boolean;
		shift: AvailableShiftSlot | null;
		isLoading?: boolean;
		onConfirm: (request: CreateBookingRequest) => void;
		onCancel: () => void;
	} = $props();

	// Local state for buddy name
	let buddyName = $state('');

	// Reset buddy name when dialog closes
	$effect(() => {
		if (!open) {
			buddyName = '';
		}
	});

	function handleConfirm() {
		if (!shift) return;

		const bookingRequest: CreateBookingRequest = {
			schedule_id: shift.schedule_id,
			start_time: shift.start_time,
			buddy_name: buddyName.trim() || undefined
		};

		onConfirm(bookingRequest);
	}

	function handleCancel() {
		onCancel();
		open = false;
	}

	function formatShiftTime(shift: AvailableShiftSlot) {
		const start = new Date(shift.start_time);
		const end = new Date(shift.end_time);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);

		let dateLabel = '';
		if (start.toDateString() === today.toDateString()) {
			dateLabel = 'Today';
		} else if (start.toDateString() === tomorrow.toDateString()) {
			dateLabel = 'Tomorrow';
		} else {
			dateLabel = start.toLocaleDateString('en-GB', {
				weekday: 'short',
				month: 'short',
				day: 'numeric'
			});
		}

		const timeRange = `${start.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		})} - ${end.toLocaleTimeString('en-GB', {
			hour: '2-digit',
			minute: '2-digit'
		})}`;

		return `${dateLabel} • ${timeRange}`;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Commit to Shift</Dialog.Title>
			<Dialog.Description>Confirm your commitment to this patrol shift.</Dialog.Description>
		</Dialog.Header>

		{#if shift}
			<div class="space-y-4 py-4">
				<!-- Shift Details -->
				<div class="p-4 bg-muted rounded-lg">
					<h4 class="font-medium text-sm mb-1">{shift.schedule_name}</h4>
					<p class="text-sm text-muted-foreground">{formatShiftTime(shift)}</p>
				</div>

				<!-- Optional Buddy Name -->
				<div class="space-y-2">
					<Label for="buddy-name">Buddy Name (Optional)</Label>
					<Input
						id="buddy-name"
						bind:value={buddyName}
						placeholder="Enter your patrol buddy's name"
						disabled={isLoading}
					/>
					<p class="text-xs text-muted-foreground">
						If you're patrolling with someone, add their name here
					</p>
				</div>
			</div>

			<Dialog.Footer>
				<Button variant="outline" onclick={handleCancel} disabled={isLoading}>Cancel</Button>
				<Button onclick={handleConfirm} disabled={isLoading}>
					{isLoading ? 'Committing...' : 'Commit to Shift'}
				</Button>
			</Dialog.Footer>
		{/if}
	</Dialog.Content>
</Dialog.Root>
