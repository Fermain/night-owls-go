<script lang="ts">
  import cronstrue from 'cronstrue';

  let { cronExpression }: { cronExpression: string } = $props();

  let humanReadable = $state('');
  let error = $state('');

  $effect(() => {
    try {
      humanReadable = cronstrue.toString(cronExpression);
      error = '';
    } catch (e: any) {
      humanReadable = ''; // Clear previous valid string
      error = e.message || 'Invalid CRON expression';
      // Optionally, log the full error for debugging: console.error("Failed to parse CRON:", cronExpression, e);
    }
  });
</script>

{#if error}
  <span title={error} class="text-red-500 cursor-help">
    <code>{cronExpression}</code> (Error: {error})
  </span>
{:else if humanReadable}
  <span title={cronExpression}>{humanReadable}</span>
{:else}
  <!-- Fallback for initial state or unexpected issues -->
  <code>{cronExpression}</code>
{/if} 