<!--
Error Boundary Component for Svelte
Provides graceful error handling and recovery for child components
-->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import type { AppError } from '$lib/utils/errorHandling';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import BugIcon from '@lucide/svelte/icons/bug';
	import type { Snippet } from 'svelte';

	interface Props {
		fallbackMessage?: string;
		showDetails?: boolean;
		onError?: (error: AppError) => void;
		allowRetry?: boolean;
		children: Snippet;
	}

	let {
		fallbackMessage = 'Something went wrong',
		showDetails = false,
		onError,
		allowRetry = true,
		children
	}: Props = $props();

	let hasError = $state(false);
	let error = $state<AppError | null>(null);
	let errorId = $state<string>('');
	let retryCount = $state(0);
	const maxRetries = 3;

	// Generate unique error ID for debugging
	function generateErrorId(): string {
		return `ERR_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
	}

	// Handle errors from child components
	function handleError(error: Error | unknown, errorInfo?: Record<string, unknown>) {
		console.error('Error caught by ErrorBoundary:', error);
		if (errorInfo) {
			console.error('Error info:', errorInfo);
		}

		// Log error to our error handling service
		const errorMessage = error instanceof Error ? error.message : 'Unknown error';
		const errorStack = error instanceof Error ? error.stack : undefined;

		// You can add error reporting here
		// await reportError(errorMessage, errorStack, errorInfo);
	}

	// Retry functionality
	function retry(): void {
		if (retryCount < maxRetries) {
			hasError = false;
			error = null;
			errorId = '';
			retryCount++;
		}
	}

	// Reset error state
	function reset(): void {
		hasError = false;
		error = null;
		errorId = '';
		retryCount = 0;
	}

	// Set up global error listeners
	onMount(() => {
		// Listen for unhandled JavaScript errors
		window.addEventListener('error', handleError);

		// Listen for unhandled promise rejections
		window.addEventListener('unhandledrejection', handleError);

		return () => {
			window.removeEventListener('error', handleError);
			window.removeEventListener('unhandledrejection', handleError);
		};
	});

	// Clean up on component destroy
	onDestroy(() => {
		// Clean up is handled by the onMount return function
	});
</script>

{#if hasError && error}
	<!-- Error Fallback UI -->
	<div class="flex items-center justify-center min-h-[200px] p-4">
		<Card.Root class="w-full max-w-md">
			<Card.Header class="text-center">
				<div class="flex justify-center mb-2">
					<div class="p-3 bg-destructive/10 rounded-full">
						<AlertTriangleIcon class="h-6 w-6 text-destructive" />
					</div>
				</div>
				<Card.Title class="text-destructive">Component Error</Card.Title>
				<Card.Description>
					{fallbackMessage}
				</Card.Description>
			</Card.Header>

			<Card.Content class="space-y-4">
				{#if showDetails && error}
					<div class="text-xs text-muted-foreground space-y-2">
						<div>
							<strong>Error Type:</strong>
							{error.type}
						</div>
						<div>
							<strong>Message:</strong>
							{error.message}
						</div>
						{#if errorId}
							<div>
								<strong>Error ID:</strong>
								<code class="bg-muted px-1 rounded text-xs">{errorId}</code>
							</div>
						{/if}
						{#if error.code}
							<div>
								<strong>Code:</strong>
								{error.code}
							</div>
						{/if}
						<div>
							<strong>Retry Count:</strong>
							{retryCount}/{maxRetries}
						</div>
					</div>
				{/if}

				<div class="flex gap-2 justify-center">
					{#if allowRetry && retryCount < maxRetries}
						<Button size="sm" onclick={retry} variant="outline">
							<RefreshCwIcon class="h-4 w-4 mr-2" />
							Try Again ({maxRetries - retryCount} left)
						</Button>
					{/if}

					<Button size="sm" onclick={reset} variant="secondary">Reset Component</Button>
				</div>

				{#if retryCount >= maxRetries}
					<div class="text-center text-xs text-muted-foreground">
						<p>Maximum retry attempts reached.</p>
						<p>Please refresh the page or contact support if the problem persists.</p>
					</div>
				{/if}
			</Card.Content>

			<Card.Footer class="pt-2">
				<div class="text-center text-xs text-muted-foreground w-full">
					<BugIcon class="h-3 w-3 inline mr-1" />
					Error ID: <code class="text-xs">{errorId}</code>
				</div>
			</Card.Footer>
		</Card.Root>
	</div>
{:else}
	<!-- Render children when no error -->
	{@render children()}
{/if}
