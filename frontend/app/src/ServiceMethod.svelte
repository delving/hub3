<script>
  import JsonExample from "./JsonExample.svelte";
  import Fields from "./Fields.svelte";

  export let objects;
  export let service;
  export let method;

  const inputTypeID = method.inputObject.typeID
  const outputTypeID = method.outputObject.typeID
</script>

<div class="content">
  <div class="prose"><h1><a href="#service:{service.name}">{service.name}</a>.{method.name}
  </h1>
    <p class="comment">{method.comment}</p>

    <h2>JSON input</h2>
    <JsonExample typeID={inputTypeID} objects={objects}></JsonExample>
    <Fields typeID={inputTypeID} objects={objects}></Fields>

    <h2>JSON output</h2>
    <JsonExample typeID={outputTypeID} objects={objects}></JsonExample>
    <Fields typeID={inputTypeID} objects={objects}></Fields>
  </div>
</div>

<style>
  h1, h2 {
    color: #eee;
  }
</style>