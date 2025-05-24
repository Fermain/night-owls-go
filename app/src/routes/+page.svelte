<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { Button } from '$lib/components/ui/button';
	import { isAuthenticated } from '$lib/services/userService';
	import { goto } from '$app/navigation';
	import type { Schedule } from '$lib/types';
	import ShieldIcon from 'lucide-svelte/icons/shield';
	import CalendarIcon from 'lucide-svelte/icons/calendar';
	import UsersIcon from 'lucide-svelte/icons/users';
	import ClockIcon from 'lucide-svelte/icons/clock';

	// Redirect authenticated users to admin panel
	$effect(() => {
		if ($isAuthenticated) {
			goto('/admin', { replaceState: true });
		}
	});

	const fetchSchedules = async (): Promise<Schedule[]> => {
		const response = await fetch('/schedules');
		if (!response.ok) {
			throw new Error('Network response was not ok');
		}
		return response.json();
	};

	const query = createQuery<Schedule[], Error>({
		queryKey: ['schedules'],
		queryFn: fetchSchedules
	});

	let data = $derived($query.data ?? []);
</script>

<svelte:head>
	<title>Community Watch - Keeping Our Neighborhood Safe</title>
	<meta
		name="description"
		content="Join the Community Watch and help keep our neighborhood safe. View patrol schedules and sign up for shifts."
	/>
</svelte:head>

<div class="min-h-screen bg-background">
	<!-- Header -->
	<header class="border-b">
		<div class="container mx-auto flex h-16 items-center justify-between px-4">
			<div class="flex items-center gap-2">
				<div
					class="bg-primary text-primary-foreground flex size-8 items-center justify-center rounded-md"
				>
					<ShieldIcon class="size-5" />
				</div>
				<span class="text-xl font-semibold">Community Watch</span>
			</div>

			<div class="flex items-center gap-2">
				<Button variant="ghost" onclick={() => goto('/login')}>Sign In</Button>
				<Button onclick={() => goto('/register')}>Join Us</Button>
			</div>
		</div>
	</header>

	<!-- Hero Section -->
	<section class="py-20 px-4">
		<div class="container mx-auto text-center">
			<div class="mx-auto max-w-3xl">
				<h1 class="mb-6 text-4xl font-bold tracking-tight sm:text-6xl">
					Protecting Our Community
					<span class="text-primary">Together</span>
				</h1>
				<p class="mb-8 text-xl text-muted-foreground">
					Join the Community Watch and be part of a network dedicated to keeping our neighborhood
					safe. Sign up for patrol shifts, connect with neighbors, and make a difference.
				</p>
				<div class="flex flex-col gap-4 sm:flex-row sm:justify-center">
					<Button size="lg" class="text-lg" onclick={() => goto('/register')}>
						<UsersIcon class="mr-2 h-5 w-5" />
						Join the Watch
					</Button>
					<Button size="lg" variant="outline" class="text-lg" onclick={() => goto('/login')}>
						Sign In
					</Button>
				</div>
			</div>
		</div>
	</section>

	<!-- Features Section -->
	<section class="py-16 px-4 bg-muted/50">
		<div class="container mx-auto">
			<div class="text-center mb-12">
				<h2 class="text-3xl font-bold mb-4">How It Works</h2>
				<p class="text-muted-foreground text-lg">
					Simple steps to get involved in community safety
				</p>
			</div>

			<div class="grid gap-8 md:grid-cols-3">
				<div class="text-center">
					<div
						class="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-primary/10"
					>
						<UsersIcon class="h-8 w-8 text-primary" />
					</div>
					<h3 class="mb-2 text-xl font-semibold">1. Join the Community</h3>
					<p class="text-muted-foreground">
						Sign up with your phone number and become part of our safety network.
					</p>
				</div>

				<div class="text-center">
					<div
						class="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-primary/10"
					>
						<CalendarIcon class="h-8 w-8 text-primary" />
					</div>
					<h3 class="mb-2 text-xl font-semibold">2. Choose Your Shifts</h3>
					<p class="text-muted-foreground">
						Browse available patrol schedules and sign up for times that work for you.
					</p>
				</div>

				<div class="text-center">
					<div
						class="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-primary/10"
					>
						<ShieldIcon class="h-8 w-8 text-primary" />
					</div>
					<h3 class="mb-2 text-xl font-semibold">3. Keep Us Safe</h3>
					<p class="text-muted-foreground">
						Patrol your neighborhood and help maintain a safe environment for everyone.
					</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Current Schedules Section -->
	<section class="py-16 px-4">
		<div class="container mx-auto">
			<div class="text-center mb-12">
				<h2 class="text-3xl font-bold mb-4">Current Patrol Schedules</h2>
				<p class="text-muted-foreground text-lg">
					See what patrol schedules are active in our community
				</p>
			</div>

			{#if $query.isLoading}
				<div class="text-center">
					<div
						class="mx-auto mb-4 h-8 w-8 animate-spin rounded-full border-2 border-current border-t-transparent"
					></div>
					<p class="text-muted-foreground">Loading schedules...</p>
				</div>
			{:else if $query.isError}
				<div class="text-center text-destructive">
					<p>Error loading schedules: {$query.error?.message}</p>
				</div>
			{:else if data && data.length > 0}
				<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
					{#each data as schedule (schedule.schedule_id)}
						<div class="rounded-lg border bg-card p-6 text-card-foreground shadow-sm">
							<div class="flex items-start justify-between">
								<div>
									<h3 class="font-semibold text-lg mb-2">{schedule.name}</h3>
									<div class="flex items-center gap-2 text-sm text-muted-foreground mb-2">
										<ClockIcon class="h-4 w-4" />
										<span>{schedule.duration_minutes} minutes per shift</span>
									</div>
									<div class="flex items-center gap-2 text-sm text-muted-foreground">
										<CalendarIcon class="h-4 w-4" />
										<span>{schedule.start_date} to {schedule.end_date}</span>
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>

				<div class="mt-8 text-center">
					<p class="text-muted-foreground mb-4">Ready to help keep our community safe?</p>
					<Button size="lg" onclick={() => goto('/register')}>
						<UsersIcon class="mr-2 h-5 w-5" />
						Join Community Watch
					</Button>
				</div>
			{:else}
				<div class="text-center">
					<p class="text-muted-foreground mb-4">No active patrol schedules at the moment.</p>
					<p class="text-sm text-muted-foreground">
						Join our community to be notified when new schedules are available.
					</p>
				</div>
			{/if}
		</div>
	</section>

	<!-- Footer -->
	<footer class="border-t py-8 px-4 bg-muted/30">
		<div class="container mx-auto text-center text-sm text-muted-foreground">
			<p>&copy; 2025 Community Watch. Making our neighborhood safer, together.</p>
		</div>
	</footer>
</div>
