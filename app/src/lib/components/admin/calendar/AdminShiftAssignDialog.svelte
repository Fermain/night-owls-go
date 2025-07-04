<script lang="ts">
	import type { AdminShiftSlot } from '$lib/types';
	import { Dialog, DialogContent, DialogHeader, DialogTitle } from '$lib/components/ui/dialog';
	import { CalendarIcon, UserPlusIcon, ClockIcon } from 'lucide-svelte';
	import { formatTimeSlot } from '$lib/utils/dateFormatting';
	import { getRelativeTime } from '$lib/utils/shifts';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent } from '$lib/components/ui/card';
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

	// Comprehensive date/time formatting
	const shiftDate = $derived.by(() => {
		if (!shift) return null;
		const date = new Date(shift.start_time);
		return {
			// Full date: "Wednesday, January 15, 2025"
			fullDate: date.toLocaleDateString('en-US', {
				weekday: 'long',
				year: 'numeric',
				month: 'long',
				day: 'numeric'
			}),
			// Time range: "18:00 - 06:00"
			timeRange: formatTimeSlot(shift.start_time, shift.end_time),
			// Relative: "in 2 days", "tomorrow", "in 3 hours"
			relative: getRelativeTime(shift.start_time)
		};
	});

	function handleBookingSuccess() {
		open = false;
		onAssignSuccess?.();
	}
</script>

<Dialog bind:open>
	<DialogContent class="max-w-2xl max-h-[90vh] overflow-y-auto">
		{#if shift && shiftDate}
			<DialogHeader>
				<DialogTitle class="flex items-center gap-2">
					<UserPlusIcon class="h-5 w-5" />
					Assign Team to Shift
				</DialogTitle>
			</DialogHeader>

			<div class="space-y-4">
				<!-- Date & Time Information (same as details dialog) -->
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
								<Badge variant="secondary" class="bg-orange-100 text-orange-700 border-orange-200">
									Needs Assignment
								</Badge>
								{#if shift.is_recurring_reservation}
									<Badge variant="outline" class="text-xs">Recurring</Badge>
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
