<script lang="ts">
	let { monthsToShow }: { monthsToShow: number } = $props();
</script>

<div class="px-4">
	<div class="flex flex-wrap gap-4 text-xs text-muted-foreground">
		<div class="flex items-center gap-1">
			<div class="w-3 h-3 bg-primary/10 border border-primary/30 rounded"></div>
			<span>Available slot</span>
		</div>
		<div class="flex items-center gap-1">
			<span class="text-base">🦉</span>
			<span>My shift</span>
		</div>
		<div class="flex items-center gap-1">
			<div class="w-3 h-3 bg-green-600 rounded"></div>
			<span>Active now</span>
		</div>
	</div>
</div>
