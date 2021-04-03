<script>
  export let figures = [
    {
      title: 'Image A',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image B',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image C',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image D',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image E',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image F',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image G',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image H',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image I',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image J',
      link: '/detail/NK3189-A',
      image: 'https:/media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    },
    {
      title: 'Image K',
      link: '/detail/NK3189-A',
      image: 'https://media.delving.org/thumbnail/dcn/nkcollectie/b7918d46-69b9-1589-2140-b0ad8bf7e58a/1000'
    }
  ]
</script>

<section>
  <header>
    <h1>Series</h1>
  </header>
  <div>
    {#each figures as figure}
      <a href={figure.link}>
        <figure>
          <img src={figure.image} alt="{figure.title}"/>
          <figcaption>{figure.title}</figcaption>
        </figure>
      </a>
    {/each}
  </div>
</section>

<style>
  div {
    display: grid;
    gap: 0.5em;
    max-width: 100%;
    grid-template-columns: repeat(5, 1fr);
  }

  figure, img {
    margin: 0;
    max-width: 100%;
  }

  a {
    font-weight: bold;
    border: 0;
  }
</style>