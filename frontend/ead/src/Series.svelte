<script>
  import Metadata from "./Metadata.svelte";

  export let config;
  export let series;
</script>

<section>
  <header>
    <h2>Objecten</h2>
  </header>

  {#each series as serie}
    <div>
      {#each config.sections as section}
        <div>
          <Metadata display={section} context={serie}/>
        </div>
      {/each}
    </div>
  {/each}

</section>

<style>
  section > div {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    max-width: 100%;
    gap: 0.5rem;
  }

  div > div:first-child {
    grid-column: 1 / 4;
  }


</style>