<script>
  export let config
  import {configure} from './nav.js'

  configure(config)

  import Ead from "./Ead.svelte";
  import EadSearch from "./EadSearch.svelte";
</script>

<main>
  <EadSearch/>
  <!--  <Ead/>-->
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>