<script>
  import './archive-detail.scss'
  import NavTree from "./NavTree.svelte";
  import VirtualScroller from "./VirtualScroller.svelte";

  export let tree;
  let treeContainer;
</script>

<div id="description">
  <div class="left">
    <div class="tree">
      <NavTree {treeContainer} navigationTree={tree.navigationTree}/>
    </div>
  </div>

  <div class="center" bind:this={treeContainer}>
    {#if treeContainer}
      <div bind:this={treeContainer} class="tree">
        <VirtualScroller scrollContainer={treeContainer} match={tree.match} pages={tree.pages}/>
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
</style>

