<script lang="ts">
	import type { AdminShiftSlot } from '$lib/types';
	import { Dialog, DialogContent, DialogHeader, DialogTitle } from '$lib/components/ui/dialog';
	import { CalendarIcon, UserPlusIcon, ClockIcon } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent } from '$lib/components/ui/card';
	import {
		generateDialogDateInfo,
		getStatusBadgeInfo,
		validateShiftForDialog,
		getShiftValidationError
	} from '$lib/utils/adminDialogs';
	import ShiftBookingForm from '$lib/components/admin/shifts/ShiftBookingForm.svelte';

	let {
		shift,
		open = $bindable(),
		onAssignSuccess
	}: {
		shift: AdminShiftSlot | null;
		open: boolean;
		onAssignSuccess?: () => void;
	} = $props();

	// Use utility functions for derived values
	const shiftDate = $derived(shift ? generateDialogDateInfo(shift) : null);
	const isValid = $derived(validateShiftForDialog(shift));
	const validationError = $derived(getShiftValidationError(shift));
	const statusBadge = $derived(shift ? getStatusBadgeInfo(shift) : null);

	function handleBookingSuccess() {
		open = false;
		onAssignSuccess?.();
	}
</script>

<Dialog bind:open>
	<DialogContent class="max-w-2xl max-h-[90vh] overflow-y-auto">
		{#if !isValid}
			<!-- Error State -->
			<DialogHeader>
				<DialogTitle class="flex items-center gap-2 text-destructive">
					<UserPlusIcon class="h-5 w-5" />
					Cannot Assign - Invalid Data
				</DialogTitle>
			</DialogHeader>
			<div class="p-4 bg-destructive/10 border border-destructive/20 rounded-md">
				<p class="text-sm text-destructive">{validationError}</p>
			</div>
		{:else if shift && shiftDate}
			<!-- Valid Shift Assignment -->
			<DialogHeader>
				<DialogTitle class="flex items-center gap-2">
					<UserPlusIcon class="h-5 w-5" />
					Assign Team to Shift
				</DialogTitle>
			</DialogHeader>

			<div class="space-y-4">
				<!-- Date & Time Context -->
				<Card>
					<CardContent class="pt-6 space-y-4">
						<!-- Full date display -->
						<div class="text-center border-b pb-4">
							<div class="text-lg font-semibold">{shiftDate.fullDate}</div>
							<div class="text-muted-foreground text-sm">{shiftDate.relative}</div>
						</div>

						<!-- Shift details -->
						<div class="space-y-3">
							<div class="flex items-center gap-2 text-sm">
								<ClockIcon class="h-4 w-4 text-muted-foreground" />
								<span class="font-medium">{shiftDate.timeRange}</span>
							</div>
							<div class="flex items-center gap-2 text-sm">
								<CalendarIcon class="h-4 w-4 text-muted-foreground" />
								<span>{shift.schedule_name}</span>
							</div>
							<div class="flex items-center gap-2">
								{#if statusBadge}
									<Badge variant={statusBadge.variant} class={statusBadge.className}>
										{statusBadge.text}
									</Badge>
								{/if}
							</div>
						</div>
					</CardContent>
				</Card>

				<!-- Assignment Form (with header hidden to avoid duplication) -->
				<ShiftBookingForm
					selectedShift={shift}
					onBookingSuccess={handleBookingSuccess}
					hideHeader={true}
					noPadding={true}
				/>
			</div>
		{/if}
	</DialogContent>
</Dialog>
