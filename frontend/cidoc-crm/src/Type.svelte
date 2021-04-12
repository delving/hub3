<script>
  import {store} from "./store";

  export let type;
  export let remove;
  export let property;
  export let rootElement;

  let latest = property.latest
  delete property.latest
  let uuidElement;

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

  function truncateClasses(classes) {
    let displayString = classes[0]
    for (let i = 1; i < classes.length; i++) {
      let type = classes[i]
      if (type.length + displayString.length > 35) return displayString + ", ..."
      displayString += `, ${type}`
    }
    return displayString
  }

  function initiateCopy(e) {
    uuidElement.removeAttribute("disabled")
    uuidElement.select()
    uuidElement.setSelectionRange(0, 99999)
    document.execCommand("copy")
    uuidElement.setAttribute("disabled", "disabled")
  }

  function jumpToParent(e) {
    let parent = rootElement.parentNode
    while (!parent.classList.contains('header')) {
      parent = parent.parentNode
    }
    console.log(parent)
    parent.scrollIntoView({behavior: 'smooth'})
  }
</script>

<div bind:this={rootElement} class:latest={latest} class="header" class:root={!property.type}
     class:property={property.type}>
  {#if property.about !== '#root'}
    <div class="jump-to-parent">
      <button on:click={jumpToParent} type="button">
        <img src="assets/icons/caret-up-fill.svg"/>
      </button>
    </div>
  {/if}
  <div>
    <button type="button" on:click={remove}>
      <img src="assets/icons/x-circle-fill.svg"/>
    </button>
    {property.about}

    <button on:click={add}>
      <img src="assets/icons/plus.svg"/>
    </button>
    <input bind:this={uuidElement} class="uuid" disabled value="#[{property.uuid}]">
    <button on:click={initiateCopy}>
      <img src="assets/icons/clipboard.svg"/>
    </button>
  </div>

  <div class="content">
    <div>
    <span class="classes">
      <img src="assets/icons/chevron-double-down.svg"/>
      <ul class="list-group">
        {#each type as c}
          <li class="list-group-item">{c}</li>
        {/each}
      </ul>
    </span>
      <span class="bracket">[</span>
      {truncateClasses(property.type)}
      <span class="bracket">]</span>
    </div>
    <div class="right">
      <div class="inputs">
        <input placeholder="Id" class="form-control id" bind:value={property.id} required/>
        <input placeholder="Source" class="form-control" bind:value={property.value} required/>
      </div>
      <label>
        Gen:
        <input bind:checked={property.gen} class="form-check-inline" type="checkbox"/>
      </label>
    </div>
  </div>
  <ul class="list-group type-list">
    {#each property.properties as property}
      <li class="list-group-item">
        <div>
          <svelte:self type={property.type} {property}
                       remove={() => removeChild(property)}/>
        </div>
      </li>
    {/each}
  </ul>
</div>

<style>
  .root, .root li {
    background-color: #CEC6C0;
  }

  label {
    font-weight: bold;
  }

  input {
    display: inline;
    width: auto;
  }

  .property, .property li {
    background-color: #f7f7f7;
  }

  .header {
    padding: 0.5rem;
  }

  .latest {
    border: 3px solid blue;
  }

  .classes:hover ul {
    display: block;
    z-index: 1;
  }

  .classes ul {
    display: none;
    position: absolute;
    border: 1px solid black;
  }

  .right {
    display: flex;
    flex-grow: 1;
    gap: 0.5rem;
  }

  .content .bracket {
    font-weight: bold;
  }

  .content {
    padding-top: 0.2rem;
    display: inline-flex;
    width: 100%;
    gap: 1rem;
  }

  .inputs {
    display: inline-flex;
    flex-grow: 1;
    gap: 1rem;
  }

  .inputs input {
    width: 50%;
  }

  .uuid {
    color: blue;
  }

  .header,
  .type-list > li {
    padding-right: 0;
  }

  .type-list > li {
    border-width: 0.5rem;
    border-bottom: 0;
    border-right: 0;
  }

  .jump-to-parent {
    top: -0.9rem;
    left: -0.9rem;
    position: absolute;
  }

  .jump-to-parent button {
    border-radius: 50%;
  }
</style>

