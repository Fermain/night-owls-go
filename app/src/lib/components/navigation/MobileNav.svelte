<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { isAuthenticated } from '$lib/services/userService';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import UserIcon from '@lucide/svelte/icons/user';
	import MenuIcon from '@lucide/svelte/icons/menu';
	import XIcon from '@lucide/svelte/icons/x';
	import ShieldIcon from '@lucide/svelte/icons/shield';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	function closeMenu() {
		open = false;
	}
</script>

{#if $isAuthenticated}
	<!-- For authenticated users, mobile navigation is handled by the main layout -->
{:else}
	<!-- Mobile Menu Button for Unauthenticated Users -->
	<div class="md:hidden">
		<Button variant="ghost" size="icon" onclick={() => (open = !open)}>
			<MenuIcon class="h-5 w-5" />
		</Button>
	</div>

	<!-- Slide-out Menu for Unauthenticated Users -->
	{#if open}
		<div class="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm md:hidden">
			<div class="fixed inset-y-0 left-0 z-50 w-3/4 max-w-sm border-r bg-background">
				<!-- Header -->
				<div class="flex items-center justify-between p-4 border-b">
					<div class="flex items-center gap-2">
						<div class="h-8 w-8 flex items-center justify-center">
							<img src="/logo.png" alt="Mount Moreland Night Owls" class="h-6 w-6 object-contain" />
						</div>
						<span class="font-semibold">Mount Moreland Night Owls</span>
					</div>
					<Button variant="ghost" size="icon" onclick={closeMenu}>
						<XIcon class="h-5 w-5" />
					</Button>
				</div>

				<!-- Navigation Links -->
				<div class="p-4 space-y-2">
					<a
						href="/login"
						class="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
						onclick={closeMenu}
					>
						<UserIcon class="h-5 w-5" />
						<span>Sign In</span>
					</a>
					<a
						href="/register"
						class="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
						onclick={closeMenu}
					>
						<ShieldIcon class="h-5 w-5" />
						<span>Join Community</span>
					</a>
					<a
						href="/bookings"
						class="flex items-center gap-3 p-3 rounded-lg hover:bg-accent transition-colors"
						onclick={closeMenu}
					>
						<CalendarIcon class="h-5 w-5" />
						<span>View Shifts</span>
					</a>
				</div>

				<!-- Footer -->
				<div class="absolute bottom-4 left-4 right-4 text-center">
					<p class="text-xs text-muted-foreground">Night Owls Control Platform</p>
				</div>
			</div>
		</div>
	{/if}
{/if}
