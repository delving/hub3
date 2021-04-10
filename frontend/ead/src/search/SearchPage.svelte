<script>
  import {queryStore} from "./queryStore";
  import {config} from "./../config"
  import Pager from "../Pager.svelte";
  import Grid from "../Grid.svelte";
  import {searchStore} from "../searchStore";
  import EadSearch from "../ead/search/EadSearch.svelte";
  import Search from "./Search.svelte";
  import EadDetail from "../ead/detail/EadDetail.svelte";
  import IndexSearch from "../IndexSearch.svelte";
  import Index from "../Index.svelte";
  import IndexDetailPage from "../IndexDetailPage.svelte";
  import {linkTo, updateRoute, route} from "../router";
  import EadDescription from "../ead/detail/EadDescription.svelte";
  import RDFDetail from "../RDFDetail.svelte";

  let routeConfig = route.config;
  let search;
  let query;
  let tabsContainer;

  queryStore.subscribe(currValue => {
    query = currValue;
    searchStore.prepare(query)
  })

  searchStore.subscribe(currValue => {
    search = currValue
  })

  function show(e) {
    history.pushState(null, document.title, e.target.href)
    search = null
    routeConfig = updateRoute().config
  }

  function createNavLink(component) {
    return linkTo(component.routes[0], {});
  }
</script>

<section data-tab-type={routeConfig.type}>
  <Search {routeConfig} {search} {query}/>
  {#if routeConfig.navigation !== false}
    <div class="tabs">
      <ul bind:this={tabsContainer} class="nav nav-tabs">
        {#each config.components as component}
          {#if component.navigation !== false}
            <li class="nav-item">
              <a on:click|preventDefault={show}
                 href={createNavLink(component)}
                 class:active={component.active} class="nav-link" aria-current="page">{component.tabLabel}</a>
            </li>
          {/if}
        {/each}
      </ul>
    </div>
  {/if}
  {#if search}
    {#if routeConfig.type === 'grid'}
      <Grid config={routeConfig} {search}/>
    {:else if routeConfig.type === 'archive'}
      <EadSearch {search}/>
    {:else if routeConfig.type === 'findingAid'}
      <EadDetail tree={search}/>
    {:else if routeConfig.type === 'findingAidDescription'}
      <EadDescription description={search}/>
    {:else if routeConfig.type === 'indexSearch'}
      <IndexSearch config={routeConfig.display} {search}/>
    {:else if routeConfig.type === 'index'}
      <Index config={routeConfig.display} {search}/>
    {:else if routeConfig.type === 'indexDetail'}
      <IndexDetailPage config={routeConfig.display} {search}/>
    {:else if routeConfig.type === 'image'}
      <RDFDetail config={routeConfig.display} new_record={search}/>
    {/if}
    {#if routeConfig.pagination !== false}
      <div class="pager">
        <Pager/>
      </div>
    {/if}
  {/if}
</section>

<style type="text/scss">
  section {
    display: flex;
    flex-direction: column;
    max-height: 100vh;
    height: 100vh;
  }
</style>