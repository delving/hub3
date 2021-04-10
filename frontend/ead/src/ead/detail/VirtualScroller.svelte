<svelte:options immutable/>
<script>
  import {afterUpdate, onMount} from "svelte";
  import {searchStore} from "../../searchStore";

  export let pages;
  export let match;
  export let scrollContainer;

  let containers = []
  let prevMatchContainer;

  afterUpdate(() => {
    if (match && scrollContainer) {
      for (const container of containers) {
        const page = +container.dataset.index;
        if (page === match.page) {
          const matchContainers = container.querySelectorAll('em');
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
  })

  function onScroll() {
    const scrollTop = Math.abs(scrollContainer.getBoundingClientRect().y)
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
      searchStore.prependPage(pages[index].index - 1)
    } else if (index == 2) {
      containers = []
      searchStore.appendPage(pages[index].index + 1)
    }
  }

  onMount(() => {
    scrollContainer.addEventListener('scroll', onScroll, {passive: true})
  })
</script>

{#each pages as page, index (page.index)}
  {#if page.html}
    <div data-index={page.index} bind:this={containers[index]} class="page">
      {@html page.html}
    </div>
  {/if}
{/each}
