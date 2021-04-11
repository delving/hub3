<script lang="ts">
  import Type from "./Type.svelte";
  import "./starter.scss"
  import {crm} from "./crm"
  import Add from "./Add.svelte";
  import {store} from "./store";
  import {onMount} from "svelte";

  let selected = []
  let state
  let formElement
  let isValid;
  let filename

  store.subscribe(currValue => state = currValue)

  function checkValidity() {
    isValid = formElement.checkValidity()
  }

  function createBaseType() {
    result.root = selected.map(i => i.about)
  }

  let models = []
  let result = {
    root: [],
    properties: []
  };

  async function save() {
    if (!result.filename) return;
    const request = {
      method: 'post',
      body: JSON.stringify(result),
      headers: {
        'Content-Type': 'application/json'
      }
    };
    console.log(JSON.stringify(result, undefined, 2))
    try {
      const response = await fetch('http://localhost:3000/save', request)
      if (response.status !== 200) {
        console.error('Failed to save changes')
      }
    } catch (e) {
      console.error('Failed to save changes')
    }
  }

  onMount(async () => {
    const response = await fetch('http://localhost:3000/models', {
      method: 'post',
      body: JSON.stringify(result),
      headers: {
        'Content-Type': 'application/json'
      }
    })
    models = await response.json()
  })

  function addNewModel() {
    if (!filename.endsWith(".json")) {
      result.filename = filename + ".json"
    } else {
      result.filename = filename
    }
  }

  async function fetchModel(filename) {
    const response = await fetch('http://localhost:3000/models', {
      method: 'post',
      body: JSON.stringify({
        filename
      }),
      headers: {
        'Content-Type': 'application/json'
      }
    })
    result = await response.json()
    console.log(result)
  }

  setInterval(save, 10000)
</script>

<main>
  {#if !result.filename}
    <input bind:value={filename} class="form-control"/>
    <button on:click={addNewModel}>Add New Model</button>
    <hr/>
    <h1>Existing Models</h1>
    <ul class="list-group">
      {#each models as model}
        <li on:click={() => fetchModel(model)} class="list-group-item">
          <a href="#">{model}</a>
        </li>
      {/each}
    </ul>
  {:else if !state.addition}
    {#if result.root.length === 0}
      <form bind:this={formElement}>
        <button disabled={!isValid} type="button" class="btn btn-dark" on:click={createBaseType}>Create base type
        </button>
        <label>
          Select classes
          <select on:change={checkValidity} required size="90" multiple class="form-select" bind:value={selected}>
            {#each crm.classes as value}
              <option {value}>{value.labels.en}</option>
            {/each}
          </select>
        </label>
      </form>
    {:else}
      <div>
        <div>
          <button type="button" class="btn btn-dark" on:click={save}>Save</button>
        </div>
        <hr/>
        <Type type={result.root} property={result} remove={() => {}}/>
      </div>
    {/if}
  {:else}
    <Add addition={state.addition}/>
  {/if}
</main>

<style>
  label {
    width: 100%;
  }
</style>