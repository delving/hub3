<script>
  import {doOnce} from "./doOnce";
  import {linkTo} from "./router";

  export let config;
  export let search;
  const events = {
    facetClicked: (facet, node) => {
      node.classList.toggle('open');
      doOnce(document.body, 'click', () => node.classList.remove('open'));
    }
  };

  let items = search.items

  function createLink(item) {
    return linkTo(config.linkTo, item.meta)
  }
</script>

<div class="grid">
  {#each items as item}
    <a href={linkTo(config.linkTo, item.meta)} class="item" let:value>
      {#each config.display as property}
        {#if value = item.summary[property.value]}
          {#if property.label}
            <label>{property.label}</label>
          {/if}
          {#if property.type === 'image'}
            <div class="image-container">
              <img class="rounded mx-auto d-block" src={value} alt="Geen foto"/>
            </div>
          {:else}
            <span>{value}</span>
          {/if}
        {:else if property.type === 'image'}
          <div class="image-container">
            Geen foto
          </div>
        {/if}
      {/each}
    </a>
  {/each}
</div>

<style>
  .grid {
    grid-area: grid;
    overflow-y: auto;
    display: grid;
    gap: 2rem;
    max-height: 100%;

    grid-template-columns: repeat(5, 1fr);
    grid-auto-rows: auto;
    background-color: #e9e9e9;
  }

  label, span {
    display: inline;
  }

  a {
    color: black;
    text-decoration: none;
  }

  span {
    padding: 0.3rem;
  }

  .image-container {
    display: flex;
    overflow: hidden;
    min-height: 220px;
    max-height: 220px;
  }

  .image-container > * {
    display: block;
    max-width: 100%;
    align-self: flex-end;
    text-align: center;
  }

  .item {
    background-color: white;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3), 0 0 40px rgba(128, 128, 128, 0.1) inset;
  }
</style>