<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Label } from '$lib/components/ui/label';
	import { PhoneInput } from '$lib/components/ui/phone-input';
	import * as InputOTP from '$lib/components/ui/input-otp';
	import { Button } from '$lib/components/ui/button';
	import { authService } from '$lib/services/authService';
	import { toast } from 'svelte-sonner';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import { formStore, saveUserData, clearUserData } from '$lib/stores/formStore';
	import { onboardingActions, onboardingState } from '$lib/stores/onboardingStore';
	import type { E164Number } from 'svelte-tel-input/types';
	import { getPageOpenGraph } from '$lib/utils/opengraph';

	// OpenGraph tags for this page
	const ogTags = getPageOpenGraph('login');

	// Redirect if already authenticated
	$effect(() => {
		if ($isAuthenticated) {
			// Check if user needs onboarding
			const needsOnboarding = onboardingActions.needsOnboarding($onboardingState);
			if (needsOnboarding) {
				goto('/onboarding', { replaceState: true });
			} else {
				// Role-based redirect: admin to admin area, others to main dashboard
				const userRole = $currentUser?.role;
				if (userRole === 'admin') {
					goto('/admin', { replaceState: true });
				} else {
					goto('/', { replaceState: true });
				}
			}
		}
	});

	// Get phone from URL params (from register redirect)
	const urlPhone = page.url.searchParams.get('phone');

	// Simple state management for pure login flow
	let step = $state<'request' | 'verify'>(urlPhone ? 'verify' : 'request');
	let phoneNumber: E164Number | null = $state(
		(urlPhone as E164Number) || ($formStore.lastPhoneNumber as E164Number) || null
	);
	let phoneValid = $state(true);
	let otpValue = $state('');
	let isLoading = $state(false);
	let isAutoVerifying = $state(false);
	let verificationSucceeded = $state(false); // Track if verification already succeeded

	const OTP_LENGTH = 6;

	// Auto-verify when OTP is complete
	$effect(() => {
		if (
			otpValue.length === OTP_LENGTH &&
			!isLoading &&
			!isAutoVerifying &&
			!verificationSucceeded &&
			step === 'verify'
		) {
			handleVerification(); // Don't set isAutoVerifying here - let handleVerification manage it
		} else if (otpValue.length < OTP_LENGTH) {
			isAutoVerifying = false;
		}
	});

	// Request OTP for existing user login
	async function handleLoginRequest(event: SubmitEvent) {
		event.preventDefault();

		if (!phoneNumber || !phoneValid) {
			toast.error('Please enter a valid phone number.');
			return;
		}

		isLoading = true;

		try {
			// Save phone number for convenience
			saveUserData(phoneNumber, ''); // No name for pure login

			// Request OTP for login (using register endpoint which handles both cases)
			await authService.register({
				phone: phoneNumber
				// No name - this is pure login
			});

			// Clean toast messaging
			toast.success('Verification code sent to your phone');
			step = 'verify';
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : 'Failed to send verification code';

			// Handle specific error cases with better messaging
			if (errorMessage.includes('Too many requests') || errorMessage.includes('rate limit')) {
				toast.error('Too many login attempts. Please wait a few minutes before trying again.', {
					duration: 5000
				});
			} else if (
				errorMessage.includes('user not found') ||
				errorMessage.includes('please register first')
			) {
				toast.error('Account not found. Please create an account first.', {
					action: {
						label: 'Register',
						onClick: () => goto('/register')
					}
				});
			} else {
				toast.error(errorMessage, { duration: 5000 });
			}
			console.error('Login request error:', error);
		} finally {
			isLoading = false;
		}
	}

	// Verify OTP and complete login
	async function handleVerification(event?: SubmitEvent) {
		event?.preventDefault();

		// Prevent duplicate verification calls or calls after success
		if (isLoading || verificationSucceeded) {
			return;
		}

		if (otpValue.length !== OTP_LENGTH) {
			if (event) {
				toast.error('Please enter the complete verification code.');
			}
			return;
		}

		if (!phoneNumber) {
			toast.error('Phone number is missing.');
			return;
		}

		// Determine if this is auto-verification (no event) and set state accordingly
		const isAutoCall = !event;
		if (isAutoCall) {
			isAutoVerifying = true;
		}

		// Set loading state to prevent duplicate calls
		isLoading = true;
		let toastId: string | number | undefined;

		try {
			// Show loading toast for auto-verification
			if (isAutoCall) {
				toastId = toast.loading('Verifying code...');
			}

			// Login with phone and OTP (no name needed)
			await authService.login(phoneNumber, '', otpValue);

			// Mark verification as succeeded to prevent any further attempts
			verificationSucceeded = true;

			// Clean up loading toast and show success
			if (toastId) {
				toast.dismiss(toastId);
			}
			toast.success('Welcome back!');

			// Navigate based on onboarding status and user role
			const needsOnboarding = onboardingActions.needsOnboarding($onboardingState);
			const userRole = $currentUser?.role;

			if (needsOnboarding) {
				goto('/onboarding', { replaceState: true });
			} else {
				// Role-based redirect: admin to admin area, others to main dashboard
				if (userRole === 'admin') {
					goto('/admin', { replaceState: true });
				} else {
					goto('/', { replaceState: true });
				}
			}
		} catch (error) {
			console.error('🚨 Login verification failed:', error);

			// Clean up loading toast
			if (toastId) {
				toast.dismiss(toastId);
			}

			const errorMessage = error instanceof Error ? error.message : 'Verification failed';
			if (errorMessage.includes('Too many requests') || errorMessage.includes('rate limit')) {
				toast.error(
					'Too many verification attempts. Please wait a few minutes before trying again.',
					{
						duration: 5000
					}
				);
			} else {
				toast.error('Invalid verification code. Please try again.');
			}

			// Clear OTP for retry
			otpValue = '';
		} finally {
			isLoading = false;
			isAutoVerifying = false;
		}
	}

	// Go back to phone entry
	function goBackToPhoneEntry() {
		step = 'request';
		otpValue = '';
		isAutoVerifying = false;
		verificationSucceeded = false; // Reset success flag when going back
	}

	// Clear saved data
	function handleClearData() {
		clearUserData();
		phoneNumber = null;
		toast.success('Saved data cleared');
	}
