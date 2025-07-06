<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from 'svelte-sonner';
	import { Calendar, Copy, CheckCircle, ExternalLink, Info } from 'lucide-svelte';
	import { onMount } from 'svelte';

	interface CalendarFeedResponse {
		feed_url: string;
		webcal_url: string;
		token: string;
		expires_at: string;
		description: string;
	}

	let feedData: CalendarFeedResponse | null = null;
	let isGenerating = false;
	let copied = false;
	let isLocalhost = false;

	async function generateFeed() {
		isGenerating = true;
		try {
			const response = await fetch('/api/calendar/generate-token', {
				method: 'POST',
				headers: {
					Authorization: `Bearer ${localStorage.getItem('authToken')}`
				}
			});

			if (!response.ok) {
				throw new Error('Failed to generate calendar feed');
			}

			feedData = await response.json();
			toast.success('Calendar subscription created!');
		} catch (error) {
			console.error('Error generating calendar feed:', error);
			toast.error('Failed to create calendar subscription');
		} finally {
			isGenerating = false;
		}
	}

	async function copyToClipboard(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			copied = true;
			toast.success('Calendar URL copied!');
			setTimeout(() => (copied = false), 2000);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
			toast.error('Failed to copy to clipboard');
		}
	}

	function openCalendarApp(url: string) {
		window.open(url, '_blank');
	}

	onMount(() => {
		isLocalhost =
			window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

		// Check if user already has a feed token stored
		const existingFeed = localStorage.getItem('calendarFeed');
		if (existingFeed) {
			try {
				feedData = JSON.parse(existingFeed);
			} catch (error) {
				console.error('Invalid stored calendar feed data');
				localStorage.removeItem('calendarFeed');
			}
		}
	});

	$: if (feedData) {
		localStorage.setItem('calendarFeed', JSON.stringify(feedData));
	}
</script>

<Card.Root class="w-full">
	<Card.Header>
		<div class="flex items-center gap-2">
			<Calendar class="h-5 w-5 text-primary" />
			<Card.Title>Automatic Calendar Sync</Card.Title>
		</div>
		<Card.Description>Add your shifts to your phone's calendar app automatically</Card.Description>
	</Card.Header>

	<Card.Content class="space-y-4">
		{#if isLocalhost}
			<!-- Localhost Warning -->
			<div class="flex items-start gap-2 p-3 bg-amber-50 border border-amber-200 rounded-lg">
				<Info class="h-4 w-4 text-amber-600 mt-0.5 flex-shrink-0" />
				<div class="text-sm text-amber-800">
					<p class="font-medium">Development Mode</p>
					<p>Calendar sync only works when the app is deployed online, not on localhost.</p>
				</div>
			</div>
		{/if}

		{#if !feedData}
			<!-- Generate Feed -->
			<div class="text-center space-y-3">
				<Button onclick={generateFeed} disabled={isGenerating || isLocalhost} class="w-full">
					{#if isGenerating}
						<div
							class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent mr-2"
						></div>
						Setting up...
					{:else}
						<Calendar class="h-4 w-4 mr-2" />
						Set Up Calendar Sync
					{/if}
				</Button>
				<p class="text-sm text-muted-foreground">
					Your shifts will automatically appear in your calendar app
				</p>
			</div>
		{:else}
			<!-- Feed Generated -->
			<div class="space-y-4">
				<div class="flex items-center justify-center">
					<Badge variant="default" class="flex items-center gap-1">
						<CheckCircle class="h-3 w-3" />
						Calendar Sync Ready
					</Badge>
				</div>

				<div class="text-center space-y-3">
					<div class="flex gap-2">
						<Button
							variant="outline"
							onclick={() => feedData && openCalendarApp(feedData.webcal_url)}
							class="flex-1"
							disabled={isLocalhost}
						>
							<ExternalLink class="h-4 w-4 mr-2" />
							Open in Calendar
						</Button>
						<Button
							variant="outline"
							onclick={() => feedData && copyToClipboard(feedData.webcal_url)}
							class="flex-1"
						>
							{#if copied}
								<CheckCircle class="h-4 w-4 mr-2" />
								Copied!
							{:else}
								<Copy class="h-4 w-4 mr-2" />
								Copy Link
							{/if}
						</Button>
					</div>

					<p class="text-xs text-muted-foreground">
						Paste this link in your calendar app to automatically sync your shifts
					</p>
				</div>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
