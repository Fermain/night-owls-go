<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		class: className,
		children,
		...restProps
	}: {
		class?: string;
		children: any;
		[key: string]: any;
	} = $props();
</script>

<div
	class={cn(
		'grid min-w-[8rem] items-start gap-1.5 rounded-lg border border-border/50 bg-background px-2.5 py-1.5 text-xs shadow-xl',
		className
	)}
	{...restProps}
>
	{@render children()}
</div>
