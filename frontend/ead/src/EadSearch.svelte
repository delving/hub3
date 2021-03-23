<script>
  import {onMount} from "svelte";
  import {linkEad, linkCLevel, linkEadDescription} from './nav'

  let search;
  let searchRequest = {}
  onMount(async () => {
    await update()
  })

  async function loadDetails(archive) {
    const response = await fetch(`https://hub3.nl-hana.delving.io/api/ead/search/${archive.inventoryID}`)
    archive.details = await response.json()
    search = search;
  }

  async function update() {
    const queryBuilder = []
    for (const [key, value] of Object.entries(searchRequest)) {
      if (value) {
        queryBuilder.push(`${key}=${value}`);
      }
    }
    const query = queryBuilder.join('&');

    const response = await fetch('https://hub3.nl-hana.delving.io/api/ead/search?' + query)
    let body = await response.json()
    search = body;
  }
</script>

{#if search}
  <div class="archive-search">
    <div class="stats">Some stats</div>
    <div class="search">
      <input bind:value={searchRequest.q}/>
      <button on:click={update}>Zoeken</button>
      Sorteren op
      <select>
        <option>Relevantie</option>
        <option>Nummer Toegang</option>
        <option>Periode</option>
      </select>
      Volgorde
      <select>
        <option>Oplopend</option>
        <option>Aflopend</option>
      </select>
      Resultaten
      <select>
        <option>10</option>
        <option>20</option>
        <option>50</option>
      </select>
    </div>
    <div class="facets">
      {#each search.facets as facet}
        {#if facet.links.length > 0}
          <div class="facet">
            <p class="title">{facet.name} {facet.total}</p>
            <div class="options">
              {#each facet.links as link}
                <p>
                  <input type="checkbox" name={link.name} checked={link.isSelected}/>
                  <label for={link.name}>{link.displayString}</label>
                </p>
              {/each}
            </div>
          </div>
        {/if}
      {/each}
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
  @import "variables";

  button, select, input, .facet {
    background-color: $DEFAULT_COMPONENT_BG_COLOR;
    padding: 10px;
  }

  .options {
    border-top: 1px solid $DEFAULT_TEXT_COLOR;
  }

  .facet {
    margin-bottom: 10px;
  }

  table, .title {
    font-weight: bold;
  }

  .archive-search {
    display: grid;
    grid-template-columns: repeat(6, 1fr);
    grid-auto-rows: auto;
    grid-template-areas:
      "facets stats stats stats stats stats"
      "facets search search search search search"
      "facets content content content content content";
  }

  .stats {
    grid-area: stats;
  }

  .search {
    grid-area: search;
  }

  .facets {
    grid-area: facets;
  }

  .content {
    grid-area: content;
  }

  input, label {
    display: inline;
  }
</style>