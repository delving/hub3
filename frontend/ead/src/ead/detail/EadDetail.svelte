<script>
  import './archive-detail.scss'
  import {afterUpdate, onMount} from "svelte";
  import {dom} from '../../dom'
  import {treeStore} from './treeStore'
  import NavTree from "./NavTree.svelte";
  import DescriptionSections from "./DescriptionSections.svelte";
  import {descriptionStore} from "./descriptionStore";
  import VirtualScroller from "./VirtualScroller.svelte";

  export let route;

  let showTree = route.routeId !== 'eadDescription';
  let query;
  let pager;
  let searchButton;
  let match;
  let hitCount;
  let tree, description;

  treeStore.subscribe(currValue => tree = currValue);
  descriptionStore.subscribe(currValue => description = currValue);

  function refitUI() {
    searchButton.scrollIntoView();
  }

  async function search() {
    pager = await (showTree ? treeStore : descriptionStore).search(query)
    hitCount = showTree ? tree.hitCount : description.hitCount;
    match = pager.firstMatch()
  }

  async function displayTree() {
    await treeStore.prepare()
    showTree = true;
  }

  async function displayDescription() {
    await descriptionStore.prepare()
    showTree = false;
  }

  async function previousPage() {
    match = await pager.previous();
  }

  async function nextPage() {
    match = await pager.next();
  }

  onMount(async () => {
    if (showTree)
      await treeStore.prepare({
        cLevelId: route.values.cLevelPath
      })
    else
      await descriptionStore.prepare()
  })

  afterUpdate(refitUI)
</script>

<div id="description">
  <div>
    <input bind:value={query} type="text"/>
    <button bind:this={searchButton} disabled={!query} on:click={search}>Zoeken</button>
    {#if match}
      {#if hitCount}
        <span>{match.displayString} / {hitCount}</span>
        <button on:click={previousPage} disabled={match.isFirst}>Previous</button>
        <button on:click={nextPage} disabled={match.isLast}>Next</button>
      {:else}
        Geen resultaten gevonden
      {/if}
    {/if}
  </div>
  <div class="left">
    <div class="menu">
      <button on:click={displayDescription}>Beschrijving</button>
      {#if !showTree && description}
        <DescriptionSections/>
      {/if}
      <button on:click={displayTree}>Archiefbestanddelen</button>
    </div>
    {#if showTree}
      <div class="tree">
        <NavTree/>
      </div>
    {/if}
  </div>

  <div class="center" class:desc-panel={!showTree}>
    {#if !showTree && description}
      <VirtualScroller match={match} pages={description.pages} pager={descriptionStore}/>
    {/if}
    {#if showTree}
      <div bind:this={dom.treeContainer} class="tree">
        <VirtualScroller match={match} pages={tree.pages} pager={treeStore}/>
      </div>
    {/if}
  </div>
</div>

<style type="text/scss">

  #description {
    overflow: hidden;
    max-height: 300px;
    margin-bottom: 30px;
  }

  #description {
    margin: 0;
    padding: 0;
    display: grid;
    max-height: 100vh;
  }

  .left {
    grid-column: 1 / span 1;
    margin-right: 10px;
  }

  .center {
    max-height: 100%;
    grid-column: 2 / 8;
  }

  .center, .left {
    max-height: 100vh;
    min-height: 100%;
    overflow-y: scroll;
  }

  .menu {
    button {
      font-weight: bold;
      text-align: left;
      width: 100%;
      display: block;
    }
  }
</style>

