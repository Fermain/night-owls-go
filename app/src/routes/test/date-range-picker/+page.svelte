<script lang="ts">
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';

	let testPageStartDate = $state<string | null>(null);
	let testPageEndDate = $state<string | null>(null);

	// This function will be passed as the 'change' prop to DateRangePicker
	function handlePickerChange(detail: { start: string | null; end: string | null }) {
		console.log('[Test Page] DateRangePicker emitted change:', detail);
		testPageStartDate = detail.start;
		testPageEndDate = detail.end;
	}

	// Functions to simulate parent component updating props
	function setDatesProgrammatically() {
		testPageStartDate = '2024-08-01';
		testPageEndDate = '2024-08-05';
	}

	function clearDatesProgrammatically() {
		testPageStartDate = null;
		testPageEndDate = null;
	}
</script>

<svelte:head>
	<title>Date Range Picker Test</title>
</svelte:head>

<div class="container mx-auto p-8 space-y-6">
	<h1 class="text-2xl font-bold">Date Range Picker Test Page (Updated API)</h1>

	<div class="p-4 border rounded-md">
		<h2 class="text-lg font-semibold mb-2">Component Under Test:</h2>
		<DateRangePicker
			initialStartDate={testPageStartDate}
			initialEndDate={testPageEndDate}
			change={handlePickerChange}
			placeholderText="Select a test date range"
		/>
	</div>

	<div class="p-4 border rounded-md bg-slate-50">
		<h2 class="text-lg font-semibold mb-2">Current State on Test Page:</h2>
		<p>Start Date: <code data-testid="current-start-date">{testPageStartDate || 'null'}</code></p>
		<p>End Date: <code data-testid="current-end-date">{testPageEndDate || 'null'}</code></p>
	</div>

	<div class="space-x-2">
		<button
			class="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
			onclick={setDatesProgrammatically}>Set Dates (Aug 1-5, 2024)</button
		>
		<button
			class="rounded bg-gray-500 px-4 py-2 text-white hover:bg-gray-600"
			onclick={clearDatesProgrammatically}>Clear Dates</button
		>
	</div>

	<div>
		<h3 class="text-md font-semibold mt-4">Instructions:</h3>
		<ol class="list-decimal list-inside">
			<li>Interact with the date picker.</li>
			<li>Observe if the "Current State on Test Page" updates.</li>
			<li>
				Use the buttons to set or clear dates and see if the picker reflects these prop changes.
			</li>
		</ol>
	</div>
</div>
