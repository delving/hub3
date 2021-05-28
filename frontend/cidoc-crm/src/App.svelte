<script lang="ts">
  import Type from "./Type.svelte";
  import "./starter.scss"
  import {crm} from "./crm"
  import Add from "./Add.svelte";
  import {store} from "./store";
  import {onMount} from "svelte";
  import {getModel} from "./import";

  let selected = []
  let state
  let formElement
  let isValid;
  let filename
  let lastSavedElement
  let lastSaved
  let errorMessageElement

  store.subscribe(currValue => state = currValue)

  function checkValidity() {
    isValid = formElement.checkValidity()
  }

  function createBaseType() {
    root.type = selected.map(i => i.about)
  }

  let models = []
  let root = createRoot()

  function createRoot() {
    return {
      about: '#root',
      uuid: Math.floor(new Date().getTime() / 1000),
      type: [],
      properties: []
    }
  }

  function remove() {
    if (confirm("Do you really want to delete the root node and start over?")) {
      const filename = root.filename
      root = createRoot()
      root.filename = filename
    }
  }

  async function save() {
    if (!root.filename) return;
    const request = {
      method: 'post',
      body: JSON.stringify(root),
      headers: {
        'Content-Type': 'application/json'
      }
    };
    console.log(JSON.stringify(root, undefined, 2))
    const response = await fetch('http://localhost:3000/save', request)
    if (response.status !== 200) {
      console.error('Failed to save changes')
    } else {
      lastSaved = new Date().getTime()
    }
  }

  onMount(async () => {
    const response = await fetch('http://localhost:3000/models', {
      method: 'post',
      body: JSON.stringify(root),
      headers: {
        'Content-Type': 'application/json'
      }
    })
    models = await response.json()
  })

  function addNewModel() {
    if (!filename.endsWith(".json")) {
      root.filename = filename + ".json"
    } else {
      root.filename = filename
    }
  }

  async function fetchModel(filename) {
    lastSaved = null
    const response = await fetch('http://localhost:3000/models', {
      method: 'post',
      body: JSON.stringify({
        filename
      }),
      headers: {
        'Content-Type': 'application/json'
      }
    })
    root = await response.json()
    console.log(root)
  }

  function lastSavedUpdater() {
    if (!lastSavedElement) return
    if (!lastSaved) {
      lastSavedElement.innerHTML = 'Last saved: Never'
    } else {
      const secondsElapsed = Math.floor((new Date().getTime() - lastSaved) / 1000) + 1
      lastSavedElement.innerHTML = `Last saved: ${secondsElapsed} seconds ago`
    }
  }

  function browseModels() {
    save()
    root = createRoot()
  }

  async function importModel() {
    const [model, err] = await getModel(true)
    if(!err) {
      model.filename = root.filename
      root = model
    } else {
      errorMessageElement.textContent = err
    }
  }

  setInterval(lastSavedUpdater, 900)
  setInterval(save, 8000)
</script>

<main>
  {#if !root.filename}
    <input bind:value={filename} class="form-control"/>
    <button type="button" class="btn btn-dark" on:click={addNewModel}>Add New Model</button>
    <hr/>
    <h1>Existing Models</h1>
    <ul class="list-group">
      {#each models as model}
        <li on:click={() => fetchModel(model)} class="list-group-item">
          <a href="#">{model}</a>
        </li>
      {/each}
    </ul>
  {:else if !state.change}
    {#if root.type.length === 0}
      <form bind:this={formElement}>
        <button disabled={!isValid} type="button" class="btn btn-dark" on:click={createBaseType}>Create base type
        </button>
        or
        <button type="button" class="btn btn-dark" on:click={importModel}>Import an existing model
        </button>
        <pre bind:this={errorMessageElement}></pre>
        <label>
          Select classes
          <select on:change={checkValidity} required size="90" multiple class="form-select" bind:value={selected}>
            {#each crm.classes as value}
              <option {value}>{value.about}</option>
            {/each}
          </select>
        </label>
      </form>
    {:else}
      <div class="last-saved">
        <button type="button" class="btn btn-dark" on:click={save}>Save</button>
        <div bind:this={lastSavedElement}></div>
      </div>
      <div>
        <div>
          <h2>{root.filename}</h2>
          <button type="button" class="btn btn-dark" on:click={browseModels}>Browse existing models</button>
        </div>
        <hr/>
        <ul class="root list-group type-list">
          <li class="list-group-item">
            <Type type={root.type} property={root} {remove}/>
          </li>
        </ul>
      </div>
    {/if}
  {:else}
    <Add change={state.change}/>
  {/if}
</main>

<style>
  .last-saved {
    position: absolute;
    top: 5px;
    right: 5px;
  }

  label {
    width: 100%;
  }

  h2 {
    display: inline;
  }

  .root, .root > li {
    background-color: darkgray;
  }

  pre {
    margin-top: 0.5rem;
  }
</style>