<script lang="ts">
	import { CronExpressionParser, type CronExpression, type CronField } from 'cron-parser';

	export let cronExpr: string;

	let activeMonthIndices: number[] = [];
	let activeDayOfWeekIndices: number[] = [];
	let activeHourValues: number[] = []; // Still stores 0-23

	let error: string | null = null;
	let isLoading = true;

	const monthNames = [
		'Jan',
		'Feb',
		'Mar',
		'Apr',
		'May',
		'Jun',
		'Jul',
		'Aug',
		'Sep',
		'Oct',
		'Nov',
		'Dec'
	];
	const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

	$: displayedMonths = activeMonthIndices.map((i) => monthNames[i]);
	$: displayedDays = activeDayOfWeekIndices.map((i) => dayNames[i]);

	$: displayedHoursAMPM = activeHourValues.map((h) => {
		const hour12 = h % 12 || 12;
		const ampm = h < 12 || h === 24 ? 'AM' : 'PM';
		return `${hour12} ${ampm}`;
	});

	$: {
		isLoading = true;
		error = null;
		activeMonthIndices = [];
		activeDayOfWeekIndices = [];
		activeHourValues = [];

		if (cronExpr && cronExpr.trim() !== '') {
			try {
				const parts = cronExpr.split(' ').filter((p) => p !== '');
				if (parts.length < 5) {
					throw new Error('Invalid CRON: Expected at least 5 parts.');
				}

				const expression: CronExpression = CronExpressionParser.parse(cronExpr);

				const getValuesFromField = (field: CronField, fieldName: string): number[] => {
					if (field && typeof (field as any).getValues === 'function') {
						return (field as any).getValues();
					} else if (field && Array.isArray((field as any).values)) {
						return (field as any).values;
					} else {
						console.warn(
							`Could not directly access ${fieldName} values via .getValues() or .values, attempting iteration as fallback.`
						);
						const fallbackValues: number[] = [];
						if (field) {
							for (const val of field as any) {
								fallbackValues.push(Number(val));
							}
						}
						return fallbackValues;
					}
				};

				if (parts[3] === '*') {
					activeMonthIndices = monthNames.map((_, i: number) => i);
				} else {
					const monthValues = getValuesFromField(expression.fields.month, 'month');
					activeMonthIndices = monthValues.map((m: number) => m - 1).sort((a, b) => a - b);
				}

				if (parts[4] === '*') {
					activeDayOfWeekIndices = dayNames.map((_, i: number) => i);
				} else {
					const dayOfWeekValues = getValuesFromField(expression.fields.dayOfWeek, 'dayOfWeek');
					activeDayOfWeekIndices = dayOfWeekValues
						.map((d: number) => {
							if (d === 0 || d === 7) return 6; // Sun
							return d - 1; // Mon-Sat
						})
						.sort((a, b) => a - b);
				}

				if (parts[1] === '*') {
					activeHourValues = Array.from({ length: 24 }, (_, i: number) => i);
				} else {
					const hourValues = getValuesFromField(expression.fields.hour, 'hour');
					activeHourValues = hourValues.sort((a, b) => a - b);
				}
			} catch (e: any) {
				error = e.message || 'Invalid CRON expression';
				console.error('Error parsing CRON:', cronExpr, e);
			} finally {
				isLoading = false;
			}
		} else if (cronExpr === '' || cronExpr === null || cronExpr === undefined) {
			error = 'CRON expression is empty.';
			isLoading = false;
		} else {
			isLoading = false;
		}
	}
</script>

<div class="cron-visualizer p-1 border rounded text-xs">
	{#if isLoading}
		<p>Loading...</p>
	{:else if error}
		<p class="text-red-500">Error: {error}</p>
	{:else if cronExpr && cronExpr.trim() !== ''}
		<div class="flex flex-wrap gap-1 items-center">
			{#if displayedMonths.length > 0}
				{#each displayedMonths as monthName (monthName)}
					<div
						class="px-1.5 py-0.5 text-center border rounded bg-blue-500 text-white hover:bg-blue-600 transition-colors duration-150"
						title={`Month: ${monthName}`}
					>
						{monthName}
					</div>
				{/each}
			{/if}

			{#if displayedDays.length > 0}
				{#each displayedDays as dayName (dayName)}
					<div
						class="px-1.5 py-0.5 text-center border rounded bg-green-500 text-white hover:bg-green-600 transition-colors duration-150"
						title={`Day: ${dayName}`}
					>
						{dayName}
					</div>
				{/each}
			{/if}

			{#if displayedHoursAMPM.length > 0}
				{#each displayedHoursAMPM as hourStr (hourStr)}
					<div
						class="px-1.5 py-0.5 text-center border rounded bg-purple-500 text-white hover:bg-purple-600 transition-colors duration-150 min-w-[3.5em]"
						title={`Hour: ${hourStr}`}
					>
						{hourStr}
					</div>
				{/each}
			{/if}

			{#if displayedMonths.length === 0 && displayedDays.length === 0 && displayedHoursAMPM.length === 0 && !error}
				<p class="italic">
					Expression implies all instances or is too broad for this condensed view.
				</p>
			{/if}
		</div>
	{:else if !cronExpr || cronExpr.trim() === ''}
		<p>No CRON expression provided.</p>
	{/if}
</div>
