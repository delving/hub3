<script>
  export let services;
</script>

{#each services as service}
  <h2><a href="#service:{service.name}">{service.name}</a></h2>
  <ul>
    {#each service.methods as method}
      <li><a href="#serviceMethod:{service.name}.{method.name}">{method.name}</a></li>
    {/each}
  </ul>
{/each}