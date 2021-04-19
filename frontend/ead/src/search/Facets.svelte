<svelte:options immutable/>
<script>
  import {queryStore} from "./queryStore";

  export let facets;
  export let facetConfig;

  console.log(facets, facetConfig)

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

<section>
  <div>

    <ul>
      <span>Filteren op:</span>
      {#each facets as facet (facet.field)}
        <li>
          <a on:click={toggleOptions}>{facet.displayString}</a>
          {#if facet.links.length > 0}
            <ul class="list-group">
              {#each facet.links as link (link.value)}
                <li class="list-group-item">
                  <input type="checkbox" class="form-check-input" on:change={e => change(facet, link, e)}
                         name={link.name}
                         checked={link.isSelected}/>
                  <label class="form-check-label" for={link.name}>{link.displayString}</label>
                </li>
              {/each}
            </ul>
          {/if}
        </li>
      {/each}
    </ul>
  </div>
</section>

<style type="text/scss">
  section {
    background: white;
    padding: 1rem 0;
  }

  section > div {
    width: 75%;
    margin: 0 auto;
  }

  a {
    color: black;
    font-style: italic;
  }

  div > ul {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: 0.6rem;
    margin: 0;
    padding: 0;
    list-style-type: none;
  }

  ul {
    ul {
      position: absolute;
      z-index: 1;
      color: yellow;
      display: none;
    }
  }
</style>