<script lang="ts">
	import { TelInput } from 'svelte-tel-input';
	import type { E164Number, CountryCode } from 'svelte-tel-input/types';

	// Props
	let {
		value = $bindable(),
		valid = $bindable(),
		disabled = false,
		readonly = false,
		required = false,
		class: className = '',
		placeholder,
		...restProps
	} = $props<{
		value?: E164Number | string | null;
		valid?: boolean;
		disabled?: boolean;
		readonly?: boolean;
		required?: boolean;
		class?: string;
		placeholder?: string;
		[key: string]: unknown;
	}>();

	// Internal state - hardcode South Africa to avoid international formatting issues
	let selectedCountry = $state('ZA' as CountryCode);
	let internalValid = $state(true);

	// Sync internal validity with parent
	$effect(() => {
		if (valid !== undefined) {
			valid = internalValid;
		}
	});

	// Standard configuration to avoid bugs and ensure consistency
	const phoneInputOptions = {
		autoPlaceholder: true,
		spaces: true,
		format: 'national' as const, // Use national to avoid country code insertion issues
		invalidateOnCountryChange: false
	};
</script>

<div class="relative">
	<div
		class="absolute left-3 top-1/2 transform -translate-y-1/2 flex items-center gap-2 text-sm text-muted-foreground pointer-events-none z-10"
	>
		<span class="text-base">🇿🇦</span>
		<span class="font-mono">+27</span>
	</div>
	<TelInput
		bind:country={selectedCountry}
		bind:value
		bind:valid={internalValid}
		{disabled}
		{readonly}
		{required}
		{placeholder}
		options={phoneInputOptions}
		class="phone-input {!internalValid && value ? 'phone-input-invalid' : ''} {className}"
		{...restProps}
	/>
</div>

<style type="text/postcss">
	:global(.phone-input) {
		@apply border-input placeholder:text-muted-foreground focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent py-1 text-base shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm;
		padding-left: 4.5rem !important; /* Make room for flag and country code */
		padding-right: 0.75rem !important;
	}

	:global(.phone-input-invalid) {
		@apply border-destructive focus-visible:ring-destructive;
	}
</style>
