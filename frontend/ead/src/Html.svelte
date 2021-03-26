<svelte:options immutable/>
<script>
  export let node;
</script>

<span class:text={node.text} class="{node.tagName || ''}">
{#if node.inline}
  {#each node.inline as child}
    <svelte:self node={child}/>
  {/each}
  {:else}
  {node.text}
{/if}
</span>
