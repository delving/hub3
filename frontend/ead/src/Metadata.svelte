<svelte:options immutable/>
<script>
  import Value from "./Value.svelte";

  export let display;
  export let context;

  console.log(display, context['rdf_label'])
</script>

<section>
  {#if display.image}
    <figure>
      {#if display.image.src in context}
        <img src={context[display.image.src][0]} alt={display.image.alt}/>
      {:else}
        <figcaption>Afbeelding niet beschikbaar</figcaption>
      {/if}
    </figure>
  {/if}
  <div>
    {#if display.label}
      <header>
        <h3>
          {#if context[display.label]}
            {context[display.label][0]}
          {:else}
            {display.label}
          {/if}
        </h3>
      </header>
    {/if}
    <ul>
      {#each display.items as item}
        <li>
          {#if item.label}
            <header>{item.label}</header>
          {/if}
          <p>
            <Value {context} path={item.path} value={item.values} style={item.style}/>
          </p>

        </li>
      {/each}
    </ul>
  </div>
</section>

<style lang="scss">
  section {
    display: flex;
    gap: 0.5rem;
    flex-direction: row;
  }

  section > div {
    flex-grow: 1;
  }

  figcaption {
    padding: 1rem;
    text-align: center;
    background-color: #e9e9e9;
    height: 100%;
  }

  img {
    min-width: 25%;
  }

  div > header {
    padding: 0.25rem;
    background-color: #e9e9e9;
  }

  ul {
    width: 100%;
    padding: 0;
    list-style-type: none;
  }

  p {
    margin: 0;
  }

  li > header {
    font-weight: bold;
    color: #999;
  }

  label > p {
    display: block;
    font-weight: normal;
    color: black;
  }

  li {
    padding: 0.25rem 0;
    border-bottom: 1px solid black;
  }
</style>