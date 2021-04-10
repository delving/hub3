<svelte:options immutable/>
<script>
  import {tick} from "svelte";
  import {searchStore} from "../../searchStore";

  export let navigationTree
  export let treeContainer;
  let navTree
  $: if (navigationTree) navTree = navigationTree

  async function scrollTo(id) {
    const domQuery = `.c[data-identifier="${id}"]`;
    let cLevel = treeContainer.querySelector(domQuery);
    if (!cLevel) {
      await searchStore.prepare({findById: `@${id}`});
      await tick()
      cLevel = treeContainer.querySelector(domQuery);
    }
    cLevel.scrollIntoView();
  }

  async function navTreeClicked(e) {
    let target = e.target;
    while (target && !target.classList.contains('c')) {
      target = target.parentNode;
    }
    if (target.classList.contains('c')) {
      await scrollTo(target.dataset.identifier);
      target.classList.add('open');
    }
  }
</script>

<div class="nav-tree" on:click={e => navTreeClicked(e)}>{@html navTree}</div>