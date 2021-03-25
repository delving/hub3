<script>
  import {getRoute} from './nav.js'
  import Ead from "./ead/detail/EadDetail.svelte";
  import EadSearch from "./ead/search/EadSearch.svelte"

  const route = getRoute();
</script>

<main>
  {#if route.routeId}
    <Ead {route} />
  {:else}
    <EadSearch/>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>