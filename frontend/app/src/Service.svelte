<script>
  export let service
</script>

<div class="prose"><h1>{service.name}</h1>
  <p>{service.comment}</p>
  <h2 class="pb-4">Methods</h2>
  <p>{service.name} exposes the following methods:</p>
  <ul class="text-gray-700">
    {#each service.methods as method}
      <li>
        <a href="#serviceMethod:{service.name}.{method.name}">
          <span>{method.name}</span>
        </a>—{method.comment}
      </li>
    {/each}
  </ul>
</div>