<script>
  import {getService} from "../../gen/def";
  import {linkServiceMethod} from "./link";

  export let serviceName

  let service;
  $: service = getService(serviceName)
</script>

<div><h1 class="bright-color">{service.name}</h1>
  <p>{service.comment}</p>
  <h2 class="bright-color">Methods</h2>
  <p>{service.name} exposes the following methods:</p>
  <ul>
    {#each service.methods as method}
      <li>
        <a href={linkServiceMethod(service, method)}>
          <span>{method.name}</span>
        </a>—{method.comment}
      </li>
    {/each}
  </ul>
</div>