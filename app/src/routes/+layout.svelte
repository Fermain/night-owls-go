<script lang="ts">
  import '../app.css'; // Import global styles (Tailwind)

  // For client-side only SPA as per SvelteKit SPA guide
  export const ssr = false;
  export const prerender = true;

  // QueryClient setup will be done here later for TanStack Query
  // import { QueryClient } from '@tanstack/svelte-query';
  // import { setQueryClientContext } from '@tanstack/svelte-query';
  // const queryClient = new QueryClient();
  // setQueryClientContext(queryClient);

</script>

<div class="min-h-screen flex flex-col">
  <!-- 
    Can add common layout elements here later like a Navbar or Footer 
    The main page content will be rendered in the slot.
  -->
  <slot />
</div> 