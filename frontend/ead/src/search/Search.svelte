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

<nav class="search navbar navbar-expand-lg navbar-light border-bottom">
  <form>
    <input class="form-control" bind:value={q}/>
    <button type="button" class="btn btn-dark" on:click={update}>Zoeken</button>
  </form>
  {#if search}
    {#if search.hitPager}
      <HitPager pager={search.hitPager}/>
    {/if}
  {/if}
</nav>

<style type="text/scss">
  .search {
    background-color: #ffb612;
    grid-area: search;
  }

  form {
    display: flex;
    width: 75%;
    padding: 0;
    margin: 0 auto;
  }
</style>
