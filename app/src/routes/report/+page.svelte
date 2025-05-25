<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import InfoIcon from '@lucide/svelte/icons/info';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import SendIcon from '@lucide/svelte/icons/send';
	import { toast } from 'svelte-sonner';

	// Form state
	let selectedSeverity = $state('');
	let message = $state('');
	let isSubmitting = $state(false);

	// Mock current shift data (would come from user session)
	const mockCurrentShift = {
		id: 1,
		schedule_name: 'Night Patrol',
		start_time: '2025-05-25T00:00:00Z',
		end_time: '2025-05-25T02:00:00Z',
		location: 'Main Street Area'
	};

	const severityOptions = [
		{
			value: '0',
			label: 'Low Priority',
			description: 'Routine patrol notes, minor observations',
			color:
				'text-green-700 bg-green-50 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800',
			icon: CheckCircleIcon
		},
		{
			value: '1',
			label: 'Normal',
			description: 'General incidents, noise complaints, suspicious activity',
			color:
				'text-blue-700 bg-blue-50 border-blue-200 dark:bg-blue-950 dark:text-blue-300 dark:border-blue-800',
			icon: InfoIcon
		},
		{
			value: '2',
			label: 'High Priority',
			description: 'Security threats, property damage, immediate attention needed',
			color:
				'text-red-700 bg-red-50 border-red-200 dark:bg-red-950 dark:text-red-300 dark:border-red-800',
			icon: AlertTriangleIcon
		}
	];

	const quickTemplates = [
		'All quiet during patrol. No incidents observed.',
		'Routine patrol completed. No issues to report.',
		'Minor disturbance resolved peacefully.',
		'Suspicious activity observed - [describe details]',
		'Property damage discovered - [describe location and extent]',
		'Noise complaint from residents - [provide details]'
	];

	// Get current time formatted
	function getCurrentTime() {
		return new Date().toLocaleString('en-GB', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getSeverityOption(value: string) {
		return severityOptions.find((opt) => opt.value === value);
	}

	function useTemplate(template: string) {
		message = template;
	}

	async function handleSubmit() {
		if (!selectedSeverity) {
			toast.error('Please select incident severity');
			return;
		}

		if (!message.trim()) {
			toast.error('Please enter incident details');
			return;
		}

		isSubmitting = true;

		// Simulate API call
		try {
			await new Promise((resolve) => setTimeout(resolve, 1000));

			// Mock successful submission
			toast.success('Incident report submitted successfully');

			// Reset form
			selectedSeverity = '';
			message = '';
		} catch (error) {
			toast.error('Failed to submit report. Please try again.');
		} finally {
			isSubmitting = false;
		}
	}

	function handleEmergency() {
		// Handle emergency call
		if (confirm('This will call emergency services immediately. Continue?')) {
			window.location.href = 'tel:999';
		}
	}
</script>

<svelte:head>
	<title>Report Incident - Night Owls</title>
</svelte:head>

<div
	class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800"
>
	<!-- Header -->
	<header
		class="bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border-b border-slate-200 dark:border-slate-700 sticky top-0 z-40"
	>
		<div class="px-4 py-3">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-lg font-semibold text-slate-900 dark:text-slate-100">Report Incident</h1>
					<p class="text-sm text-slate-600 dark:text-slate-400">Document patrol observations</p>
				</div>
				<Button variant="destructive" size="sm" class="gap-2" onclick={handleEmergency}>
					<PhoneIcon class="h-4 w-4" />
					Emergency
				</Button>
			</div>
		</div>
	</header>

	<div class="px-4 py-6 space-y-6">
		<!-- Current Shift Context -->
		{#if mockCurrentShift}
			<Card.Root class="bg-blue-50 dark:bg-blue-950/50 border-blue-200 dark:border-blue-800">
				<Card.Content class="p-4">
					<div class="flex items-center gap-3 mb-2">
						<ClockIcon class="h-5 w-5 text-blue-600 dark:text-blue-400" />
						<div>
							<h3 class="font-medium text-slate-900 dark:text-slate-100">Current Shift</h3>
							<p class="text-sm text-slate-600 dark:text-slate-400">
								{mockCurrentShift.schedule_name} • {mockCurrentShift.location}
							</p>
						</div>
					</div>
					<div class="text-xs text-slate-500 dark:text-slate-400">
						Report time: {getCurrentTime()}
					</div>
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Report Form -->
		<Card.Root>
			<Card.Header>
				<Card.Title class="text-base">Incident Details</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-6">
				<!-- Severity Selection -->
				<div class="space-y-3">
					<Label class="text-base font-medium">Incident Severity *</Label>
					<div class="space-y-2">
						{#each severityOptions as severity}
							{@const IconComponent = severity.icon}
							<button
								type="button"
								class="w-full p-4 rounded-lg border-2 text-left transition-all
									{selectedSeverity === severity.value
									? severity.color
									: 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'}"
								onclick={() => (selectedSeverity = severity.value)}
							>
								<div class="flex items-start gap-3">
									<IconComponent
										class="h-5 w-5 mt-0.5 {selectedSeverity === severity.value
											? ''
											: 'text-slate-400'}"
									/>
									<div class="flex-1">
										<div class="font-medium text-sm">{severity.label}</div>
										<div class="text-xs text-slate-600 dark:text-slate-400 mt-1">
											{severity.description}
										</div>
									</div>
									{#if selectedSeverity === severity.value}
										<CheckCircleIcon class="h-5 w-5 text-current" />
									{/if}
								</div>
							</button>
						{/each}
					</div>
				</div>

				<!-- Quick Templates -->
				<div class="space-y-3">
					<Label class="text-base font-medium">Quick Templates</Label>
					<div class="grid grid-cols-1 gap-2">
						{#each quickTemplates as template}
							<Button
								variant="outline"
								size="sm"
								class="h-auto p-3 text-left text-xs justify-start whitespace-normal"
								onclick={() => useTemplate(template)}
							>
								{template}
							</Button>
						{/each}
					</div>
				</div>

				<!-- Message Input -->
				<div class="space-y-2">
					<Label for="message" class="text-base font-medium">Incident Description *</Label>
					<Textarea
						id="message"
						bind:value={message}
						placeholder="Describe what happened, where, and any relevant details..."
						rows={6}
						class="resize-none"
					/>
					<div class="flex justify-between text-xs text-slate-500 dark:text-slate-400">
						<span>Be specific about location, time, and circumstances</span>
						<span>{message.length}/1000</span>
					</div>
				</div>

				<!-- Submit Button -->
				<Button
					onclick={handleSubmit}
					disabled={isSubmitting || !selectedSeverity || !message.trim()}
					class="w-full"
					size="lg"
				>
					{#if isSubmitting}
						Submitting...
					{:else}
						<SendIcon class="h-4 w-4 mr-2" />
						Submit Report
					{/if}
				</Button>
			</Card.Content>
		</Card.Root>

		<!-- Emergency Notice -->
		<Card.Root class="border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950/50">
			<Card.Content class="p-4">
				<div class="flex items-start gap-3">
					<AlertTriangleIcon class="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
					<div class="flex-1">
						<h3 class="font-medium text-red-900 dark:text-red-100 text-sm">Emergency Situations</h3>
						<p class="text-xs text-red-700 dark:text-red-300 mt-1">
							For immediate threats or emergencies requiring police, medical, or fire response, call
							999 directly rather than submitting a report.
						</p>
						<Button variant="destructive" size="sm" class="mt-2" onclick={handleEmergency}>
							<PhoneIcon class="h-4 w-4 mr-2" />
							Call Emergency Services
						</Button>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	</div>
</div>
