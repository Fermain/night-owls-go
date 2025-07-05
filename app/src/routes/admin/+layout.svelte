<!-- Main admin layout with error boundaries for robust error handling -->
<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { currentUser } from '$lib/services/userService';
	import MobileAdminHeader from '$lib/components/admin/MobileAdminHeader.svelte';
	import UnifiedSidebar from '$lib/components/layout/UnifiedSidebar.svelte';
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
	<!-- Sidebar with error boundary -->
	<ErrorBoundary fallbackMessage="Navigation temporarily unavailable" showDetails={false}>
		<UnifiedSidebar />
	</ErrorBoundary>

	<!-- Main content area -->
	<div class="lg:pl-72">
		<!-- Header with error boundary -->
		<ErrorBoundary fallbackMessage="Header temporarily unavailable" showDetails={false}>
			<MobileAdminHeader />
		</ErrorBoundary>

		<!-- Page content with comprehensive error boundary -->
		<main class="py-10">
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
				<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
					<slot />
				</div>
			</ErrorBoundary>
		</main>
	</div>
</div>
