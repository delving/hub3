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

<style>
  div {
    box-sizing: border-box;
    max-height: 100vh;
    height: 100vh;
    padding: 1em 1em 0 1em;
    background-color: rgb(243, 243, 243);
  }

  ul {
    list-style-type: none;
    padding: 1em 0 0 0;
  }
</style>