<script>
  import {onMount} from "svelte";

  let search;
  let searchRequest = {}
  onMount(async () => {
    await update()
  })

  async function update() {
    const queryBuilder = []
    for (const [key, value] of Object.entries(searchRequest)) {
      if(value) {
        queryBuilder.push(`${key}=${value}`);
      }
    }
    const query = queryBuilder.join('&');

    const response = await fetch('https://hub3.nl-hana.delving.io/api/ead/search?' + query)
    search = await response.json()
  }
</script>

{#if search}
  <div>
    <input bind:value={searchRequest.q} />
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
          {facet.name} {facet.total}
          {#each facet.links as link}
            <div>
              <input type="checkbox" name={link.name} checked={link.isSelected}/>
              <label for={link.name}>{link.displayString}</label>
            </div>
          {/each}
        </div>
      {/if}
    {/each}
  </div>
  <div>
    <ul>
      <li>
        <span>Nr. archiefinventaris</span>
        <span>Archiefnaam</span>
        <span>Periode</span>
        <span></span>
      </li>
      {#each search.archives as archive}
        <li>
          <span>{archive.inventoryID}</span>
          <span>{archive.title}</span>
          <span>{archive.period.join(' ')}</span>
          <span>
            <button>Detailresultaten</button>
            <button>Beschrijving</button>
          </span>
        </li>
      {/each}
    </ul>
  </div>
{/if}

<style>
  ul {
    display: table;
  }

  li {
    display: table-row;
  }

  li span {
    display: table-cell;
  }

  .facet {
    border: 2px solid black;
  }

  input, label {
    display: inline;
  }
</style>