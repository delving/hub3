<script>
  import xml from './4.OSK.xml'
  import './global.scss'
  import {onMount, tick} from "svelte";
  import {getTree} from "./api";


  let container, treeContainer, navigationTree;
  let treePages = []

  // let description = ead.descriptions[0].html;

  // function showDescription(i) {
  //   description = ead.descriptions[i].html;
  //   console.log(description)
  // }

  async function navTreeClicked(e) {
    let target = e.target;
    while (target && !target.classList.contains('c')) {
      target = target.parentNode;
    }
    if (target.classList.contains('c')) {
      const identifier = target.dataset.identifier;
      let partner = treeContainer.querySelector(`.c[data-identifier="${identifier}"]`);
      if(!partner) {
        const result = await getTree({cLevelId: identifier})
        console.log(result)
        treePages = result.pages
        await tick()
        partner = treeContainer.querySelector(`.c[data-identifier="${identifier}"]`);
      }
      partner.scrollIntoView();
      target.classList.add('open');
    }
  }

  onMount(async () => {
    const result = await getTree({
      navigationTree: !navigationTree
    })
    navigationTree = result.navigationTree
    treePages = result.pages
  })
</script>

<div bind:this={container} id="description">
  <div class="left">
    <!--    <ul>-->
    <!--      {#each ead.descriptions as description, i}-->
    <!--        <li><a href="#" on:click={() => showDescription(i)}>{description.title}</a></li>-->
    <!--      {/each}-->
    <!--    </ul>-->
    {#if navigationTree}
      <div class="nav-tree" on:click={e => navTreeClicked(e)}>{@html navigationTree}</div>
    {/if}
  </div>

  <div class="center">
    <!--    <div class="description">{@html description}</div>-->
    <div bind:this={treeContainer} class="tree">
      {#each treePages as page, i}
        <div class="page" class:red={i % 2 === 0}>
          {@html page}
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
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
    grid-column: 2 / 8;
  }

  .center, .left {
    max-height: 100vh;
    min-height: 100%;
    overflow-y: scroll;
  }
</style>

