<svelte:options immutable/>
<script>
  import {queryStore} from "./queryStore";

  export let facets;

  function change(link, event) {
    queryStore.setFacet(link.url, event.target.checked)
  }
</script>

{#each facets as facet}
  {#if facet.links.length > 0}
    <div class="facet">
      <p class="title">{facet.name} {facet.total}</p>
      <div class="options">
        {#each facet.links as link}
          <p>
            <input type="checkbox" on:change={e => change(link, e)} name={link.name} checked={link.isSelected}/>
            <label for={link.name}>{link.displayString}</label>
          </p>
        {/each}
      </div>
    </div>
  {/if}
{/each}

<style type="text/scss">
  @import "variables";

  .title {
    font-weight: bold;
  }

  .facet {
    background-color: $DEFAULT_COMPONENT_BG_COLOR;
    padding: 10px;
    margin-bottom: 10px;
  }

  .options {
    border-top: 1px solid $DEFAULT_TEXT_COLOR;
  }

  input, label {
    display: inline;
  }
</style>