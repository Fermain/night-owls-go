<script lang="ts">
	import { cn } from '$lib/utils';
	import type { LoadingProps } from '$lib/types/ui';

	let {
		isLoading,
		loadingText = 'Loading...',
		size = 'md',
		className,
		id,
		'data-testid': testId,
		...props
	}: LoadingProps = $props();

	const sizeClasses = {
		sm: 'h-4 w-4',
		md: 'h-8 w-8',
		lg: 'h-12 w-12'
	};

	const textSizeClasses = {
		sm: 'text-sm',
		md: 'text-base',
		lg: 'text-lg'
	};
</script>

{#if isLoading}
	<div
		{id}
		data-testid={testId}
		class={cn('flex flex-col items-center justify-center gap-3 py-8', className)}
		{...props}
	>
		<div
			class={cn(
				'animate-spin rounded-full border-2 border-muted border-t-primary',
				sizeClasses[size]
			)}
		></div>
		{#if loadingText}
			<p class={cn('text-muted-foreground', textSizeClasses[size])}>
				{loadingText}
			</p>
		{/if}
	</div>
{/if}
