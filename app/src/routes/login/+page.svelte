<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { TelInput } from 'svelte-tel-input';
	import * as InputOTP from '$lib/components/ui/input-otp';
	import { Button } from '$lib/components/ui/button';
	import { authService } from '$lib/services/authService';
	import { toast } from 'svelte-sonner';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import type { E164Number, CountryCode } from 'svelte-tel-input/types';

	$effect(() => {
		if ($isAuthenticated) {
			toast.info(`Already logged in as ${$currentUser?.name || 'user'}. Redirecting...`);
			goto('/admin', { replaceState: true });
		}
	});

	// Get URL parameters for pre-filling form
	const urlPhone = $page.url.searchParams.get('phone');
	const urlName = $page.url.searchParams.get('name');

	// State management
	let step = $state<'register' | 'verify'>(urlPhone ? 'verify' : 'register');
	let phoneNumber: E164Number | null = $state(urlPhone as E164Number || null);
	let selectedCountry: CountryCode | null = $state('ZA');
	let phoneValid = $state(true);
	let name = $state(urlName || '');
	let otpValue = $state('');
	let isLoading = $state(false);

	const OTP_LENGTH = 6;

	// Step 1: Submit phone number and name to register/request OTP
	async function handleRegistration(event: SubmitEvent) {
		event.preventDefault();

		if (!phoneNumber || !phoneValid) {
			toast.error('Please enter a valid phone number.');
			return;
		}

		isLoading = true;
		try {
			await authService.register({
				phone: phoneNumber, // E164 format ready for API
				name: name.trim() || undefined
			});

			toast.success('OTP sent! Check sms_outbox.log for the code.');
			step = 'verify';
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Registration failed');
			console.error('Registration error:', error);
		} finally {
			isLoading = false;
		}
	}

	// Step 2: Verify OTP and get JWT token
	async function handleVerification(event: SubmitEvent) {
		event.preventDefault();
		if (otpValue.length !== OTP_LENGTH) {
			toast.error('Please enter the complete OTP.');
			return;
		}

		if (!phoneNumber) {
			toast.error('Phone number is missing.');
			return;
		}

		isLoading = true;
		try {
			await authService.login(phoneNumber, name, otpValue);
			toast.success('Login successful!');
			goto('/admin', { replaceState: true });
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Verification failed');
			otpValue = ''; // Clear OTP on error
			console.error('Verification error:', error);
		} finally {
			isLoading = false;
		}
	}

	function goBackToRegistration() {
		step = 'register';
		otpValue = '';
	}
</script>

<svelte:head>
	<title>Login - Community Watch</title>
</svelte:head>

{#if !$isAuthenticated}
	<div class="flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
		<div class="flex w-full max-w-sm flex-col gap-6">
			<!-- Header -->
			<div class="flex flex-col gap-2 text-center">
				<h1 class="text-2xl font-semibold tracking-tight">
					{step === 'register' ? 'Welcome to Community Watch' : 'Enter verification code'}
				</h1>
				<p class="text-sm text-muted-foreground">
					{step === 'register'
						? 'Enter your phone number to get started'
						: `We sent a verification code to ${phoneNumber}`}
				</p>
			</div>

			<!-- Registration Form -->
			{#if step === 'register'}
				<form onsubmit={handleRegistration} class="flex flex-col gap-6">
					<div class="flex flex-col gap-2">
						<Label for="name">Name (optional)</Label>
						<Input
							id="name"
							type="text"
							placeholder="Your name"
							bind:value={name}
							disabled={isLoading}
						/>
					</div>

					<div class="flex flex-col gap-2 relative pb-6">
						<Label for="phone">Phone Number</Label>
						<TelInput
							bind:country={selectedCountry}
							bind:value={phoneNumber}
							bind:valid={phoneValid}
							disabled={isLoading}
							required
							class="tel-input {!phoneValid && phoneNumber ? 'tel-input-invalid' : ''}"
						/>
						<p class="text-xs text-muted-foreground mt-1">
							We'll send you a verification code via SMS
						</p>
						{#if !phoneValid && phoneNumber}
							<p class="text-xs text-destructive mt-1">
								Please enter a valid phone number
							</p>
						{/if}
					</div>

					<Button type="submit" class="w-full" disabled={isLoading || !phoneNumber || !phoneValid}>
						{#if isLoading}
							<div
								class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
							></div>
							Sending code...
						{:else}
							Send verification code
						{/if}
					</Button>
				</form>
			{/if}

			<!-- Verification Form -->
			{#if step === 'verify'}
				<form onsubmit={handleVerification} class="flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<Label>Verification Code</Label>
						<InputOTP.Root maxlength={OTP_LENGTH} bind:value={otpValue} disabled={isLoading}>
							{#snippet children({ cells })}
								<InputOTP.Group class="justify-center">
									{#each cells.slice(0, 3) as cell, i (i)}
										<InputOTP.Slot {cell} />
									{/each}
								</InputOTP.Group>
								<InputOTP.Separator />
								<InputOTP.Group class="justify-center">
									{#each cells.slice(3, OTP_LENGTH) as cell, i (i)}
										<InputOTP.Slot {cell} />
									{/each}
								</InputOTP.Group>
							{/snippet}
						</InputOTP.Root>
					</div>

					<Button
						type="submit"
						class="w-full"
						disabled={isLoading || otpValue.length !== OTP_LENGTH}
					>
						{#if isLoading}
							<div
								class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
							></div>
							Verifying...
						{:else}
							Verify & Continue
						{/if}
					</Button>

					<div class="text-center">
						<button
							type="button"
							class="text-sm text-muted-foreground underline-offset-4 hover:underline"
							onclick={goBackToRegistration}
							disabled={isLoading}
						>
							Wrong phone number? Go back
						</button>
					</div>
				</form>
			{/if}

			<!-- Help text -->
			<div class="text-center text-xs text-muted-foreground">
				{#if step === 'verify'}
					<p>
						Didn't receive the code? Check the <code class="font-mono">sms_outbox.log</code> file
					</p>
				{:else}
					<p>By continuing, you agree to our terms of service</p>
				{/if}
			</div>
		</div>
	</div>
{:else}
	<div class="flex items-center justify-center min-h-screen bg-background p-4">
		<div class="text-center">
			<div
				class="mb-4 h-8 w-8 animate-spin rounded-full border-2 border-current border-t-transparent mx-auto"
			></div>
			<p class="text-muted-foreground">Redirecting...</p>
		</div>
	</div>
{/if}

<style>
	:global(.tel-input) {
		@apply border-input placeholder:text-muted-foreground focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-base shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm;
	}
	
	:global(.tel-input-invalid) {
		@apply border-destructive focus-visible:ring-destructive;
	}
</style>