</script>

<svelte:head>
	<title>{ogTags.title}</title>
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.description}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImageAlt}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogType}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogSiteName}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterCard}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImageAlt}
</svelte:head>

<div class="flex flex-1 flex-col items-center justify-center gap-6 p-4 md:p-10">
	<div class="w-full max-w-md mx-auto flex flex-col gap-6">
		<!-- Header -->
		<div class="flex flex-col gap-2 text-center">
			<div class="flex justify-center mb-4">
				<div class="h-12 w-12 p-2 flex items-center justify-center">
					<img src="/logo.png" alt="Mount Moreland Night Owls" class="object-contain" />
				</div>
			</div>
			<h1 class="text-2xl font-semibold tracking-tight">
				{step === 'request' ? 'Sign In to Night Owls' : 'Enter verification code'}
			</h1>
			<p class="text-sm text-muted-foreground">
				{step === 'request'
					? 'Enter your phone number to receive a verification code'
					: `We sent a verification code to ${phoneNumber}`}
			</p>
		</div>

		<!-- Phone Number Entry -->
		{#if step === 'request'}
			<form onsubmit={handleLoginRequest} class="flex flex-col gap-6">
				<div class="flex flex-col gap-2">
					<Label for="phone">Phone Number</Label>
					<PhoneInput
						bind:value={phoneNumber}
						bind:valid={phoneValid}
						disabled={isLoading}
						required
						placeholder="Enter your phone number"
					/>
					<div class="text-xs text-muted-foreground mt-1">
						<p>We'll send you a verification code via SMS</p>
						{#if $formStore.lastPhoneNumber && phoneNumber === $formStore.lastPhoneNumber}
							<div class="flex items-center gap-2 mt-1">
								<span class="text-primary">Using saved phone number</span>
								<button
									type="button"
									class="text-muted-foreground hover:text-foreground underline"
									onclick={handleClearData}
								>
									Clear
								</button>
							</div>
						{/if}
					</div>
					{#if !phoneValid && phoneNumber}
						<p class="text-xs text-destructive">Please enter a valid phone number</p>
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

		<!-- OTP Verification -->
		{#if step === 'verify'}
			<form onsubmit={handleVerification} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label class="text-center">Verification Code</Label>
					<div class="flex justify-center">
						<InputOTP.Root
							maxlength={OTP_LENGTH}
							bind:value={otpValue}
							disabled={isLoading || isAutoVerifying}
						>
							{#snippet children({ cells })}
								<InputOTP.Group>
									{#each cells.slice(0, 3) as cell, i (i)}
										<InputOTP.Slot {cell} />
									{/each}
								</InputOTP.Group>
								<InputOTP.Separator />
								<InputOTP.Group>
									{#each cells.slice(3, OTP_LENGTH) as cell, i (i)}
										<InputOTP.Slot {cell} />
									{/each}
								</InputOTP.Group>
							{/snippet}
						</InputOTP.Root>
					</div>
					<p class="text-xs text-muted-foreground text-center">
						Code will verify automatically when complete
					</p>
				</div>

				<Button
					type="submit"
					class="w-full"
					disabled={isLoading || isAutoVerifying || otpValue.length !== OTP_LENGTH}
				>
					{#if isLoading || isAutoVerifying}
						<div
							class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
						></div>
						{isAutoVerifying ? 'Verifying...' : 'Verifying...'}
					{:else}
						Sign In
					{/if}
				</Button>

				<div class="text-center">
					<button
						type="button"
						class="text-sm text-muted-foreground underline-offset-4 hover:underline"
						onclick={goBackToPhoneEntry}
						disabled={isLoading || isAutoVerifying}
					>
						Wrong phone number? Go back
					</button>
				</div>
			</form>
		{/if}

		<!-- Navigation -->
		{#if step === 'request'}
			<div class="text-center text-sm text-muted-foreground">
				<p>
					Don't have an account?
					<a href="/register" class="underline underline-offset-4 hover:text-primary font-medium">
						Create one here
					</a>
				</p>
			</div>
		{/if}

		<!-- Help text -->
		<div class="text-center text-xs text-muted-foreground">
			{#if step === 'verify'}
				<p>Didn't receive the code? Check your phone for SMS</p>
			{:else}
				<p>By continuing, you agree to our terms of service and privacy policy</p>
			{/if}
		</div>
	</div>
</div>
