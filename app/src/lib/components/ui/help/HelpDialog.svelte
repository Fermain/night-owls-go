<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import TrophyIcon from '@lucide/svelte/icons/trophy';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import SmartphoneIcon from '@lucide/svelte/icons/smartphone';
	import StarIcon from '@lucide/svelte/icons/star';
	import RocketIcon from '@lucide/svelte/icons/rocket';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-w-2xl max-h-[90vh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<SmartphoneIcon class="h-5 w-5 text-primary" />
				Welcome to the Night Owls App!
			</Dialog.Title>
			<Dialog.Description>
				Your guide to transitioning from WhatsApp coordination to our digital platform
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4 py-4">
			<!-- How to use sections -->
			<div class="space-y-6">
				<!-- Booking shifts -->
				<section class="space-y-3">
					<h4 class="font-semibold flex items-center gap-2">
						<CalendarIcon class="h-5 w-5 text-primary" />
						Booking Your Shifts
					</h4>
					<ul class="ml-7 space-y-2 text-sm text-muted-foreground list-disc list-inside">
						<li><strong>Browse available shifts</strong> on the main dashboard</li>
						<li><strong>Book instantly</strong> - no waiting for WhatsApp confirmations</li>
						<li><strong>View your schedule</strong> - all your upcoming shifts in one place</li>
					</ul>
				</section>

				<!-- Check-in process -->
				<section class="space-y-3">
					<h4 class="font-semibold flex items-center gap-2">
						<CheckCircleIcon class="h-5 w-5 text-primary" />
						Starting Your Shift
					</h4>
					<ul class="ml-7 space-y-2 text-sm text-muted-foreground list-disc list-inside">
						<li><strong>Check in</strong> when you arrive - just tap the check-in button</li>
						<li><strong>Get GPS verification</strong> - confirms you're actually on location</li>
						<li><strong>Early bird bonus</strong> - arrive 15+ minutes early for extra points!</li>
						<li><strong>No more paper sign-ins</strong> - everything's digital and secure</li>
					</ul>
				</section>

				<!-- Reporting -->
				<section class="space-y-3">
					<h4 class="font-semibold flex items-center gap-2">
						<AlertTriangleIcon class="h-5 w-5 text-primary" />
						Incident Reporting
					</h4>
					<ul class="ml-7 space-y-2 text-sm text-muted-foreground list-disc list-inside">
						<li><strong>Quick reports</strong> - tap the Report button in the header</li>
						<li>
							<strong>Three severity levels</strong> - from routine observations to serious incidents
						</li>
						<li>
							<strong>GPS location capture</strong> - automatically record where incidents occur
						</li>
						<li>
							<strong>Detailed documentation</strong> - comprehensive incident logging for coordinators
						</li>
					</ul>
				</section>

				<!-- Points and achievements -->
				<section class="space-y-3">
					<h4 class="font-semibold flex items-center gap-2">
						<TrophyIcon class="h-5 w-5 text-primary" />
						Points & Recognition
					</h4>
					<ul class="ml-7 space-y-2 text-sm text-muted-foreground list-disc list-inside">
						<li><strong>Earn points</strong> for every check-in and shift completion</li>
						<li>
							<strong>Bonus points</strong> for weekend shifts, early check-ins, and detailed reports
						</li>
						<li>
							<strong>Achievement badges</strong> - unlock Owlet, Solid Owl, Wise Owl, and Super Owl!
						</li>
						<li><strong>Leaderboard</strong> - see how you rank among fellow community members</li>
					</ul>
					<div class="ml-7 mt-3">
						<p class="text-sm font-bold text-foreground mb-2">Achievement Badges:</p>
						<div class="flex gap-2 flex-wrap">
							<Badge class="text-lg" variant="secondary">🐣 Owlet (25 pts)</Badge>
							<Badge class="text-lg" variant="secondary">🦉 Solid Owl (500 pts)</Badge>
							<Badge class="text-lg" variant="secondary">🦅 Wise Owl (1,250 pts)</Badge>
							<Badge class="text-lg" variant="secondary">🐉 Super Owl (2,500 pts)</Badge>
						</div>
					</div>
				</section>

				<!-- Emergency features -->
				<section class="space-y-3">
					<h4 class="font-semibold flex items-center gap-2">
						<PhoneIcon class="h-5 w-5 text-primary" />
						Emergency Support
					</h4>
					<ul class="ml-7 space-y-2 text-sm text-muted-foreground list-disc list-inside">
						<li><strong>Emergency button</strong> always visible in the header</li>
						<li><strong>Direct contact</strong> to current emergency coordinators</li>
						<li><strong>GPS location sharing</strong> - help arrives faster when needed</li>
					</ul>
				</section>
			</div>

			<Separator />

			<!-- Why it's better -->
			<section class="space-y-3">
				<h3 class="text-lg font-semibold text-primary flex items-center gap-2">
					<StarIcon class="h-5 w-5" />
					Why This Beats WhatsApp
				</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
					<div class="space-y-2">
						<p class="font-medium text-green-700 dark:text-green-400">✅ What you get:</p>
						<ul class="space-y-1 text-muted-foreground ml-4 list-disc list-inside">
							<li>No missed messages</li>
							<li>Clear shift schedules</li>
							<li>Automatic record keeping</li>
							<li>Instant emergency alerts</li>
							<li>Recognition for your service</li>
						</ul>
					</div>
					<div class="space-y-2">
						<p class="font-medium text-red-700 dark:text-red-400">❌ What you avoid:</p>
						<ul class="space-y-1 text-muted-foreground ml-4 list-disc list-inside">
							<li>WhatsApp message chaos</li>
							<li>Lost paper rosters</li>
							<li>Forgotten shift times</li>
							<li>Unclear who's on duty</li>
							<li>Manual coordination stress</li>
						</ul>
					</div>
				</div>
			</section>

			<Separator />

			<!-- Getting started -->
			<section class="space-y-3">
				<h3 class="text-lg font-semibold text-primary flex items-center gap-2">
					<RocketIcon class="h-5 w-5" />
					Ready to Start?
				</h3>
				<ol class="ml-7 space-y-2 text-sm text-muted-foreground list-decimal list-inside">
					<li><strong>Browse available shifts</strong> on your dashboard</li>
					<li><strong>Book a shift</strong> that works for you</li>
					<li><strong>Show up and check in</strong> when your shift starts</li>
					<li><strong>Submit a report</strong> when you finish</li>
					<li>
						<strong>Become a <strong>legendary owl</strong> protecting your community</strong>
					</li>
				</ol>
			</section>
		</div>

		<Dialog.Footer>
			<Button onclick={() => (open = false)} class="w-full sm:w-auto">Got it, thanks!</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
