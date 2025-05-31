<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import GPSCapture from '$lib/components/ui/gps-capture/GPSCapture.svelte';
	import { createMutation } from '@tanstack/svelte-query';
	import { UserApiService } from '$lib/services/api/user';
	import { toast } from 'svelte-sonner';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import InfoIcon from '@lucide/svelte/icons/info';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';

	// Define and export the GeolocationData type
	export interface GeolocationData {
		latitude: number;
		longitude: number;
		accuracy: number;
		timestamp: string;
	}

	// Props
	let {
		open = $bindable(false),
		bookingId = null,
		onReportSubmitted = () => {}
	}: {
		open?: boolean;
		bookingId?: number | null;
		onReportSubmitted?: () => void;
	} = $props();

	// Form state
	let selectedSeverity = $state<string>('0');
	let message = $state('');
	let locationData = $state<GeolocationData | null>(null);

	// Severity options matching the original backup
	const severityOptions = [
		{
			value: '0',
			label: 'Normal',
			description: 'Routine patrol notes, minor observations',
			icon: CheckCircleIcon,
			color:
				'text-green-700 bg-green-50 border-green-200 dark:bg-green-950 dark:text-green-300 dark:border-green-800'
		},
		{
			value: '1',
			label: 'Suspicion',
			description: 'General incidents, noise complaints, suspicious activity',
			icon: InfoIcon,
			color:
				'text-orange-700 bg-orange-50 border-orange-200 dark:bg-orange-950 dark:text-orange-300 dark:border-orange-800'
		},
		{
			value: '2',
			label: 'Incident',
			description: 'Security threats, property damage, immediate attention needed',
			icon: AlertTriangleIcon,
			color:
				'text-red-700 bg-red-50 border-red-200 dark:bg-red-950 dark:text-red-300 dark:border-red-800'
		}
	];

	// Create report mutation
	const reportMutation = createMutation({
		mutationFn: (data: any) => {
			if (bookingId) {
				return UserApiService.createShiftReport(bookingId, data);
			} else {
				return UserApiService.createOffShiftReport(data);
			}
		},
		onSuccess: () => {
			toast.success('Report submitted successfully');
			resetForm();
			open = false;
			onReportSubmitted();
		},
		onError: (error) => {
			toast.error(`Failed to submit report: ${error.message}`);
		}
	});

	function resetForm() {
		selectedSeverity = '0';
		message = '';
		locationData = null;
	}

	function handleSubmit() {
		if (!selectedSeverity || !message.trim()) {
			toast.error('Please select severity and provide a message');
			return;
		}

		const reportData = {
			severity: parseInt(selectedSeverity),
			message: message.trim(),
			latitude: locationData?.latitude || null,
			longitude: locationData?.longitude || null,
			accuracy: locationData?.accuracy || null
		};

		$reportMutation.mutate(reportData);
	}

	function handleLocationCaptured(location: GeolocationData) {
		locationData = location;
	}

	function handleLocationError(error: string) {
		console.log('Location capture failed:', error);
		// Don't show error toast since location is optional
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Submit Incident Report</Dialog.Title>
			<Dialog.Description>
				Report any incidents, observations, or concerns from your shift or area.
			</Dialog.Description>
		</Dialog.Header>

		<div class="space-y-4">
			<!-- Location Capture -->
			<GPSCapture onLocationCaptured={handleLocationCaptured} onError={handleLocationError} />

			<!-- Severity Selection -->

			<div class="space-y-2">
				{#each severityOptions as severity (severity.value)}
					{@const IconComponent = severity.icon}
					<button
						type="button"
						class="w-full p-2 rounded-lg border-2 text-left transition-all
								{selectedSeverity === severity.value
							? severity.color
							: severity.value === '0'
								? 'border-slate-200 dark:border-slate-700 hover:border-green-300 dark:hover:border-green-600'
								: severity.value === '1'
									? 'border-slate-200 dark:border-slate-700 hover:border-orange-300 dark:hover:border-orange-600'
									: 'border-slate-200 dark:border-slate-700 hover:border-red-300 dark:hover:border-red-600'}"
						onclick={() => (selectedSeverity = severity.value)}
					>
						<div class="flex items-start gap-2">
							<IconComponent
								class="h-4 w-4 mt-0.5 {selectedSeverity === severity.value ? '' : 'text-slate-400'}"
							/>
							<div class="flex-1">
								<div class="font-medium text-sm">{severity.label}</div>
								<div class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">
									{severity.description}
								</div>
							</div>
							{#if selectedSeverity === severity.value}
								<CheckCircleIcon class="h-4 w-4 text-current" />
							{/if}
						</div>
					</button>
				{/each}
			</div>

			<!-- Incident Details -->
			<div class="space-y-2">
				<Textarea
					id="message"
					bind:value={message}
					placeholder="Describe what happened, where, and any relevant details..."
					rows={4}
					class="resize-none"
				/>
			</div>
		</div>

		<Dialog.Footer class="flex gap-2">
			<Button variant="outline" onclick={() => (open = false)} disabled={$reportMutation.isPending}>
				Cancel
			</Button>
			<Button onclick={handleSubmit} disabled={$reportMutation.isPending} class="flex-1">
				{$reportMutation.isPending ? 'Submitting...' : 'Submit Report'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
