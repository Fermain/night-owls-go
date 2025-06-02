<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Badge } from '$lib/components/ui/badge';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import UserIcon from '@lucide/svelte/icons/user';
	import { createMutation } from '@tanstack/svelte-query';
	import {
		UserApiService,
		type AvailableShiftSlot,
		type CreateBookingRequest
	} from '$lib/services/api/user';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		shifts: AvailableShiftSlot[];
		onSuccess: () => void;
		onCancel: () => void;
	}

	let { open = $bindable(false), shifts, onSuccess, onCancel }: Props = $props();

	// State
	let selectedShiftIds = $state<Set<string>>(new Set());

	// Create bulk booking mutation
	const bulkBookingMutation = createMutation({
		mutationFn: async (bookingRequests: CreateBookingRequest[]) => {
			const results = [];
			for (const request of bookingRequests) {
				try {
					const booking = await UserApiService.createBooking(request);
					results.push({ success: true, booking, request });
				} catch (error) {
					results.push({
						success: false,
						error: error instanceof Error ? error.message : 'Unknown error',
						request
					});
				}
			}
			return results;
		},
		onSuccess: (results) => {
			const successCount = results.filter((r) => r.success).length;
			const errorCount = results.filter((r) => !r.success).length;

			if (errorCount === 0) {
				toast.success(`Successfully committed to ${successCount} shifts!`);
				selectedShiftIds.clear();
				onSuccess();
				open = false;
			} else {
				toast.error(`${successCount} shifts booked, ${errorCount} failed`);
				// Show failed requests
				results
					.filter((r) => !r.success)
					.forEach((r) => {
						toast.error(`Failed to book shift: ${r.error}`);
					});
			}
		},
		onError: (error) => {
			toast.error(`Failed to book shifts: ${error.message}`);
		}
	});

	// Derived values
	const selectedShifts = $derived(
		shifts.filter((shift) => selectedShiftIds.has(getShiftId(shift)))
	);

	const hasSelection = $derived(selectedShiftIds.size > 0);

	// Helper functions
	function getShiftId(shift: AvailableShiftSlot): string {
		return `${shift.schedule_id}-${shift.start_time}`;
	}

	function toggleShift(shift: AvailableShiftSlot) {
		const shiftId = getShiftId(shift);
		if (selectedShiftIds.has(shiftId)) {
			selectedShiftIds.delete(shiftId);
		} else {
			selectedShiftIds.add(shiftId);
		}
		selectedShiftIds = new Set(selectedShiftIds); // Trigger reactivity
	}

	function toggleSelectAll() {
		const allShiftIds = shifts.map((shift) => getShiftId(shift));
		const allSelected = allShiftIds.every((id) => selectedShiftIds.has(id));

		if (allSelected) {
			selectedShiftIds.clear();
		} else {
			allShiftIds.forEach((id) => selectedShiftIds.add(id));
		}
		selectedShiftIds = new Set(selectedShiftIds); // Trigger reactivity
	}

	function formatShiftTime(startTime: string, endTime: string): string {
		const start = new Date(startTime);
		const end = new Date(endTime);
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

	function formatShiftDate(startTime: string): string {
		const date = new Date(startTime);
		return date.toLocaleDateString('en-GB', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function handleBulkAssign() {
		if (!hasSelection) return;

		const bookingRequests: CreateBookingRequest[] = selectedShifts.map((shift) => ({
			schedule_id: shift.schedule_id,
			start_time: shift.start_time
		}));

		$bulkBookingMutation.mutate(bookingRequests);
	}

	function handleCancel() {
		selectedShiftIds.clear();
		onCancel();
		open = false;
	}

	// Group shifts by date for better organization
	const groupedShifts = $derived.by(() => {
		const groups = new Map<string, AvailableShiftSlot[]>();
		shifts.forEach((shift) => {
			const date = formatShiftDate(shift.start_time);
			if (!groups.has(date)) {
				groups.set(date, []);
			}
			groups.get(date)!.push(shift);
		});

		// Sort by date
		return Array.from(groups.entries())
			.sort(([dateA], [dateB]) => new Date(dateA).getTime() - new Date(dateB).getTime())
			.map(([date, shifts]) => ({
				date,
				shifts: shifts.sort(
					(a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
				)
			}));
	});

	const allSelected = $derived(
		shifts.length > 0 && shifts.every((shift) => selectedShiftIds.has(getShiftId(shift)))
	);

	const someSelected = $derived(selectedShiftIds.size > 0 && !allSelected);
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-w-2xl max-h-[90vh] overflow-hidden">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<CalendarIcon class="h-5 w-5" />
				Bulk Assign Shifts
			</Dialog.Title>
			<Dialog.Description>Select multiple shifts to commit to all at once.</Dialog.Description>
		</Dialog.Header>

		<div class="flex flex-col gap-4 max-h-[60vh] overflow-hidden">
			<!-- Selection Summary -->
			<div class="flex items-center justify-between p-3 bg-muted/30 rounded-lg border">
				<div class="flex items-center gap-2">
					<CheckCircleIcon class="h-4 w-4 text-primary" />
					<span class="text-sm font-medium">
						{selectedShiftIds.size} of {shifts.length} shifts selected
					</span>
				</div>

				<div class="flex items-center gap-2">
					<Checkbox
						checked={allSelected}
						indeterminate={someSelected}
						onCheckedChange={toggleSelectAll}
					/>
					<span class="text-sm">
						{allSelected ? 'Deselect All' : 'Select All'}
					</span>
				</div>
			</div>

			<!-- Shifts List -->
			<div class="flex-1 overflow-y-auto space-y-4">
				{#if shifts.length === 0}
					<div class="text-center py-8">
						<CalendarIcon class="h-12 w-12 mx-auto mb-2 text-muted-foreground" />
						<p class="text-muted-foreground">No shifts available for bulk assignment</p>
					</div>
				{:else}
					{#each groupedShifts as { date, shifts: dateShifts } (date)}
						<div class="space-y-2">
							<h3 class="text-sm font-semibold text-muted-foreground border-b pb-1">
								{date}
							</h3>

							<div class="space-y-2">
								{#each dateShifts as shift (getShiftId(shift))}
									{@const shiftId = getShiftId(shift)}
									{@const isSelected = selectedShiftIds.has(shiftId)}

									<div
										class="flex items-center gap-3 p-3 border rounded-lg transition-colors cursor-pointer hover:bg-accent/50 {isSelected
											? 'border-primary bg-primary/10'
											: ''}"
										onclick={() => toggleShift(shift)}
										role="button"
										tabindex="0"
										onkeydown={(e) => e.key === 'Enter' && toggleShift(shift)}
									>
										<Checkbox checked={isSelected} onCheckedChange={() => toggleShift(shift)} />

										<div class="flex-1 min-w-0">
											<div class="flex items-center gap-2 mb-1">
												<ClockIcon class="h-4 w-4 text-muted-foreground" />
												<span class="font-medium text-sm">
													{formatShiftTime(shift.start_time, shift.end_time)}
												</span>
											</div>

											<div class="flex items-center gap-2">
												<Badge variant="secondary" class="text-xs">
													{shift.schedule_name}
												</Badge>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				{/if}
			</div>
		</div>

		<!-- Summary and Actions -->
		{#if hasSelection}
			<div class="border-t pt-4">
				<div class="bg-primary/10 border border-primary/20 rounded-lg p-3 mb-4">
					<h4 class="text-sm font-medium mb-2">Commitment Summary</h4>
					<div class="text-xs text-muted-foreground space-y-1">
						<p>
							You will be committed to {selectedShiftIds.size} shift{selectedShiftIds.size === 1
								? ''
								: 's'}:
						</p>
						<ul class="list-disc list-inside space-y-0.5 max-h-20 overflow-y-auto">
							{#each selectedShifts.slice(0, 5) as shift}
								<li>{shift.schedule_name} - {formatShiftTime(shift.start_time, shift.end_time)}</li>
							{/each}
							{#if selectedShifts.length > 5}
								<li class="text-primary">... and {selectedShifts.length - 5} more</li>
							{/if}
						</ul>
					</div>
				</div>
			</div>
		{/if}

		<Dialog.Footer>
			<Button variant="outline" onclick={handleCancel} disabled={$bulkBookingMutation.isPending}>
				Cancel
			</Button>
			<Button onclick={handleBulkAssign} disabled={!hasSelection || $bulkBookingMutation.isPending}>
				{#if $bulkBookingMutation.isPending}
					Committing...
				{:else}
					Commit to {selectedShiftIds.size} Shift{selectedShiftIds.size === 1 ? '' : 's'}
				{/if}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
