<script lang="ts">
	import Button, { type ButtonProps } from './button.svelte';

	interface ButtonTestWrapperProps {
		variant?: ButtonProps['variant'];
		size?: ButtonProps['size'];
		disabled?: boolean;
		href?: string;
		onclick?: (event: MouseEvent) => void;
		text?: string;
		[key: string]: unknown;
	}

	let {
		variant = 'default',
		size = 'default',
		disabled = false,
		href = undefined,
		onclick = undefined,
		text = 'Click me',
		...restProps
	}: ButtonTestWrapperProps = $props();
</script>

<Button {variant} {size} {disabled} {href} {onclick} {...restProps}>
	{text}
</Button>
