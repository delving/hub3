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

<header class="search">
  <div class="form">
    <input bind:value={q}/>
    <button on:click={update}>Zoeken</button>
  </div>
</header>

<style type="text/scss">
  .search {
    display: grid;
    grid-area: search;
  }
  .form {
    grid-column: 2 / 8;
    display: flex;
  }
  input {
    flex-grow: 1;
  }
</style>
