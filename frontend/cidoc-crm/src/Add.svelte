<script>
  import {getAllowedProperties, getAllowedTypes} from "./crm";
  import {store} from "./store";

  export let addition;

  let allowedProperties = getAllowedProperties(addition.context)
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
    addition.property.properties = [...addition.property.properties, {
      about: selectedProperty.about,
      type: range.map(i => i.about),
      properties: [],
    }]
    store.set({})
  }
</script>

<form bind:this={formElement}>
  <button disabled={!isValid} on:click={addProperty}
          type="button" class="btn btn-dark">Add property</button>
  <select required class="form-select" bind:value={selectedProperty} on:change={change}>
    <option value="">--Pick a property--</option>
    {#each allowedProperties as value}
      <option {value}>{value.labels['en']}</option>
    {/each}
  </select>
  <label>Pick at least one class
    <select
      bind:value={range}
      on:change={checkValidity}
      disabled={!allowedTypes}
      required size="30" multiple class="form-select">
      {#if allowedTypes}
        {#each allowedTypes as value}
          <option {value}>{value.labels['en']}</option>
        {/each}
      {/if}
    </select>
  </label>
</form>

<style>
  label {
    width: 100%;
  }
</style>