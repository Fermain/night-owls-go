<script lang="ts">
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Switch } from '$lib/components/ui/switch';
	import * as Select from '$lib/components/ui/select';
	import SendIcon from '@lucide/svelte/icons/send';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { UserData } from '$lib/schemas/user';

	// Form state
	let message = $state('');
	let selectedAudience = $state<string>('all');
	let enablePushNotifications = $state(false);
	let isScheduled = $state(false);
	let scheduledDateTime = $state('');

	// Broadcast types
	const audienceOptions = [
		{ value: 'all', label: 'All Users' },
		{ value: 'admins', label: 'Admins Only' },
		{ value: 'owls', label: 'Owls Only' },
		{ value: 'active', label: 'Active Users (last 30 days)' }
	];

	// Fetch users for audience size estimation
	const usersQuery = $derived(
		createQuery<UserData[], Error>({
			queryKey: ['allUsersForBroadcast'],
			queryFn: async () => {
				const response = await authenticatedFetch('/api/admin/users');
				if (!response.ok) {
					throw new Error('Failed to fetch users');
				}
				return response.json();
			}
		})
	);

	// Calculate audience size
	const audienceSize = $derived.by(() => {
		const users = $usersQuery.data ?? [];
		switch (selectedAudience) {
			case 'all':
				return users.length;
			case 'admins':
				return users.filter((user) => user.role === 'admin').length;
			case 'owls':
				return users.filter((user) => user.role === 'owl' || !user.role).length;
			case 'active':
				// For now, return all users. In future, filter by last activity
				return users.length;
			default:
				return 0;
		}
	});

	// Send broadcast mutation
	const sendBroadcastMutation = createMutation({
		mutationFn: async (broadcastData: {
			message: string;
			audience: string;
			pushEnabled: boolean;
			scheduledAt?: string;
		}) => {
			// This would be implemented when the backend supports it
			// For now, simulate API call
			await new Promise((resolve) => setTimeout(resolve, 1000));

			if (Math.random() > 0.1) {
				return { success: true, id: Date.now(), sentCount: audienceSize };
			} else {
				throw new Error('Failed to send broadcast');
			}
		},
		onSuccess: (result) => {
			toast.success(`Broadcast sent successfully to ${result.sentCount} users!`);
			// Reset form
			message = '';
			selectedAudience = 'all';
			enablePushNotifications = false;
			isScheduled = false;
			scheduledDateTime = '';
		},
		onError: (error: Error) => {
			toast.error(`Failed to send broadcast: ${error.message}`);
		}
	});

	function handleSendBroadcast() {
		if (!message.trim()) {
			toast.error('Please enter a message');
			return;
		}

		if (isScheduled && !scheduledDateTime) {
			toast.error('Please select a scheduled time');
			return;
		}

		$sendBroadcastMutation.mutate({
			message: message.trim(),
			audience: selectedAudience,
			pushEnabled: enablePushNotifications,
			scheduledAt: isScheduled ? scheduledDateTime : undefined
		});
	}
</script>

<svelte:head>
	<title>Admin - Broadcasts</title>
</svelte:head>

<div class="p-6">
	<div class="max-w-2xl mx-auto">
		<div class="mb-6">
			<h1 class="text-2xl font-semibold mb-2">Send Broadcast</h1>
			<p class="text-muted-foreground">Compose and send a message to your users</p>
		</div>

		<!-- Send Broadcast Form -->
		<Card.Root class="p-6">
			<Card.Content class="space-y-6">
				<!-- Message -->
				<div class="space-y-2">
					<Label for="message">Message</Label>
					<textarea
						id="message"
						bind:value={message}
						placeholder="Type your message here..."
						rows={4}
						class="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 resize-none"
					></textarea>
					<div class="text-xs text-muted-foreground text-right">
						{message.length}/500 characters
					</div>
				</div>

				<!-- Audience Selection -->
				<div class="space-y-2">
					<Label>Audience</Label>
					<Select.Root type="single" bind:value={selectedAudience}>
						<Select.Trigger>
							{audienceOptions.find((opt) => opt.value === selectedAudience)?.label ??
								'Select audience'}
						</Select.Trigger>
						<Select.Content>
							{#each audienceOptions as option (option.value)}
								<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
					<div class="text-xs text-muted-foreground">
						{#if $usersQuery.isLoading}
							Calculating audience size...
						{:else}
							Will reach approximately {audienceSize} users
						{/if}
					</div>
				</div>

				<!-- Push Notifications -->
				<div class="flex items-center space-x-2">
					<Switch id="push" bind:checked={enablePushNotifications} />
					<Label for="push" class="text-sm cursor-pointer">Enable push notifications (SMS)</Label>
				</div>

				<!-- Scheduling -->
				<div class="space-y-2">
					<div class="flex items-center space-x-2">
						<Switch id="scheduled" bind:checked={isScheduled} />
						<Label for="scheduled" class="text-sm cursor-pointer">Schedule for later</Label>
					</div>
					{#if isScheduled}
						<div class="space-y-2">
							<Label for="scheduledTime">Send at</Label>
							<input
								id="scheduledTime"
								type="datetime-local"
								bind:value={scheduledDateTime}
								class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
							/>
						</div>
					{/if}
				</div>

				<!-- Send Button -->
				<Button
					onclick={handleSendBroadcast}
					disabled={$sendBroadcastMutation.isPending || !message.trim()}
					class="w-full"
					size="lg"
				>
					{#if $sendBroadcastMutation.isPending}
						Sending...
					{:else}
						<SendIcon class="h-4 w-4 mr-2" />
						{isScheduled ? 'Schedule Broadcast' : 'Send Now'}
					{/if}
				</Button>
			</Card.Content>
		</Card.Root>
	</div>
</div>
