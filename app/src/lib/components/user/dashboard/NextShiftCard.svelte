<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import PlayIcon from '@lucide/svelte/icons/play';
	import SquareIcon from '@lucide/svelte/icons/square';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import { formatTime, getTimeUntil } from '$lib/utils/bookings';

	interface NextShift {
		start_time: string;
		end_time: string;
		schedule_name: string;
		can_checkin: boolean;
		is_active: boolean;
	}

	let {
		nextShift,
		onCheckIn,
		onCheckOut,
		onQuickReport
	}: {
		nextShift: NextShift;
		onCheckIn?: () => void;
		onCheckOut?: () => void;
		onQuickReport?: () => void;
	} = $props();
</script>

<Card.Root class="border-l-4 border-l-primary">
	<Card.Header class="pb-3">
		<div class="flex items-center justify-between">
			<Card.Title class="text-base">My next shift</Card.Title>
			<Badge variant={nextShift.is_active ? 'default' : 'secondary'}>
				{nextShift.is_active ? 'Active' : getTimeUntil(nextShift.start_time)}
			</Badge>
		</div>
	</Card.Header>
	<Card.Content class="pt-0 space-y-3">
		<div class="text-sm text-muted-foreground">
			{nextShift.schedule_name}
		</div>

		<div class="flex items-center text-sm">
			<ClockIcon class="h-4 w-4 mr-2 text-muted-foreground" />
			<span>{formatTime(nextShift.start_time)} - {formatTime(nextShift.end_time)}</span>
		</div>

		<div class="flex gap-2">
			{#if nextShift.is_active}
				<Button onclick={onCheckOut} variant="destructive" class="flex-1">
					<SquareIcon class="h-4 w-4 mr-2" />
					Check Out
				</Button>
				<Button onclick={onQuickReport} variant="outline">
					<AlertTriangleIcon class="h-4 w-4 mr-2" />
					Report
				</Button>
			{:else if nextShift.can_checkin}
				<Button onclick={onCheckIn} class="flex-1">
					<PlayIcon class="h-4 w-4 mr-2" />
					Check In
				</Button>
			{/if}
		</div>
	</Card.Content>
</Card.Root>
