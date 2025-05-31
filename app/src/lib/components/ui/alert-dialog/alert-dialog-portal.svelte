<script lang="ts">
	import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';

	type $$_Props = AlertDialogPrimitive.PortalProps;
</script>

<AlertDialogPrimitive.Portal {...$$restProps}>
	<slot />
</AlertDialogPrimitive.Portal>
