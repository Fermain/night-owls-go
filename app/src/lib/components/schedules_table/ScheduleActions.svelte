<script lang="ts">
    import { buttonVariants } from "$lib/components/ui/button"; // Using buttonVariants for styling
    import type { HTMLAnchorAttributes } from "svelte/elements";

    export let scheduleId: number;

    let href: HTMLAnchorAttributes['href'] = `/admin/schedules/${scheduleId}/edit`;
</script>

<a {href} class={buttonVariants({ variant: "default", size: "sm" })} role="button">
    Edit
</a>

<!-- Placeholder for Delete button -->
<!-- <Button variant="destructive" size="sm" on:click={() => console.log('Delete', scheduleId)} class="ml-2">Delete</Button> --> 