<script>
  import {getService} from "../../gen/def";
  import {API} from "./api";
  import Editor from "./Editor.svelte";
  import {onMount} from "svelte";
  import {alert} from './alert'

  export let serviceName;
  let service, api, items, fields, item;
  $: {
    service = getService(serviceName)
    api = new API(service)
    fields = api.getSearchResponseFields()
  }

  async function search() {
    let [result, ok] = await api.search({})
    if(ok) {
      const hits = result.hits
      hits.sort((hitA, hitB) => hitA.prefix.localeCompare(hitB.prefix))
      items = hits
    } else {
      alert('Failed to load items', true);
    }
  }

  async function setItem(itemId) {
    const [result, ok] = await api.get(itemId)
    if (ok) item = result; else {
      alert('Failed to load item', true);
    }
  }

  async function deleteItem(itemId) {
    const [_, ok] = await api.delete(itemId)
    if(ok) {
      alert('Item deleted', false);
      await search()
    } else {
      alert('Failed to delete item', true);
    }
  }

  async function saveItem() {
    const [_, ok] = await api.put(item)
    if(ok) {
      alert('Item saved', false);
      item = null
      await search();
    } else {
      alert('Failed to save item', true);
    }
  }

  function cancel() {
    item = null;
  }

  onMount(async () => {
    await search()
  })
</script>

{#if items}
  {#if item}
    <Editor inputDescription="item" input={item}/>
    <button on:click={cancel}>Cancel</button>
    <button on:click={() => saveItem(item)}>Save</button>
  {/if}
  <div class="table-container">
    <table>
      <tr class="header">
        {#each fields as field}
          <th>{field.name}</th>
        {/each}
        <th>Edit</th>
        <th>Delete</th>
      </tr>
      {#each items as item, i (item.uuid)}
        <tr class:odd={i % 2 !== 0}>
          {#each fields as field}
            <td>{item[field.name]}</td>
          {/each}
          <td>
            <button class="icon edit" on:click={setItem(item.uuid)}></button>
          </td>
          <td>
            <button class="icon delete" on:click={deleteItem(item.uuid)}></button>
          </td>
        </tr>
      {/each}
    </table>
  </div>
{/if}

<style>
  .table-container {
    max-height: 100vh;
    overflow: auto;
  }

  table {
    text-align: left;
    width: 100%;
    border-collapse: collapse;
  }

  .header {
    border-radius: 1em;
  }

  th, td {
    padding: 1em;
  }

  th {
    background-color: #40b0ff;
    top: 0;
    position: sticky;
    font-weight: bold;
  }

  .odd {
    background-color: rgb(243, 243, 243);
  }

  .icon {
    width: 24px;
    height: 24px;
    border: none;
    cursor: pointer;
  }

  .edit {
    background: url('data:image/svg+xml;utf8,<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" > <path fill-rule="evenodd" clip-rule="evenodd" d="M21.2635 2.29289C20.873 1.90237 20.2398 1.90237 19.8493 2.29289L18.9769 3.16525C17.8618 2.63254 16.4857 2.82801 15.5621 3.75165L4.95549 14.3582L10.6123 20.0151L21.2189 9.4085C22.1426 8.48486 22.338 7.1088 21.8053 5.99367L22.6777 5.12132C23.0682 4.7308 23.0682 4.09763 22.6777 3.70711L21.2635 2.29289ZM16.9955 10.8035L10.6123 17.1867L7.78392 14.3582L14.1671 7.9751L16.9955 10.8035ZM18.8138 8.98525L19.8047 7.99429C20.1953 7.60376 20.1953 6.9706 19.8047 6.58007L18.3905 5.16586C18 4.77534 17.3668 4.77534 16.9763 5.16586L15.9853 6.15683L18.8138 8.98525Z" fill="currentColor" /> <path d="M2 22.9502L4.12171 15.1717L9.77817 20.8289L2 22.9502Z" fill="currentColor" /> </svg>') center center no-repeat;
    border-bottom: 2px solid blue;
  }

  .delete {
    background: url('data:image/svg+xml;utf8,<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"> <path d="M8 11C7.44772 11 7 11.4477 7 12C7 12.5523 7.44772 13 8 13H16C16.5523 13 17 12.5523 17 12C17 11.4477 16.5523 11 16 11H8Z" fill="currentColor" /> <path fill-rule="evenodd" clip-rule="evenodd" d="M23 12C23 18.0751 18.0751 23 12 23C5.92487 23 1 18.0751 1 12C1 5.92487 5.92487 1 12 1C18.0751 1 23 5.92487 23 12ZM21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" fill="currentColor" /> </svg>') center center no-repeat;
    border-bottom: 2px solid red;
  }
</style>