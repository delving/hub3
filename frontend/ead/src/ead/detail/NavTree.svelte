<script>
  import {treeStore} from "./treeStore";
  import {tick} from "svelte";
  import {dom} from "../../dom";
  import Html from "../../Html.svelte";

  let navTree;
  treeStore.subscribe(currValue => navTree = currValue.navigationTree);

  async function scrollTo(id) {
    const domQuery = `.c[data-identifier="${id}"]`;
    let cLevel = dom.treeContainer.querySelector(domQuery);
    if (!cLevel) {
      await treeStore.prepare({cLevelId: `@${id}`});
      await tick()
      cLevel = dom.treeContainer.querySelector(domQuery);
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

{#if navTree}
  <div class="nav-tree" on:click={e => navTreeClicked(e)}>
    <Html node={navTree}/>
  </div>
{/if}