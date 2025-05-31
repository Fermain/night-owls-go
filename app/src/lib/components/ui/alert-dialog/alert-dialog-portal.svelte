<script lang="ts">
	import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
	import type { Snippet } from 'svelte';

	let { children, ...restProps }: { children: Snippet } = $props();
</script>

<AlertDialogPrimitive.Portal {...restProps}>
	{@render children()}
</AlertDialogPrimitive.Portal>
