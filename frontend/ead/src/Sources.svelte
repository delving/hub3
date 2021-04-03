<script>
  let sources = [
    {
      title: 'Awefknwefwenlfwknwef',
      link: 'http://www.nationaalarchief.nl'
    },
    {
      title: 'Awefknwefwenlfwknwef',
      link: 'http://www.nationaalarchief.nl'
    },
    {
      title: 'Awefknwefwenlfwknwef',
      link: 'http://www.nationaalarchief.nl'
    },
    {
      title: 'Awefknwefwenlfwknwef',
      link: 'http://www.nationaalarchief.nl'
    },
    {
      title: 'Awefknwefwenlfwknwef',
      link: 'http://www.nationaalarchief.nl'
    },
    {
      title: 'Awefknwefwenlfwknwef',
      link: 'http://www.nationaalarchief.nl'
    }
  ]

</script>

<section>
  <h1>Sources</h1>
  <ul>
    {#each sources as source}
      <li>
        <a href="#">{source.title}</a>
      </li>
    {/each}
  </ul>
</section>

<style type="text/scss">
  a {
    border: 0;
  }

  ul {
    border-top: 1px solid black;
  }

  li {
    padding: 1em 0;
    border-bottom: 1px solid black;
  }
</style>