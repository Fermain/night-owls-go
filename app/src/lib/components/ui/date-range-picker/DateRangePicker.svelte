<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import type { DateRange as BitsDateRange } from 'bits-ui';
	import {
		DateFormatter,
		type DateValue,
		getLocalTimeZone,
		today,
		CalendarDate
	} from '@internationalized/date';
	import { cn } from '$lib/utils';
	import { buttonVariants } from '$lib/components/ui/button';
	import { RangeCalendar } from '$lib/components/ui/range-calendar';
	import * as Popover from '$lib/components/ui/popover';
	import { parseYyyyMmDdToCalendarDate, formatCalendarDateToYyyyMmDd } from '$lib/utils/date';

	const {
		initialStartDate,
		initialEndDate,
		placeholderText = 'Select a date range',
		change
	} = $props<{
		initialStartDate?: string | null;
		initialEndDate?: string | null;
		placeholderText?: string;
		change?: (r: { start: string | null; end: string | null }) => void;
	}>();

	/* — utilities & state — */
	const df = new DateFormatter('en-ZA', { dateStyle: 'medium', timeZone: 'UTC' });

	let currentRange = $state<BitsDateRange | undefined>(undefined);
	let calendarPlaceholder = $state<DateValue>(today(getLocalTimeZone()));
	let popoverOpen = $state(false);

	$effect(() => {
		if (popoverOpen) return; 

		const s = parseYyyyMmDdToCalendarDate(initialStartDate);
		const e = parseYyyyMmDdToCalendarDate(initialEndDate);

		const propStartStr = formatCalendarDateToYyyyMmDd(s);
		const propEndStr = formatCalendarDateToYyyyMmDd(e);
		const currentStartStr = formatCalendarDateToYyyyMmDd(currentRange?.start);
		const currentEndStr = formatCalendarDateToYyyyMmDd(currentRange?.end);

		if (propStartStr !== currentStartStr || propEndStr !== currentEndStr) {
			currentRange =
				s && e && s.compare(e) <= 0
					? { start: s, end: e }
					: s
						? { start: s, end: undefined }
						: undefined;
		}

		if (!currentRange?.start) {
			calendarPlaceholder = s ?? today(getLocalTimeZone());
		} else {
			calendarPlaceholder = currentRange.start;
		}
	});

	function buttonLabel(): string {
		if (currentRange?.start && currentRange?.end) {
			return `${df.format(currentRange.start.toDate('UTC'))} – ${df.format(
				currentRange.end.toDate('UTC')
			)}`;
		}
		if (currentRange?.start) return df.format(currentRange.start.toDate('UTC'));
		return placeholderText;
	}
</script>

<!-- — markup — -->
<Popover.Root bind:open={popoverOpen}>
	<Popover.Trigger
		class={cn(
			'w-full justify-start text-left font-normal',
			!currentRange?.start && 'text-muted-foreground',
			buttonVariants({ variant: 'outline' })
		)}
	>
		<CalendarIcon class="mr-2 h-4 w-4" />
		<span>{buttonLabel()}</span>
	</Popover.Trigger>

	<Popover.Content class="w-auto p-0" align="start">
		<RangeCalendar
			bind:value={currentRange}
			bind:placeholder={calendarPlaceholder}
			numberOfMonths={2}
			minValue={new CalendarDate(1900, 1, 1)}
			onValueChange={(v) => {
				if (v?.start && v.end && v.start.compare(v.end) <= 0) {
					popoverOpen = false;
					if (change) {
						const startStr = formatCalendarDateToYyyyMmDd(v.start);
						const endStr = formatCalendarDateToYyyyMmDd(v.end);
						if (startStr !== initialStartDate || endStr !== initialEndDate) {
							change({ start: startStr, end: endStr });
						}
					}
				} 
				// Add handling for partial selection or clearing if needed in the future
				// else if (v?.start && !v.end) { /* partial selection */ }
				// else if (!v?.start && !v.end && currentRange) { /* cleared */ }
			}}
		/>
	</Popover.Content>
</Popover.Root>
