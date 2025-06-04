<script lang="ts">
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import type { EmptyStateProps } from '$lib/types/ui';

	let {
		title,
		description,
		icon: Icon,
		action,
		className,
		id,
		'data-testid': testId,
		...props
	}: EmptyStateProps = $props();
</script>

<div
	{id}
	data-testid={testId}
	class={cn('flex flex-col items-center justify-center gap-4 py-12 text-center', className)}
	{...props}
>
	<div class="flex flex-col items-center gap-4">
		{#if Icon}
			<div class="p-4 bg-muted rounded-full">
				<Icon class="h-12 w-12 text-muted-foreground" />
			</div>
		{/if}

		<div class="space-y-2 max-w-md">
			<h3 class="text-lg font-semibold text-foreground">
				{title}
			</h3>

			{#if description}
				<p class="text-sm text-muted-foreground">
					{description}
				</p>
			{/if}
		</div>
	</div>

	{#if action}
		<Button onclick={action.onClick} class="mt-2">
			{action.label}
		</Button>
	{/if}
</div>
