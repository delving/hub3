<script>
  import {linkCLevel, linkEad, linkEadDescription} from '../../nav'

  export let search;
  let searchRequest = {}

  async function loadDetails(archive) {
    const response = await fetch(`https://hub3.nl-hana.delving.io/api/ead/search/${archive.inventoryID}`)
    archive.details = await response.json()
    search = search;
  }
</script>

{#if search}
  <table class="table">
    <thead>
    <tr>
      <th scope="col">Nr. archiefinventaris</th>
      <th scope="col">Archiefnaam</th>
      <th scope="col">Periode</th>
      <th scope="col"></th>
    </tr>
    </thead>
    <tbody>
    {#each search.archives as archive}
      <tr>
        <td>
          <a href={linkEad(archive)}>{archive.inventoryID}</a>
          <p>
            <button class="btn btn-light" on:click={async() => await loadDetails(archive)}>Detailresultaten</button>
            {#if archive.details}
              <ul class="list-group">
                {#each archive.details.cLevels as cLevel}
                  <li class="list-group-item">
                    <a href={linkCLevel(archive, cLevel)}>{cLevel.label}</a>
                  </li>
                {/each}
              </ul>
            {/if}
            <a class="btn btn-light" href={linkEadDescription(archive, searchRequest)}>Beschrijving</a>
          </p>
        </td>
        <td>{archive.title}</td>
        <td>{archive.period.join(' ')}</td>
      </tr>
    {/each}
    </tbody>
  </table>
{/if}

<style type="text/scss">
p {
  position: relative;
}

ul {
  position: absolute;
  z-index: 1;
}
</style>