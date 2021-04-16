<script>
  export let context;
  export let path;

  let value = path.length === 0 ? context : context[path[0]]
</script>
{#if value}
  {#if Array.isArray(value)}
    {#if value.length > 1}
      <ul>
        {#each value as item}
          <li>
            <svelte:self context={item} path={path.slice(1)}/>
          </li>
        {/each}
      </ul>
    {:else if value.length === 1}
      <svelte:self context={value[0]} path={path.slice(1)}/>
    {/if}
  {:else if typeof value === 'object'}
    <svelte:self context={value} path={path.slice(1)}/>
  {:else if typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean'}
    <span>{value}</span>
  {/if}
{/if}

<style>
  ul {
    padding: 0;
    list-style-type: none;
  }
</style>