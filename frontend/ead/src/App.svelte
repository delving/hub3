<script>
  import {configure, getRouteId} from './nav.js'
  import Ead from "./ead/detail/EadDetail.svelte";
  import EadSearch from "./ead/search/EadSearch.svelte"

  export let config;

  configure(config);
  const routeId = getRouteId();
</script>

<main>
  {#if routeId}
    <Ead routeId={routeId} />
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