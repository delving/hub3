<script>
  export let search;
</script>

<ul class="list-group">
  {#each search.resources[0].entries as entry}
    {#if "@value" in entry}
      <li class="list-group-item">
        <label>{entry['searchLabel']}<p>{entry['@value']}</p></label>
      </li>
    {/if}
  {/each}
</ul>