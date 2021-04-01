<script>
  import './rdf.scss'
  import './collectienederland.scss'
  import * as response from './v2.json'
  import {rdfToHtml} from "./rdf";
  import SearchPage from "./search/SearchPage.svelte";
  import {doOnce} from "./doOnce";

  const search = {...response}
  const events = {
    facetClicked: (facet, node) => {
      node.classList.toggle('open');
      doOnce(document.body, 'click', () => node.classList.remove('open'));
    }
  };

  const searchConfig = {
    display: [
      {type: 'image', searchLabel: ['edm_hasView', 'nave_thumbSmall']},
      {searchLabel: ['dc_creator']},
      {label: 'Bron', searchLabel: ['edm_provider']},
    ]
  };

  rdfToHtml(response.items, searchConfig);
  let items = response.items;
</script>

<div class="search-page">
  <SearchPage {search} {events}>
    {#each items as item}
      <a href="detail/{item.id}" class="item">
        {@html item.html}
      </a>
    {/each}
  </SearchPage>
</div>

<style>

  .item {
    display: flex;
    flex-direction: column;
    border: 1px solid black;
  }
</style>