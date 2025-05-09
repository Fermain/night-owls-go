<script lang="ts">
  import '../app.css'; // Import global styles (Tailwind)
  import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
  import { SvelteQueryDevtools } from '@tanstack/svelte-query-devtools';
  import { dev } from '$app/environment'; // SvelteKit environment module

  // For client-side only SPA as per SvelteKit SPA guide
  export const ssr = false;
  // export const prerender = true; // Prerendering might not be what we want for a fully dynamic SPA with client-side data fetching.
                                  // Let's keep it simple for now and potentially revisit if specific prerendering needs arise.

  const queryClient = new QueryClient();

</script>

<QueryClientProvider client={queryClient}>
  <div class="min-h-screen flex flex-col">
    <!-- 
      Can add common layout elements here later like a Navbar or Footer 
      The main page content will be rendered in the slot.
    -->
    <slot />
  </div>
  {#if dev}
    <SvelteQueryDevtools initialIsOpen={false} />
  {/if}
</QueryClientProvider> 