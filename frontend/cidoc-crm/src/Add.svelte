<script context="module">
  let latestAddition
</script>

<script>
  import {getAllowedProperties, getAllowedTypes} from "./crm";
  import {store} from "./store";

  export let addition;

  console.log(addition)
  let allowedProperties = getAllowedProperties(addition.property.about, addition.context)
  let isValid
  let formElement
  let allowedTypes;
  let selectedProperty;
  let range

  function change() {
    if (!selectedProperty) {
      allowedTypes = null;
      return;
    }
    allowedTypes = getAllowedTypes(selectedProperty.about)
  }

  function checkValidity() {
    isValid = formElement.checkValidity()
  }

  function addProperty() {
    const newProperty = {
      latest: true,
      about: selectedProperty.about,
      type: range.map(i => i.about),
      uuid: Math.floor(new Date().getTime() / 1000),
      properties: [],
    };
    if (latestAddition) {
      latestAddition.latest = false
    }
    latestAddition = newProperty
    addition.property.properties = [...addition.property.properties, newProperty]
    store.set({})
  }

  function toggleRestrictions(e) {
    const restrictionsDisabled = e.target.checked
    allowedTypes = getAllowedTypes(selectedProperty.about, restrictionsDisabled)
  }

  function cancel() {
    store.set({})
  }
</script>

<form bind:this={formElement}>
  <button disabled={!isValid} on:click={addProperty}
          type="button" class="btn btn-dark">Add property</button>
  <button on:click={cancel}
          type="button" class="btn btn-dark">Cancel</button>
  <select required class="form-select" bind:value={selectedProperty} on:change={change}>
    <option value="">--Pick a property--</option>
    {#each allowedProperties as value}
      <option {value}>{value.about}</option>
    {/each}
  </select>
  <label>Pick at least one class
    <label>
      <input disabled={!selectedProperty} on:change={toggleRestrictions} type="checkbox"/>
      Disable restrictions
    </label>
    <select
      bind:value={range}
      on:change={checkValidity}
      disabled={!allowedTypes}
      required size="30" multiple class="form-select">
      {#if allowedTypes}
        {#each allowedTypes as value}
          <option {value}>{value.about}</option>
        {/each}
      {/if}
    </select>
  </label>
</form>

<style>
  label {
    width: 100%;
  }

  label label {
    width: auto;
  }
</style>