<script>
  import "highlight.js/styles/darcula.css"
  import {def} from '../gen/docs'
  import ServiceMethod from './ServiceMethod.svelte'
  import NavigationTree from "./NavigationTree.svelte";
  import Service from "./Service.svelte";
  import Services from "./Services.svelte";
  import APITree from "./APITree.svelte";
  import Markdown from "./Markdown.svelte";

  console.log(def);
  export let topics;

  let service;
  let method;
  let markdown;
  let routeId;

  const objectsById = {}
  for (const object of def.objects) {
    objectsById[object.typeID] = object
  }

  function route(hash) {
    const parts = hash.substring(1).split(':')
    if (parts <= 1) return null
    const type = parts[0]
    if (type === "service") {
      service = def.services.find(service => service.name === parts[1]) || service
      method = null
    } else if (type === "serviceMethod") {
      const serviceAndMethodName = parts[1].split(".")
      if (serviceAndMethodName.length !== 2) return

      service = def.services.find(service => service.name === serviceAndMethodName[0]) || service
      method = service.methods.find(method => method.name === serviceAndMethodName[1]) || method
    } else if (type === "markdown") {
      markdown = null
      for (const topic of topics) {
        for(const link of topic.links) {
          if(link.id === parts[1]) {
            markdown = link.markdown
            break;
          }
        }
      }
    }
    return parts[0]
  }

  routeId = route(window.location.hash)

  window.onhashchange = () => {
    routeId = route(window.location.hash)
  }
</script>

<main>
  <div class="docs">
    <div class="left">
      <NavigationTree topics={topics}></NavigationTree>
      <APITree services={def.services}></APITree>
    </div>
    <div class="center">
      {#if routeId === "service"}
        <Service service="{service}"></Service>
      {:else if routeId === "serviceMethod"}
        <ServiceMethod service={service} method={method} objects={objectsById}></ServiceMethod>
      {:else if routeId === "markdown"}
        <Markdown markdown={markdown}></Markdown>
      {:else}
        <Services services={def.services}></Services>
      {/if}
    </div>
  </div>
</main>

<style>
  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    -webkit-text-size-adjust: 100%;
    font-family: SourceSansPro, sans-serif;
    font-size: 16px;
    line-height: 1.5;
    background: black linear-gradient(180deg, #1F2543 0%, #161824 100%) fixed;
    color: #99a;
    tab-size: 3;
    height: 100%;
  }

  :global(.bright-color) {
    color: white;
  }

  :global(a, a:visited, a:hover, a:active) {
    color: #99a;
    text-underline: white;
  }

  .docs {
    border-radius: .5em;
    background-color: black;
    width: 80%;
    padding: 10px;
    margin: auto;
    display: grid;
  }

  .left {
    grid-column: 1 / span 1;
    margin-right: 10px;
  }

  .center {
    grid-column: 2 / 8;
  }
</style>