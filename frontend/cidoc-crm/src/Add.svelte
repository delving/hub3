<script context="module">
  let latestAddition
</script>

<script lang="ts">
  import {getAllowedProperties, getAllowedTypes} from "./crm";
  import {store} from "./store";
  import {afterUpdate} from "svelte";
  import {getModel} from "./import";

  export let change;

  let selectedProperty;
  let allowedProperties = getAllowedProperties(change.property.about, change.context)
  let isValid
  let formElement
  let allowedTypes;
  let errorMessageElement;
  let classesDisplayLimit = 1;
  let range

  $: if(change.type === "update") {
    selectedProperty = allowedProperties.find(p => change.property.about === p.about)
    propertySelectionChanged()
  }

  function propertySelectionChanged() {
    if (!selectedProperty) {
      allowedTypes = null;
      updateClassList()
      return;
    }
    allowedTypes = getAllowedTypes(selectedProperty.about)
    updateClassList()
  }

  function checkValidity() {
    if (formElement) {
      isValid = formElement.checkValidity()
    }
  }

  function addProperty() {
    const newProperty = {
      latest: true,
      about: selectedProperty.about,
      type: range.map(i => i.about),
      uuid: Math.floor(new Date().getTime() / 1000),
      properties: [],
    };
    insertNewProperty(newProperty)
  }

  function insertNewProperty(newProperty) {
    if (latestAddition) {
      latestAddition.latest = false
    }
    latestAddition = newProperty
    change.property.properties = [newProperty, ...change.property.properties]
    store.set({})
  }

  function toggleRestrictions(e) {
    const restrictionsDisabled = e.target.checked
    allowedTypes = getAllowedTypes(selectedProperty.about, restrictionsDisabled)
    updateClassList()
  }

  function updateClassList() {
    const allowedTypesCount = allowedTypes ? allowedTypes.length : 1;
    classesDisplayLimit = Math.min(30, allowedTypesCount)
    if (allowedTypesCount === 1) {
      range = [allowedTypes[0]]
    }
  }

  function cancel() {
    store.set({})
  }

  async function importModel() {
    const [model, err] = await getModel(false)
    if (!err) {
      insertNewProperty(model)
    } else {
      errorMessageElement.textContent = err
    }
  }

  afterUpdate(() => checkValidity())
</script>

<form bind:this={formElement}>
  <button disabled={!isValid} on:click={addProperty}
          type="button" class="btn btn-dark">Add property
  </button>
  <button on:click={cancel}
          type="button" class="btn btn-dark">Cancel
  </button>
  <div class="property">
    <select required class="form-select" bind:value={selectedProperty} on:change={propertySelectionChanged}>
      <option value="">--Pick a property--</option>
      {#each allowedProperties as value}
        <option {value}>{value.about}</option>
      {/each}
    </select>
    <div>
      or <button type="button" class="btn btn-dark" on:click={importModel}>Import existing nodes from clipboard</button>
    </div>
  </div>
  <div>
    <label>Pick at least one class
      <label>
        <input disabled={!selectedProperty} on:change={toggleRestrictions} type="checkbox"/>
        Disable restrictions
      </label>
      <select
        bind:value={range}
        on:change={checkValidity}
        disabled={!allowedTypes}
        size="{classesDisplayLimit}"
        required multiple class="form-select">
        {#if allowedTypes}
          {#each allowedTypes as value}
            <option {value}>{value.about}</option>
          {/each}
        {/if}
      </select>

    </label>
  </div>
  <div>
    <pre bind:this={errorMessageElement}></pre>
  </div>
</form>

<style>
  .property {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
  }

  .property select {
    flex-grow: 1;
  }

  .property div {
    white-space: nowrap;
  }

  label {
    width: 100%;
  }

  label label {
    width: auto;
  }

  textarea {
    min-height: 1000px;
  }
</style>