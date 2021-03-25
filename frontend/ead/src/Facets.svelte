<svelte:options immutable/>
<script>
  import {config} from './config';
  import {queryStore} from "./queryStore";

  const facetConfig = config.facets;

  export let facets;

  $: {
    facets = facets
      .filter(f => f.field in facetConfig)
      .map(f => {
        const config = facetConfig[f.field];
        return {
        ...f,
          displayString: config.label,
          order: config.order
        }
      })
      .sort((a, b) => a.order - b.order);
    console.log(facets)
  }

  function change(facet, link, event) {
    queryStore.setFacet(facet, link, event.target.checked)
  }
</script>

{#each facets as facet (facet.field)}
  {#if facet.links.length > 0}
    <div class="facet">
      <p class="title">{facet.displayString} ({facet.total})</p>
      <div class="options">
        {#each facet.links as link (link.value)}
          <p>
            <input type="checkbox" on:change={e => change(facet, link, e)} name={link.name} checked={link.isSelected}/>
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