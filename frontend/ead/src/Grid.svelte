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

  let items = search['dcterms:hasParts']

  const display = [
    {
      type: 'image',
      value: 'edm:isShownBy'
    },
    {
      value: 'dc:identifier'
    },
    {
      value: 'dc:title'
    }
  ]

  function createLink(item) {
    return linkTo(config.linkTo, item.meta)
  }
</script>

<div class="grid">
  {#each items as item}
    <a href={linkTo(config.linkTo, item)} class="item" let:value>
      {#each display as property}
        {#if value = item[property.value]}
          <p>
            {#if property.label}
              <label>{property.label}</label>
            {/if}
            {#if property.type === 'image'}
              <img class="rounded mx-auto d-block" src={value} alt="Geen foto"/>
            {:else}
              <span>{value}</span>
            {/if}
          </p>
        {:else if property.type === 'image'}
          <p>
            <img alt="Geen foto"/>
          </p>
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
    gap: 10px;
    max-height: 100%;

    grid-template-columns: repeat(4, 1fr);
    grid-auto-rows: auto;
  }

  label, span {
    display: inline;
  }

  img {
    display: block;
    min-height: 25vh;
    max-height: 25vh;
  }

  .item {
    display: flex;
    flex-direction: column;
    border: 1px solid black;
  }
</style>