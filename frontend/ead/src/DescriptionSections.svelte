<script>
  import {onMount} from "svelte";
  import {description} from "./description";

  async function showSection(i) {
    await description.patchSection(i)
  }

  onMount(async () => {
    await description.prepare()
  })
</script>

{#if $description.sections}
  {#each $description.sections as section, i}
    <li><a href="#" on:click={() => showSection(i)}>{section.title}</a></li>
  {/each}
{/if}