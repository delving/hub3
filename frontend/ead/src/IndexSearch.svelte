<script>
  import Index from "./Index.svelte";
  export let config;
  export let search;
  search.items = search.collapse.map(collapsedIndex => ({
    ...collapsedIndex.items[0]
  }))
</script>

<Index {config} {search}/>
