<script lang="ts">
	import { cn } from '$lib/utils';
	import { getErrorMessage, isRetryableError } from '$lib/utils/errors';
	import { Button } from '$lib/components/ui/button';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import type { ErrorProps } from '$lib/types/ui';

	let {
		error,
		title = 'Something went wrong',
		showRetry = true,
		onRetry,
		className,
		id,
		'data-testid': testId,
		...props
	}: ErrorProps = $props();

	const errorMessage = $derived(error ? getErrorMessage(error) : '');
	const canRetry = $derived(error ? isRetryableError(error) : false);
	const shouldShowRetry = $derived(showRetry && canRetry && onRetry);

	function handleRetry() {
		onRetry?.();
	}
</script>

{#if error}
	<div
		{id}
		data-testid={testId}
		class={cn('flex flex-col items-center justify-center gap-4 py-8 text-center', className)}
		{...props}
	>
		<div class="flex flex-col items-center gap-3">
			<div class="p-3 bg-destructive/10 rounded-full">
				<AlertTriangleIcon class="h-8 w-8 text-destructive" />
			</div>

			<div class="space-y-2">
				<h3 class="text-lg font-semibold text-foreground">
					{title}
				</h3>

				{#if errorMessage}
					<p class="text-sm text-muted-foreground max-w-md">
						{errorMessage}
					</p>
				{/if}
			</div>
		</div>

		{#if shouldShowRetry}
			<Button variant="outline" onclick={handleRetry} class="gap-2">
				<RefreshCwIcon class="h-4 w-4" />
				Try again
			</Button>
		{/if}
	</div>
{/if}
