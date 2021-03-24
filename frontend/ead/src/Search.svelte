<svelte:options immutable/>
<script>
  import {queryStore} from "./queryStore";

  export let query;
  let q = query.q

  function update() {
    queryStore.updateQuery({q})
    q = ''
  }
</script>

<input bind:value={q}/>
<button on:click={update}>Zoeken</button>
Some stats
