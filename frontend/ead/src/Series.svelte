<script>
  import Metadata from "./Metadata.svelte";

  export let config;
  export let series;
</script>

<section>
  <header>
    <h2>Objecten</h2>
  </header>
  <div>
    {#each series as serie}
      {#each config.sections as section}
        <Metadata display={section} context={serie}/>
      {/each}
    {/each}
  </div>
</section>