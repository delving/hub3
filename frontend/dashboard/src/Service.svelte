<script>
  import {getService} from "../../gen/def";
  import {API} from "./api";

  export let serviceName;
  let service, api, search, fields;
  $: {
    service = getService(serviceName)
    api = new API(service)
    fields = api.getSearchResponseFields()
    search = api.search({})
  }
</script>

{#await search then result}
  <table>
    <tr>
      {#each fields as field}
        <th>{field.name}</th>
      {/each}
    </tr>
    {#each result.hits as hit}
      <tr>
        {#each fields as field}
          <td>{hit[field.name]}</td>
        {/each}
      </tr>
    {/each}
  </table>
{/await}