<script>
  import Editor from "./Editor.svelte";

  export let inputDescription;
  export let input;
  let properties;
  $: {
    properties = []
    if (input && typeof input === "object") {
      for (const [key, value] of Object.entries(input)) {
        const type = Array.isArray(value) ? "array" : typeof value
        properties.push({
          name: key,
          value: value,
          type: type
        })
      }
    } else {
      properties.push({
        name: inputDescription,
        value: input,
        type: Array.isArray(input) ? "array" : typeof input
      })
    }
  }
</script>
{#if properties}
  <div>
    {#each properties as property}
      {#if property.type === "object"}
        <Editor inputDescription={property.name} input={property.value}/>
      {:else if property.type === "array"}
        <fieldset>
          <h2>{property.name}</h2>
          {#each property.value as propertyValue}
            <p>{propertyValue}</p>
          {/each}
        </fieldset>
      {:else}
        <label for={property.name}>{property.name}</label>
        <input name={property.name} bind:value={input[property.name]}/>
      {/if}
    {/each}
  </div>
{/if}

<style>
  div {
    padding-top: 1em;
  }

  label, input {
    display: block;
  }

  label {
    font-weight: bold;
  }
</style>