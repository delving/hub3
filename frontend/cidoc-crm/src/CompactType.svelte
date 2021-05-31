<script>
  export let parent;
  export let property;
  export let path = []
  export let index = 0;
  export let swapProperties

  let open = !parent
  let up
  let down
  let isView = false

  export let setView = (selectedProperty, selectedPath) => {
    property = selectedProperty
    path = selectedPath
    isView = true
  }

  $: if (parent) {
    up = null
    down = null

    if (index > 0) {
      up = parent.properties[index - 1]
    }

    if (index < parent.properties.length - 1) {
      down = parent.properties[index + 1]
    }
  }

  function swap(indexOfA, indexOfB) {
    const a = property.properties[indexOfA]
    const b = property.properties[indexOfB]
    property.properties[indexOfA] = b
    property.properties[indexOfB] = a
  }

  function moveUp() {
    swapProperties(index, index - 1)
  }

  function moveDown() {
    swapProperties(index, index + 1)
  }

  function closeAllNodes(node) {
    for (const property of node.properties) {
      closeAllNodes(property)
    }
    node.open = false
  }

  function viewAncestor(pathEntry) {
    setView(pathEntry.property, pathEntry.path)
  }

  function toggleNode() {
    setView(property, path)
  }
</script>

<div>
  {#if property.properties.length > 0 && !isView && property.about !== "#root"}
    <button on:click={toggleNode} type="button" class="toggle-node">
      {#if !open}
        <img title="Expand sub properties" src="assets/icons/file-minus-fill.svg"/>
      {:else}
        <img title="Collapse sub properties" src="assets/icons/file-plus-fill.svg"/>
      {/if}
      ...{property.properties.length}
    </button>
  {/if}
  {#if isView}
    {#each path as entry}
      <a  href="#" on:click={() => viewAncestor(entry)}>{entry.property.about}</a>->
    {/each}
  {/if}
  <strong>{property.about}</strong>
  {#if property.about !== "#root"}
    {#if up}
      <button type="button" on:click={moveUp} title="Move Up">
        <img src="assets/icons/arrow-bar-up.svg"/>
      </button>
    {/if}
    {#if down}
      <button type="button" on:click={moveDown} title="Move Down">
        <img src="assets/icons/arrow-bar-down.svg"/>
      </button>
    {/if}
  {/if}

</div>
{#if open}
  <ul class="list-group type-list">
    {#each property.properties as child, i}
      <li class="list-group-item">
        <svelte:self parent={property}
                     {setView}
                     path={path.concat({property, path})}
                     property={child}
                     index={i}
                     swapProperties={swap}/>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .toggle-node {
    border: none;
    background: none;
  }
</style>

