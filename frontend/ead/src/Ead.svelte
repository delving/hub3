<script>
  import './global.scss'
  import {onMount, tick} from "svelte";
  import {getTree} from "./api";

  let container;
  let centerContainer;
  let treeContainer;
  let navigationTree;
  let query;
  let indexOfLastPage;
  let matches;
  let treePages = []

  // let description = ead.descriptions[0].html;

  // function showDescription(i) {
  //   description = ead.descriptions[i].html;
  //   console.log(description)
  // }

  async function scrollTo(id) {
    const domQuery = `.c[data-identifier="${id}"]`;
    let cLevel = treeContainer.querySelector(domQuery);
    if (!cLevel) {
      const result = await getTree({cLevelId: id});
      treePages = result.pages;
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

  async function treeScrolled(e) {
    const firstPage = treePages[0]
    const lastPage = treePages[treePages.length - 1]
    if (!firstPage.container || !lastPage.container) return;
    const firstPageHeight = firstPage.container.getBoundingClientRect().height
    const lastPageTop = lastPage.container.getBoundingClientRect().top

    const scrollTop = centerContainer.scrollTop;
    if (scrollTop < firstPageHeight && firstPage.index !== 0) {
      const result = await getTree({
        page: firstPage.index - 1,
        query
      })
      console.log('prepended pages', result.pages.map(p => p.index), 'to', ...treePages.slice(0, treePages.length - 1).map(p => p.index));
      treePages = [...result.pages, ...treePages.slice(0, treePages.length - 1)]
    } else if (lastPageTop <= 0 && lastPage.index < indexOfLastPage) {
      const result = await getTree({
        page: lastPage.index + 1,
        query
      })
      treePages = treePages.slice(1)
      await tick()
      treePages = [...treePages, ...result.pages]
      console.log('appended page', result.pages.map(p => p.index), 'to', treePages.map(p => p.index));
      centerContainer.scrollTop = scrollTop - firstPageHeight;
    }
  }

  async function search() {
    if (!query) return;
    const result = await getTree({
      navigationTree: !navigationTree,
      search: true,
      query
    })
    console.log(result)
    treePages = result.pages;
    matches = result.matches;
  }

  onMount(async () => {
    const result = await getTree({
      navigationTree: !navigationTree,
      query
    })
    navigationTree = result.navigationTree
    treePages = result.pages
    indexOfLastPage = result.pageCount - 1;
    centerContainer.addEventListener('scroll', treeScrolled, {passive: true})
  })
</script>

<div bind:this={container} id="description">
  <input bind:value={query} type="text"/>
  <button on:click={search}>Zoeken</button>
  {#if matches}
    {matches.length}
  {/if}
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

  <div bind:this={centerContainer} class="center">
    <!--    <div class="description">{@html description}</div>-->
    <div bind:this={treeContainer} class="tree">
      {#each treePages as page, i (page.index)}
        <div bind:this={page.container} class="page p{page.index}">
          {@html page.html}
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

