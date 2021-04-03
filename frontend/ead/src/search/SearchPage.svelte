<script>
  import Facets from "./Facets.svelte";
  import Search from "./Search.svelte";
  import Sort from "./Sort.svelte";
  import {queryStore} from "./queryStore";
  import {noop} from "svelte/internal";
  import Pager from "../Pager.svelte";

  export let events;
  export let search;
  let query;

  $: events = {
    facetClicked: noop,
    ...events
  };

  queryStore.subscribe(async currValue => {
    query = currValue;
  })
</script>

{#if search}
  <Search {query}/>
  <div class="sort">
    <Sort {query}/>
  </div>

  <div class="facets">
    <Facets {events} facets={search.facets}/>
  </div>

  <div class="content">
    <slot></slot>
  </div>

  <div class="pager">
    <Pager/>
  </div>
{/if}

<style type="text/scss">
  @import "src/variables";

  button, select, input {
    background-color: $DEFAULT_COMPONENT_BG_COLOR;
    padding: 10px;
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

  .facets {
    grid-area: facets;
  }

  .content {
    grid-area: content;
  }

  .pager {
    grid-area: pager;
  }
</style>