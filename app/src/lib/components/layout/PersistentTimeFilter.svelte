<script lang="ts">
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import {
		selectedDayRange,
		dayRangeOptions,
		getDayRangeLabel
	} from '$lib/stores/shiftFilterStore';
	import { userSession } from '$lib/stores/authStore';
	import { page } from '$app/state';

	// Only show for authenticated users on home page
	const currentUser = $derived($userSession);
	const isHomePage = $derived(page?.url?.pathname === '/');
	const shouldShow = $derived(currentUser.isAuthenticated && isHomePage);

	const currentValue = $derived($selectedDayRange);
	const currentLabel = $derived(getDayRangeLabel(currentValue));

	function handleValueChange(value: string) {
		selectedDayRange.set(value);
	}
</script>

{#if shouldShow}
	<div class="px-4 py-2 border-b bg-muted/20">
		<div class="flex items-center gap-2">
			<Label class="text-sm font-medium whitespace-nowrap">Time Range:</Label>
			<Select.Root type="single" value={currentValue} onValueChange={handleValueChange}>
				<Select.Trigger class="h-8 text-sm">
					{currentLabel}
				</Select.Trigger>
				<Select.Content>
					{#each dayRangeOptions as option (option.value)}
						<Select.Item value={option.value} label={option.label}>
							{option.label}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>
	</div>
{/if}
