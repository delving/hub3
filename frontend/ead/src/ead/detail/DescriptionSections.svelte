<script>
  import {onMount} from "svelte";
  import {descriptionStore} from "./descriptionStore";

  async function showSection(i) {
    await descriptionStore.patchSection(i)
  }

  onMount(async () => {
    await descriptionStore.prepare()
  })
</script>

{#if $descriptionStore.sections}
  {#each $descriptionStore.sections as section, i}
    <li><a href="#" on:click={() => showSection(i)}>{section.title}</a></li>
  {/each}
{/if}