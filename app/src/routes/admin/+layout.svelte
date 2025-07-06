<!-- Main admin layout with error boundaries for robust error handling -->
<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { currentUser } from '$lib/services/userService';
	import MobileAdminHeader from '$lib/components/admin/MobileAdminHeader.svelte';
	import ErrorBoundary from '$lib/components/ErrorBoundary.svelte';
	import { ErrorLogger } from '$lib/utils/errorHandling';

	// Track navigation for monitoring
	afterNavigate(({ to, from, type }) => {
		// Log navigation for audit trail in development only
		if (import.meta.env.DEV) {
			console.log('Admin navigation:', { from: from?.url, to: to?.url, type });
		}
	});
</script>

<div class="min-h-screen bg-background">
	<!-- Main content area -->
	<div>
		<!-- Header with error boundary -->
		<ErrorBoundary fallbackMessage="Header temporarily unavailable" showDetails={false}>
			<MobileAdminHeader />
		</ErrorBoundary>

		<!-- Page content with comprehensive error boundary -->
		<main>
			<ErrorBoundary
				fallbackMessage="Admin dashboard temporarily unavailable"
				showDetails={$currentUser?.role === 'admin'}
				onError={(error) => {
					// Log admin page errors with additional context
					ErrorLogger.logError({
						...error,
						details: {
							...error.details,
							adminPage: 'admin-layout',
							userRole: $currentUser?.role,
							userId: $currentUser?.id
						}
					});
				}}
			>
				<slot />
			</ErrorBoundary>
		</main>
	</div>
</div>
