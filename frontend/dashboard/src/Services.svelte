<script>
  import {services} from '../../gen/def'
  import {linkService} from "./link";
</script>

<div>
  <h2>Services</h2>
  <ul>
    {#each services as service}
      <li><a href={linkService(service)}>{service.name}</a></li>
    {/each}
  </ul>
</div>