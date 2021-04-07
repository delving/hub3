<script>
  import {doOnce} from "./doOnce";

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
</script>

{#each items as item}
  <a href="/nkDetail/{item['@id']}" class="item" let:value>
    {#each display as property}
      {#if value = item[property.value]}
        <p>
          {#if property.label}
            <label>{property.label}</label>
          {/if}
          {#if property.type === 'image'}
            <img src={value} alt="Geen foto"/>
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

<style>
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