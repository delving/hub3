<script>
  import {linkService, linkServiceMethod} from "./link";

  export let services;
</script>

<h2 class="bright-color">API Reference</h2>
{#each services as service}
  <h2><a href={linkService(service)}>{service.name}</a></h2>
  <ul>
    {#each service.methods as method}
      <li><a href={linkServiceMethod(service, method)}>{method.name}</a></li>
    {/each}
  </ul>
{/each}

<style>
  ul {
    margin: 0;
    padding: 0;
    list-style-type: none;
  }

  li {
    margin-left: 20px;
  }
</style>