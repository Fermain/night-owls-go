# Svelte 5: A Guide to New Features (Especially Runes)

Svelte 5 introduces a significant evolution in how reactivity is handled, primarily through a new system called **Runes**. This guide summarizes the key aspects based on official Svelte announcements and documentation.

## What are Runes?

Runes are special symbols (functions with a `$` prefix, e.g., `$state()`) that provide more explicit and fine-grained control over reactivity within Svelte applications. They are designed to:

*   Make reactivity more transparent and less "magical."
*   Allow Svelte's reactivity to be used outside of `.svelte` component files (in `.svelte.js` or `.svelte.ts` files).
*   Simplify Svelte's core concepts by making many older patterns and syntaxes obsolete.
*   Improve performance through fine-grained updates powered by an underlying signal system.

Runes are keywords in the Svelte language. They don't need to be imported and are not regular JavaScript functions (they can't be assigned to variables or passed as arguments).

**Important Note:** Svelte 5 aims to be a drop-in replacement for most existing Svelte 4 applications. Runes are an **opt-in** feature; your existing components will continue to work.

## Key Runes

Here are some of the primary runes introduced in Svelte 5:

### 1. `$state` - Declaring Reactive State

Replaces `let` for top-level reactive declarations in components. Makes it explicit what is reactive state.

*   **Svelte 4:**
    ```html
    <script>
      let count = 0;
    </script>
    ```
*   **Svelte 5 with Runes:**
    ```html
    <script>
      let count = $state(0);

      function increment() {
        count += 1;
      }
    </script>
    <button on:click={increment}>
      clicks: {count}
    </button>
    ```
    Objects and arrays wrapped in `$state` become deeply reactive.

    *   **Equality with `$state`**: When dealing with objects or arrays managed by `$state`, direct equality checks (`===`) might not behave as expected if you're comparing a raw object to the reactive proxy. It's generally better to work with the reactive proxies directly or use specific properties for comparison. Svelte 5 also introduces `$state.raw()` to get the underlying raw object if needed for interop with non-Svelte code or for specific equality checks, though direct mutation of the raw object won't trigger reactivity.

### 2. `$derived` - Declaring Derived State

Replaces `$: ` (reactive statements) for creating values that depend on other reactive state. Dependencies are tracked at runtime.

*   **Svelte 4:**
    ```html
    <script>
      export let width;
      export let height;
      $: area = width * height;
    </script>
    ```
*   **Svelte 5 with Runes:**
    ```html
    <script>
      let { width, height } = $props(); // See $props below
      const area = $derived(width * height);
    </script>
    ```
    `$derived` expressions re-evaluate only when their direct dependencies change.

### 3. `$effect` - Running Side Effects

Replaces `$: ` for running code in response to state changes (e.g., logging, data fetching, manual DOM manipulation) and also some lifecycle functions like `onMount`, `afterUpdate`. Effects track their own dependencies and re-run when those change.

*   **Svelte 4 (logging example):**
    ```html
    <script>
      // ...
      $: console.log(area);
    </script>
    ```
*   **Svelte 5 with Runes:**
    ```html
    <script>
      // ...
      const area = $derived(width * height);
      $effect(() => {
        console.log(area);
      });
    </script>
    ```
    `$effect` runs after the DOM has been updated. For effects that need to run *before* DOM updates (rare), or need cleanup, `$effect.pre` and returning a cleanup function from `$effect` or `$effect.pre` are available.

### 4. `$props` - Declaring Component Props

Replaces `export let` for declaring component properties. It offers a clearer way to define what a component accepts.

*   **Svelte 4:**
    ```html
    <script>
      export let name;
      export let age = 0; // Default value
    </script>
    ```
*   **Svelte 5 with Runes:**
    ```html
    <script>
      let { name, age = 0 } = $props();
    </script>
    ```
    This rune makes prop handling more explicit and consistent.

### 5. `$bindable` (formerly `$prop`)

Used in conjunction with `$props` to make a prop two-way bindable from the parent component using the `bind:` directive.

```html
<script lang="ts">
  // Child.svelte
  let { value = $bindable() } = $props<{ value?: string }>();
</script>

<input bind:value />
```

### 6. Other Runes
*   `$inspect()`: A debugging tool to inspect reactive dependencies and values.
*   `$host()`: Used for interacting with the host element of a custom element.

## Benefits of Runes

*   **Fine-grained Reactivity:** Svelte 5's reactivity, powered by signals (an internal implementation detail), allows for more precise updates. Changes to one piece of state only update the specific parts of the DOM that depend on it, leading to potentially significant performance improvements.
*   **Universal Reactivity:** Runes like `$state`, `$derived`, and `$effect` can be used not only in `.svelte` files but also in `.svelte.js` and `.svelte.ts` files. This allows you to create reactive logic (like custom stores in Svelte 4) more naturally and consistently.
*   **Simpler Concepts:** Runes aim to simplify Svelte by making many existing concepts and syntaxes obsolete or less central:
    *   The distinction between `let` at the component's top level vs. elsewhere.
    *   `export let` for props.
    *   The `$: ` reactive statements and their quirks.
    *   The need for `<script context="module">` in many cases.
    *   `$$props` and `$$restProps`.
    *   Many lifecycle functions (e.g., `onMount`, `afterUpdate` can often be replaced with `$effect`).
    *   The traditional Svelte store API and the `$` store prefix (while stores are not deprecated, runes offer a more integrated way to manage shared reactive state).
*   **Improved Developer Experience:** By making reactivity explicit, it becomes easier to understand how data flows and when updates occur, especially in complex components or when refactoring.

## Signals: The Power Under the Hood

While developers interact with Runes, Svelte 5's reactivity is internally powered by a system of signals. This is an implementation detail, meaning you don't directly manipulate signals. Svelte's compiler uses runes to create and manage these signals optimally. This approach allows Svelte to maximize both efficiency and ergonomics.

## Migrating from Svelte 4

Svelte 5 is designed to be largely backward-compatible. You can adopt runes incrementally. Existing Svelte 4 code should continue to work. The official documentation will provide a migration guide.

## Conclusion

Runes represent a major step forward for Svelte, aiming to make the framework more powerful, consistent, and easier to use, especially for complex applications. They provide a more explicit and fine-grained approach to reactivity that extends beyond component boundaries, all while building on Svelte's core strengths of performance and a great developer experience.

---
*This guide is based on information available from the Svelte team (blog posts, Svelte 5 preview documentation) as of early 2024/late 2023. Specific APIs and features might evolve before the final Svelte 5 release.* 