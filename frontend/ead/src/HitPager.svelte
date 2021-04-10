<script>
  export let pager;
  let hitCount = pager.hitCount;
  let match = pager.firstMatch();

  function previousPage() {
    match = pager.previous();
  }

  function nextPage() {
    match = pager.next();
  }
</script>

<div>
  {#if match}
    {#if hitCount}
      <span>{match.displayString} / {hitCount}</span>
      <button on:click={previousPage} disabled={match.isFirst}>Previous</button>
      <button on:click={nextPage} disabled={match.isLast}>Next</button>
    {:else}
      Geen resultaten gevonden
    {/if}
  {/if}
</div>