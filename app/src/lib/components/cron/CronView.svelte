<script lang="ts">
	import cronstrue from 'cronstrue';

	let { cronExpr }: { cronExpr: string } = $props();

	const humanizedCron = $derived.by(() => {
		if (!cronExpr || cronExpr.trim() === '') {
			return null;
		}
		try {
			return cronstrue.toString(cronExpr);
		} catch (error) {
			return 'Invalid CRON expression';
		}
	});
</script>

{#if humanizedCron}
	<div class="text-sm text-muted-foreground bg-muted/50 p-2 rounded border">
		<span class="font-medium">Schedule:</span>
		{humanizedCron}
	</div>
{/if}
