<script>
  import './global.scss'
  import {onMount, tick} from "svelte";
  import {Pager} from "./pager";
  import {fetchTree, fetchDescription} from "./api";

  let container;
  let centerContainer;
  let treeContainer;
  let navigationTree;
  let query;
  let indexOfLastPage;
  let pager;
  let description;
  let showTree = true;
  let section;
  let searchResult;
  let searchButton;
  let treePages = []
  let matchIndex;
  let scrollEventsEnabled;

  async function scrollTo(id) {
    const domQuery = `.c[data-identifier="${id}"]`;
    let cLevel = treeContainer.querySelector(domQuery);
    if (!cLevel) {
      const result = await fetchTree({cLevelId: id});
      treePages = result.pages;
      await tick()
      cLevel = treeContainer.querySelector(domQuery);
    }
    cLevel.scrollIntoView();
    refitUI();
  }

  function refitUI() {
    searchButton.scrollIntoView();
    tick()
    setTimeout(() => scrollEventsEnabled = true);
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

  async function treeScrolled() {
    if (!scrollEventsEnabled) return;

    const firstPage = treePages[0]
    const lastPage = treePages[treePages.length - 1]
    if (!firstPage.container || !lastPage.container) return;
    const firstPageHeight = firstPage.container.getBoundingClientRect().height
    const lastPageTop = lastPage.container.getBoundingClientRect().top

    const scrollTop = centerContainer.scrollTop;
    if (scrollTop < firstPageHeight && firstPage.index !== 0) {
      const result = await fetchTree({
        page: firstPage.index - 1,
        query
      })
      // console.log('prepended pages', result.pages.map(p => p.index), 'to', ...treePages.slice(0, treePages.length - 1).map(p => p.index));
      treePages = [...result.pages, ...treePages.slice(0, treePages.length - 1)]
    } else if (lastPageTop <= 0 && lastPage.index < indexOfLastPage) {
      const result = await fetchTree({
        page: lastPage.index + 1,
        query
      })
      treePages = treePages.slice(1)
      await tick()
      treePages = [...treePages, ...result.pages]
      // console.log('appended page', result.pages.map(p => p.index), 'to', treePages.map(p => p.index));
      centerContainer.scrollTop = scrollTop - firstPageHeight;
    }
  }

  async function search() {
    if (!query) return;
    scrollEventsEnabled = false;
    pager = new Pager(treeContainer)
    searchResult = await pager.search({
      navigationTree: !navigationTree,
      search: true,
      query
    })
    treePages = searchResult.pages;
    matchIndex = pager.matchIndex;
    await tick();
    await pager.jump();
    refitUI();
  }

  async function displayDescription() {
    description = await fetchDescription({})
    await showSection(0)
    showTree = false;
  }

  async function displayTree() {
    if (!navigationTree) {
      await getTree()
    }
    showTree = true;
  }

  async function showSection(i) {
    if (!description.sections[i].html) {
      const result = await fetchDescription({index: i})
      description.sections[i].html = result.html;
    }
    description.activeIndex = i;
  }

  async function getTree() {
    const result = await fetchTree({
      navigationTree: !navigationTree,
      query
    })
    navigationTree = result.navigationTree
    treePages = result.pages
    indexOfLastPage = result.pageCount - 1;
  }

  function displayMatch(result) {
    if (result) {
      treePages = result.pages;
      waitForDOM(treePages, () => {
        pager.jump();
        refitUI();
      });
    } else {
      refitUI();
    }
  }

  async function previousPage() {
    scrollEventsEnabled = false;
    const result = await pager.previous();
    matchIndex = pager.matchIndex;
    displayMatch(result)
  }

  function waitForDOM(pages, done) {
    tick()
    if (pages) {
      for (const page of pages) {
        if (!page.container) {
          setTimeout(() => waitForDOM(pages, done))
          return;
        }
      }
    }
    done();
  }

  async function nextPage() {
    scrollEventsEnabled = false;
    const result = await pager.next();
    console.log(result);
    matchIndex = pager.matchIndex;
    displayMatch(result)
  }

  onMount(async () => {
    await getTree()
    centerContainer.addEventListener('scroll', treeScrolled, {passive: true})
  })
</script>

<div bind:this={container} id="description">
  <div>
    <input bind:value={query} type="text"/>
    <button bind:this={searchButton} on:click={search}>Zoeken</button>
    {#if searchResult}
      {#if searchResult.hitCount}
        <span>{matchIndex + 1} / {searchResult.hitCount}</span>
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
      {#if !showTree && description}
        {#each description.sections as section, i}
          <li><a href="#" on:click={() => showSection(i)}>{section.title}</a></li>
        {/each}
      {/if}
      <button on:click={displayTree}>Archiefbestanddelen</button>
    </div>
    {#if showTree && navigationTree}
      <div class="nav-tree" on:click={e => navTreeClicked(e)}>{@html navigationTree}</div>
    {/if}
  </div>

  <div bind:this={centerContainer} class="center">
    {#if !showTree && description}
      <div class="description">{@html description.sections[description.activeIndex].html}</div>
    {/if}
    {#if showTree && navigationTree}
      <div bind:this={treeContainer} class="tree">
        {#each treePages as page, i (page.index)}
          <div bind:this={page.container} data-index="{page.index}" class="page">
            {@html page.html}
          </div>
        {/each}
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

