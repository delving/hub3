<script>
  import {queryStore} from "./queryStore";
  import Facets from "./Facets.svelte";
  import HitPager from "../HitPager.svelte";

  export let query;
  export let routeConfig;
  export let search;

  let q = query.q

  function update() {
    queryStore.updateQuery({
      q,
      search: !!q
    })
    q = ''
  }
</script>

<nav class="search navbar navbar-expand-lg navbar-light bg-light border-bottom">
  <div class="container-fluid">
    <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent"
            aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation">
      <span class="navbar-toggler-icon"></span>
    </button>
    <div class="collapse navbar-collapse" id="navbarSupportedContent">
      <form class="d-flex">
        <input class="form-control" bind:value={q}/>
        <button type="button" class="btn btn-dark" on:click={update}>Zoeken</button>
      </form>
      {#if search}
        {#if search.facets && routeConfig.facets}
          <Facets facetConfig={routeConfig.facets} facets={search.facets}/>
        {/if}
        {#if search.hitPager}
          <HitPager pager={search.hitPager}/>
        {/if}
      {/if}

    </div>
  </div>
</nav>

<style type="text/scss">
  .search {
    display: grid;
    grid-area: search;
  }
</style>
