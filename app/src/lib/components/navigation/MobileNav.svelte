<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import MenuIcon from '@lucide/svelte/icons/menu';

	let { isOpen = $bindable(false) }: { isOpen?: boolean } = $props();
</script>

<!-- Mobile navigation placeholder -->
<div class="md:hidden">
	<Button variant="ghost" size="icon" onclick={() => isOpen = !isOpen}>
		<MenuIcon class="h-5 w-5" />
	</Button>
</div>

{#if isOpen}
	<div class="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm md:hidden">
		<div class="fixed inset-y-0 left-0 z-50 w-3/4 border-r bg-background p-6">
			<div class="text-center py-8 text-muted-foreground">
				<p class="text-sm">Mobile navigation will be implemented here</p>
			</div>
		</div>
	</div>
{/if} 