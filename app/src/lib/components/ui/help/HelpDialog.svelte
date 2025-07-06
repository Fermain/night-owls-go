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
	import BookOpenIcon from '@lucide/svelte/icons/book-open';
	import InfoIcon from '@lucide/svelte/icons/info';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';

	// Import version from config
	import { APP_INFO } from '$lib/config/version';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="max-w-3xl max-h-[90vh] overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<BookOpenIcon class="h-5 w-5 text-primary" />
				Night Owls Application Manual
			</Dialog.Title>
			<Dialog.Description class="flex items-center justify-between">
				<span>Mount Moreland Community Watch Platform</span>
				<Badge variant="outline" class="ml-auto">v{APP_INFO.version}</Badge>
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-6 py-4">
			<!-- Shift Management -->
			<section class="space-y-3">
				<h3 class="font-semibold flex items-center gap-2 text-primary">
					<CalendarIcon class="h-5 w-5" />
					1. Shift Management
				</h3>
				<div class="ml-7 space-y-3">
					<div>
						<h4 class="font-medium text-sm mb-2">Viewing Available Shifts</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Access the main dashboard to view all available patrol shifts</li>
							<li>Shifts display date, time, schedule name, and booking status</li>
							<li>Open shifts show "Open shift" text, assigned shifts show volunteer names</li>
							<li>Use calendar view to see shifts organized by date</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Booking Process</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Click "Commit" on any available shift slot</li>
							<li>Optionally add a buddy name in the booking dialog</li>
							<li>Confirm your booking to receive confirmation</li>
							<li>Booked shifts appear in "My Shifts" section with check-in options</li>
							<li>Bulk assignment available for multiple shifts</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Cancellation Policy</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Cancel bookings by clicking "X" button on your shift cards</li>
							<li>Cancellations must be made at least 2 hours before shift start</li>
							<li>Emergency cancellations: contact coordinators via Emergency Contacts</li>
						</ul>
					</div>
				</div>
			</section>

			<!-- Check-in Procedure -->
			<section class="space-y-3">
				<h3 class="font-semibold flex items-center gap-2 text-primary">
					<CheckCircleIcon class="h-5 w-5" />
					2. Check-in Procedure
				</h3>
				<div class="ml-7 space-y-3">
					<div>
						<h4 class="font-medium text-sm mb-2">Standard Check-in</h4>
						<ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
							<li>Navigate to your upcoming shift on the dashboard</li>
							<li>Click "Check In" button when check-in window opens</li>
							<li>Receive immediate check-in confirmation with timestamp</li>
							<li>10 points automatically awarded for successful check-in</li>
						</ol>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Check-in Requirements</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Check-in window opens 30 minutes before shift start</li>
							<li>No GPS location verification required for check-in</li>
							<li>Early check-in (15+ minutes before start): +3 bonus points</li>
							<li>Weekend/holiday shifts: +5 bonus points automatically applied</li>
							<li>Late night shifts (22:00-05:00): +3 bonus points</li>
						</ul>
					</div>
				</div>
			</section>

			<!-- Incident Reporting -->
			<section class="space-y-3">
				<h3 class="font-semibold flex items-center gap-2 text-primary">
					<AlertTriangleIcon class="h-5 w-5" />
					3. Incident Reporting
				</h3>
				<div class="ml-7 space-y-3">
					<div>
						<h4 class="font-medium text-sm mb-2">Creating Reports</h4>
						<ol class="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
							<li>Access "Report" button from header navigation or shift card</li>
							<li>Select severity level: Normal (0), Suspicion (1), or Incident (2)</li>
							<li>Provide detailed description in the message field</li>
							<li>GPS coordinates captured automatically if location services enabled</li>
							<li>Submit report to receive 5 base points + completion points</li>
						</ol>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Severity Guidelines</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>
								<strong>Normal (0):</strong> Routine patrol notes, minor observations, general shift
								completion
							</li>
							<li>
								<strong>Suspicion (1):</strong> General incidents, noise complaints, suspicious activity
							</li>
							<li>
								<strong>Incident (2):</strong> Security threats, property damage, immediate attention
								needed (+10 bonus points)
							</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Report Completion</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Submitting a report marks your shift as completed</li>
							<li>15 completion points awarded in addition to report points</li>
							<li>Shift count increases, contributing toward achievement progress</li>
							<li>Off-shift reports can be submitted anytime without a booking</li>
						</ul>
					</div>
				</div>
			</section>

			<!-- Emergency Procedures -->
			<section class="space-y-3">
				<h3 class="font-semibold flex items-center gap-2 text-primary">
					<PhoneIcon class="h-5 w-5" />
					4. Emergency Procedures
				</h3>
				<div class="ml-7 space-y-3">
					<div>
						<h4 class="font-medium text-sm mb-2">Emergency Contacts</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Access Emergency Contacts dialog from application navigation</li>
							<li>Default contact: RUSA (086 123 4333) - Private Security Response Unit</li>
							<li>Secondary contacts include SAPS (10111) and ER24 (084 124)</li>
							<li>Contacts cached offline for access without internet connection</li>
							<li>Tap any contact to initiate direct phone call</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Critical Situations</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>
								For life-threatening emergencies: Call RUSA, SAPS (10111), and alert emergency
								WhatsApp groups
							</li>
							<li>Community ensures both SAPS and RUSA response on site</li>
							<li>For security incidents: Contact RUSA via Emergency Contacts</li>
							<li>
								Use Incident (Level 2) reports for urgent situations requiring coordination response
							</li>
						</ul>
					</div>
				</div>
			</section>

			<Separator />

			<!-- Points System -->
			<section class="space-y-3">
				<h3 class="font-semibold flex items-center gap-2 text-primary">
					<TrophyIcon class="h-5 w-5" />
					Points & Recognition System
				</h3>
				<div class="ml-7 space-y-3">
					<div>
						<h4 class="font-medium text-sm mb-2">Point Values</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Shift check-in: 10 points</li>
							<li>Shift completion (via report): 15 points</li>
							<li>Incident report submission: 5 points</li>
							<li>Level 2 (Incident) reports: +10 bonus points</li>
							<li>Early check-in (15+ minutes): +3 bonus points</li>
							<li>Weekend shifts: +5 bonus points</li>
							<li>Late night shifts (22:00-05:00): +3 bonus points</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Achievement Levels</h4>
						<div class="flex gap-2 flex-wrap mb-2">
							<Badge variant="secondary">🐣 Owlet (1 shift)</Badge>
							<Badge variant="secondary">🦉 Solid Owl (20 shifts)</Badge>
							<Badge variant="secondary">🦅 Wise Owl (50 shifts)</Badge>
							<Badge variant="secondary">🐉 Super Owl (100 shifts)</Badge>
						</div>
						<p class="text-sm text-muted-foreground">
							Achievement badges based on completed shift count. View progress on Leaderboard page.
						</p>
					</div>
				</div>
			</section>

			<Separator />

			<!-- Technical Information -->
			<section class="space-y-3">
				<h3 class="font-semibold flex items-center gap-2 text-primary">
					<InfoIcon class="h-5 w-5" />
					Technical Requirements
				</h3>
				<div class="ml-7 space-y-3">
					<div>
						<h4 class="font-medium text-sm mb-2">Device Requirements</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Smartphone or tablet with modern web browser</li>
							<li>Internet connection for real-time functionality</li>
							<li>Location services optional (for incident report GPS data)</li>
							<li>Browser compatibility: Chrome, Safari, Firefox, Edge</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Offline Functionality</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Emergency contacts cached for offline access</li>
							<li>Incident reports queued when offline, sync automatically</li>
							<li>Shift information cached for recent viewing</li>
							<li>Check-ins require internet connection</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Data Privacy</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>GPS data only captured for incident reports when location services enabled</li>
							<li>Personal information limited to community coordinators</li>
							<li>All communications encrypted using secure protocols</li>
							<li>Data retained according to community governance policies</li>
						</ul>
					</div>
				</div>
			</section>
		</div>

		<Dialog.Footer class="flex-col sm:flex-row gap-2">
			<div class="text-xs text-muted-foreground flex items-center gap-2">
				<MapPinIcon class="h-3 w-3" />
				<span>Mount Moreland Community Watch</span>
			</div>
			<Button onclick={() => (open = false)} class="w-full sm:w-auto">Close Manual</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
