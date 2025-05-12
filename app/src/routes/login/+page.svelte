<script lang="ts">
	import { goto } from '$app/navigation';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input'; // For phone number input
	import * as InputOTP from '$lib/components/ui/input-otp'; // Use namespaced import
	import { Button } from '$lib/components/ui/button';
	import { fakeLogin } from '$lib/stores/authStore';
	import { toast } from 'svelte-sonner';
	import { isAuthenticated, currentUser } from '$lib/services/userService'; // Import from userService

	$effect(() => {
		if ($isAuthenticated) {
			toast.info(`Already logged in as ${$currentUser?.name || 'user'}. Redirecting...`);
			goto('/admin', { replaceState: true });
		}
	});

	let phoneNumber = $state('');
	let otpValue = $state(''); // This will be bound to InputOTP.Root
	let isLoading = $state(false);

	const OTP_LENGTH = 6;

	async function handleLoginSubmit() {
		if (phoneNumber.trim() === '' || otpValue.length !== OTP_LENGTH) {
			toast.error('Please enter a valid phone number and complete OTP.');
			return;
		}
		isLoading = true;
		await new Promise((resolve) => setTimeout(resolve, 1000));
		const fakeToken = `fake-jwt-token-for-${phoneNumber.replace(/\D/g, '')}-${Date.now()}`;
		fakeLogin(phoneNumber, fakeToken);
		// isLoading = false; // fakeLogin will update isAuthenticated, triggering the $effect to redirect
		// toast.success('Logged in successfully!'); // Toast is now handled by the $effect or can be shown before redirect
		// goto('/admin'); // Redirection is handled by the $effect
	}
</script>

<svelte:head>
	<title>Login - OTP</title>
</svelte:head>

<!-- Only render the form if not authenticated, to avoid flash of content before redirect -->
{#if !$isAuthenticated}
	<div class="flex items-center justify-center min-h-screen bg-background p-4">
		<div
			class="w-full max-w-md p-8 space-y-6 bg-card text-card-foreground rounded-lg shadow-md border"
		>
			<div class="text-center">
				<h1 class="text-3xl font-bold">Enter OTP</h1>
				<p class="text-muted-foreground">We've sent a one-time password to your phone.</p>
			</div>

			<form on:submit|preventDefault={handleLoginSubmit} class="space-y-6">
				<div>
					<Label for="phone">Phone Number</Label>
					<Input
						id="phone"
						type="tel"
						placeholder="+1234567890"
						bind:value={phoneNumber}
						disabled={isLoading}
						required
					/>
				</div>

				<div>
					<Label>One-Time Password</Label>
					<!-- Removed for="otp" as id is on Root -->
					<InputOTP.Root maxlength={OTP_LENGTH} bind:value={otpValue} disabled={isLoading}>
						{#snippet children({ cells })}
							<InputOTP.Group>
								{#each cells.slice(0, 3) as cell, i (i)}
									<InputOTP.Slot {cell} />
								{/each}
							</InputOTP.Group>
							{#if OTP_LENGTH > 3}
								<InputOTP.Separator />
								<InputOTP.Group>
									{#each cells.slice(3, OTP_LENGTH) as cell, i (i)}
										<InputOTP.Slot {cell} />
									{/each}
								</InputOTP.Group>
							{/if}
						{/snippet}
					</InputOTP.Root>
				</div>

				<Button
					type="submit"
					class="w-full"
					disabled={isLoading || otpValue.length !== OTP_LENGTH || phoneNumber.trim() === ''}
				>
					{#if isLoading}
						<svg
							class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
						>
							<circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							></circle>
							<path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
							></path>
						</svg>
						Verifying...
					{:else}
						Verify & Login
					{/if}
				</Button>
			</form>
			<p class="text-center text-sm text-muted-foreground">
				Didn't receive the code? 
				<button type="button" class="underline hover:text-primary focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 p-0 m-0 bg-transparent border-none cursor-pointer">
					Resend OTP
				</button> 
				(not implemented)
			</p>
		</div>
	</div>
{:else}
	<!-- Optional: Show a loading message or minimal content while redirecting -->
	<div class="flex items-center justify-center min-h-screen bg-background p-4">
		<p>Loading...</p>
	</div>
{/if}
