<svelte:options immutable/>
<script>
  import {config} from '../config';
  import {queryStore} from "./queryStore";

  const facetConfig = config.facets;

  export let facets;

  $: facetNodes = {}
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
  }

  function change(facet, link, event) {
    queryStore.setFacet(facet, link, event.target.checked)
  }

  function toggleOptions(e) {
    e.target.parentNode.querySelector('ul').classList.toggle('d-block');
  }
</script>

<ul class="navbar-nav me-auto mb-2 mb-lg-0">
  {#each facets as facet (facet.field)}
    <li class="nav-item">
      <a class="nav-link" on:click={toggleOptions}>{facet.displayString}</a>
      {#if facet.links.length > 0}
        <ul class="list-group">
          {#each facet.links as link (link.value)}
            <li class="list-group-item">
              <input type="checkbox" class="form-check-input" on:change={e => change(facet, link, e)} name={link.name}
                     checked={link.isSelected}/>
              <label class="form-check-label" for={link.name}>{link.displayString}</label>
            </li>
          {/each}
        </ul>
      {/if}
    </li>
  {/each}
</ul>

<style type="text/scss">
  ul ul {
    position: absolute;
    z-index: 1;
    display: none;
  }
</style>