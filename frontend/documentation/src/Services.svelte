<script>
  import {linkService} from "./link";

  export let services;
</script>

<div>
  <h2>Services</h2>
  <p>The clients and JSON/HTTP RPC API model their capabilities through a
    small selection of services.</p>
  <ul>
    {#each services as service}
      <li><a href={linkService(service)}>{service.name}</a>—{service.comment}</li>
    {/each}
  </ul>
</div>