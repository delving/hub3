<script>
  export let topics
</script>

{#each topics as topic}
  <h2 class="bright-color">{topic.subject}</h2>
  <ul>
    {#each topic.links as link}
      <li><a href="#markdown:{link.id}">{link.text}</a></li>
    {/each}
  </ul>
{/each}

<style>
  ul {
    margin: 0;
    padding: 0;
    list-style-type: none;
  }
</style>