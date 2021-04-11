<script>
  import {store} from "./store";

  export let type;
  export let remove;
  export let property;

  let isLiteral = type.find(i => i.indexOf('#Literal') >= 0)

  function removeChild(child) {
    property.properties = property.properties.filter(p => p !== child)
  }

  function add() {
    store.set({
      addition: {
        context: type,
        property
      }
    })
  }

  function toggleText() {
    property.hasTextContent = !property.hasTextContent
    if(!property.hasTextContent) {
      property.value = undefined;
    }
  }
</script>

<div class="header" class:root={!property.type} class:property={property.type}>

  {#if property.type}
    <button type="button" on:click={remove}>
      <img src="assets/icons/x-circle-fill.svg"/>
    </button>
    {property.about}
  {:else}
    <input bind:value={property.id} class="id" placeholder="id"/>
    {type}
  {/if}
  {#if !isLiteral}
    <button on:click={add}>
      <img src="assets/icons/plus.svg"/>
    </button>
    or
    <button on:click={toggleText}>
      {#if property.hasTextContent}
        Remove Text
      {:else}
        Add Text
      {/if}
    </button>
  {/if}

  <ul class="list-group-flush property-type">
    {#if property.type}
      <li class="list-group-item">
        <input bind:value={property.id} class="id" placeholder="id"/>
        <span class="type">[{property.type}]</span>
        <div>
          {#if property.hasTextContent || isLiteral}
            <input class="form-control" bind:value={property.value}/>
          {/if}
        </div>
      </li>
    {/if}
    <ul class="list-group-flush">
      {#each property.properties as property}
        <li class="list-group-item">
          <div>
            <svelte:self type={property.type} {property}
                         remove={() => removeChild(property)}/>
          </div>
        </li>
      {/each}
    </ul>
  </ul>
</div>

<style>
  .root, .root li {
    background-color: #CEC6C0;
  }

  .id {
    width: 3rem;
  }

  .property, .property li {
    background-color: #f7f7f7;
  }

  .property-type, .property-type li {
    background-color: #dadada;
    padding-left: 1rem;
  }

  .header {
    padding: 0.5rem;
  }

  li {
    border: none;
  }
</style>

