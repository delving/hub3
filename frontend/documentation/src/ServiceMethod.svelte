<script>
  import JsonExample from "./JsonExample.svelte";
  import Fields from "./Fields.svelte";
  import {getService, getServiceMethod} from "../../gen/def";
  import {linkService} from "./link";

  export let serviceName;
  export let methodName;

  let service, method;
  $: service = getService(serviceName)
  $: method = getServiceMethod(service, methodName)
</script>

<div>
  <div><h1><a href={linkService(service)}>{service.name}</a>.{method.name}
  </h1>
    <p>{method.comment}</p>

    <h2>JSON input</h2>
    <JsonExample type={method.input}></JsonExample>
    <Fields type={method.input}></Fields>

    <h2>JSON output</h2>
    <JsonExample type={method.output}></JsonExample>
    <Fields type={method.output}></Fields>
  </div>
</div>

<style>
  h1, h2 {
    color: #eee;
  }
</style>