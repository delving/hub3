<script>
  import {getRoute} from './nav.js'
  import Ead from "./ead/detail/EadDetail.svelte";
  import DetailPage from "./RDFDetail.svelte";
  import SearchPage from "./search/SearchPage.svelte";
  import {configReady} from "./config"
  import {queryStore} from "./search/queryStore"
  import "bootstrap/dist/css/bootstrap-reboot.min.css"
  import "bootstrap/dist/css/bootstrap.min.css"

  let route;
  let component;

  const components = {
    grid: SearchPage,
    archive: SearchPage,
    detail: DetailPage,
    archiveDetail: Ead
  }

  configReady(async () => {
    queryStore.parseUrl()
    route = getRoute()
    component = components[route.component]
  })
</script>

{#if component}
  <main id="delving">
    <svelte:component {route} this={component}/>
  </main>
{/if}