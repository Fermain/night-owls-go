<script lang="ts">
	import { goto } from '$app/navigation';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { TelInput } from 'svelte-tel-input';
	import { Button } from '$lib/components/ui/button';
	import { authService } from '$lib/services/authService';
	import { toast } from 'svelte-sonner';
	import { isAuthenticated, currentUser } from '$lib/services/userService';
	import ShieldIcon from 'lucide-svelte/icons/shield';
	import type { E164Number, CountryCode } from 'svelte-tel-input/types';

	$effect(() => {
		if ($isAuthenticated) {
			toast.info(`Already logged in as ${$currentUser?.name || 'user'}. Redirecting...`);
			goto('/admin', { replaceState: true });
		}
	});

	// State management
	let phoneNumber: E164Number | null = $state(null);
	let selectedCountry: CountryCode | null = $state('ZA');
	let phoneValid = $state(true);
	let name = $state('');
	let isLoading = $state(false);

	async function handleRegistration(event: SubmitEvent) {
		event.preventDefault();

		if (!phoneNumber || !phoneValid) {
			toast.error('Please enter a valid phone number.');
			return;
		}

		if (!name.trim()) {
			toast.error('Please enter your name.');
			return;
		}

		isLoading = true;
		try {
			await authService.register({
				phone: phoneNumber, // E164 format ready for API
				name: name.trim()
			});

			toast.success('Registration successful! Check sms_outbox.log for your verification code.');
			// Redirect to login with pre-filled phone number
			goto(`/login?phone=${encodeURIComponent(phoneNumber)}&name=${encodeURIComponent(name)}`);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Registration failed');
			console.error('Registration error:', error);
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Join Community Watch</title>
</svelte:head>

{#if !$isAuthenticated}
	<div class="grid min-h-svh lg:grid-cols-2">
		<!-- Left side - Form -->
		<div class="flex flex-col gap-4 p-6 md:p-10">
			<div class="flex justify-center gap-2 md:justify-start">
				<a href="/" class="flex items-center gap-2 font-medium">
					<div
						class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md"
					>
						<ShieldIcon class="size-4" />
					</div>
					Community Watch
				</a>
			</div>

			<div class="flex flex-1 items-center justify-center">
				<div class="w-full max-w-xs">
					<div class="flex flex-col gap-6">
						<!-- Header -->
						<div class="flex flex-col gap-2 text-center">
							<h1 class="text-2xl font-semibold tracking-tight">Join the Community Watch</h1>
							<p class="text-sm text-muted-foreground">
								Help keep our neighborhood safe. Create your account to get started.
							</p>
						</div>

						<!-- Registration Form -->
						<form onsubmit={handleRegistration} class="flex flex-col gap-6">
							<div class="flex flex-col gap-2">
								<Label for="name">Full Name</Label>
								<Input
									id="name"
									type="text"
									placeholder="John Doe"
									bind:value={name}
									disabled={isLoading}
									required
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

							<Button
								type="submit"
								class="w-full"
								disabled={isLoading || !phoneNumber || !phoneValid || !name.trim()}
							>
								{#if isLoading}
									<div
										class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
									></div>
									Creating account...
								{:else}
									Create account
								{/if}
							</Button>
						</form>

						<!-- Footer -->
						<div class="text-center text-sm text-muted-foreground">
							<p>
								Already have an account?
								<a href="/login" class="underline underline-offset-4 hover:text-primary">
									Sign in
								</a>
							</p>
						</div>

						<div class="text-center text-xs text-muted-foreground">
							<p>By creating an account, you agree to our terms of service and privacy policy.</p>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Right side - Hero/Image -->
		<div class="bg-muted relative hidden lg:block">
			<div class="absolute inset-0 bg-gradient-to-br from-primary/20 to-primary/5"></div>
			<div class="absolute inset-0 flex flex-col justify-center p-10 text-center">
				<div class="mx-auto max-w-md">
					<ShieldIcon class="mx-auto mb-6 h-16 w-16 text-primary" />
					<h2 class="mb-4 text-2xl font-bold">Stronger Together</h2>
					<p class="text-muted-foreground">
						Join your neighbors in keeping our community safe. Coordinate patrols, share important
						updates, and build lasting connections.
					</p>
				</div>
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
