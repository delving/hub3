<script>
  import './rdf.scss'
  import './collectienederland.scss'
  import response from './NK3189.json'
  import SearchPage from "./search/SearchPage.svelte";
  import {doOnce} from "./doOnce";

  const search = {...response}
  const events = {
    facetClicked: (facet, node) => {
      node.classList.toggle('open');
      doOnce(document.body, 'click', () => node.classList.remove('open'));
    }
  };

  let items = response['dcterms:hasParts']

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

<div class="search-page">
  <SearchPage {search} {events}>
    {#each items as item}
      <a href="detail/{item['@id']}" class="item" let:value>
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
  </SearchPage>
</div>

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