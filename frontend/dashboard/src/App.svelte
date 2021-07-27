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
  <div class="services">
    <Services></Services>
  </div>
  {#if route.params.service}
    <div class="service">
      <Service serviceName={route.params.service}></Service>
    </div>
  {/if}
  <div id="alert"></div>
</main>

<style>
  main {
    display: flex;
    gap: 4em;
  }

  .services {
    order: 1;
  }

  .service {
    order: 2;
    flex-grow: 1;
  }

  #alert {
    position: absolute;
    box-sizing: border-box;
    width: 100%;
    min-height: 2em;
    padding: 1em;
    bottom: 0;
    color: white;
    font-weight: bold;
    z-index: 1;
  }
</style>