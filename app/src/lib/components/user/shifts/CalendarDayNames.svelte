<script lang="ts">
	const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
</script>

<div class="px-4">
	<div class="grid grid-cols-7 gap-1 text-center">
		{#each dayNames as dayName, index (index)}
			<div class="text-xs font-medium text-muted-foreground p-2">
				{dayName}
			</div>
		{/each}
	</div>
</div>
