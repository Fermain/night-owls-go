<script lang="ts">
    import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
    import type { CreateMutationResult } from '@tanstack/svelte-query';
    import type { UserData } from '$lib/schemas/user'; // Updated import path

    let {
        open = $bindable(false),
        user,
        mutation
    } = $props<{
        open?: boolean;
        user: UserData | undefined | null; // User object to display name/phone
        mutation: CreateMutationResult<any, Error, number, unknown>; // The delete mutation store
    }>();

    function confirmDelete() {
        if (user?.id) {
            $mutation.mutate(user.id);
        }
    }
</script>

{#if user} <!-- Only render if user is provided -->
    <AlertDialog.Root bind:open>
        <AlertDialog.Content>
            <AlertDialog.Header>
                <AlertDialog.Title>Are you sure you want to delete this user?</AlertDialog.Title>
                <AlertDialog.Description>
                    This action cannot be undone. This will permanently delete the user
                    {user.name ? ` "${user.name}"` : ''}
                    {user.phone ? ` (${user.phone})` : ''}.
                </AlertDialog.Description>
            </AlertDialog.Header>
            <AlertDialog.Footer>
                <AlertDialog.Cancel disabled={$mutation.isPending}>Cancel</AlertDialog.Cancel>
                <AlertDialog.Action
                    onclick={confirmDelete}
                    disabled={$mutation.isPending}
                    class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                    {#if $mutation.isPending}Deleting...{:else}Yes, delete user{/if}
                </AlertDialog.Action>
            </AlertDialog.Footer>
        </AlertDialog.Content>
    </AlertDialog.Root>
{/if} 