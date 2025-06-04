<script lang="ts">
	// === IMPORTS ===
	// UI Components (centralized imports)
	import {
		Button,
		Card,
		CardContent,
		Input,
		Label,
		Switch,
		Textarea,
		LoadingState,
		ErrorState
	} from '$lib/components/ui';
	import * as Select from '$lib/components/ui/select';
	import SendIcon from '@lucide/svelte/icons/send';

	// Utilities with new patterns
	import { apiGet, apiPost } from '$lib/utils/api';
	import { classifyError, getErrorMessage } from '$lib/utils/errors';
	import { toast } from 'svelte-sonner';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';

	// Types using our new domain types and API mappings
	import type { CreateBroadcastData, BroadcastAudience, User, Broadcast } from '$lib/types/domain';
	import type { BaseComponentProps } from '$lib/types/ui';
	import type { components } from '$lib/types/api';
	import {
		mapCreateBroadcastToAPIRequest,
		mapAPIBroadcastToDomain,
		mapAPIUserArrayToDomain
	} from '$lib/types/api-mappings';

	// Validation
	import { BROADCAST_AUDIENCE_LABELS } from '$lib/types/domain';

	// === COMPONENT PROPS ===
	interface BroadcastFormProps extends BaseComponentProps {
		onSuccess?: (broadcast: Broadcast) => void;
	}

	let { onSuccess, className, id, 'data-testid': testId, ...props }: BroadcastFormProps = $props();

	// === STATE MANAGEMENT ===
	const queryClient = useQueryClient();

	// Form values using our domain types
	let formValues = $state<CreateBroadcastData>({
		title: '',
		message: '',
		audience: 'all',
		pushEnabled: true,
		scheduledAt: undefined
	});

	let isScheduled = $state(false);
	let scheduledDateTime = $state('');

	// Fetch users for audience size estimation using our new API utilities
	const usersQuery = $derived(
		createQuery<User[], Error>({
			queryKey: ['allUsersForBroadcast'],
			queryFn: async () => {
				const apiUsers =
					await apiGet<components['schemas']['api.UserAPIResponse'][]>('/admin/users');
				return mapAPIUserArrayToDomain(apiUsers);
			},
			staleTime: 1000 * 60 * 5, // 5 minutes
			retry: 2
		})
	);

	// Calculate audience size
	const audienceSize = $derived.by(() => {
		const users = $usersQuery.data ?? [];
		switch (formValues.audience) {
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

	// Audience options for the select
	const audienceOptions = Object.entries(BROADCAST_AUDIENCE_LABELS).map(([value, label]) => ({
		value: value as BroadcastAudience,
		label
	}));

	// Send broadcast mutation using our new API utilities
	const sendBroadcastMutation = createMutation({
		mutationFn: async (broadcastData: CreateBroadcastData) => {
			const requestData = mapCreateBroadcastToAPIRequest(broadcastData);
			// Use the manual broadcast response type since it's not in the generated API types yet
			type BroadcastAPIResponse = {
				broadcast_id: number;
				title: string;
				message: string;
				audience: BroadcastAudience;
				sender_user_id: number;
				sender_name?: string;
				push_enabled: boolean;
				scheduled_at?: string | null;
				sent_at?: string | null;
				status: string;
				recipient_count: number;
				sent_count: number;
				failed_count: number;
				created_at: string;
			};
			const apiResponse = await apiPost<typeof requestData, BroadcastAPIResponse>(
				'/admin/broadcasts',
				requestData
			);
			return mapAPIBroadcastToDomain(apiResponse);
		},
		onSuccess: (result) => {
			toast.success(`Broadcast sent successfully to ${result.recipientCount} users!`);
			// Reset form
			formValues = {
				title: '',
				message: '',
				audience: 'all',
				pushEnabled: true,
				scheduledAt: undefined
			};
			isScheduled = false;
			scheduledDateTime = '';
			// Invalidate broadcasts list to refresh sidebar
			queryClient.invalidateQueries({ queryKey: ['broadcasts'] });
			onSuccess?.(result);
		},
		onError: (error: Error) => {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
		}
	});

	function handleSendBroadcast() {
		if (!formValues.title.trim()) {
			toast.error('Please enter a title');
			return;
		}

		if (!formValues.message.trim()) {
			toast.error('Please enter a message');
			return;
		}

		if (isScheduled && !scheduledDateTime) {
			toast.error('Please select a scheduled time');
			return;
		}

		// Update scheduledAt if scheduled
		if (isScheduled && scheduledDateTime) {
			formValues.scheduledAt = scheduledDateTime;
		} else {
			formValues.scheduledAt = undefined;
		}

		$sendBroadcastMutation.mutate(formValues);
	}
</script>

<!-- Send Broadcast Form -->
<Card {id} data-testid={testId} class={className} {...props}>
	<CardContent class="space-y-4">
		{#if $sendBroadcastMutation.isPending}
			<LoadingState isLoading={true} loadingText="Sending broadcast..." />
		{:else}
			<!-- Title -->
			<div class="space-y-2">
				<Label for="title">Title</Label>
				<Input
					id="title"
					bind:value={formValues.title}
					placeholder="Enter alert title..."
					maxlength={100}
				/>
				<div class="text-xs text-muted-foreground text-right">
					{formValues.title.length}/100 characters
				</div>
			</div>

			<!-- Message -->
			<div class="space-y-2">
				<Label for="message">Message</Label>
				<Textarea
					id="message"
					bind:value={formValues.message}
					placeholder="Type your message here..."
					rows={4}
					class="resize-none"
				/>
				<div class="text-xs text-muted-foreground text-right">
					{formValues.message.length}/500 characters
				</div>
			</div>

			<!-- Audience Selection -->
			<div class="space-y-2">
				<Label>Audience</Label>
				<Select.Root type="single" bind:value={formValues.audience}>
					<Select.Trigger>
						<span class="truncate">
							{audienceOptions.find((opt) => opt.value === formValues.audience)?.label ??
								'Select audience'}
						</span>
					</Select.Trigger>
					<Select.Content>
						{#each audienceOptions as option (option.value)}
							<Select.Item value={option.value}>{option.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
				<div class="text-xs text-muted-foreground">
					{#if $usersQuery.isLoading}
						Calculating audience size...
					{:else if $usersQuery.isError}
						<ErrorState error={$usersQuery.error} title="Failed to load users" showRetry={false} />
					{:else}
						Will reach approximately {audienceSize} users
					{/if}
				</div>
			</div>

			<!-- Push Notifications -->
			<div class="flex items-center space-x-2">
				<Switch id="push" bind:checked={formValues.pushEnabled} />
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
						<Input id="scheduledTime" type="datetime-local" bind:value={scheduledDateTime} />
					</div>
				{/if}
			</div>

			<!-- Send Button -->
			<Button
				onclick={handleSendBroadcast}
				disabled={$sendBroadcastMutation.isPending ||
					!formValues.title.trim() ||
					!formValues.message.trim()}
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

			<!-- Error State for Send Operation -->
			{#if $sendBroadcastMutation.isError}
				<ErrorState
					error={$sendBroadcastMutation.error}
					title="Failed to send broadcast"
					showRetry={true}
					onRetry={() => $sendBroadcastMutation.mutate(formValues)}
				/>
			{/if}
		{/if}
	</CardContent>
</Card>
