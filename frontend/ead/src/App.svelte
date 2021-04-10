<script>
  import {updateRoute} from './router.js'
  import SearchPage from "./search/SearchPage.svelte";
  import {configReady} from "./config"
  import "bootstrap/dist/css/bootstrap-reboot.min.css"
  import "bootstrap/dist/css/bootstrap.min.css"

  let route;
  let component;

  configReady(() => {
    route = updateRoute()
  })
  console.log('App here!')
</script>

{#if route}
  <main id="delving">
    <SearchPage/>
  </main>
{/if}