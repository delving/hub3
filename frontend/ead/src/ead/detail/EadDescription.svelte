<script>
  import './archive-detail.scss'
  import {afterUpdate} from "svelte";
  import DescriptionSections from "./DescriptionSections.svelte";
  import {descriptionStore} from "./descriptionStore";
  import VirtualScroller from "./VirtualScroller.svelte";

  let showTree = false;
  let query;
  let pager;
  let searchButton;
  let match;
  let hitCount;
  export let description;

  // treeStore.subscribe(currValue => tree = currValue);
  // descriptionStore.subscribe(currValue => description = currValue);

  function refitUI() {
    searchButton.scrollIntoView();
  }

  // async function search() {
  //   pager = await (showTree ? treeStore : descriptionStore).search(query)
  //   // hitCount = showTree ? tree.hitCount : description.hitCount;
  //   match = pager.firstMatch()
  // }
  //
  // async function displayTree() {
  //   await treeStore.prepare()
  //   showTree = true;
  // }

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

  // onMount(async () => {
  //   if (showTree)
  //     await treeStore.prepare({
  //       cLevelId: route.values.cLevelPath
  //     })
  //   else
  //     await descriptionStore.prepare()
  // })

  afterUpdate(refitUI)
</script>

<div id="description">
  <div class="left">
    <div class="menu">
      {#if description}
        <DescriptionSections sections={description.sections}/>
      {/if}
    </div>
  </div>

  <div class="center" class:desc-panel={!showTree}>
    {#if description}
      <VirtualScroller match={match} pages={description.pages} pager={descriptionStore}/>
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
</style>