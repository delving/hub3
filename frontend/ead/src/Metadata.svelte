<svelte:options immutable/>
<script>
  import Value from "./Value.svelte";

  export let display;
  export let context;
</script>

<section>
  {#if display.label}
    <header><h3>{display.label}</h3></header>
  {/if}
  <div>
    {#if display.image && display.image.src in context}
      <img src={context[display.image.src][0]} alt={display.image.alt}/>
    {/if}
    <ul>
      {#each display.items as item}
        <li>
          <label>{item.label}
            <p>
              <Value {context} path={item.path}/>
            </p>
          </label>
        </li>
      {/each}
    </ul>
  </div>
</section>

<style lang="scss">
  section > div {
    display: flex;
    flex-direction: row;
  }

  img {
    width: 25%;
  }

  header {
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

  label {
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