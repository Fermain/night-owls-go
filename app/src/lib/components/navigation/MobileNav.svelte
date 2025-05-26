<script lang="ts">
	import { page } from '$app/state';
	import { contextualNavigation } from '$lib/stores/navigation';
	
	// Get navigation items based on user role
	const navItems = $derived($contextualNavigation.public);
</script>

<!-- Mobile bottom navigation -->
<nav class="fixed bottom-0 left-0 right-0 bg-background border-t border-border md:hidden">
	<div class="flex items-center justify-around h-16">
		{#each navItems as item (item.url)}
			{@const IconComponent = item.icon}
			<a
				href={item.url}
				class="flex flex-col items-center gap-1 px-2 py-1 text-xs
				{page.url.pathname === item.url || 
				 (item.url !== '/' && page.url.pathname.startsWith(item.url))
					? 'text-primary' 
					: 'text-muted-foreground'}"
			>
				<IconComponent class="w-5 h-5" />
				<span>{item.title}</span>
			</a>
		{/each}
	</div>
</nav> 