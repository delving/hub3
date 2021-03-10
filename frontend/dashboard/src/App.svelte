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

<style>
	main {
		text-align: center;
		padding: 1em;
		max-width: 240px;
		margin: 0 auto;
	}

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 4em;
		font-weight: 100;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
</style>