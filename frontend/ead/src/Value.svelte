<script>
  export let context;
  export let path;
  export let value;
  export let style = {}

  function slicePath() {
    if (path) return path.slice(1)
    return []
  }

  $: if(path) {
    value = path.length === 0 ? context : context[path[0]]
  }
</script>
{#if value}
  {#if Array.isArray(value)}
    {#if value.length > 1}
      <ul class:hor={style.horizontal}>
        {#each value as item}
          <li>
            <svelte:self context={item} path={slicePath()}/>
          </li>
        {/each}
      </ul>
    {:else if value.length === 1}
      <svelte:self context={value[0]} path={slicePath()}/>
    {/if}
  {:else if typeof value === 'object'}
    <svelte:self context={value} path={slicePath()}/>
  {:else if typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean'}
    <span>{value}</span>
  {/if}
{/if}

<style>
  ul {
    display: flex;
    flex-direction: column;
    padding: 0;
    list-style-type: none;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .hor {
    flex-direction: row;
  }
</style>