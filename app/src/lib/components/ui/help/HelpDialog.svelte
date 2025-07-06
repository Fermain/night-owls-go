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
							<li>Shifts display date, time, duration, and required number of volunteers</li>
							<li>Green indicators show available slots, yellow shows partial booking</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Booking Process</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Click "Book" on any available shift slot</li>
							<li>Confirm your booking in the dialog that appears</li>
							<li>Booked shifts appear in "My Shifts" section on dashboard</li>
							<li>You will be automatically paired with other volunteers when applicable</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Cancellation Policy</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Cancel bookings from "My Shifts" section</li>
							<li>Cancellations must be made at least 2 hours before shift start</li>
							<li>Emergency cancellations: use Emergency button to contact coordinators</li>
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
							<li>Arrive at your designated patrol area</li>
							<li>Tap "Check In" button on your shift card</li>
							<li>Allow GPS location access when prompted</li>
							<li>System validates you are within the patrol zone</li>
							<li>Check-in confirmation appears with timestamp</li>
						</ol>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Check-in Requirements</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Must be within 100 meters of designated patrol area</li>
							<li>GPS location services must be enabled</li>
							<li>Check-in window opens 30 minutes before shift start</li>
							<li>Late check-ins (>15 minutes) are flagged for coordinator review</li>
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
							<li>Tap "Report" button in header or on shift card</li>
							<li>Select severity level: Low, Medium, or High</li>
							<li>Provide detailed description of incident or observation</li>
							<li>GPS coordinates are automatically captured</li>
							<li>Submit report for coordinator review</li>
						</ol>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Severity Guidelines</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li><strong>Low:</strong> Routine observations, minor issues, general notes</li>
							<li>
								<strong>Medium:</strong> Suspicious activity, property damage, noise complaints
							</li>
							<li>
								<strong>High:</strong> Safety threats, criminal activity, emergency situations
							</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Off-Shift Reporting</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Reports can be submitted anytime, not just during shifts</li>
							<li>Use "Off-Shift Report" option for incidents outside patrol hours</li>
							<li>Include approximate time and date of incident</li>
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
						<h4 class="font-medium text-sm mb-2">Emergency Button</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Red Emergency button is always visible in application header</li>
							<li>Displays current emergency coordinator contact information</li>
							<li>Shows backup contacts if primary coordinator unavailable</li>
							<li>GPS coordinates shared automatically when emergency button is accessed</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Critical Situations</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>For life-threatening emergencies: Call 10111 (SAPS) immediately</li>
							<li>Use Emergency button for community-specific coordination needs</li>
							<li>Emergency contacts have direct line to security response teams</li>
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
							<li>Shift completion: 15 points</li>
							<li>Incident report submission: 5 points</li>
							<li>Detailed incident reports: +5 bonus points</li>
							<li>Early check-in (15+ minutes): +3 bonus points</li>
							<li>Weekend/holiday shifts: +5 bonus points</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Achievement Levels</h4>
						<div class="flex gap-2 flex-wrap mb-2">
							<Badge variant="secondary">🐣 Owlet (25 pts)</Badge>
							<Badge variant="secondary">🦉 Solid Owl (500 pts)</Badge>
							<Badge variant="secondary">🦅 Wise Owl (1,250 pts)</Badge>
							<Badge variant="secondary">🐉 Super Owl (2,500 pts)</Badge>
						</div>
						<p class="text-sm text-muted-foreground">
							Achievement badges are displayed on your profile and visible in community
							leaderboards.
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
							<li>GPS-enabled smartphone or tablet</li>
							<li>Internet connection (cellular or Wi-Fi)</li>
							<li>Modern web browser (Chrome, Safari, Firefox, Edge)</li>
							<li>Location services enabled for this application</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Offline Functionality</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>Application data cached for offline viewing</li>
							<li>Check-ins and reports queued when offline</li>
							<li>Automatic synchronization when connection restored</li>
							<li>Emergency contacts accessible offline</li>
						</ul>
					</div>
					<div>
						<h4 class="font-medium text-sm mb-2">Data Privacy</h4>
						<ul class="text-sm text-muted-foreground space-y-1 list-disc list-inside">
							<li>GPS data used only for shift verification and incident reporting</li>
							<li>Personal information limited to community coordinators</li>
							<li>All communications encrypted in transit</li>
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
