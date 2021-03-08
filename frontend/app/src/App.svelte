<script>
  import "highlight.js/styles/darcula.css"
  import {def} from '../../example/client.gen'
  import ServiceMethod from './ServiceMethod.svelte'
  import NavigationTree from "./NavigationTree.svelte";
  import Service from "./Service.svelte";
  import Services from "./Services.svelte";

  console.log(def);

  let service;
  let method;

  const objectsById = {}
  for (const object of def.objects) {
    objectsById[object.typeID] = object
  }

  function route(hash) {
    const parts = hash.substring(1).split(':')
    if (parts <= 1) return
    const type = parts[0]
    if (type === "service") {
      service = def.services.find(service => service.name === parts[1]) || service
      method = null
    } else if (type === "serviceMethod") {
      const serviceAndMethodName = parts[1].split(".")
      if (serviceAndMethodName.length !== 2) return

      service = def.services.find(service => service.name === serviceAndMethodName[0]) || service
      method = service.methods.find(method => method.name === serviceAndMethodName[1]) || method
    } else if (type === "object") {
      service = null
      method = null
    }
  }

  route(window.location.hash)

  window.onhashchange = () => {
    route(window.location.hash)
    console.log(window.location.hash);
  }
</script>

<main>
  <div class="left">
    <NavigationTree services={def.services}></NavigationTree>
  </div>
  <div class="center">
    {#if !service && !method}
      <Services services={def.services}></Services>
    {:else if method}
      <ServiceMethod service={service} method={method} objects={objectsById}></ServiceMethod>
    {:else}
      <Service service="{service}"></Service>
    {/if}
  </div>
</main>

<style>
  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    -webkit-text-size-adjust: 100%;
    font-family: SourceSansPro, sans-serif;
    line-height: 1.5;
    background: black linear-gradient(180deg, #1F2543 0%, #161824 100%) fixed;
    color: #99a;
    tab-size: 3;
    height: 100%;
  }

  main {
    display: grid;
  }

  .left {
    grid-column: 1 / span 1;
    margin-right: 10px;
  }

  .center {
    grid-column: 2 / 6;
  }
</style>