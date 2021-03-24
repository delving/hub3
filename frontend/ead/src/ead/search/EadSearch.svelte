<script>
  import {onMount} from "svelte";
  import {linkEad, linkCLevel, linkEadDescription} from '../../nav'
  import Facets from "../../Facets.svelte";
  import Search from "../../Search.svelte";
  import Sort from "../../Sort.svelte";
  import {queryStore} from "../../queryStore";

  let search;
  let searchRequest = {}
  let query;

  queryStore.subscribe(async currValue => {
    query = currValue;
    await update()
  })

  async function loadDetails(archive) {
    const response = await fetch(`https://hub3.nl-hana.delving.io/api/ead/search/${archive.inventoryID}${query.searchString}`)
    archive.details = await response.json()
    search = search;
  }

  async function update() {
    const response = await fetch(`https://hub3.nl-hana.delving.io/api/ead/search${query.searchString}`)
    search = await response.json()
  }
</script>

{#if search}
  <div class="archive-search">
    <div class="search">
      <Search {query}/>
    </div>
    <div class="sort">
      <Sort {query} />
    </div>

    <div class="facets">
      <Facets facets={search.facets}/>
    </div>

    <div class="content">
      <table>
        <tr>
          <td>Nr. archiefinventaris</td>
          <td>Archiefnaam</td>
          <td>Periode</td>
          <td></td>
        </tr>
        {#each search.archives as archive}
          <tr>
            <td><a href={linkEad(archive)}>{archive.inventoryID}</a></td>
            <td>{archive.title}</td>
            <td>{archive.period.join(' ')}</td>
          </tr>
          <tr>
            <td>
              <button on:click={async() => await loadDetails(archive)}>Detailresultaten</button>
              {#if archive.details}
                <ul>
                  {#each archive.details.cLevels as cLevel}
                    <li><a href={linkCLevel(archive, cLevel)}>{cLevel.label}</a></li>
                  {/each}
                </ul>
              {/if}

            </td>
            <td>
              <button><a href={linkEadDescription(archive, searchRequest)}>Beschrijving</a></button>
            </td>
            <td></td>
          </tr>
        {/each}
      </table>
    </div>
  </div>
{/if}

<style type="text/scss">
  @import "../../variables";

  button, select, input {
    background-color: $DEFAULT_COMPONENT_BG_COLOR;
    padding: 10px;
  }

  table {
    font-weight: bold;
  }

  .archive-search {
    display: grid;
    grid-template-columns: repeat(6, 1fr);
    grid-auto-rows: auto;
    grid-template-areas:
      "facets search search search search search"
      "facets sort sort sort sort sort"
      "facets content content content content content";
  }

  .stats {
    grid-area: stats;
  }

  .search {
    grid-area: search;
  }

  .sort {
    grid-area: sort;
  }

  .facets {
    grid-area: facets;
  }

  .content {
    grid-area: content;
  }
</style>