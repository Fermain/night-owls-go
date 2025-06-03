<script lang="ts">
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import { Switch } from '$lib/components/ui/switch';
	import * as Select from '$lib/components/ui/select';
	import SendIcon from '@lucide/svelte/icons/send';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { UserData } from '$lib/schemas/user';
	import { createBroadcast } from '$lib/queries/admin/broadcasts';
	import type { CreateBroadcastData } from '$lib/schemas/broadcast';
	import { AUDIENCE_OPTIONS } from '$lib/utils/broadcasts';

	// Form state
	let title = $state('');
	let message = $state('');
	let selectedAudience = $state<string>('all');
	let enablePushNotifications = $state(true);
	let isScheduled = $state(false);
	let scheduledDateTime = $state('');

	const queryClient = useQueryClient();

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
		mutationFn: async (broadcastData: CreateBroadcastData) => {
			return await createBroadcast(broadcastData);
		},
		onSuccess: (result) => {
			toast.success(`Broadcast sent successfully to ${result.recipient_count} users!`);
			// Reset form
			title = '';
			message = '';
			selectedAudience = 'all';
			enablePushNotifications = true;
			isScheduled = false;
			scheduledDateTime = '';
			// Invalidate broadcasts list to refresh sidebar
			queryClient.invalidateQueries({ queryKey: ['broadcasts'] });
		},
		onError: (error: Error) => {
			toast.error(`Failed to send broadcast: ${error.message}`);
		}
	});

	function handleSendBroadcast() {
		if (!title.trim()) {
			toast.error('Please enter a title');
			return;
		}

		if (!message.trim()) {
			toast.error('Please enter a message');
			return;
		}

		if (isScheduled && !scheduledDateTime) {
			toast.error('Please select a scheduled time');
			return;
		}

		const data: CreateBroadcastData = {
			title: title.trim(),
			message: message.trim(),
			audience: selectedAudience as 'all' | 'admins' | 'owls' | 'active',
			push_enabled: enablePushNotifications,
			scheduled_at: isScheduled ? scheduledDateTime : undefined
		};

		$sendBroadcastMutation.mutate(data);
	}
</script>

<!-- Send Broadcast Form -->
<Card.Root>
	<Card.Content class="space-y-4">
		<!-- Title -->
		<div class="space-y-2">
			<Label for="title">Title</Label>
			<input
				id="title"
				bind:value={title}
				placeholder="Enter alert title..."
				maxlength={100}
				class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
			/>
			<div class="text-xs text-muted-foreground text-right">
				{title.length}/100 characters
			</div>
		</div>

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
					{AUDIENCE_OPTIONS.find((opt) => opt.value === selectedAudience)?.label ??
						'Select audience'}
				</Select.Trigger>
				<Select.Content>
					{#each AUDIENCE_OPTIONS as option (option.value)}
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
			<Label
				for="push"
				class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
				>Send push alerts to mobile devices</Label
			>
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
			disabled={$sendBroadcastMutation.isPending || !title.trim() || !message.trim()}
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
