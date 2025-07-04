# Frontend Architecture & Libraries (Night Owls Go - SvelteKit App)

This document outlines the frontend architecture, key libraries, and development conventions for the SvelteKit application located in the `app/` directory.

## 1. Core Framework: SvelteKit with Svelte 5

*   **SvelteKit Version:** `^2.16.0` (Confirmed from `app/package.json`)
*   **Svelte Version:** `^5.0.0` (Confirmed from `app/package.json`)
*   **Rendering Mode:** Client-Side Rendered (CSR) SPA (`export const ssr = false;` in `app/src/routes/+layout.svelte`).
*   **Key Svelte 5 Features:** This project aims to leverage Svelte 5 features such as Runes (`$state`, `$derived`, `$effect`) for efficient and readable reactive UI state management where appropriate.
*   **Official Documentation:**
    *   SvelteKit: [https://kit.svelte.dev/docs](https://kit.svelte.dev/docs)
    *   Svelte 5: [https://svelte.dev/docs/runes](https://svelte.dev/docs/runes)

## 2. Server State Management: TanStack Svelte Query

*   **Library:** `@tanstack/svelte-query`
*   **Version:** `^5.75.7` (Confirmed from `app/package.json`)
*   **Purpose:** Provides robust server state management, including data fetching, caching, background updates, mutations, and more.
*   **Setup:**
    *   A `QueryClient` is instantiated in `app/src/routes/+layout.svelte`.
    *   The `<QueryClientProvider client={queryClient}>` wraps the main application slot in `app/src/routes/+layout.svelte`, making the client accessible to all pages.
*   **Basic Usage (`createQuery`):
    ```svelte
    <script lang="ts">
      import { createQuery } from '@tanstack/svelte-query';

      const fetchSomeData = async () => {
        const response = await fetch('/api/some-data');
        if (!response.ok) throw new Error('Network response was not ok');
        return response.json();
      };

      const query = createQuery({
        queryKey: ['someData'],
        queryFn: fetchSomeData,
      });
    </script>

    {#if $query.isLoading}Loading...{/if}
    {#if $query.error}Error: {$query.error.message}{/if}
    {#if $query.data}Data: {JSON.stringify($query.data)}{/if}
    ```
*   **Mutations (`createMutation`):** (Details to be added after implementation)
*   **Query Keys:** (Conventions to be established)
*   **Devtools:** 
    *   Library: `@tanstack/svelte-query-devtools`
    *   Version: `^5.75.7` (Confirmed from `app/package.json`)
    *   Installed via `pnpm add @tanstack/svelte-query-devtools --save-dev` (in `app/` directory).
    *   Imported and rendered in `app/src/routes/+layout.svelte`:
        ```svelte
        <script lang="ts">
          // ... other imports
          import { SvelteQueryDevtools } from '@tanstack/svelte-query-devtools';
          import { dev } from '$app/environment';
          // ...
        </script>
        
        // ... QueryClientProvider wrapper ...
        {#if dev}
          <SvelteQueryDevtools initialIsOpen={false} />
        {/if}
        ```
    *   Accessible in development mode by clicking the Svelte Query logo icon on the page.
*   **Official Documentation:** [https://tanstack.com/query/latest/docs/svelte/overview](https://tanstack.com/query/latest/docs/svelte/overview)

*   **Fetching with Dynamic Query Parameters (e.g., Search):**
    A common pattern is to fetch a list of items based on a dynamic search term. This involves making the Svelte Query `queryKey` reactive to the search term and passing the term to the `queryFn`.

    *Example from `app/src/routes/admin/users/+layout.svelte` (User Search):*
    ```svelte
    <script lang="ts">
      import { Input } from '$lib/components/ui/input/index.js';
      import { createQuery } from '@tanstack/svelte-query';
      // ... other imports ...
      type UserData = { id: number; name: string; /* ... other fields */ };

      let searchTerm = $state(''); // Svelte 5 rune for reactive search term

      const fetchUsers = async (currentSearchTerm: string) => {
        let url = '/api/admin/users';
        if (currentSearchTerm) {
          // Ensure the backend API expects a 'search' query parameter
          url += `?search=${encodeURIComponent(currentSearchTerm)}`;
        }
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error('Failed to fetch users');
        }
        return response.json() as Promise<UserData[]>;
      };

      const usersQuery = createQuery<UserData[], Error, UserData[], [string, string]>({
        // The queryKey array includes the reactive searchTerm.
        // When searchTerm changes, Svelte Query treats it as a new query and will refetch.
        queryKey: ['adminUsers', searchTerm],
        // The queryFn is an arrow function to capture the current searchTerm value.
        queryFn: () => fetchUsers(searchTerm)
      });
    </script>

    <!-- Search Input -->
    <Input type="search" placeholder="Search users..." bind:value={searchTerm} />

    <!-- Displaying results -->
    {#if $usersQuery.isLoading}
      <p>Loading users...</p>
    {:else if $usersQuery.isError}
      <p class="text-destructive">Error: {$usersQuery.error.message}</p>
    {:else if $usersQuery.data}
      <ul>
        {#each $usersQuery.data as user (user.id)}
          <li>{user.name}</li>
        {/each}
      </ul>
      {#if $usersQuery.data.length === 0}
        <p>No users found.</p>
      {/if}
    {/if}
    ```
    **Key Takeaways:**
    *   Use a Svelte 5 rune (e.g., `let searchTerm = $state('');`) for the input value.
    *   Include this reactive variable in the `queryKey` array: `queryKey: ['uniqueQueryName', searchTerm]`. This is crucial for Svelte Query to automatically refetch when the search term changes.
    *   The `queryFn` should be an arrow function that captures the current value of the reactive search term: `queryFn: () => fetchFunction(searchTerm)`.
    *   The backend API must be prepared to handle the corresponding query parameter (e.g., `/api/endpoint?search=value`).

## 3. UI Components: shadcn-svelte@next

*   **Library Approach:** `shadcn-svelte` provides UI components by copying code into your project via its CLI. It is based on TailwindCSS and Radix Svelte.
*   **Version/Source:** CLI using `@next` tag (e.g., `npx shadcn-svelte@next add <component>`). Schema version in `app/components.json` is `https://next.shadcn-svelte.com/schema.json`.
*   **Initialization:** Configured via `app/components.json`.
    *   Style: "new-york"
    *   Tailwind Config: `tailwind.config.js`
    *   Tailwind CSS Base: `src/app.css`
    *   Base Color: "neutral"
    *   Aliases: `$lib/components`, `$lib/utils`, `$lib/components/ui` (for added components).
*   **Adding Components:**
    *   Use the CLI from the `app/` directory:
    ```bash
    # Example for pnpm users:
    pnpm dlx shadcn-svelte@next add button
    ```
    *   Components are added to the path specified in `components.json` (e.g., `$lib/components/ui/`).
*   **Importing Components:**
    *   Components are typically imported directly from their directories after installation:
    ```svelte
    <script lang="ts">
      import { Button } from "$lib/components/ui/button";
      import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
    </script>
    ```
*   **Event Handling Note (Svelte 5 & `shadcn-svelte@next`):**
    *   For some components like `Button`, there have been community reports and observed behavior where Svelte's standard `on:click={...}` directive might cause linter errors (e.g., `Argument of type '"click"' is not assignable to parameter of type 'never'`) or may not function as expected in a Svelte 5 context with `@next` versions of `shadcn-svelte`.
    *   A successful workaround has been to use the standard HTML `onclick={...}` attribute instead:
    ```svelte
    <script lang="ts">
      import { Button } from "$lib/components/ui/button";
      function handleClick() { console.log("Clicked!"); }
    </script>

    <Button onclick={handleClick}>Click Me</Button> 
    ```
    *   While this resolves linter issues and ensures functionality, it's a point to monitor for future `shadcn-svelte` updates, as `on:click` is the more idiomatic Svelte approach.
*   **Usage:** (Examples to be added after component integration)
*   **Theming & Customization:** (Details on customizing the "new-york" style and using CSS variables/Tailwind utilities to be added).
*   **Official Documentation:** [https://www.shadcn-svelte.com/docs](https://www.shadcn-svelte.com/docs) (Note: ensure to check `@next` specific docs if available, or infer from CLI/schema. The `@next` version is specifically for Svelte 5).

## 4. Styling

*   **Framework:** TailwindCSS
*   **Version:** `^3.4.17` (Confirmed from `app/package.json`)
*   **Configuration:** `tailwind.config.js`, `postcss.config.cjs`.
*   **Global Styles/Tailwind Layers:** Imported/defined in `src/app.css`.
*   **Utility Classes:** Primary method for styling components.
*   **`clsx` / `tailwind-merge`:** Available for conditional and merged class names (see `app/package.json`).

## 5. State Management (UI/Client-Side)

*   **Primary Mechanism:** Svelte 5 Runes (`$state`, `$derived`, `$effect`) for local and cross-component UI state.
*   **Svelte Stores:** Still available if needed for more complex scenarios not covered by runes or Svelte Query.

## 6. Web Research & Findings Log

*(This section will be updated with summaries of web research, links to articles, gists, or discussions that inform design decisions or troubleshooting.)*

---