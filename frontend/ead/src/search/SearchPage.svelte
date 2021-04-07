<script>
  import Search from "./Search.svelte";
  import Sort from "./Sort.svelte";
  import {queryStore} from "./queryStore";
  import {noop} from "svelte/internal";
  import Pager from "../Pager.svelte";
  import Grid from "../Grid.svelte";
  import {searchStore} from "../searchStore";
  import EadSearch from "../ead/search/EadSearch.svelte";

  export let events;
  export let route;
  let search;
  let query;

  $: events = {
    facetClicked: noop,
    ...events
  };

  queryStore.subscribe(currValue => {
    query = currValue;
    searchStore.prepare(query)
  })

  searchStore.subscribe(currValue => {
    search = currValue
  })
</script>

{#if search}
  <section data-component-type={route.component}>
    <Search facets={search.facets} {query}/>
    <div class="sort">
      <Sort {query}/>
    </div>
<!--    <Facets {events} facets={search.facets}/>-->
    {#if route.component === 'grid'}
      <div class="grid">
        <Grid {search}/>
      </div>
    {:else if route.component === 'archive'}
      <div class="archive">
        <EadSearch {search}/>
      </div>
    {/if}
    <div class="pager">
      <Pager/>
    </div>
  </section>
{/if}

<style type="text/scss">
  section {
    display: grid;
    grid-template-columns: 3em 1fr 1fr 1fr 1fr 1fr 1fr 3em;
  }

  section[data-component-type="archive"] {
    grid-template-areas:
    "search search  search  search  search  search  search  search"
    ".      facets  facets  facets  facets  facets  facets  ."
    ".      sort    sort    sort    sort    sort    sort    ."
    ".      archive archive archive archive archive archive ."
    ".      pager   pager   pager   pager   pager   pager   ."
  }

  section[data-component-type="grid"] {
    grid-template-areas:
    "search search search search search search search search"
    ".      facets facets facets facets facets facets ."
    ".      sort   sort   sort   sort   sort   sort   ."
    ".      grid   grid   grid   grid   grid   grid   ."
    ".      pager  pager  pager  pager  pager  pager  .";
  }

  table {
    font-weight: bold;
  }

  .stats {
    grid-area: stats;
    display: flex;
  }

  .sort {
    grid-area: sort;
  }

  .grid {
    grid-area: grid;
  }

  .pager {
    grid-area: pager;
  }

  .archive {
    grid-area: archive;
  }
</style>