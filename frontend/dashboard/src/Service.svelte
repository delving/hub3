<script>
  import {getService} from "../../gen/def";
  import {API} from "./api";
  import Editor from "./Editor.svelte";
  import {onMount} from "svelte";

  export let serviceName;
  let service, api, items, fields, item;
  $: {
    service = getService(serviceName)
    api = new API(service)
    fields = api.getSearchResponseFields()
  }

  async function search() {
    let result = await api.search({})
    const hits = result.hits
    hits.sort((hitA, hitB) => hitA.prefix.localeCompare(hitB.prefix))
    items = hits
  }

  async function setItem(itemId) {
    item = await api.get(itemId)
  }

  async function deleteItem(itemId) {
    await api.delete(itemId)
    await search()
  }

  async function saveItem() {
    await api.put(item)
    await search()
    item = null
  }

  onMount(async () => {
    await search()
  })
</script>

{#if items}
  {#if item}
    <Editor inputDescription="item" input={item}/>
    <button on:click={saveItem(item)}>Save</button>
  {/if}
  <table>
    <tr>
      {#each fields as field}
        <th>{field.name}</th>
      {/each}
      <th>Edit</th>
      <th>Delete</th>
    </tr>
    {#each items as item (item.uuid)}
      <tr>
        {#each fields as field}
          <td>{item[field.name]}</td>
        {/each}
        <td>
          <button on:click={setItem(item.uuid)}>Edit</button>
        </td>
        <td>
          <button on:click={deleteItem(item.uuid)}>Delete</button>
        </td>
      </tr>
    {/each}
  </table>
{/if}