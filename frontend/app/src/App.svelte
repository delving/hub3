<script>
  import "highlight.js/styles/darcula.css"
  import ServiceMethod from './ServiceMethod.svelte'
  import NavigationTree from "./NavigationTree.svelte";
  import Service from "./Service.svelte";
  import Services from "./Services.svelte";
  import APITree from "./APITree.svelte";
  import Topic from "./Topic.svelte";
  import {services} from '../../gen/def'

  export let topics;

  let routeId;
  let params = {};

  function route(hash) {
    const parts = hash.substring(1).split(':')
    if (parts <= 1) return null
    params = {}
    const properties = parts[1].split("&")
    for (const property of properties) {
      const splitProperty = property.split("=")
      params[splitProperty[0]] = splitProperty.length == 2 ? splitProperty[1] : ''
    }
    routeId = parts[0]
  }

  route(window.location.hash)

  window.onhashchange = () => {
    route(window.location.hash)
  }
</script>

<main>
  <div class="docs">
    <div class="left">
      <NavigationTree topics={topics}/>
      <APITree services={services}/>
    </div>
    <div class="center">
      {#if routeId === "service"}
        <Service serviceName={params.service}/>
      {:else if routeId === "serviceMethod"}
        <ServiceMethod serviceName={params.service} methodName={params.method}/>
      {:else if routeId === "topic"}
        <Topic {topics} topicId={params.topic}/>
      {:else}
        <Services services={services}/>
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