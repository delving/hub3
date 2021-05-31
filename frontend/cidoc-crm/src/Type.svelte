<script>
  import {store} from "./store";
  import {afterUpdate} from "svelte";

  export let type;
  export let remove;
  export let property;
  export let parent;
  export let isParentHidden

  let flash
  let showComment
  let latest = property.latest
  delete property.latest
  let uuidElement;
  let commentElement
  let hidden = false

  $: { hidden = isParentHidden }
  $: hasComment = property.comment && property.comment.trim()

  function removeChild(child) {
    if (confirm(`Do you really want to delete ${child.uuid}?`)) {
      property.properties = property.properties.filter(p => p !== child)
    }
  }

  function add() {
    store.set({
      change: {
        type: "create",
        context: type,
        property
      }
    })
  }

  function editProperty() {
    store.set({
      change: {
        type: "update",
        parentProperty: parent,
        context: parent ? parent.type : null,
        property
      }
    })
  }

  function truncateClasses(classes) {
    let displayString = classes[0]
    for (let i = 1; i < classes.length; i++) {
      let type = classes[i]
      if (type.length + displayString.length > 50) return displayString + ", ..."
      displayString += `, ${type}`
    }
    return displayString
  }

  function copyNode() {
    const inputElement = document.createElement("textarea")
    document.body.appendChild(inputElement)
    inputElement.style.position = "absolute";
    inputElement.style.left = "-9000px";
    inputElement.value = JSON.stringify(property, undefined, 2)
    inputElement.select()
    inputElement.setSelectionRange(0, 99999)
    document.execCommand("copy")
    flash = true
    setTimeout(() => document.body.removeChild(inputElement))
    setTimeout(() => flash = false, 500)
  }

  function initiateCopy() {
    uuidElement.removeAttribute("disabled")
    uuidElement.select()
    uuidElement.setSelectionRange(0, 99999)
    document.execCommand("copy")
    uuidElement.setAttribute("disabled", "disabled")
  }

  function toggleNode() {
    hidden = !hidden
  }

  function updateComment() {
    if (commentElement) {
      commentElement.style.height = `${commentElement.scrollHeight}px`;
    }
  }

  function toggleComment() {
    showComment = !showComment
  }

  afterUpdate(() => {
    updateComment()
  })
</script>

<div class="header"
     class:latest={latest}
     class:root={!property.type}
     class:property={property.type}
     class:flash={flash}>
  <div class="first-line">
    <div class="left">
      {#if parent}
        <button type="button" on:click={editProperty} title="Edit">
          <img src="assets/icons/pencil.svg"/>
        </button>
      {/if}
      <button type="button" on:click={copyNode} title="Copy as JSON">
        <img src="assets/icons/clipboard-plus.svg"/>
      </button>
      <button on:click={add} title="Add Node">
        <img src="assets/icons/plus.svg"/>
      </button>
      <span>{property.about}</span>
      <input bind:this={uuidElement} class="uuid" disabled value="#[{property.uuid}]">
      <button on:click={initiateCopy} title="Copy UUID">
        <img src="assets/icons/clipboard.svg"/>
      </button>
    </div>

    <div class="right">
      <strong>IF:&nbsp;</strong>
      <input type="text" class="form-control" placeholder="condition" bind:value={property.if}/>
      <button type="button" on:click={toggleComment} title="Toggle comment">
        {#if hasComment}<span>1</span>{/if}
        <img src="assets/icons/chat-text-fill.svg"/>
      </button>
      <button type="button" on:click={remove} title="Delete">
        <img src="assets/icons/x-circle-fill.svg"/>
      </button>
    </div>
  </div>

  <textarea bind:this={commentElement}
            bind:value={property.comment}
            class:hidden={!showComment}
            on:input={updateComment}></textarea>

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

  {#if property.properties.length > 0}
    <div class="toggle-node">
      <button on:click={toggleNode} type="button">
        {#if hidden}
          <img title="Expand sub properties" src="assets/icons/file-minus-fill.svg"/>
        {:else}
          <img title="Collapse sub properties" src="assets/icons/file-plus-fill.svg"/>
        {/if}
      </button>
    </div>
  {/if}
  {#if !hidden}
    <ul class="list-group type-list">
      {#each property.properties as child}
        <li class="list-group-item">
          <div>
            <svelte:self parent={property}
                         type={child.type}
                         property={child}
                         isParentHidden={hidden}
                         remove={() => removeChild(child)}/>
          </div>
        </li>
      {/each}
    </ul>
  {:else}
    <strong>...{property.properties.length} properties hidden</strong>
  {/if}
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

  textarea {
    width: 100%;
    resize: none;
    overflow: hidden;
    min-height: 50px;
  }

  .content, textarea {
    padding-top: 0.2rem;
  }

  .content {
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

  .toggle-node {
    top: -1rem;
    left: -1rem;
    position: absolute;
  }

  .toggle-node button {
    border: none;
    background: none;
  }

  .toggle-node img {
    width: 1rem;
    height: 1rem;
  }

  .first-line {
    display: flex;
    gap: 0.5rem;
    flex-direction: row;
  }

  .first-line .left, .first-line .right {
    display: inline-block;
  }

  .first-line .right {
    display: flex;
    flex-direction: row;
  }

  .first-line .right input {
    flex-grow: 1;
  }

  .hidden {
    display: none;
  }
</style>

