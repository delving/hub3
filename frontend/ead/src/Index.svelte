<script>
  import Value from "./Value.svelte";
  import {linkTo} from "./router";

  export let search;
  console.log(search)
  export let config;

  function createLink(item) {
    return linkTo(config.linkTo, item.meta)
  }
</script>
<table class="table">
  <thead>
  <tr>
    <th>#</th>
    {#each config.items as property}
      <th scope="column">{property.label}</th>
    {/each}
  </tr>
  </thead>
  <tbody>
  {#each search.items as item, index}
      <tr>
        <th scope="row"><a href={createLink(item)}>{index}</a></th>
        {#each config.items as property}
          <td>
            <Value context={item.fields} path={property.path}/>
          </td>
        {/each}
      </tr>
  {/each}
  </tbody>
</table>