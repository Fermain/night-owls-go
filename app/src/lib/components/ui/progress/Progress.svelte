<script lang="ts">
	export let value: number = 0;
	export let max: number = 100;
	export let variant: 'default' | 'success' | 'warning' | 'destructive' = 'default';

	let className = '';
	export { className as class };

	$: percentage = Math.min(Math.max(value, 0), max);
	$: progressPercentage = (percentage / max) * 100;

	function getVariantClasses(variant: string) {
		switch (variant) {
			case 'success':
				return 'bg-green-500';
			case 'warning':
				return 'bg-yellow-500';
			case 'destructive':
				return 'bg-red-500';
			default:
				return 'bg-primary';
		}
	}
</script>

<div
	class="relative h-2 w-full overflow-hidden rounded-full bg-muted {className}"
	role="progressbar"
	aria-valuenow={value}
	aria-valuemax={max}
>
	<div
		class="h-full transition-all duration-300 ease-in-out {getVariantClasses(variant)}"
		style="width: {progressPercentage}%"
	></div>
</div>
