<script>
  import Services from "./Services.svelte";
  import Service from "./Service.svelte";
  import {routeChanged} from "../../src/router";
  import {onDestroy} from "svelte";

  window.baseUrl = "http://localhost:3001"
  let route = routeChanged(newRoute => route = newRoute)
  onDestroy(route.unsubscribe)
</script>

<main>
  <Services></Services>
  <Service serviceName={route.params.service}></Service>
</main>