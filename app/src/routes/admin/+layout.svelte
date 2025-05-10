<div class="admin-layout p-4">
	<header class="mb-4">
		<h1 class="text-xl font-bold">Admin Section</h1>
		<!-- Basic Admin Navigation can go here later -->
		<nav class="mt-2">
			<a href="/admin" class="mr-2 hover:underline">Dashboard</a>
			<a href="/admin/schedules" class="mr-2 hover:underline">Manage Schedules</a>
			<!-- Add other admin links (Users, Reports) here as they are built -->
		</nav>
		<hr class="my-4" />
	</header>
	<main>
		<slot />
	</main>
</div>
