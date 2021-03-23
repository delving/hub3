<script>
  import './archive-detail.scss'
  import {afterUpdate, onMount} from "svelte";
  import {dom} from './dom'
  import {treeStore} from './treeStore'
  import NavTree from "./NavTree.svelte";
  import DescriptionSections from "./DescriptionSections.svelte";
  import {description} from "./description";
  import VirtualScroller from "./VirtualScroller.svelte";

  export let route;

  let showTree = route.routeId !== 'eadDescription';
  let query;
  let pager;
  let searchButton;
  let match;
  let tree;

  treeStore.subscribe(currValue => tree = currValue);

  function refitUI() {
    searchButton.scrollIntoView();
  }

  async function search() {
    pager = await treeStore.search(query)
    match = pager.firstMatch()
  }

  async function displayTree() {
    await treeStore.prepare()
    showTree = true;
  }

  function displayDescription() {
    showTree = false;
  }

  async function previousPage() {
    match = await pager.previous();
  }

  async function nextPage() {
    match = await pager.next();
  }

  onMount(async () => {
    await treeStore.prepare()
  })

  afterUpdate(refitUI)
</script>

<div id="description">
  <div>
    <input bind:value={query} type="text"/>
    <button bind:this={searchButton} disabled={!query} on:click={search}>Zoeken</button>
    {#if match}
      {#if tree.hitCount}
        <span>{match.index + 1} / {tree.hitCount}</span>
        <button on:click={previousPage}>Previous</button>
        <button on:click={nextPage}>Next</button>
      {:else}
        Geen resultaten gevonden
      {/if}
    {/if}
  </div>
  <div class="left">
    <div class="menu">
      <button on:click={displayDescription}>Beschrijving</button>
      {#if !showTree}
        <DescriptionSections/>
      {/if}
      <button on:click={displayTree}>Archiefbestanddelen</button>
    </div>
    {#if showTree}
      <NavTree/>
    {/if}
  </div>

  <div class="center">
    {#if !showTree && $description.activeSection}
      <div class="description">{@html $description.activeSection}</div>
    {/if}
    {#if showTree}
      <div bind:this={dom.treeContainer} class="tree">
        <VirtualScroller match={match} pages={tree.pages} pager={treeStore}/>
      </div>
    {/if}
  </div>
</div>

<style type="text/scss">
  .description {
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

