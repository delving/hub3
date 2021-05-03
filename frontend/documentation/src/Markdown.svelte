<script>
  import MarkdownIt from 'markdown-it'

  export let markdown = ''

  // Initialize `markdown-it`
  const md = new MarkdownIt()

  // Render to an html string
  let rendered;
  $: rendered = md.render(markdown)
</script>

<!-- Render with the `@html` directive -->
<div>
  {@html rendered}
</div>