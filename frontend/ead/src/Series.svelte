<script>
  export let series;
</script>

<section>
  <header>
    <h1>Series</h1>
  </header>
  <div>
    {#each series as serie}
      <a href="/detail/{serie['@id']}">
        <figure>
          <img src={serie['edm:isShownBy']} alt="{serie['@id']} {serie['dc:title'][0]}"/>
          <figcaption>{serie['dc:title'][0]}</figcaption>
        </figure>
      </a>
    {/each}
  </div>
</section>

<style>
  div {
    display: grid;
    gap: 0.5em;
    max-width: 100%;
    grid-template-columns: repeat(5, 1fr);
  }

  figure {
    display: flex;
    flex-direction: column;
  }

  figure {
    margin: 0;
  }

  img {
    margin: auto;
    max-width: 100%;
    height: 20vh;
  }

  a {
    font-weight: bold;
    border: 1px solid black;
  }
</style>