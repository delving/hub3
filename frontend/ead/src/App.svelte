<script>
  import {getRoute} from './nav.js'
  import Ead from "./ead/detail/EadDetail.svelte";
  import EadSearch from "./ead/search/EadSearch.svelte"
  import RDF from "./RDF.svelte";
  import * as response from './v2.json'
  import RDFDetail from "./RDFDetail.svelte";

  response.items.forEach((i, index) => i.id = index);

  const route = getRoute();
</script>

<main>
  {#if route.routeId}
<!--    <Ead {route} />-->
    <RDFDetail {route}/>
  {:else}
<!--    <EadSearch/>-->
    <RDF/>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>