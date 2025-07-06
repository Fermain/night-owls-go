<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Dialog, DialogContent, DialogHeader, DialogTitle } from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { UserIcon, ClockIcon, CalendarIcon, PhoneIcon } from 'lucide-svelte';
	import {
		generateDialogDateInfo,
		isShiftFilled,
		getStatusBadgeInfo,
		validateShiftForDialog,
		getShiftValidationError
	} from '$lib/utils/adminDialogs';
	import type { AdminShiftSlot } from '$lib/types';

	let {
		shift,
		open = $bindable(),
		onReassignClick
	}: {
		shift: AdminShiftSlot | null;
		open: boolean;
		onReassignClick?: (shift: AdminShiftSlot) => void;
	} = $props();

	// Use utility functions for derived values
	const shiftDate = $derived(shift ? generateDialogDateInfo(shift) : null);
	const isValid = $derived(validateShiftForDialog(shift));
	const validationError = $derived(getShiftValidationError(shift));
	const shiftFilled = $derived(shift ? isShiftFilled(shift) : false);
	const statusBadge = $derived(shift ? getStatusBadgeInfo(shift) : null);
</script>

<Dialog bind:open>
	<DialogContent class="max-w-md">
		{#if !isValid}
			<!-- Error State -->
			<DialogHeader>
				<DialogTitle class="flex items-center gap-2 text-destructive">
					<CalendarIcon class="h-5 w-5" />
					Invalid Shift Data
				</DialogTitle>
			</DialogHeader>
			<div class="p-4 bg-destructive/10 border border-destructive/20 rounded-md">
				<p class="text-sm text-destructive">{validationError}</p>
			</div>
		{:else if shift && shiftDate}
			<!-- Valid Shift Display -->
			<DialogHeader>
				<DialogTitle class="flex items-center gap-2">
					<CalendarIcon class="h-5 w-5" />
					Shift Details
				</DialogTitle>
			</DialogHeader>

			<div class="space-y-4">
				<!-- Date & Time Information -->
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

				<!-- Assignment Info (if filled) -->
				{#if shiftFilled}
					<Card>
						<CardHeader class="pb-3">
							<CardTitle class="text-sm flex items-center gap-2">
								<UserIcon class="h-4 w-4" />
								Current Assignment
							</CardTitle>
						</CardHeader>
						<CardContent class="space-y-3">
							<div>
								<p class="font-medium">{shift.user_name}</p>
								{#if shift.user_phone}
									<div class="flex items-center gap-2 text-sm text-muted-foreground mt-1">
										<PhoneIcon class="h-3 w-3" />
										<span>{shift.user_phone}</span>
									</div>
								{/if}
							</div>

							{#if shift.buddy_name}
								<div class="pt-2 border-t">
									<p class="text-sm text-muted-foreground">Buddy</p>
									<p class="font-medium">{shift.buddy_name}</p>
								</div>
							{/if}

							<!-- Action Buttons -->
							<div class="flex gap-2 pt-2">
								<Button
									variant="outline"
									size="sm"
									onclick={() => {
										open = false;
										onReassignClick?.(shift);
									}}
								>
									Reassign
								</Button>
							</div>
						</CardContent>
					</Card>
				{/if}
			</div>
		{/if}
	</DialogContent>
</Dialog>
