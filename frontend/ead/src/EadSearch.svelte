<script>
  import {onMount} from "svelte";
  import {linkEad,linkCLevel,linkEadDescription} from './nav'

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
          <span><a href={linkEad(archive)}>{archive.inventoryID}</a></span>
          <span>{archive.title}</span>
          <span>{archive.period.join(' ')}</span>
          <span>
            <button on:click={async() => await loadDetails(archive)}>Detailresultaten</button>
            {#if archive.details}
              <ul>
                {#each archive.details.cLevels as cLevel}
                  <li><a href={linkCLevel(archive, cLevel)}>{cLevel.label}</a></li>
                {/each}
              </ul>
            {/if}
            <button><a href={linkEadDescription(archive, searchRequest)}>Beschrijving</a></button>
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