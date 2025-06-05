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

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-w-2xl max-h-[90vh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<SmartphoneIcon class="h-5 w-5 text-blue-600" />
				Welcome to the Night Owls App!
			</Dialog.Title>
			<Dialog.Description>
				Your guide to transitioning from WhatsApp coordination to our digital platform
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-6 py-4">
			<!-- Introduction -->
			<div class="space-y-3">
				<h3 class="text-lg font-semibold text-primary">🦉 No More WhatsApp Chaos!</h3>
				<p class="text-muted-foreground">
					Gone are the days of missed messages, paper rosters, and confusion about who's on shift.
					This app keeps everything organized and makes community watch coordination seamless.
				</p>
			</div>

			<Separator />

			<!-- How to use sections -->
			<div class="space-y-6">
				<!-- Booking shifts -->
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<CalendarIcon class="h-5 w-5 text-green-600" />
						<h4 class="font-semibold">📅 Booking Your Shifts</h4>
					</div>
					<div class="ml-7 space-y-2 text-sm text-muted-foreground">
						<p>• <strong>Browse available shifts</strong> on the main dashboard</p>
						<p>• <strong>Book instantly</strong> - no waiting for WhatsApp confirmations</p>
						<p>• <strong>View your schedule</strong> - all your upcoming shifts in one place</p>
					</div>
				</div>

				<!-- Check-in process -->
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<CheckCircleIcon class="h-5 w-5 text-blue-600" />
						<h4 class="font-semibold">✅ Starting Your Shift</h4>
					</div>
					<div class="ml-7 space-y-2 text-sm text-muted-foreground">
						<p>• <strong>Check in</strong> when you arrive - just tap the check-in button</p>
						<p>• <strong>Get GPS verification</strong> - confirms you're actually on location</p>
						<p>• <strong>Early bird bonus</strong> - arrive 15+ minutes early for extra points!</p>
						<p>• <strong>No more paper sign-ins</strong> - everything's digital and secure</p>
					</div>
				</div>

				<!-- Reporting -->
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<AlertTriangleIcon class="h-5 w-5 text-orange-600" />
						<h4 class="font-semibold">📝 Incident Reporting</h4>
					</div>
					<div class="ml-7 space-y-2 text-sm text-muted-foreground">
						<p>• <strong>Quick reports</strong> - tap the Report button in the header</p>
						<p>
							• <strong>Three severity levels</strong> - from routine observations to serious incidents
						</p>
						<p>• <strong>Photo evidence</strong> - attach images directly to reports</p>
						<p>
							• <strong>Instant alerts</strong> - serious incidents notify coordinators immediately
						</p>
					</div>
				</div>

				<!-- Points and achievements -->
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<TrophyIcon class="h-5 w-5 text-yellow-600" />
						<h4 class="font-semibold">🏆 Points & Recognition</h4>
					</div>
					<div class="ml-7 space-y-2 text-sm text-muted-foreground">
						<p>• <strong>Earn points</strong> for every check-in and shift completion</p>
						<p>
							• <strong>Bonus points</strong> for weekend shifts, early check-ins, and detailed reports
						</p>
						<p>
							• <strong>Achievement badges</strong> - unlock Owlet, Solid Owl, Wise Owl, and Super Owl!
						</p>
						<p>• <strong>Leaderboard</strong> - see how you rank among fellow community members</p>
						<div class="flex gap-2 mt-2">
							<Badge variant="secondary">🦜 Owlet (1 shift)</Badge>
							<Badge variant="secondary">🦉 Solid Owl (20 shifts)</Badge>
							<Badge variant="secondary">🦅 Wise Owl (50 shifts)</Badge>
							<Badge variant="secondary">🔥 Super Owl (100 shifts)</Badge>
						</div>
					</div>
				</div>

				<!-- Emergency features -->
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<PhoneIcon class="h-5 w-5 text-red-600" />
						<h4 class="font-semibold">🚨 Emergency Support</h4>
					</div>
					<div class="ml-7 space-y-2 text-sm text-muted-foreground">
						<p>• <strong>Emergency button</strong> always visible in the header</p>
						<p>• <strong>Direct contact</strong> to current emergency coordinators</p>
						<p>• <strong>GPS location sharing</strong> - help arrives faster when needed</p>
						<p>• <strong>Backup contacts</strong> - multiple people notified for reliability</p>
					</div>
				</div>
			</div>

			<Separator />

			<!-- Why it's better -->
			<div class="space-y-3">
				<h3 class="text-lg font-semibold text-primary">🌟 Why This Beats WhatsApp</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
					<div class="space-y-2">
						<p class="font-medium text-green-700">✅ What you get:</p>
						<ul class="space-y-1 text-muted-foreground ml-4">
							<li>• No missed messages</li>
							<li>• Clear shift schedules</li>
							<li>• Automatic record keeping</li>
							<li>• Instant emergency alerts</li>
							<li>• Recognition for your service</li>
						</ul>
					</div>
					<div class="space-y-2">
						<p class="font-medium text-red-700">❌ What you avoid:</p>
						<ul class="space-y-1 text-muted-foreground ml-4">
							<li>• WhatsApp message chaos</li>
							<li>• Lost paper rosters</li>
							<li>• Forgotten shift times</li>
							<li>• Unclear who's on duty</li>
							<li>• Manual coordination stress</li>
						</ul>
					</div>
				</div>
			</div>

			<Separator />

			<!-- Getting started -->
			<div class="space-y-3">
				<h3 class="text-lg font-semibold text-primary">🚀 Ready to Start?</h3>
				<div class="space-y-2 text-sm text-muted-foreground">
					<p><strong>1.</strong> Browse available shifts on your dashboard</p>
					<p><strong>2.</strong> Book a shift that works for your schedule</p>
					<p><strong>3.</strong> Show up and check in when your shift starts</p>
					<p><strong>4.</strong> Submit a report when you finish</p>
					<p><strong>5.</strong> Watch your points grow and badges unlock!</p>
				</div>
				<div class="mt-4 p-3 bg-blue-50 rounded-lg border-l-4 border-blue-400">
					<p class="text-sm text-blue-800">
						<strong>Need help?</strong> Contact your community coordinator or use the emergency button
						if you have any questions during your shift.
					</p>
				</div>
			</div>
		</div>

		<Dialog.Footer>
			<Button onclick={() => (open = false)} class="w-full sm:w-auto">Got it, thanks!</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
