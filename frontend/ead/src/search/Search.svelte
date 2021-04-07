<svelte:options immutable/>
<script>
  import {queryStore} from "./queryStore";
  import Facets from "./Facets.svelte";

  export let query;
  export let facets;
  let q = query.q

  function update() {
    queryStore.updateQuery({q})
    q = ''
  }
</script>

<nav class="search navbar navbar-expand-lg navbar-light bg-light border-bottom">
  <div class="container-fluid">
    <a class="navbar-brand" href="#">Filters</a>
    <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent"
            aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
      <span class="navbar-toggler-icon"></span>
    </button>
    <div class="collapse navbar-collapse" id="navbarSupportedContent">
      <Facets {facets} />
      <form class="d-flex">
        <input class="form-control" bind:value={q}/>
        <button class="btn btn-dark" on:click={update}>Zoeken</button>
      </form>
    </div>
  </div>
</nav>

<style type="text/scss">
  .search {
    display: grid;
    grid-area: search;
  }
</style>
