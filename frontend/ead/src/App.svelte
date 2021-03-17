<script>
  import {configure, parseUrl} from './nav.js'
  import Ead from "./Ead.svelte";
  import EadSearch from "./EadSearch.svelte";

  export let config;

  configure(config);
  const route = parseUrl();
</script>

<main>
  {#if route.routeId}
    <Ead/>
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