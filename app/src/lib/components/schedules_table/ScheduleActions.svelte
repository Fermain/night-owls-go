<script lang="ts">
    import { buttonVariants, Button } from "$lib/components/ui/button"; // Using buttonVariants for styling, added Button for Delete
    import type { HTMLAnchorAttributes } from "svelte/elements";
    import { createMutation, useQueryClient } from "@tanstack/svelte-query";
    import { toast } from "svelte-sonner";

    export let scheduleId: number;

    let href: HTMLAnchorAttributes['href'] = `/admin/schedules/${scheduleId}/edit`;

    const queryClient = useQueryClient();

    const deleteMutation = createMutation<
        Response, // Response type from fetch
        Error,    // Error type
        number    // Variables type (scheduleId)
    >({
        mutationFn: async (idToDelete) => {
            const response = await fetch(`/api/admin/schedules/${idToDelete}`, {
                method: 'DELETE'
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ message: 'Failed to delete schedule' }));
                throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
            }
            return response;
        },
        onSuccess: () => {
            toast.success('Schedule deleted successfully!');
            queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
        },
        onError: (error) => {
            toast.error(`Error deleting schedule: ${error.message}`);
        }
    });

    function handleDelete() {
        if (window.confirm(`Are you sure you want to delete schedule ID ${scheduleId}? This action cannot be undone.`)) {
            $deleteMutation.mutate(scheduleId);
        }
    }
</script>

<div class="flex space-x-2">
    <a {href} class={buttonVariants({ variant: "outline", size: "sm" })} role="button">
        Edit
    </a>
    <Button 
        variant="destructive" 
        size="sm" 
        onclick={handleDelete} 
        disabled={$deleteMutation.isPending}
    >
        {#if $deleteMutation.isPending}
            Deleting...
        {:else}
            Delete
        {/if}
    </Button>
</div>
