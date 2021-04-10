<script>
  import {onMount} from "svelte";
  import {descriptionStore} from "./descriptionStore";

  export let sections;

  async function showSection(i) {
    await descriptionStore.patchSection(i)
  }

  onMount(async () => {
    await descriptionStore.prepare()
  })
</script>

<nav>
  <ul class="list-group">
    {#each sections as section, i}
      <li class="list-group-item"><a href="#" on:click={() => showSection(i)}>{section.title}</a></li>
    {/each}
  </ul>
</nav>