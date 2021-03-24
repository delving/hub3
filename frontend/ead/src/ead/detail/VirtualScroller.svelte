<svelte:options immutable/>
<script>
  import {onMount} from "svelte";

  export let pages;
  export let pager;
  export let match;

  let scrollContainer;
  let containers = []
  let prevMatchContainer;

  $: {
    if (match && scrollContainer) {
      for (const container of containers) {
        const page = +container.dataset.index;
        if (page === match.page) {
          const matchContainers = container.querySelectorAll('.dhcl');
          const matchContainer = matchContainers[match.index];
          if (prevMatchContainer)
            prevMatchContainer.classList.remove('active');
          if (matchContainer) {
            matchContainer.classList.add('active');
            matchContainer.scrollIntoView();
          }
          prevMatchContainer = matchContainer;
          break;
        }
      }
    } else {
      prevMatchContainer = null;
    }
  }

  function onScroll() {
    const scrollTop = scrollContainer.scrollTop

    let item;
    let index;
    for (let i = 2; i >= 0; i--) {
      item = pages[i]
      const container = containers[i]
      if (!container) return;
      index = i;
      if (scrollTop >= container.offsetTop) break;
    }
    if (index == 0 && item.index > 0) {
      containers = []
      pager.prependPage(pages[index].index - 1)
    } else if (index == 2) {
      containers = []
      pager.appendPage(pages[index].index + 1)
    }
  }

  onMount(() => {
    scrollContainer.addEventListener('scroll', onScroll, {passive: true})
  })
</script>

<div bind:this={scrollContainer} class="scroll">
  {#each pages as page, index (page.index)}
    <div data-index={page.index} bind:this={containers[index]} class="page">
      {@html page.html}
    </div>
  {/each}
</div>

<style>
  .scroll {
    max-height: 90vh;
    overflow-y: scroll;
  }
</style>